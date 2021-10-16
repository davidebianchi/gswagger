package apirouter

import "github.com/gorilla/mux"

type gorillaRouter struct {
	router *mux.Router
}

func (r gorillaRouter) AddRoute(path string, method string, handler Handler) Route {
	return r.router.HandleFunc(path, handler).Methods(method)
}

func NewGorillaMuxRouter(router *mux.Router) Router {
	return gorillaRouter{
		router: router,
	}
}
