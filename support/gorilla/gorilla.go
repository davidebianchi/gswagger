package gorilla

import (
	"github.com/davidebianchi/gswagger/apirouter"

	"net/http"

	"github.com/gorilla/mux"
)

// HandlerFunc is the http type handler used by gorilla/mux
type HandlerFunc func(w http.ResponseWriter, req *http.Request)
type Route = *mux.Route

type gorillaRouter struct {
	router *mux.Router
}

func (r gorillaRouter) AddRoute(method string, path string, handler HandlerFunc) Route {
	return r.router.HandleFunc(path, handler).Methods(method)
}

func (r gorillaRouter) SwaggerHandler(contentType string, blob []byte) HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		w.Write(blob)
	}
}

func (r gorillaRouter) TransformPathToOasPath(path string) string {
	return path
}

func NewRouter(router *mux.Router) apirouter.Router[HandlerFunc, Route] {
	return gorillaRouter{
		router: router,
	}
}
