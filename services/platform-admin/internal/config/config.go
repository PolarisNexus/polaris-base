package config

import (
	"fmt"
	"os"
)

type Config struct {
	HTTPAddr         string
	ApisixAdminURL   string
	ApisixAdminKey   string
	ElasticURL       string
	ElasticUser      string
	ElasticPassword  string
	CrowdsecLAPIURL    string
	CrowdsecUser       string
	CrowdsecPassword   string
	CrowdsecBouncerKey string

	// OIDC（Authentik）
	OIDCIssuer   string // 如 http://authentik-server:9000/application/o/platform-admin/
	OIDCClientID string // Authentik Provider.client_id（blueprint 固定 polaris-platform-admin）
	AuthDisabled bool   // 仅本地调试用；生产严禁开启
	AllowedCORS  string // 前端 origin，CORS 放行
}

func Load() (*Config, error) {
	c := &Config{
		HTTPAddr:         getenv("HTTP_ADDR", ":8080"),
		ApisixAdminURL:   getenv("APISIX_ADMIN_URL", "http://apisix:9180"),
		ApisixAdminKey:   os.Getenv("APISIX_ADMIN_KEY"),
		ElasticURL:       getenv("ELASTIC_URL", "http://elasticsearch:9200"),
		ElasticUser:      getenv("ELASTIC_USERNAME", "elastic"),
		ElasticPassword:  os.Getenv("ELASTIC_PASSWORD"),
		CrowdsecLAPIURL:    getenv("CROWDSEC_LAPI_URL", "http://crowdsec:8080"),
		CrowdsecUser:       getenv("CROWDSEC_USERNAME", "platform-admin"),
		CrowdsecPassword:   os.Getenv("CROWDSEC_PASSWORD"),
		CrowdsecBouncerKey: os.Getenv("CROWDSEC_BOUNCER_KEY"),
		OIDCIssuer:       getenv("OIDC_ISSUER", "http://authentik-server:9000/application/o/platform-admin/"),
		OIDCClientID:     getenv("OIDC_CLIENT_ID", "polaris-platform-admin"),
		AuthDisabled:     os.Getenv("AUTH_DISABLED") == "1",
		AllowedCORS:      getenv("ALLOWED_CORS_ORIGIN", "http://localhost:5173"),
	}
	if c.ApisixAdminKey == "" {
		return nil, fmt.Errorf("APISIX_ADMIN_KEY required")
	}
	if c.CrowdsecPassword == "" {
		return nil, fmt.Errorf("CROWDSEC_PASSWORD required")
	}
	if c.CrowdsecBouncerKey == "" {
		return nil, fmt.Errorf("CROWDSEC_BOUNCER_KEY required (LAPI /v1/decisions 走 bouncer 凭据)")
	}
	return c, nil
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
