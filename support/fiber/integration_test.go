package fiber_test

import (
	"context"
	"net/http"
	"testing"

	swagger "github.com/davidebianchi/gswagger"
	oasFiber "github.com/davidebianchi/gswagger/support/fiber"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

type SwaggerRouter = swagger.Router[oasFiber.HandlerFunc, oasFiber.Route]

const (
	swaggerOpenapiTitle   = "test swagger title"
	swaggerOpenapiVersion = "test swagger version"
)

func TestWithFiber(t *testing.T) {
	require.True(t, true)
}

func setupSwagger(t *testing.T) (*fiber.App, *SwaggerRouter) {
	t.Helper()

	context := context.Background()
	fiberRouter := fiber.New()

	router, err := swagger.NewRouter(oasFiber.NewRouter(fiberRouter), swagger.Options{
		Context: context,
		Openapi: &openapi3.T{
			Info: &openapi3.Info{
				Title:   swaggerOpenapiTitle,
				Version: swaggerOpenapiVersion,
			},
		},
	})
	require.NoError(t, err)

	operation := swagger.Operation{}

	_, err = router.AddRawRoute(http.MethodGet, "/hello", okHandler, operation)
	require.NoError(t, err)

	err = router.GenerateAndExposeSwagger()
	require.NoError(t, err)

	return fiberRouter, router
}

func okHandler(c *fiber.Ctx) error {
	c.Status(http.StatusOK)
	return c.SendString("OK")
}
