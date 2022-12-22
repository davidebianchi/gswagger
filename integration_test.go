package swagger_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/davidebianchi/gswagger/apirouter"
	"github.com/davidebianchi/gswagger/support/gorilla"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

const (
	swaggerOpenapiTitle   = "test swagger title"
	swaggerOpenapiVersion = "test swagger version"
)

type GorillaSwaggerRouter = swagger.Router[gorilla.HandlerFunc, gorilla.Route]
type echoSwaggerRouter = swagger.Router[http.HandlerFunc, *echo.Route]

func TestIntegration(t *testing.T) {
	t.Run("router works correctly", func(t *testing.T) {
		muxRouter, _ := setupSwagger(t)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/hello", nil)

		muxRouter.ServeHTTP(w, r)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)

		body := readBody(t, w.Result().Body)
		require.Equal(t, "OK", body)

		t.Run("and generate swagger", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, swagger.DefaultJSONDocumentationPath, nil)

			muxRouter.ServeHTTP(w, r)

			require.Equal(t, http.StatusOK, w.Result().StatusCode)

			body := readBody(t, w.Result().Body)
			require.Equal(t, "{\"components\":{},\"info\":{\"title\":\"test swagger title\",\"version\":\"test swagger version\"},\"openapi\":\"3.0.0\",\"paths\":{\"/hello\":{\"get\":{\"responses\":{\"default\":{\"description\":\"\"}}}}}}", body)
		})
	})

	t.Run("router works correctly - echo", func(t *testing.T) {
		echoRouter, _ := setupEchoSwagger(t)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/hello", nil)

		echoRouter.ServeHTTP(w, r)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)

		body := readBody(t, w.Result().Body)
		require.Equal(t, "OK", body)

		t.Run("and generate swagger", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, swagger.DefaultJSONDocumentationPath, nil)

			echoRouter.ServeHTTP(w, r)

			require.Equal(t, http.StatusOK, w.Result().StatusCode)

			body := readBody(t, w.Result().Body)
			require.Equal(t, "{\"components\":{},\"info\":{\"title\":\"test swagger title\",\"version\":\"test swagger version\"},\"openapi\":\"3.0.0\",\"paths\":{\"/hello\":{\"get\":{\"responses\":{\"default\":{\"description\":\"\"}}}}}}", body)
		})
	})

	t.Run("works correctly with subrouter - handles path prefix - gorilla mux", func(t *testing.T) {
		muxRouter, swaggerRouter := setupSwagger(t)

		muxSubRouter := muxRouter.NewRoute().Subrouter()
		subRouter, err := swaggerRouter.SubRouter(gorilla.NewRouter(muxSubRouter), swagger.SubRouterOptions{
			PathPrefix: "/prefix",
		})
		require.NoError(t, err)

		_, err = subRouter.AddRoute(http.MethodGet, "/foo", okHandler, swagger.Definitions{})
		require.NoError(t, err)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/hello", nil)

		muxRouter.ServeHTTP(w, r)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)

		body := readBody(t, w.Result().Body)
		require.Equal(t, "OK", body)

		t.Run("correctly call sub router", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/prefix/foo", nil)

			muxRouter.ServeHTTP(w, r)

			require.Equal(t, http.StatusOK, w.Result().StatusCode)

			body := readBody(t, w.Result().Body)
			require.Equal(t, "OK", body)
		})

		t.Run("and generate swagger", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, swagger.DefaultJSONDocumentationPath, nil)

			muxRouter.ServeHTTP(w, r)

			require.Equal(t, http.StatusOK, w.Result().StatusCode)

			body := readBody(t, w.Result().Body)
			require.Equal(t, "{\"components\":{},\"info\":{\"title\":\"test swagger title\",\"version\":\"test swagger version\"},\"openapi\":\"3.0.0\",\"paths\":{\"/hello\":{\"get\":{\"responses\":{\"default\":{\"description\":\"\"}}}}}}", body)
		})
	})

	t.Run("works correctly with subrouter - handles path prefix - echo", func(t *testing.T) {
		eRouter, swaggerRouter := setupEchoSwagger(t)

		subRouter, err := swaggerRouter.SubRouter(echoRouter{router: eRouter}, swagger.SubRouterOptions{
			PathPrefix: "/prefix",
		})
		require.NoError(t, err)

		_, err = subRouter.AddRoute(http.MethodGet, "/foo", okHandler, swagger.Definitions{})
		require.NoError(t, err)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/hello", nil)

		eRouter.ServeHTTP(w, r)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)

		body := readBody(t, w.Result().Body)
		require.Equal(t, "OK", body)

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
			require.Equal(t, "{\"components\":{},\"info\":{\"title\":\"test swagger title\",\"version\":\"test swagger version\"},\"openapi\":\"3.0.0\",\"paths\":{\"/hello\":{\"get\":{\"responses\":{\"default\":{\"description\":\"\"}}}}}}", body)
		})
	})
}

func readBody(t *testing.T, requestBody io.ReadCloser) string {
	t.Helper()

	body, err := io.ReadAll(requestBody)
	require.NoError(t, err)

	return string(body)
}

func setupSwagger(t *testing.T) (*mux.Router, *GorillaSwaggerRouter) {
	t.Helper()

	context := context.Background()
	muxRouter := mux.NewRouter()

	router, err := swagger.NewRouter(gorilla.NewRouter(muxRouter), swagger.Options{
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

	return muxRouter, router
}

func okHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`OK`))
}

func setupEchoSwagger(t *testing.T) (*echo.Echo, *echoSwaggerRouter) {
	t.Helper()

	context := context.Background()
	e := echo.New()

	router, err := swagger.NewRouter(newEchoRouter(e), swagger.Options{
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

	_, err = router.AddRawRoute(http.MethodGet, "/hello", func(w http.ResponseWriter, req *http.Request) {
		ctx := e.NewContext(req, w)
		echoOkHandler(ctx)
	}, operation)
	require.NoError(t, err)

	err = router.GenerateAndExposeSwagger()
	require.NoError(t, err)

	return e, router
}

func echoOkHandler(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func newEchoRouter(e *echo.Echo) apirouter.Router[http.HandlerFunc, *echo.Route] {
	return echoRouter{router: e}
}

type echoRouter struct {
	router *echo.Echo
}

func (r echoRouter) AddRoute(method, path string, handler http.HandlerFunc) *echo.Route {
	return r.router.Add(method, path, echo.WrapHandler(http.HandlerFunc(handler)))
}

func (r echoRouter) SwaggerHandler(contentType string, blob []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		w.Write(blob)
	}
}
