package authweb

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"log/slog"
	"net/http"
)

type state struct {
	Provider  string `json:"provider"`
	ReturnUrl string `json:"returnUrl"`
	Challenge string `json:"challenge"`
}

type ProviderConfig struct {
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
}

type Config struct {
	Providers map[string]ProviderConfig
}

func RegisterHandlers(mux *http.ServeMux, config Config) {
	mux.Handle("GET /auth/{provider}/signin", signinHandler(config.Providers))
	mux.Handle("GET /auth/redirect", redirectHandler())
}

func generateUrl(provider ProviderConfig, domain, state string) string {
	return provider.AuthURL +
		"?client_id=" + provider.ClientID +
		"&redirect_uri=" + domain + "/auth/redirect" +
		"&response_type=code" +
		"&scope=profile" +
		"&state=" + state
}

func generateRandomString() string {
	bytes := make([]byte, 12)
	rand.Read(bytes)
	return base32.StdEncoding.EncodeToString(bytes)
}

func generateState(provider, returnUrl string) (string, error) {
	challenge := generateRandomString()
	state := state{
		Provider:  provider,
		ReturnUrl: returnUrl,
		Challenge: challenge,
	}
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(state); err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(b.Bytes()), nil
}

func parseState(s string) (state, error) {
	b, err := base32.StdEncoding.DecodeString(s)
	if err != nil {
		return state{}, err
	}
	var st state
	if err := json.NewDecoder(bytes.NewReader(b)).Decode(&st); err != nil {
		return state{}, err
	}
	return st, nil
}

func signinHandler(providers map[string]ProviderConfig) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		providerID := r.PathValue("provider")
		provider, ok := providers[providerID]
		if !ok {
			http.Error(w, "unknown provider", http.StatusBadRequest)
			return
		}
		state, err := generateState(providerID, r.URL.Query().Get("returnUrl"))
		if err != nil {
			slog.Error("failed to generate state", "error", err)
			http.Error(w, "failed to generate state", http.StatusInternalServerError)
			return
		}
		url := generateUrl(provider, "http://localhost:7331", state)
		http.SetCookie(w, &http.Cookie{
			Name:     "state",
			Value:    state,
			HttpOnly: true,
			Path:     "/",
			// Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, url, http.StatusFound)
	})
}

func redirectHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		stateString := r.URL.Query().Get("state")

		cookie, err := r.Cookie("state")
		if err != nil {
			slog.Error("failed to get state cookie", "error", err)
			http.Error(w, "failed to get state cookie", http.StatusInternalServerError)
			return
		}
		if stateString == "" || stateString != cookie.Value {
			slog.Error("state mismatch", "state", stateString, "cookie", cookie.Value)
			http.Error(w, "state mismatch", http.StatusBadRequest)
			return
		}
		state, err := parseState(stateString)
		if err != nil {
			slog.Error("failed to parse state", "error", err)
			http.Error(w, "failed to parse state", http.StatusInternalServerError)
			return
		}
		slog.Info("logged in", "code", code, "state", state.Provider)
	})
}
