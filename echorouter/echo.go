package echorouter

import (
	"fmt"
	"net/http"

	"github.com/davidebianchi/gswagger/apirouter"

	echo "github.com/labstack/echo/v4"
)

type echoRouter struct {
	router *echo.Echo
}

func (r echoRouter) AddRoute(method string, path string, handler interface{}) (apirouter.Route, error) {
	switch h := handler.(type) {
	case func(c echo.Context) error:
		return r.router.Add(method, path, h), nil
	default:
		return nil, fmt.Errorf("%w: handler type for echo is not handled: method %s, path %s", apirouter.ErrInvalidHandler, method, path)
	}
}

func (r echoRouter) SwaggerHandler(contentType string, json []byte) interface{} {
	return func(c echo.Context) error {
		return c.Blob(http.StatusOK, contentType, json)
	}
}

func New(router *echo.Echo) apirouter.Router {
	return echoRouter{
		router: router,
	}
}
