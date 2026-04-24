package main

import (
	"log"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/PolarisNexus/polaris-base/services/platform-admin/internal/ai_gateway"
	"github.com/PolarisNexus/polaris-base/services/platform-admin/internal/auth"
	"github.com/PolarisNexus/polaris-base/services/platform-admin/internal/bot"
	"github.com/PolarisNexus/polaris-base/services/platform-admin/internal/config"
	"github.com/PolarisNexus/polaris-base/services/platform-admin/internal/gateway"
	"github.com/PolarisNexus/polaris-base/services/platform-admin/internal/server"
	"github.com/PolarisNexus/polaris-base/services/platform-admin/internal/waf"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	esClient := waf.NewESClient(cfg.ElasticURL, cfg.ElasticUser, cfg.ElasticPassword)
	gwSvc := gateway.NewService(gateway.NewClient(cfg.ApisixAdminURL, cfg.ApisixAdminKey))
	wfSvc := waf.NewService(esClient)
	btSvc := bot.NewService(bot.NewClient(cfg.CrowdsecLAPIURL, cfg.CrowdsecUser, cfg.CrowdsecPassword, cfg.CrowdsecBouncerKey))
	aiSvc := ai_gateway.NewService(esClient)

	opt := server.Options{AllowOrigin: cfg.AllowedCORS}
	if cfg.AuthDisabled {
		log.Printf("WARNING: AUTH_DISABLED=1, OIDC check bypassed")
	} else {
		opt.Verifier = auth.NewVerifier(cfg.OIDCIssuer, cfg.OIDCClientID)
		log.Printf("OIDC verifier: issuer=%s audience=%s", cfg.OIDCIssuer, cfg.OIDCClientID)
	}
	handler := server.Mux(gwSvc, wfSvc, btSvc, aiSvc, opt)

	// h2c 以支持 gRPC 明文（APISIX 背后已终止 TLS）。
	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: h2c.NewHandler(handler, &http2.Server{}),
	}
	log.Printf("platform-admin listening on %s", cfg.HTTPAddr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("serve: %v", err)
	}
}
