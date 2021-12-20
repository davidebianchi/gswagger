package apirouter

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type gorillaRouter struct {
	router *mux.Router
}

func (r gorillaRouter) AddRoute(method string, path string, handler interface{}) (Route, error) {
	switch h := handler.(type) {
	case func(w http.ResponseWriter, req *http.Request):
		return r.router.HandleFunc(path, h).Methods(method), nil
	default:
		return nil, fmt.Errorf("%w: handler type for gorilla is not handled", ErrInvalidHandler)
	}
}

func NewGorillaMuxRouter(router *mux.Router) Router {
	return gorillaRouter{
		router: router,
	}
}
