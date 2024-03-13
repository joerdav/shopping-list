package routing

import "net/http"

func RegisterRoute(
	mux *http.ServeMux,
	path string,
	handler http.Handler,
	middlewares ...func(http.Handler) http.Handler,
) {
	h := handler
	for _, m := range middlewares {
		h = m(h)
	}
	mux.Handle(path, h)
}
