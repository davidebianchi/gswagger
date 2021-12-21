package echorouter

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidebianchi/gswagger/apirouter"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestGorillaMuxRouter(t *testing.T) {
	echoRouter := echo.New()
	ar := New(echoRouter)

	h := func(c echo.Context) error {
		return c.String(200, "")
	}

	t.Run("create a new api router", func(t *testing.T) {
		require.Implements(t, (*apirouter.Router)(nil), ar)
	})

	t.Run("add new route with func(w http.ResponseWriter, req *http.Request)", func(t *testing.T) {
		route, err := ar.AddRoute(http.MethodGet, "/foo", h)
		require.NoError(t, err)

		_, ok := route.(*echo.Route)
		require.True(t, ok)

		t.Run("router exposes correctly api", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/foo", nil)

			echoRouter.ServeHTTP(w, r)

			require.Equal(t, http.StatusOK, w.Result().StatusCode)
		})

		t.Run("router exposes api only to the specific method", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/foo", nil)

			echoRouter.ServeHTTP(w, r)

			require.Equal(t, http.StatusMethodNotAllowed, w.Result().StatusCode)
		})
	})

	t.Run("create a correct swagger handler", func(t *testing.T) {
		echoRouter := echo.New()
		ar := New(echoRouter)

		_, err := ar.AddRoute(http.MethodGet, "/path", ar.SwaggerHandler("application/json", []byte("{}")))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/path", nil)
		w := httptest.NewRecorder()

		echoRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.Equal(t, "application/json", w.Result().Header.Get("Content-Type"))
		b, err := ioutil.ReadAll(w.Result().Body)
		require.NoError(t, err)
		require.Equal(t, []byte("{}"), b)
	})

	t.Run("add new route fails if handler is not handled", func(t *testing.T) {
		route, err := ar.AddRoute(http.MethodGet, "/foo", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(200)
			w.Write(nil)
		})
		require.Nil(t, route)
		require.EqualError(t, err, fmt.Sprintf("%s: handler type for echo is not handled: method GET, path /foo", apirouter.ErrInvalidHandler))
	})
}
