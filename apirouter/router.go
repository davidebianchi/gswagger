package apirouter

import "net/http"

// Handler is the http type handler
type Handler func(w http.ResponseWriter, req *http.Request)

type Router interface {
	AddRoute(path string, method string, handler Handler) Route
}

type Route interface{}
