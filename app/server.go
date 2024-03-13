package app

import (
	"database/sql"
	"net/http"

	"github.com/joerdav/shopping-list/app/authweb"
	"github.com/joerdav/shopping-list/app/itemsweb"
	"github.com/joerdav/shopping-list/app/listsweb"
	"github.com/joerdav/shopping-list/app/recipesweb"
	"github.com/joerdav/shopping-list/app/shopsweb"
)

type server struct {
	mux  *http.ServeMux
	conn *sql.DB
}

func NewServer(conn *sql.DB) *server {
	mux := http.NewServeMux()
	s := &server{mux, conn}
	s.Routes()
	return s
}

func (s *server) Routes() {
	listsweb.RegisterHandlers(s.mux, listsweb.Config{Conn: s.conn})
	shopsweb.RegisterHandlers(s.mux, shopsweb.Config{Conn: s.conn})
	itemsweb.RegisterHandlers(s.mux, itemsweb.Config{Conn: s.conn})
	recipesweb.RegisterHandlers(s.mux, recipesweb.Config{Conn: s.conn})
	authweb.RegisterHandlers(s.mux, authweb.Config{Conn: s.conn})

	s.mux.Handle("/public/", public())
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
