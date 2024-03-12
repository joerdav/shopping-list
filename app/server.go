package app

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/joerdav/shopping-list/app/authweb"
	"github.com/joerdav/shopping-list/app/itemsweb"
	"github.com/joerdav/shopping-list/app/listsweb"
	"github.com/joerdav/shopping-list/app/recipesweb"
	"github.com/joerdav/shopping-list/app/shopsweb"
)

type server struct {
	mux                *http.ServeMux
	conn               *sql.DB
	authProviderConfig authweb.Config
}

func NewServer(conn *sql.DB) *server {
	mux := http.NewServeMux()
	authProviderConfig := authweb.Config{Providers: map[string]authweb.ProviderConfig{}}
	authProviderConfig.Providers["google"] = authweb.ProviderConfig{
		AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:     "https://oauth2.googleapis.com/token",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	}
	s := &server{mux, conn, authProviderConfig}
	s.Routes()
	return s
}

func (s *server) Routes() {
	listsweb.RegisterHandlers(s.mux, listsweb.Config{Conn: s.conn})
	shopsweb.RegisterHandlers(s.mux, shopsweb.Config{Conn: s.conn})
	itemsweb.RegisterHandlers(s.mux, itemsweb.Config{Conn: s.conn})
	recipesweb.RegisterHandlers(s.mux, recipesweb.Config{Conn: s.conn})
	authweb.RegisterHandlers(s.mux, s.authProviderConfig)

	s.mux.Handle("/public/", public())
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
