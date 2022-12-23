package echo

import (
	"github.com/davidebianchi/gswagger/apirouter"

	"net/http"

	"github.com/labstack/echo/v4"
)

type Route = *echo.Route

type echoRouter struct {
	router *echo.Echo
}

func (r echoRouter) AddRoute(method string, path string, handler echo.HandlerFunc) Route {
	return r.router.Add(method, path, handler)
}

func (r echoRouter) SwaggerHandler(contentType string, blob []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Add("Content-Type", contentType)
		return c.JSONBlob(http.StatusOK, blob)
	}
}

func (r echoRouter) TransformPathToOasPath(path string) string {
	return apirouter.TransformPathParamsWithColon(path)
}

func NewRouter(router *echo.Echo) apirouter.Router[echo.HandlerFunc, Route] {
	return echoRouter{
		router: router,
	}
}
