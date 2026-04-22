package auth

import (
	"context"
	"log"
	"net/http"
	"strings"
)

type ctxKey struct{}

// Middleware 校验 Bearer ID token；成功将 claims 注入 ctx，失败返回 401。
// skipPrefixes 里的路径（如 /healthz）直接放行。
func Middleware(v *Verifier, skipPrefixes []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, p := range skipPrefixes {
				if strings.HasPrefix(r.URL.Path, p) {
					next.ServeHTTP(w, r)
					return
				}
			}
			h := r.Header.Get("Authorization")
			if !strings.HasPrefix(h, "Bearer ") {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}
			claims, err := v.Verify(r.Context(), strings.TrimPrefix(h, "Bearer "))
			if err != nil {
				log.Printf("auth reject: path=%s err=%v", r.URL.Path, err)
				http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), ctxKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClaimsFromContext 返回当前请求的 token claims，未认证则返回 nil。
func ClaimsFromContext(ctx context.Context) *Claims {
	c, _ := ctx.Value(ctxKey{}).(*Claims)
	return c
}
