package gorillarouter

import (
	"fmt"
	"net/http"

	"github.com/davidebianchi/gswagger/apirouter"

	"github.com/gorilla/mux"
)

type gorillaRouter struct {
	router *mux.Router
}

func (r gorillaRouter) AddRoute(method string, path string, handler interface{}) (apirouter.Route, error) {
	switch h := handler.(type) {
	case func(w http.ResponseWriter, req *http.Request):
		return r.router.HandleFunc(path, h).Methods(method), nil
	default:
		return nil, fmt.Errorf("%w: handler type for gorilla is not handled: method %s, path %s", apirouter.ErrInvalidHandler, method, path)
	}
}

func (r gorillaRouter) SwaggerHandler(contentType string, blob []byte) interface{} {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		w.Write(blob)
	}
}

func New(router *mux.Router) apirouter.Router {
	return gorillaRouter{
		router: router,
	}
}
