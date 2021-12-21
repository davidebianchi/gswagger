package swagger_test

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidebianchi/gswagger/echorouter"
	"github.com/davidebianchi/gswagger/gorillarouter"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

const (
	swaggerOpenapiTitle   = "test swagger title"
	swaggerOpenapiVersion = "test swagger version"
)

func TestIntegration(t *testing.T) {
	t.Run("router works correctly - gorilla", func(t *testing.T) {
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
		muxRouter, _ := setupEchoSwagger(t)

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

	t.Run("works correctly with subrouter - handles path prefix", func(t *testing.T) {
		muxRouter, swaggerRouter := setupSwagger(t)

		muxSubRouter := muxRouter.NewRoute().Subrouter()
		subRouter, err := swaggerRouter.SubRouter(gorillarouter.New(muxSubRouter), swagger.SubRouterOptions{
			PathPrefix: "/prefix",
		})
		require.NoError(t, err)

		subRouter.AddRoute(http.MethodGet, "/foo", okHandler, swagger.Definitions{})

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
}

func readBody(t *testing.T, requestBody io.ReadCloser) string {
	t.Helper()

	body, err := ioutil.ReadAll(requestBody)
	require.NoError(t, err)

	return string(body)
}

func setupSwagger(t *testing.T) (*mux.Router, *swagger.Router) {
	t.Helper()

	context := context.Background()
	muxRouter := mux.NewRouter()

	router, err := swagger.NewRouter(gorillarouter.New(muxRouter), swagger.Options{
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

func setupEchoSwagger(t *testing.T) (*echo.Echo, *swagger.Router) {
	t.Helper()

	context := context.Background()
	echoRouter := echo.New()

	router, err := swagger.NewRouter(echorouter.New(echoRouter), swagger.Options{
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

	_, err = router.AddRawRoute(http.MethodGet, "/hello", echoOkHandler, operation)
	require.NoError(t, err)

	err = router.GenerateAndExposeSwagger()
	require.NoError(t, err)

	return echoRouter, router
}

func echoOkHandler(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}
