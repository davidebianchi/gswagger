package apirouter

import "github.com/gorilla/mux"

type gorillaRouter struct {
	router *mux.Router
}

func (r gorillaRouter) AddRoute(method string, path string, handler HandlerFunc) Route {
	return r.router.HandleFunc(path, handler).Methods(method)
}

func NewGorillaMuxRouter(router *mux.Router) Router {
	return gorillaRouter{
		router: router,
	}
}
