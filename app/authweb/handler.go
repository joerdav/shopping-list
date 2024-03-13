package authweb

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/joerdav/shopping-list/business/auth"
	"github.com/joerdav/shopping-list/business/users"
	"github.com/joerdav/shopping-list/business/users/usersdb"
)

type Config struct {
	Conn *sql.DB
}

func RegisterHandlers(mux *http.ServeMux, cfg Config) {
	authCore := auth.NewCore()
	userCore := users.NewCore(usersdb.NewStorer(cfg.Conn))

	mux.Handle("GET /auth/{provider}/signin", signinHandler(authCore))
	mux.Handle("GET /auth/redirect", redirectHandler(authCore, userCore))
}

func signinHandler(authCore *auth.Core) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		providerID := r.PathValue("provider")
		url, state, err := authCore.SigninUrl(
			providerID,
			os.Getenv("HOST"),
			r.URL.Query().Get("returnUrl"),
		)
		if err != nil {
			slog.Error("failed to get signin url", "error", err)
			http.Error(w, "failed to get signin url", http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "state",
			Value:    state,
			HttpOnly: true,
			Path:     "/",
			// Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})
		http.Redirect(w, r, url.String(), http.StatusFound)
	})
}

func redirectHandler(authCore *auth.Core, userCore *users.Core) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		stateString := r.URL.Query().Get("state")

		cookie, err := r.Cookie("state")
		if err != nil {
			slog.Error("failed to get state cookie", "error", err)
			http.Error(w, "failed to get state cookie", http.StatusInternalServerError)
			return
		}
		token, returnTo, err := authCore.ExchangeCodeForToken(
			r.Context(),
			code,
			stateString,
			cookie.Value,
		)
		if err != nil {
			slog.Error("failed to exchange code for token", "error", err)
			http.Error(w, "failed to exchange code for token", http.StatusInternalServerError)
			return
		}
		t, err := authCore.ValidateIDToken(r.Context(), token)
		if err != nil {
			slog.Error("failed to validate token", "error", err)
			http.Error(w, "failed to validate token", http.StatusInternalServerError)
			return
		}

		var b bytes.Buffer
		if err := json.NewEncoder(&b).Encode(token); err != nil {
			slog.Error("failed to encode token", "error", err)
			http.Error(w, "failed to encode token", http.StatusInternalServerError)
			return
		}
		newCookie := http.Cookie{
			Name:     "token",
			Value:    base64.StdEncoding.EncodeToString(b.Bytes()),
			HttpOnly: true,
			Path:     "/",
			// Secure:   true,
			SameSite: http.SameSiteLaxMode,
		}
		userID := t.Claims.(jwt.MapClaims)["sub"].(string)
		_, err = userCore.Query(r.Context(), userID)
		if err != nil && !errors.Is(err, users.ErrNotFound) {
			slog.Error("failed to get user", "error", err)
			http.Error(w, "failed to get user", http.StatusInternalServerError)
			return
		}
		if err == nil {
			http.SetCookie(w, &newCookie)
			http.Redirect(w, r, returnTo, http.StatusFound)
			return
		}
		_, err = userCore.Create(r.Context(), users.User{userID})
		if err != nil {
			slog.Error("failed to create user", "error", err)
			http.Error(w, "failed to create user", http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &newCookie)
		http.Redirect(w, r, returnTo, http.StatusFound)
	})
}
