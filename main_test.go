package swagger

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/davidebianchi/gswagger/apirouter"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestNewRouter(t *testing.T) {
	muxRouter := mux.NewRouter()
	mAPIRouter := apirouter.NewGorillaMuxRouter(muxRouter)

	info := &openapi3.Info{
		Title:   "my title",
		Version: "my version",
	}
	openapi := &openapi3.T{
		Info:  info,
		Paths: openapi3.Paths{},
	}

	t.Run("not ok - invalid Openapi option", func(t *testing.T) {
		r, err := NewRouter(mAPIRouter, Options{})

		require.Nil(t, r)
		require.EqualError(t, err, fmt.Sprintf("%s: swagger is required", ErrValidatingSwagger))
	})

	t.Run("ok - with default context", func(t *testing.T) {
		r, err := NewRouter(mAPIRouter, Options{
			Openapi: openapi,
		})

		require.NoError(t, err)
		require.Equal(t, &Router{
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
		require.Equal(t, &Router{
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
		require.Equal(t, &Router{
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
	t.Run("not ok - empty swagger info", func(t *testing.T) {
		swagger := &openapi3.T{}

		swagger, err := generateNewValidSwagger(swagger)
		require.Nil(t, swagger)
		require.EqualError(t, err, "swagger info is required")
	})

	t.Run("not ok - empty info title", func(t *testing.T) {
		swagger := &openapi3.T{
			Info: &openapi3.Info{},
		}

		swagger, err := generateNewValidSwagger(swagger)
		require.Nil(t, swagger)
		require.EqualError(t, err, "swagger info title is required")
	})

	t.Run("not ok - empty info version", func(t *testing.T) {
		swagger := &openapi3.T{
			Info: &openapi3.Info{
				Title: "title",
			},
		}

		swagger, err := generateNewValidSwagger(swagger)
		require.Nil(t, swagger)
		require.EqualError(t, err, "swagger info version is required")
	})

	t.Run("ok - custom swagger", func(t *testing.T) {
		swagger := &openapi3.T{
			Info: &openapi3.Info{},
		}

		swagger, err := generateNewValidSwagger(swagger)
		require.Nil(t, swagger)
		require.EqualError(t, err, "swagger info title is required")
	})

	t.Run("not ok - swagger is required", func(t *testing.T) {
		swagger, err := generateNewValidSwagger(nil)
		require.Nil(t, swagger)
		require.EqualError(t, err, "swagger is required")
	})

	t.Run("ok", func(t *testing.T) {
		info := &openapi3.Info{
			Title:   "my title",
			Version: "my version",
		}
		swagger := &openapi3.T{
			Info: info,
		}

		swagger, err := generateNewValidSwagger(swagger)
		require.NoError(t, err)
		require.Equal(t, &openapi3.T{
			OpenAPI: defaultOpenapiVersion,
			Info:    info,
			Paths:   openapi3.Paths{},
		}, swagger)
	})
}

func TestGenerateAndExposeSwagger(t *testing.T) {
	t.Run("fails swagger validation", func(t *testing.T) {
		mRouter := mux.NewRouter()
		router, err := NewRouter(apirouter.NewGorillaMuxRouter(mRouter), Options{
			Openapi: &openapi3.T{
				Info: &openapi3.Info{
					Title:   "title",
					Version: "version",
				},
				Components: openapi3.Components{
					Schemas: map[string]*openapi3.SchemaRef{
						"&%": {},
					},
				},
			},
		})
		require.NoError(t, err)

		err = router.GenerateAndExposeSwagger()
		require.Error(t, err)
		require.True(t, strings.HasPrefix(err.Error(), fmt.Sprintf("%s:", ErrValidatingSwagger)))
	})

	t.Run("correctly expose json documentation from loaded swagger file", func(t *testing.T) {
		mRouter := mux.NewRouter()

		swagger, err := openapi3.NewLoader().LoadFromFile("testdata/users_employees.json")
		require.NoError(t, err)

		router, err := NewRouter(apirouter.NewGorillaMuxRouter(mRouter), Options{
			Openapi: swagger,
		})
		require.NoError(t, err)

		err = router.GenerateAndExposeSwagger()
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, DefaultJSONDocumentationPath, nil)
		mRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.True(t, strings.Contains(w.Result().Header.Get("content-type"), "application/json"))

		body := readBody(t, w.Result().Body)
		actual, err := ioutil.ReadFile("testdata/users_employees.json")
		require.NoError(t, err)
		require.JSONEq(t, string(actual), body)
	})

	t.Run("correctly expose json documentation from loaded swagger file - custom path", func(t *testing.T) {
		mRouter := mux.NewRouter()

		swagger, err := openapi3.NewLoader().LoadFromFile("testdata/users_employees.json")
		require.NoError(t, err)

		router, err := NewRouter(apirouter.NewGorillaMuxRouter(mRouter), Options{
			Openapi:               swagger,
			JSONDocumentationPath: "/custom/path",
		})
		require.NoError(t, err)

		err = router.GenerateAndExposeSwagger()
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/custom/path", nil)
		mRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.True(t, strings.Contains(w.Result().Header.Get("content-type"), "application/json"))

		body := readBody(t, w.Result().Body)
		actual, err := ioutil.ReadFile("testdata/users_employees.json")
		require.NoError(t, err)
		require.JSONEq(t, string(actual), body)
	})

	t.Run("correctly expose yaml documentation from loaded swagger file", func(t *testing.T) {
		mRouter := mux.NewRouter()

		swagger, err := openapi3.NewLoader().LoadFromFile("testdata/users_employees.json")
		require.NoError(t, err)

		router, err := NewRouter(apirouter.NewGorillaMuxRouter(mRouter), Options{
			Openapi: swagger,
		})
		require.NoError(t, err)

		err = router.GenerateAndExposeSwagger()
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, DefaultYAMLDocumentationPath, nil)
		mRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.True(t, strings.Contains(w.Result().Header.Get("content-type"), "text/plain"))

		body := readBody(t, w.Result().Body)
		expected, err := ioutil.ReadFile("testdata/users_employees.yaml")
		require.NoError(t, err)
		require.YAMLEq(t, string(expected), body, string(body))
	})

	t.Run("correctly expose yaml documentation from loaded swagger file - custom path", func(t *testing.T) {
		mRouter := mux.NewRouter()

		swagger, err := openapi3.NewLoader().LoadFromFile("testdata/users_employees.json")
		require.NoError(t, err)

		router, err := NewRouter(apirouter.NewGorillaMuxRouter(mRouter), Options{
			Openapi:               swagger,
			YAMLDocumentationPath: "/custom/path",
		})
		require.NoError(t, err)

		err = router.GenerateAndExposeSwagger()
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/custom/path", nil)
		mRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.True(t, strings.Contains(w.Result().Header.Get("content-type"), "text/plain"))

		body := readBody(t, w.Result().Body)
		expected, err := ioutil.ReadFile("testdata/users_employees.yaml")
		require.NoError(t, err)
		require.YAMLEq(t, string(expected), body, string(body))
	})

	t.Run("ok - subrouter", func(t *testing.T) {
		mRouter := mux.NewRouter()

		router, err := NewRouter(apirouter.NewGorillaMuxRouter(mRouter), Options{
			Openapi: &openapi3.T{
				Info: &openapi3.Info{
					Title:   "test swagger title",
					Version: "test swagger version",
				},
			},
			JSONDocumentationPath: "/custom/path",
		})
		require.NoError(t, err)

		router.AddRoute(http.MethodGet, "/foo", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		}, Definitions{})

		mSubRouter := mRouter.PathPrefix("/prefix").Subrouter()
		subrouter, err := router.SubRouter(apirouter.NewGorillaMuxRouter(mSubRouter), SubRouterOptions{
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

		err = router.GenerateAndExposeSwagger()
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/custom/path", nil)
		mRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.True(t, strings.Contains(w.Result().Header.Get("content-type"), "application/json"))

		body := readBody(t, w.Result().Body)
		actual, err := ioutil.ReadFile("testdata/subrouter.json")
		require.NoError(t, err)
		require.JSONEq(t, string(actual), body)
	})

	t.Run("ok - new router with path prefix", func(t *testing.T) {
		mRouter := mux.NewRouter()

		router, err := NewRouter(apirouter.NewGorillaMuxRouter(mRouter), Options{
			Openapi: &openapi3.T{
				Info: &openapi3.Info{
					Title:   "test swagger title",
					Version: "test swagger version",
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

		err = router.GenerateAndExposeSwagger()
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/custom/path", nil)
		mRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.True(t, strings.Contains(w.Result().Header.Get("content-type"), "application/json"))

		body := readBody(t, w.Result().Body)
		actual, err := ioutil.ReadFile("testdata/router_with_prefix.json")
		require.NoError(t, err)
		require.JSONEq(t, string(actual), body)
	})
}

func readBody(t *testing.T, requestBody io.ReadCloser) string {
	t.Helper()

	body, err := ioutil.ReadAll(requestBody)
	require.NoError(t, err)

	return string(body)
}
