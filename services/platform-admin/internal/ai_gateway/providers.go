package ai_gateway

import v1 "github.com/PolarisNexus/polaris-base/api/gen/go/polaris/platform_admin/v1"

// staticProviders 是 Phase I MVP 的 provider 列表。
// 来源是 components/apisix/routes/20-ai-gateway.yaml 里 6 条路由的映射，静态同步维护。
// Phase III 可改为从 APISIX Admin API 动态拉 services/routes 推导。
//
// upstream_endpoint 在 dev 下都指向 ai-mock；生产 apply 时替换真实 URL。
// UI 展示这份常量即可，不做 CRUD。
func staticProviders() []*v1.Provider {
	return []*v1.Provider{
		{
			Id:               "openai",
			DisplayName:      "OpenAI",
			ApisixProvider:   "openai",
			BaseUrl:          "/ai/v1/openai",
			SupportedPaths:   []string{"/chat/completions", "/embeddings"},
			Status:           "dev",
			UpstreamEndpoint: "http://base-ai-mock:8080",
		},
		{
			Id:               "claude",
			DisplayName:      "Claude",
			ApisixProvider:   "anthropic",
			BaseUrl:          "/ai/v1/claude",
			SupportedPaths:   []string{"/chat/completions"},
			Status:           "dev",
			UpstreamEndpoint: "http://base-ai-mock:8080",
		},
		{
			Id:               "deepseek",
			DisplayName:      "DeepSeek",
			ApisixProvider:   "deepseek",
			BaseUrl:          "/ai/v1/deepseek",
			SupportedPaths:   []string{"/chat/completions"},
			Status:           "dev",
			UpstreamEndpoint: "http://base-ai-mock:8080",
		},
		{
			Id:               "qwen",
			DisplayName:      "Qwen (DashScope)",
			ApisixProvider:   "openai-compatible",
			BaseUrl:          "/ai/v1/qwen",
			SupportedPaths:   []string{"/chat/completions", "/embeddings"},
			Status:           "dev",
			UpstreamEndpoint: "http://base-ai-mock:8080",
		},
	}
}
