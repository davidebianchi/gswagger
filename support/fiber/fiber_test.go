package fiber

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidebianchi/gswagger/apirouter"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestFiberRouterSupport(t *testing.T) {
	fiberRouter := fiber.New()
	ar := NewRouter(fiberRouter)

	t.Run("create a new api router", func(t *testing.T) {
		require.Implements(t, (*apirouter.Router[HandlerFunc])(nil), ar)
	})

	t.Run("add new route", func(t *testing.T) {
		route := ar.AddRoute(http.MethodGet, "/foo", func(c *fiber.Ctx) error {
			return c.SendStatus(http.StatusOK)
		})

		_, ok := route.(fiber.Router)
		require.True(t, ok)

		t.Run("router exposes correctly api", func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/foo", nil)

			resp, err := fiberRouter.Test(r)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)
		})

		t.Run("router exposes api only to the specific method", func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/foo", nil)

			resp, err := fiberRouter.Test(r)
			require.NoError(t, err)
			require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		})
	})

	t.Run("create swagger handler", func(t *testing.T) {
		handlerFunc := ar.SwaggerHandler("text/html", []byte("some data"))
		fiberRouter.Get("/oas", handlerFunc)

		t.Run("responds correctly to the API", func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/oas", nil)

			resp, err := fiberRouter.Test(r)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			require.Equal(t, "text/html", resp.Header.Get("Content-Type"))

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, "some data", string(body))
		})
	})
}
