package gorilla

import (
	"github.com/davidebianchi/gswagger/apirouter"

	"github.com/gorilla/mux"
)

type gorillaRouter struct {
	router *mux.Router
}

func (r gorillaRouter) AddRoute(method string, path string, handler apirouter.HandlerFunc) apirouter.Route {
	return r.router.HandleFunc(path, handler).Methods(method)
}

func NewRouter(router *mux.Router) apirouter.Router {
	return gorillaRouter{
		router: router,
	}
}
