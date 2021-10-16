package swagger

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

const (
	swaggerOpenapiTitle   = "test swagger title"
	swaggerOpenapiVersion = "test swagger version"
)

func TestIntegration(t *testing.T) {
	t.Run("router works correctly", func(t *testing.T) {
		muxRouter := setupSwagger(t)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/hello", nil)

		muxRouter.ServeHTTP(w, r)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)

		body := readBody(t, w.Result().Body)
		require.Equal(t, "OK", body)

		t.Run("and generate swagger", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, DefaultJSONDocumentationPath, nil)

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

func setupSwagger(t *testing.T) *mux.Router {
	t.Helper()

	context := context.Background()
	muxRouter := mux.NewRouter()

	router, err := NewRouter(muxRouter, Options{
		Context: context,
		Openapi: &openapi3.T{
			Info: &openapi3.Info{
				Title:   swaggerOpenapiTitle,
				Version: swaggerOpenapiVersion,
			},
		},
	})
	require.NoError(t, err)

	handler := func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`OK`))
	}
	operation := Operation{}

	_, err = router.AddRawRoute(http.MethodGet, "/hello", handler, operation)
	require.NoError(t, err)

	err = router.GenerateAndExposeSwagger()
	require.NoError(t, err)

	return muxRouter
}
