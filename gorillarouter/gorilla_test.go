package gorillarouter

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidebianchi/gswagger/apirouter"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestGorillaMuxRouter(t *testing.T) {
	muxRouter := mux.NewRouter()
	ar := New(muxRouter)

	t.Run("create a new api router", func(t *testing.T) {
		require.Implements(t, (*apirouter.Router)(nil), ar)
	})

	t.Run("add new route with func(w http.ResponseWriter, req *http.Request)", func(t *testing.T) {
		h := func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(200)
			w.Write(nil)
		}
		route, err := ar.AddRoute(http.MethodGet, "/foo", h)
		require.NoError(t, err)

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

	t.Run("create a correct swagger handler", func(t *testing.T) {
		muxRouter := mux.NewRouter()
		ar := New(muxRouter)

		_, err := ar.AddRoute(http.MethodGet, "/path", ar.SwaggerHandler("application/json", []byte("{}")))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/path", nil)
		w := httptest.NewRecorder()

		muxRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.Equal(t, "application/json", w.Result().Header.Get("Content-Type"))
		b, err := ioutil.ReadAll(w.Result().Body)
		require.NoError(t, err)
		require.Equal(t, []byte("{}"), b)
	})

	t.Run("add new route fails if handler is not handled", func(t *testing.T) {
		type HandleFunc func(w http.ResponseWriter, req *http.Request)
		route, err := ar.AddRoute(http.MethodGet, "/foo", HandleFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(200)
			w.Write(nil)
		}))
		require.Nil(t, route)
		require.EqualError(t, err, fmt.Sprintf("%s: handler type for gorilla is not handled: method GET, path /foo", apirouter.ErrInvalidHandler))
	})
}
