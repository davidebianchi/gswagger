package fiber_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	swagger "github.com/davidebianchi/gswagger"
	oasFiber "github.com/davidebianchi/gswagger/support/fiber"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

type SwaggerRouter = swagger.Router[oasFiber.HandlerFunc, oasFiber.Route]

const (
	swaggerOpenapiTitle   = "test openapi title"
	swaggerOpenapiVersion = "test openapi version"
)

func TestFiberIntegration(t *testing.T) {
	t.Run("router works correctly", func(t *testing.T) {
		router, oasRouter := setupSwagger(t)

		err := oasRouter.GenerateAndExposeOpenapi()
		require.NoError(t, err)

		t.Run("/hello", func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/hello", nil)

			resp, err := router.Test(r)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			body := readBody(t, resp.Body)
			require.Equal(t, "OK", body)
		})

		t.Run("/hello/:value", func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/hello/something", nil)

			resp, err := router.Test(r)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			body := readBody(t, resp.Body)
			require.Equal(t, "OK", body)
		})

		t.Run("and generate swagger", func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, swagger.DefaultJSONDocumentationPath, nil)

			resp, err := router.Test(r)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			body := readBody(t, resp.Body)
			require.JSONEq(t, readFile(t, "../testdata/integration.json"), body, body)
		})
	})

	t.Run("works correctly with subrouter - handles path prefix - gorilla mux", func(t *testing.T) {
		fiberRouter, oasRouter := setupSwagger(t)

		subRouter, err := oasRouter.SubRouter(oasFiber.NewRouter(fiberRouter), swagger.SubRouterOptions{
			PathPrefix: "/prefix",
		})
		require.NoError(t, err)

		_, err = subRouter.AddRoute(http.MethodGet, "/foo", okHandler, swagger.Definitions{})
		require.NoError(t, err)

		err = oasRouter.GenerateAndExposeOpenapi()
		require.NoError(t, err)

		t.Run("correctly call /hello", func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/hello", nil)

			resp, err := fiberRouter.Test(r)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			body := readBody(t, resp.Body)
			require.Equal(t, "OK", body)
		})

		t.Run("correctly call sub router", func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/prefix/foo", nil)

			resp, err := fiberRouter.Test(r)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			body := readBody(t, resp.Body)
			require.Equal(t, "OK", body)
		})

		t.Run("and generate swagger", func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, swagger.DefaultJSONDocumentationPath, nil)

			resp, err := fiberRouter.Test(r)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			body := readBody(t, resp.Body)
			require.JSONEq(t, readFile(t, "../testdata/intergation-subrouter.json"), body, body)
		})
	})
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

	_, err = router.AddRoute(http.MethodPost, "/hello/:value", okHandler, swagger.Definitions{})
	require.NoError(t, err)

	return fiberRouter, router
}

func okHandler(c *fiber.Ctx) error {
	c.Status(http.StatusOK)
	return c.SendString("OK")
}

func readBody(t *testing.T, requestBody io.ReadCloser) string {
	t.Helper()

	body, err := io.ReadAll(requestBody)
	require.NoError(t, err)

	return string(body)
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	fileContent, err := os.ReadFile(path)
	require.NoError(t, err)

	return string(fileContent)
}
