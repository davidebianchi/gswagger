package apirouter

import "net/http"

// Handler is the http type handler
type HandlerFunc func(w http.ResponseWriter, req *http.Request)

type Router interface {
	AddRoute(method string, path string, handler HandlerFunc) Route
}

type Route interface{}
