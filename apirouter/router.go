package apirouter

type Router[HandlerFunc any] interface {
	AddRoute(method string, path string, handler HandlerFunc) Route
	SwaggerHandler(contentType string, blob []byte) HandlerFunc
}

type Route any
