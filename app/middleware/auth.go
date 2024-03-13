package middleware

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/joerdav/shopping-list/business/auth"
)

func AuthMiddleware(authCore *auth.Core) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("token")
			if err != nil {
				slog.Info("failed to get token cookie", "error", err)
				redirectToSignIn(w, r)
				return
			}
			tokenBytes, err := base64.StdEncoding.DecodeString(cookie.Value)
			if err != nil {
				slog.Info("failed to decode token", "error", err)
				redirectToSignIn(w, r)
				return
			}
			var token auth.Tokens
			if err := json.NewDecoder(bytes.NewReader(tokenBytes)).Decode(&token); err != nil {
				slog.Info("failed to decode json token", "error", err)
				redirectToSignIn(w, r)
				return
			}
			t, err := authCore.ValidateIDToken(r.Context(), token)
			if err != nil {
				slog.Info("failed to validate token", "error", err)
				redirectToSignIn(w, r)
				return
			}
			claims := t.Claims.(jwt.MapClaims)
			ctx := auth.WithUserID(r.Context(), claims["sub"].(string))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func redirectToSignIn(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/auth/google/signin", http.StatusSeeOther)
}
