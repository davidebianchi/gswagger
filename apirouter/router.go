package apirouter

type Router[HandlerFunc any, Route any] interface {
	AddRoute(method string, path string, handler HandlerFunc) Route
	SwaggerHandler(contentType string, blob []byte) HandlerFunc
}
