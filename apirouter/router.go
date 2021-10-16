package apirouter

import "net/http"

// Handler is the http type handler
type HandlerFunc func(w http.ResponseWriter, req *http.Request)

type Router interface {
	AddRoute(path string, method string, handler HandlerFunc) Route
}

type Route interface{}
