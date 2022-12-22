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
