package gorilla

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidebianchi/gswagger/apirouter"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestGorillaMuxRouter(t *testing.T) {
	muxRouter := mux.NewRouter()
	ar := NewRouter(muxRouter)

	t.Run("create a new api router", func(t *testing.T) {
		require.Implements(t, (*apirouter.Router[HandlerFunc, Route])(nil), ar)
	})

	t.Run("add new route", func(t *testing.T) {
		route := ar.AddRoute(http.MethodGet, "/foo", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(200)
			w.Write(nil)
		})
		require.IsType(t, route, &mux.Route{})

		t.Run("router exposes correctly api", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/foo", nil)

			muxRouter.ServeHTTP(w, r)

			require.Equal(t, http.StatusOK, w.Result().StatusCode)
		})

		t.Run("router exposes api only to the specific method", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/foo", nil)

			muxRouter.ServeHTTP(w, r)

			require.Equal(t, http.StatusMethodNotAllowed, w.Result().StatusCode)
		})
	})

	t.Run("create openapi handler", func(t *testing.T) {
		handlerFunc := ar.SwaggerHandler("text/html", []byte("some data"))
		muxRouter.HandleFunc("/oas", handlerFunc).Methods(http.MethodGet)

		t.Run("responds correctly to the API", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/oas", nil)

			muxRouter.ServeHTTP(w, r)

			require.Equal(t, http.StatusOK, w.Result().StatusCode)
			require.Equal(t, "text/html", w.Result().Header.Get("Content-Type"))

			body, err := io.ReadAll(w.Result().Body)
			require.NoError(t, err)
			require.Equal(t, "some data", string(body))
		})
	})
}

func TestTransformPath(t *testing.T) {
	testCases := []struct {
		name         string
		path         string
		expectedPath string
	}{
		{
			name:         "only /",
			path:         "/",
			expectedPath: "/",
		},
		{
			name:         "without params",
			path:         "/foo",
			expectedPath: "/foo",
		},
		{
			name:         "without params ending with /",
			path:         "/foo/",
			expectedPath: "/foo/",
		},
		{
			name:         "with params",
			path:         "/foo/{par1}",
			expectedPath: "/foo/{par1}",
		},
		{
			name:         "with params ending with /",
			path:         "/foo/{par1}/",
			expectedPath: "/foo/{par1}/",
		},
		{
			name:         "with multiple params",
			path:         "/{par1}/{par2}/{par3}",
			expectedPath: "/{par1}/{par2}/{par3}",
		},
		{
			name:         "with multiple params ending with /",
			path:         "/{par1}/{par2}/{par3}/",
			expectedPath: "/{par1}/{par2}/{par3}/",
		},
		{
			name:         "with multiple params in a segment",
			path:         "/foo/{par2}{par3}",
			expectedPath: "/foo/{par2}{par3}",
		},
		{
			name:         "with multiple params in a segment ending with /",
			path:         "/foo/{par2}{par3}/",
			expectedPath: "/foo/{par2}{par3}/",
		},
		{
			name:         "with regex",
			path:         "/foo/{par1:[0-9]}/{par2:[a-z]}",
			expectedPath: "/foo/{par1}/{par2}",
		},
		{
			name:         "with regex ending with /",
			path:         "/foo/{par1:[0-9]}/{par2:[a-z]}/",
			expectedPath: "/foo/{par1}/{par2}/",
		},
		{
			name:         "with multiple params in a segment and the regex",
			path:         "/foo/{par2:[0-9]}{par3:a|b}/",
			expectedPath: "/foo/{par2}{par3}/",
		},
		{
			name:         "with multiple params in a segment and the regex ending with /",
			path:         "/foo/{par2:[0-9]}{par3:\\w+}/",
			expectedPath: "/foo/{par2}{par3}/",
		},
	}

	router := NewRouter(mux.NewRouter())

	for _, test := range testCases {

		t.Run(test.name, func(t *testing.T) {
			actual := router.TransformPathToOasPath(test.path)
			require.Equal(t, test.expectedPath, actual)
		})
	}
}
