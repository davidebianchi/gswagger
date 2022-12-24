package echo_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	oasEcho "github.com/davidebianchi/gswagger/support/echo"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

const (
	swaggerOpenapiTitle   = "test swagger title"
	swaggerOpenapiVersion = "test swagger version"
)

type echoSwaggerRouter = swagger.Router[echo.HandlerFunc, *echo.Route]

func TestEchoIntegration(t *testing.T) {
	t.Run("router works correctly - echo", func(t *testing.T) {
		echoRouter, oasRouter := setupEchoSwagger(t)

		err := oasRouter.GenerateAndExposeOpenapi()
		require.NoError(t, err)

		t.Run("/hello", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/hello", nil)

			echoRouter.ServeHTTP(w, r)

			require.Equal(t, http.StatusOK, w.Result().StatusCode)

			body := readBody(t, w.Result().Body)
			require.Equal(t, "OK", body)
		})

		t.Run("/hello/:value", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/hello/something", nil)

			echoRouter.ServeHTTP(w, r)

			require.Equal(t, http.StatusOK, w.Result().StatusCode)

			body := readBody(t, w.Result().Body)
			require.Equal(t, "OK", body)
		})

		t.Run("and generate swagger", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, swagger.DefaultJSONDocumentationPath, nil)

			echoRouter.ServeHTTP(w, r)

			require.Equal(t, http.StatusOK, w.Result().StatusCode)

			body := readBody(t, w.Result().Body)
			require.JSONEq(t, readFile(t, "../testdata/integration.json"), body)
		})
	})

	t.Run("works correctly with subrouter - handles path prefix - echo", func(t *testing.T) {
		eRouter, oasRouter := setupEchoSwagger(t)

		subRouter, err := oasRouter.SubRouter(oasEcho.NewRouter(eRouter), swagger.SubRouterOptions{
			PathPrefix: "/prefix",
		})
		require.NoError(t, err)

		_, err = subRouter.AddRoute(http.MethodGet, "/foo", okHandler, swagger.Definitions{})
		require.NoError(t, err)

		err = oasRouter.GenerateAndExposeOpenapi()
		require.NoError(t, err)

		t.Run("correctly call /hello", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/hello", nil)

			eRouter.ServeHTTP(w, r)

			require.Equal(t, http.StatusOK, w.Result().StatusCode)

			body := readBody(t, w.Result().Body)
			require.Equal(t, "OK", body)
		})

		t.Run("correctly call sub router", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/prefix/foo", nil)

			eRouter.ServeHTTP(w, r)

			require.Equal(t, http.StatusOK, w.Result().StatusCode)

			body := readBody(t, w.Result().Body)
			require.Equal(t, "OK", body)
		})

		t.Run("and generate swagger", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, swagger.DefaultJSONDocumentationPath, nil)

			eRouter.ServeHTTP(w, r)

			require.Equal(t, http.StatusOK, w.Result().StatusCode)

			body := readBody(t, w.Result().Body)
			require.JSONEq(t, readFile(t, "../testdata/intergation-subrouter.json"), body, body)
		})
	})
}

func readBody(t *testing.T, requestBody io.ReadCloser) string {
	t.Helper()

	body, err := io.ReadAll(requestBody)
	require.NoError(t, err)

	return string(body)
}

func setupEchoSwagger(t *testing.T) (*echo.Echo, *echoSwaggerRouter) {
	t.Helper()

	context := context.Background()
	e := echo.New()

	router, err := swagger.NewRouter(oasEcho.NewRouter(e), swagger.Options{
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

	return e, router
}

func okHandler(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	fileContent, err := os.ReadFile(path)
	require.NoError(t, err)

	return string(fileContent)
}
