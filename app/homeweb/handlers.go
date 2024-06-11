package homeweb

import (
	"net/http"

	"github.com/joerdav/shopping-list/app/middleware"
	"github.com/joerdav/shopping-list/business/auth"
	"github.com/joerdav/shopping-list/foundation/routing"
)

type Config struct{}

func RegisterHandlers(mux *http.ServeMux, config Config) {
	authMiddleware := middleware.AuthMiddleware(auth.NewCore())

	routing.RegisterRoute(mux, "GET /{$}", homeHandler(), authMiddleware)
}

func homeHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := HomePage(r.URL.Path).Render(r.Context(), w); err != nil {
			// TODO: replace with error page
			http.Error(w, "render error", http.StatusInternalServerError)
		}
	})
}
