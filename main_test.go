package swagger

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/davidebianchi/gswagger/support/gorilla"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestNewRouter(t *testing.T) {
	muxRouter := mux.NewRouter()
	mAPIRouter := gorilla.NewRouter(muxRouter)

	info := &openapi3.Info{
		Title:   "my title",
		Version: "my version",
	}
	openapi := &openapi3.T{
		Info:  info,
		Paths: &openapi3.Paths{},
	}

	t.Run("not ok - invalid Openapi option", func(t *testing.T) {
		r, err := NewRouter(mAPIRouter, Options{})

		require.Nil(t, r)
		require.EqualError(t, err, fmt.Sprintf("%s: openapi is required", ErrValidatingOAS))
	})

	t.Run("ok - with default context", func(t *testing.T) {
		r, err := NewRouter(mAPIRouter, Options{
			Openapi: openapi,
		})

		require.NoError(t, err)
		require.Equal(t, &Router[gorilla.HandlerFunc, gorilla.Route]{
			context:               context.Background(),
			router:                mAPIRouter,
			swaggerSchema:         openapi,
			jsonDocumentationPath: DefaultJSONDocumentationPath,
			yamlDocumentationPath: DefaultYAMLDocumentationPath,
		}, r)
	})

	t.Run("ok - with custom context", func(t *testing.T) {
		type key struct{}
		ctx := context.WithValue(context.Background(), key{}, "value")
		r, err := NewRouter(mAPIRouter, Options{
			Openapi: openapi,
			Context: ctx,
		})

		require.NoError(t, err)
		require.Equal(t, &Router[gorilla.HandlerFunc, gorilla.Route]{
			context:               ctx,
			router:                mAPIRouter,
			swaggerSchema:         openapi,
			jsonDocumentationPath: DefaultJSONDocumentationPath,
			yamlDocumentationPath: DefaultYAMLDocumentationPath,
		}, r)
	})

	t.Run("ok - with custom docs paths", func(t *testing.T) {
		type key struct{}
		ctx := context.WithValue(context.Background(), key{}, "value")
		r, err := NewRouter(mAPIRouter, Options{
			Openapi:               openapi,
			Context:               ctx,
			JSONDocumentationPath: "/json/path",
			YAMLDocumentationPath: "/yaml/path",
		})

		require.NoError(t, err)
		require.Equal(t, &Router[gorilla.HandlerFunc, gorilla.Route]{
			context:               ctx,
			router:                mAPIRouter,
			swaggerSchema:         openapi,
			jsonDocumentationPath: "/json/path",
			yamlDocumentationPath: "/yaml/path",
		}, r)
	})

	t.Run("ko - json documentation path does not start with /", func(t *testing.T) {
		type key struct{}
		ctx := context.WithValue(context.Background(), key{}, "value")
		r, err := NewRouter(mAPIRouter, Options{
			Openapi:               openapi,
			Context:               ctx,
			JSONDocumentationPath: "json/path",
			YAMLDocumentationPath: "/yaml/path",
		})

		require.EqualError(t, err, "invalid path json/path. Path should start with '/'")
		require.Nil(t, r)
	})

	t.Run("ko - yaml documentation path does not start with /", func(t *testing.T) {
		type key struct{}
		ctx := context.WithValue(context.Background(), key{}, "value")
		r, err := NewRouter(mAPIRouter, Options{
			Openapi:               openapi,
			Context:               ctx,
			JSONDocumentationPath: "/json/path",
			YAMLDocumentationPath: "yaml/path",
		})

		require.EqualError(t, err, "invalid path yaml/path. Path should start with '/'")
		require.Nil(t, r)
	})
}

func TestGenerateValidSwagger(t *testing.T) {
	t.Run("not ok - empty openapi info", func(t *testing.T) {
		openapi := &openapi3.T{}

		openapi, err := generateNewValidOpenapi(openapi)
		require.Nil(t, openapi)
		require.EqualError(t, err, "openapi info is required")
	})

	t.Run("not ok - empty info title", func(t *testing.T) {
		openapi := &openapi3.T{
			Info: &openapi3.Info{},
		}

		openapi, err := generateNewValidOpenapi(openapi)
		require.Nil(t, openapi)
		require.EqualError(t, err, "openapi info title is required")
	})

	t.Run("not ok - empty info version", func(t *testing.T) {
		openapi := &openapi3.T{
			Info: &openapi3.Info{
				Title: "title",
			},
		}

		openapi, err := generateNewValidOpenapi(openapi)
		require.Nil(t, openapi)
		require.EqualError(t, err, "openapi info version is required")
	})

	t.Run("ok - custom openapi", func(t *testing.T) {
		openapi := &openapi3.T{
			Info: &openapi3.Info{},
		}

		openapi, err := generateNewValidOpenapi(openapi)
		require.Nil(t, openapi)
		require.EqualError(t, err, "openapi info title is required")
	})

	t.Run("not ok - openapi is required", func(t *testing.T) {
		openapi, err := generateNewValidOpenapi(nil)
		require.Nil(t, openapi)
		require.EqualError(t, err, "openapi is required")
	})

	t.Run("ok", func(t *testing.T) {
		info := &openapi3.Info{
			Title:   "my title",
			Version: "my version",
		}
		openapi := &openapi3.T{
			Info: info,
		}

		openapi, err := generateNewValidOpenapi(openapi)
		require.NoError(t, err)
		require.Equal(t, &openapi3.T{
			OpenAPI: defaultOpenapiVersion,
			Info:    info,
			Paths:   &openapi3.Paths{},
		}, openapi)
	})
}

func TestGenerateAndExposeSwagger(t *testing.T) {
	t.Run("fails openapi validation", func(t *testing.T) {
		mRouter := mux.NewRouter()
		router, err := NewRouter(gorilla.NewRouter(mRouter), Options{
			Openapi: &openapi3.T{
				Info: &openapi3.Info{
					Title:   "title",
					Version: "version",
				},
				Components: &openapi3.Components{
					Schemas: map[string]*openapi3.SchemaRef{
						"&%": {},
					},
				},
			},
		})
		require.NoError(t, err)

		err = router.GenerateAndExposeOpenapi()
		require.Error(t, err)
		require.True(t, strings.HasPrefix(err.Error(), fmt.Sprintf("%s:", ErrValidatingOAS)))
	})

	t.Run("correctly expose json documentation from loaded openapi file", func(t *testing.T) {
		mRouter := mux.NewRouter()

		openapi, err := openapi3.NewLoader().LoadFromFile("testdata/users_employees.json")
		require.NoError(t, err)

		router, err := NewRouter(gorilla.NewRouter(mRouter), Options{
			Openapi: openapi,
		})
		require.NoError(t, err)

		err = router.GenerateAndExposeOpenapi()
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, DefaultJSONDocumentationPath, nil)
		mRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.True(t, strings.Contains(w.Result().Header.Get("content-type"), "application/json"))

		body := readBody(t, w.Result().Body)
		actual, err := os.ReadFile("testdata/users_employees.json")
		require.NoError(t, err)
		require.JSONEq(t, string(actual), body)
	})

	t.Run("correctly expose json documentation from loaded openapi file - custom path", func(t *testing.T) {
		mRouter := mux.NewRouter()

		openapi, err := openapi3.NewLoader().LoadFromFile("testdata/users_employees.json")
		require.NoError(t, err)

		router, err := NewRouter(gorilla.NewRouter(mRouter), Options{
			Openapi:               openapi,
			JSONDocumentationPath: "/custom/path",
		})
		require.NoError(t, err)

		err = router.GenerateAndExposeOpenapi()
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/custom/path", nil)
		mRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.True(t, strings.Contains(w.Result().Header.Get("content-type"), "application/json"))

		body := readBody(t, w.Result().Body)
		actual, err := os.ReadFile("testdata/users_employees.json")
		require.NoError(t, err)
		require.JSONEq(t, string(actual), body)
	})

	t.Run("correctly expose yaml documentation from loaded openapi file", func(t *testing.T) {
		mRouter := mux.NewRouter()

		openapi, err := openapi3.NewLoader().LoadFromFile("testdata/users_employees.json")
		require.NoError(t, err)

		router, err := NewRouter(gorilla.NewRouter(mRouter), Options{
			Openapi: openapi,
		})
		require.NoError(t, err)

		err = router.GenerateAndExposeOpenapi()
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, DefaultYAMLDocumentationPath, nil)
		mRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.True(t, strings.Contains(w.Result().Header.Get("content-type"), "text/plain"))

		body := readBody(t, w.Result().Body)
		expected, err := os.ReadFile("testdata/users_employees.yaml")
		require.NoError(t, err)
		require.YAMLEq(t, string(expected), body, string(body))
	})

	t.Run("correctly expose yaml documentation from loaded openapi file - custom path", func(t *testing.T) {
		mRouter := mux.NewRouter()

		openapi, err := openapi3.NewLoader().LoadFromFile("testdata/users_employees.json")
		require.NoError(t, err)

		router, err := NewRouter(gorilla.NewRouter(mRouter), Options{
			Openapi:               openapi,
			YAMLDocumentationPath: "/custom/path",
		})
		require.NoError(t, err)

		err = router.GenerateAndExposeOpenapi()
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/custom/path", nil)
		mRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.True(t, strings.Contains(w.Result().Header.Get("content-type"), "text/plain"))

		body := readBody(t, w.Result().Body)
		expected, err := os.ReadFile("testdata/users_employees.yaml")
		require.NoError(t, err)
		require.YAMLEq(t, string(expected), body, string(body))
	})

	t.Run("ok - subrouter", func(t *testing.T) {
		mRouter := mux.NewRouter()

		router, err := NewRouter(gorilla.NewRouter(mRouter), Options{
			Openapi: &openapi3.T{
				Info: &openapi3.Info{
					Title:   "test openapi title",
					Version: "test openapi version",
				},
			},
			JSONDocumentationPath: "/custom/path",
		})
		require.NoError(t, err)

		router.AddRoute(http.MethodGet, "/foo", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		}, Definitions{})

		mSubRouter := mRouter.NewRoute().Subrouter()
		subrouter, err := router.SubRouter(gorilla.NewRouter(mSubRouter), SubRouterOptions{
			PathPrefix: "/prefix",
		})
		require.NoError(t, err)

		_, err = subrouter.AddRoute(http.MethodGet, "/taz", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		}, Definitions{})
		require.NoError(t, err)

		t.Run("add route with path equal to prefix path", func(t *testing.T) {
			_, err = subrouter.AddRoute(http.MethodGet, "", func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok"))
			}, Definitions{})
			require.NoError(t, err)
		})

		err = router.GenerateAndExposeOpenapi()
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/custom/path", nil)
		mRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.True(t, strings.Contains(w.Result().Header.Get("content-type"), "application/json"))

		body := readBody(t, w.Result().Body)
		actual, err := os.ReadFile("testdata/subrouter.json")
		require.NoError(t, err)
		require.JSONEq(t, string(actual), body)

		t.Run("test request /prefix", func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/prefix", nil)
			mRouter.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Result().StatusCode)
		})

		t.Run("test request /prefix/taz", func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/prefix/taz", nil)
			mRouter.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Result().StatusCode)
		})

		t.Run("test request /foo", func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/foo", nil)
			mRouter.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Result().StatusCode)
		})
	})

	t.Run("ok - new router with path prefix", func(t *testing.T) {
		mRouter := mux.NewRouter()

		router, err := NewRouter(gorilla.NewRouter(mRouter), Options{
			Openapi: &openapi3.T{
				Info: &openapi3.Info{
					Title:   "test openapi title",
					Version: "test openapi version",
				},
			},
			JSONDocumentationPath: "/custom/path",
			PathPrefix:            "/prefix",
		})
		require.NoError(t, err)

		router.AddRoute(http.MethodGet, "/foo", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		}, Definitions{})

		err = router.GenerateAndExposeOpenapi()
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/custom/path", nil)
		mRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.True(t, strings.Contains(w.Result().Header.Get("content-type"), "application/json"))

		body := readBody(t, w.Result().Body)
		actual, err := os.ReadFile("testdata/router_with_prefix.json")
		require.NoError(t, err)
		require.JSONEq(t, string(actual), body, body)
	})
}

func readBody(t *testing.T, requestBody io.ReadCloser) string {
	t.Helper()

	body, err := io.ReadAll(requestBody)
	require.NoError(t, err)

	return string(body)
}
