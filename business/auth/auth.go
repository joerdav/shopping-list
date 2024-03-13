package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"
)

var (
	ErrUnknownProvider = errors.New("unknown provider")
	ErrStateMismatch   = errors.New("state mismatch")
	ErrTokenExchange   = errors.New("token exchange failed")
)

type ContextKey string

func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ContextKey("userID"), id)
}

func UserID(ctx context.Context) string {
	value, ok := ctx.Value(ContextKey("userID")).(string)
	if !ok {
		return ""
	}
	return value
}

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
	KeyUrl       string
	Issuer       string
}

type Config struct {
	Providers map[string]ProviderConfig
}

type Tokens struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	IDToken     string `json:"id_token"`
	Provider    string `json:"provider"`
}

type Core struct {
	config Config
}

func NewCore() *Core {
	config := Config{
		Providers: map[string]ProviderConfig{
			"google": {
				AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
				TokenURL:     "https://oauth2.googleapis.com/token",
				KeyUrl:       "https://www.googleapis.com/oauth2/v3/certs",
				ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
				ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
				Issuer:       "https://accounts.google.com",
			},
		},
	}
	return &Core{config}
}

func (c *Core) ValidateIDToken(ctx context.Context, token Tokens) (*jwt.Token, error) {
	p, ok := c.config.Providers[token.Provider]
	if !ok {
		return nil, ErrUnknownProvider
	}
	t, err := jwt.Parse(token.IDToken, func(t *jwt.Token) (interface{}, error) {
		set, err := jwk.Fetch(context.Background(), p.KeyUrl)
		if err != nil {
			return nil, err
		}
		keyID, ok := t.Header["kid"].(string)
		if !ok {
			return nil, errors.New("expecting JWT header to have string kid")
		}
		var key jwk.Key
		if key, ok = set.LookupKeyID(keyID); !ok {
			return nil, fmt.Errorf("unable to find key %q", keyID)
		}
		var raw any
		return raw, key.Raw(&raw)
	})
	if err != nil {
		return nil, err
	}
	claims := t.Claims.(jwt.MapClaims)
	if !claims.VerifyIssuer(p.Issuer, true) {
		return nil, errors.New("invalid issuer")
	}
	if !claims.VerifyAudience(p.ClientID, true) {
		return nil, errors.New("invalid audience")
	}
	return t, nil
}

func (c *Core) SigninUrl(provider, domain, returnUrl string) (*url.URL, string, error) {
	p, ok := c.config.Providers[provider]
	if !ok {
		return nil, "", ErrUnknownProvider
	}
	if returnUrl == "" {
		returnUrl = "/"
	}
	u, err := url.Parse(returnUrl)
	if err != nil {
		return nil, "", errors.New("invalid return url")
	}
	returnUrl = u.Path
	state, err := generateState(provider, returnUrl)
	if err != nil {
		return nil, "", err
	}
	authUrl, err := url.Parse(p.AuthURL)
	if err != nil {
		return nil, "", err
	}
	authUrl.RawQuery = url.Values{
		"client_id":     {p.ClientID},
		"redirect_uri":  {domain + "/auth/redirect"},
		"response_type": {"code"},
		"scope":         {"openid email"},
		"state":         {state},
	}.Encode()
	return authUrl, state, nil
}

func (c *Core) ExchangeCodeForToken(
	ctx context.Context,
	code, state, cookieState string,
) (Tokens, string, error) {
	if state == "" || state != cookieState {
		return Tokens{}, "", ErrStateMismatch
	}
	st, err := parseState(state)
	if err != nil {
		return Tokens{}, "", err
	}
	p, ok := c.config.Providers[st.Provider]
	if !ok {
		return Tokens{}, "", ErrUnknownProvider
	}
	resp, err := http.PostForm(p.TokenURL, url.Values{
		"client_id":     {p.ClientID},
		"client_secret": {p.ClientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {os.Getenv("HOST") + "/auth/redirect"},
	})
	if err != nil {
		return Tokens{}, "", err
	}
	if resp.StatusCode != http.StatusOK {
		return Tokens{}, "", ErrTokenExchange
	}
	defer resp.Body.Close()
	var tokens Tokens
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return Tokens{}, "", err
	}
	tokens.Provider = st.Provider
	return tokens, st.ReturnUrl, nil
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
