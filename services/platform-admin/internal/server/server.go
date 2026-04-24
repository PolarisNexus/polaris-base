package server

import (
	"net/http"

	"github.com/PolarisNexus/polaris-base/api/gen/go/polaris/platform_admin/v1/platform_adminv1connect"
	"github.com/PolarisNexus/polaris-base/services/platform-admin/internal/ai_gateway"
	"github.com/PolarisNexus/polaris-base/services/platform-admin/internal/auth"
	"github.com/PolarisNexus/polaris-base/services/platform-admin/internal/bot"
	"github.com/PolarisNexus/polaris-base/services/platform-admin/internal/gateway"
	"github.com/PolarisNexus/polaris-base/services/platform-admin/internal/waf"
)

// Options 控制 Mux 组装行为。
type Options struct {
	Verifier    *auth.Verifier // nil 表示禁用鉴权（仅本地调试）
	AllowOrigin string         // CORS 放行 origin
}

// Mux 装配 Connect handler + healthz，按 Options 叠加鉴权与 CORS。
func Mux(gw *gateway.Service, wf *waf.Service, bt *bot.Service, ai *ai_gateway.Service, opt Options) http.Handler {
	mux := http.NewServeMux()

	path, handler := platform_adminv1connect.NewGatewayServiceHandler(gw)
	mux.Handle(path, handler)

	path, handler = platform_adminv1connect.NewWafServiceHandler(wf)
	mux.Handle(path, handler)

	path, handler = platform_adminv1connect.NewBotServiceHandler(bt)
	mux.Handle(path, handler)

	path, handler = platform_adminv1connect.NewAiGatewayServiceHandler(ai)
	mux.Handle(path, handler)

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	var h http.Handler = mux
	if opt.Verifier != nil {
		h = auth.Middleware(opt.Verifier, []string{"/healthz"})(h)
	}
	if opt.AllowOrigin != "" {
		h = cors(opt.AllowOrigin)(h)
	}
	return h
}

// cors 放行 Connect 所需的 CORS 头（Authorization、Connect-Protocol-Version 等）。
func cors(origin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers",
				"Authorization,Content-Type,Connect-Protocol-Version,Connect-Timeout-Ms,X-User-Agent")
			w.Header().Set("Access-Control-Expose-Headers", "Content-Encoding,Grpc-Status,Grpc-Message")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
