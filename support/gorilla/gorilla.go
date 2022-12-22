package gorilla

import (
	"github.com/davidebianchi/gswagger/apirouter"

	"net/http"

	"github.com/gorilla/mux"
)

// Handler is the http type handler
type HandlerFunc func(w http.ResponseWriter, req *http.Request)

type gorillaRouter struct {
	router *mux.Router
}

func (r gorillaRouter) AddRoute(method string, path string, handler HandlerFunc) apirouter.Route {
	return r.router.HandleFunc(path, handler).Methods(method)
}

func (r gorillaRouter) SwaggerHandler(contentType string, blob []byte) HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		w.Write(blob)
	}
}

func NewRouter(router *mux.Router) apirouter.Router[HandlerFunc] {
	return gorillaRouter{
		router: router,
	}
}
