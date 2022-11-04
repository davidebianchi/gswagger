package apirouter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestGorillaMuxRouter(t *testing.T) {
	muxRouter := mux.NewRouter()
	ar := NewGorillaMuxRouter(muxRouter)

	t.Run("create a new api router", func(t *testing.T) {
		require.Implements(t, (*Router)(nil), ar)
	})

	t.Run("add new route", func(t *testing.T) {
		route := ar.AddRoute(http.MethodGet, "/foo", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(200)
			w.Write(nil)
		})

		_, ok := route.(*mux.Route)
		require.True(t, ok)

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
}
