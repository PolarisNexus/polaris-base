// ai-mock — OpenAI 兼容的 mock API server（ADR-0014 Phase I MVP 的 dev 依赖）。
//
// 目的：让 AI Gateway 的端到端链路（APISIX ai-proxy-multi → provider）在不联网、不烧
// token 的前提下能冒烟到完整 usage 日志。不追求语义智能，只追求：
//   - 响应 schema 与 OpenAI 官方 100% 对齐（ai-proxy-multi 不需要 provider adapter）
//   - usage.prompt_tokens / completion_tokens 字段合理（粗估 = 字符数 / 4），
//     以便 elasticsearch-logger + platform-admin Usage 面板有真实可查数据
//
// 覆盖端点：POST /v1/chat/completions、POST /v1/embeddings、GET /v1/models
// 不做：SSE streaming（ai-proxy-multi 对 streaming 的适配另有路径，MVP 不验证）、
//       image / audio / fine-tuning / function-calling 真实语义。
package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type chatReq struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
	// 只解析计费相关字段；其余 ai-proxy-multi 透传不管
	MaxTokens int `json:"max_tokens"`
	Stream    bool `json:"stream"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResp struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []chatChoice `json:"choices"`
	Usage   usage        `json:"usage"`
}

type chatChoice struct {
	Index        int     `json:"index"`
	Message      message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type embReq struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embResp struct {
	Object string     `json:"object"`
	Data   []embDatum `json:"data"`
	Model  string     `json:"model"`
	Usage  usage      `json:"usage"`
}

type embDatum struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}

// countTokens 粗估 token 数：OpenAI tokenizer 约 4 chars/token 英文，中文约 1.5 chars/token。
// mock 取 4，满足 usage 日志量级合理即可。
func countTokens(s string) int {
	if s == "" {
		return 0
	}
	return (len(s) + 3) / 4
}

func genID(prefix string) string {
	var b [12]byte
	_, _ = rand.Read(b[:])
	return prefix + hex.EncodeToString(b[:])
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// dispatch 按路径分派到 chat 或 embeddings（宽松，路径含 "embedding" 走 embeddings，其他都当 chat）。
// 兼容 ai-proxy 对各 provider 上游路径的约定差异（/v1/chat/completions、/chat/completions、/、/v1/messages）。
func dispatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("unmatched: %s %s", r.Method, r.URL.Path)
		http.Error(w, "ai-mock: non-POST not handled at "+r.URL.Path, http.StatusNotFound)
		return
	}
	if strings.Contains(r.URL.Path, "embedding") {
		handleEmbeddings(w, r)
		return
	}
	handleChat(w, r)
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req chatReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Stream {
		// MVP 不支持 streaming，让上层看到明确错误便于排查
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]string{
				"message": "ai-mock: streaming not supported in MVP",
				"type":    "invalid_request_error",
			},
		})
		return
	}

	promptTok := 0
	for _, m := range req.Messages {
		promptTok += countTokens(m.Content) + countTokens(m.Role)
	}
	// 回一段固定但带上下文的文本，方便业务方肉眼确认走通
	lastUser := ""
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			lastUser = req.Messages[i].Content
			break
		}
	}
	reply := fmt.Sprintf("[ai-mock reply for model=%s] echo: %s", req.Model, truncate(lastUser, 120))
	compTok := countTokens(reply)

	writeJSON(w, http.StatusOK, chatResp{
		ID:      genID("chatcmpl-"),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []chatChoice{{
			Index:        0,
			Message:      message{Role: "assistant", Content: reply},
			FinishReason: "stop",
		}},
		Usage: usage{
			PromptTokens:     promptTok,
			CompletionTokens: compTok,
			TotalTokens:      promptTok + compTok,
		},
	})
}

func handleEmbeddings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req embReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// 只支持数组形态；input:"string" 让调用方包成 [s]（OpenAI SDK 也会）
		http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
		return
	}
	tok := 0
	out := make([]embDatum, 0, len(req.Input))
	for i, s := range req.Input {
		tok += countTokens(s)
		// 固定长度向量（mock 用 8 维即可，生产是 1536/3072）
		vec := make([]float32, 8)
		for j := range vec {
			vec[j] = float32(j) / 8.0
		}
		out = append(out, embDatum{Object: "embedding", Index: i, Embedding: vec})
	}
	writeJSON(w, http.StatusOK, embResp{
		Object: "list",
		Data:   out,
		Model:  req.Model,
		Usage:  usage{PromptTokens: tok, TotalTokens: tok},
	})
}

// handleModels 返回 ai-mock 假装支持的模型列表，方便 SDK 做 /v1/models 探测。
func handleModels(w http.ResponseWriter, r *http.Request) {
	models := []map[string]any{
		{"id": "gpt-4o-mini", "object": "model", "created": 0, "owned_by": "ai-mock"},
		{"id": "gpt-3.5-turbo", "object": "model", "created": 0, "owned_by": "ai-mock"},
		{"id": "text-embedding-3-small", "object": "model", "created": 0, "owned_by": "ai-mock"},
	}
	writeJSON(w, http.StatusOK, map[string]any{"object": "list", "data": models})
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func main() {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	// APISIX ai-proxy 对各 provider 约定的上游路径不一致：
	//   openai            → /v1/chat/completions
	//   deepseek          → /chat/completions（无 /v1）
	//   openai-compatible → endpoint 整体当 URL（不追加路径，落到 /）
	//   anthropic         → /v1/messages
	// mock 走内容匹配而非严格路径匹配：端点里有 "embedding" 走 embeddings，
	// 请求 body 含 "messages" 走 chat，其他给 404 + 日志方便排查。
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/models", handleModels)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok\n")) })
	mux.HandleFunc("/", dispatch)

	log.Printf("ai-mock listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
