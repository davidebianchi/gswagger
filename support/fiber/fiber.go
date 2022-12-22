package fiber

import (
	"github.com/davidebianchi/gswagger/apirouter"
	"github.com/gofiber/fiber/v2"
)

type HandlerFunc = fiber.Handler

type fiberRouter struct {
	router fiber.Router
}

func NewRouter(router fiber.Router) apirouter.Router[HandlerFunc] {
	return fiberRouter{
		router: router,
	}
}

func (r fiberRouter) AddRoute(method string, path string, handler HandlerFunc) apirouter.Route {
	return r.router.Add(method, path, handler)
}

func (r fiberRouter) SwaggerHandler(contentType string, blob []byte) HandlerFunc {
	return func(c *fiber.Ctx) error {
		c.Set("Content-Type", contentType)
		return c.Send(blob)
	}
}
