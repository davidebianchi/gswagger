package swagger

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestAddRoute(t *testing.T) {
	okHandler := func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}

	type User struct {
		Name        string   `json:"name" jsonschema:"title=The user name,required" jsonschema_extras:"example=Jane"`
		PhoneNumber int      `json:"phone" jsonschema:"title=mobile number of user"`
		Groups      []string `json:"groups,omitempty" jsonschema:"title=groups of the user,default=users"`
		Address     string   `json:"address" jsonschema:"title=user address"`
	}
	type Users []User
	type errorResponse struct {
		Message string `json:"message"`
	}

	type Employees struct {
		OrganizationName string `json:"organization_name"`
		Users            Users  `json:"users" jsonschema:"selected users"`
	}
	type FormData struct {
		ID      string `json:"id,omitempty"`
		Address struct {
			Street string `json:"street,omitempty"`
			City   string `json:"city,omitempty"`
		} `json:"address,omitempty"`
		ProfileImage string `json:"profileImage,omitempty" jsonschema_extras:"format=binary"`
	}

	getBaseSwagger := func() *openapi3.Swagger {
		return &openapi3.Swagger{
			Info: &openapi3.Info{
				Title:   swaggerOpenapiTitle,
				Version: swaggerOpenapiVersion,
			},
		}
	}

	tests := []struct {
		name         string
		routes       func(t *testing.T, router *Router)
		fixturesPath string
		testPath     string
		testMethod   string
	}{
		{
			name:         "no routes",
			routes:       func(t *testing.T, router *Router) {},
			fixturesPath: "testdata/empty.json",
		},
		{
			name: "empty route schema",
			routes: func(t *testing.T, router *Router) {
				_, err := router.AddRoute(http.MethodPost, "/", okHandler, Schema{})
				require.NoError(t, err)
			},
			testPath:     "/",
			testMethod:   http.MethodPost,
			fixturesPath: "testdata/empty-route-schema.json",
		},
		{
			name: "multiple real routes",
			routes: func(t *testing.T, router *Router) {
				_, err := router.AddRoute(http.MethodPost, "/users", okHandler, Schema{
					RequestBody: &SchemaValue{
						Content: User{},
					},
					Responses: map[int]SchemaValue{
						201: {
							Content: "",
						},
						401: {
							Content:     &errorResponse{},
							Description: "invalid request",
						},
					},
				})
				require.NoError(t, err)

				_, err = router.AddRoute(http.MethodGet, "/users", okHandler, Schema{
					Responses: map[int]SchemaValue{
						200: {
							Content: &Users{},
						},
					},
				})
				require.NoError(t, err)

				_, err = router.AddRoute(http.MethodGet, "/employees", okHandler, Schema{
					Responses: map[int]SchemaValue{
						200: {
							Content: &Employees{},
						},
					},
				})
				require.NoError(t, err)
			},
			testPath:     "/users",
			fixturesPath: "testdata/users_employees.json",
		},
		{
			name: "multipart request body",
			routes: func(t *testing.T, router *Router) {
				_, err := router.AddRoute(http.MethodPost, "/files", okHandler, Schema{
					RequestBody: &SchemaValue{
						Content:                   &FormData{},
						Description:               "upload file",
						ContentType:               "multipart/form-data",
						AllowAdditionalProperties: true,
					},
					Responses: map[int]SchemaValue{
						200: {Content: ""},
					},
				})
				require.NoError(t, err)
			},
			testPath:     "/files",
			testMethod:   http.MethodPost,
			fixturesPath: "testdata/multipart-requestbody.json",
		},
		{
			name: "schema with params",
			routes: func(t *testing.T, router *Router) {
				var number = 0
				_, err := router.AddRoute(http.MethodGet, "/users/{userId}", okHandler, Schema{
					PathParams: map[string]SchemaValue{
						"userId": {
							Content:     number,
							Description: "userId is a number above 0",
						},
					},
				})
				require.NoError(t, err)

				_, err = router.AddRoute(http.MethodGet, "/cars/{carId}/drivers/{driverId}", okHandler, Schema{
					PathParams: map[string]SchemaValue{
						"carId": {
							Content: "",
						},
						"driverId": {
							Content: "",
						},
					},
				})
				require.NoError(t, err)
			},
			testPath:     "/users/12",
			fixturesPath: "testdata/params.json",
		},
		{
			name: "schema with querystring",
			routes: func(t *testing.T, router *Router) {
				_, err := router.AddRoute(http.MethodGet, "/projects", okHandler, Schema{
					Querystring: map[string]SchemaValue{
						"projectId": {
							Content:     "",
							Description: "projectId is the project id",
						},
					},
				})
				require.NoError(t, err)
			},
			testPath:     "/projects",
			fixturesPath: "testdata/query.json",
		},
		{
			name: "schema with headers",
			routes: func(t *testing.T, router *Router) {
				_, err := router.AddRoute(http.MethodGet, "/projects", okHandler, Schema{
					Headers: map[string]SchemaValue{
						"foo": {
							Content:     "",
							Description: "foo description",
						},
						"bar": {
							Content: "",
						},
					},
				})
				require.NoError(t, err)
			},
			testPath:     "/projects",
			fixturesPath: "testdata/headers.json",
		},
		{
			name: "schema with cookies",
			routes: func(t *testing.T, router *Router) {
				_, err := router.AddRoute(http.MethodGet, "/projects", okHandler, Schema{
					Cookies: map[string]SchemaValue{
						"debug": {
							Content:     0,
							Description: "boolean. Set 0 to disable and 1 to enable",
						},
						"csrftoken": {
							Content: "",
						},
					},
				})
				require.NoError(t, err)
			},
			testPath:     "/projects",
			fixturesPath: "testdata/cookies.json",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context := context.Background()
			r := mux.NewRouter()

			router, err := New(r, Options{
				Context: context,
				Openapi: getBaseSwagger(),
			})
			require.NoError(t, err)
			require.NotNil(t, router)

			// Add routes to test
			test.routes(t, router)

			err = router.GenerateAndExposeSwagger()
			require.NoError(t, err)

			if test.testPath != "" {
				if test.testMethod == "" {
					test.testMethod = http.MethodGet
				}

				w := httptest.NewRecorder()
				req := httptest.NewRequest(test.testMethod, test.testPath, nil)
				r.ServeHTTP(w, req)

				require.Equal(t, http.StatusOK, w.Result().StatusCode)

				body := readBody(t, w.Result().Body)
				require.Equal(t, "OK", body)
			}

			t.Run("and generate swagger documentation in json", func(t *testing.T) {
				w := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, JSONDocumentationPath, nil)

				r.ServeHTTP(w, req)

				require.Equal(t, http.StatusOK, w.Result().StatusCode)

				body := readBody(t, w.Result().Body)
				actual, err := ioutil.ReadFile(test.fixturesPath)
				require.NoError(t, err)
				t.Log("actual json schema", body)
				require.JSONEq(t, string(actual), body)
			})
		})
	}
}

func TestGenerateAndExposeSwagger(t *testing.T) {
	t.Run("fails swagger validation", func(t *testing.T) {
		mRouter := mux.NewRouter()
		router, err := New(mRouter, Options{
			Openapi: &openapi3.Swagger{
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

		swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromFile("testdata/users_employees.json")
		require.NoError(t, err)

		router, err := New(mRouter, Options{
			Openapi: swagger,
		})

		err = router.GenerateAndExposeSwagger()
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, JSONDocumentationPath, nil)
		mRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.True(t, strings.Contains(w.Result().Header.Get("content-type"), "application/json"))

		body := readBody(t, w.Result().Body)
		actual, err := ioutil.ReadFile("testdata/users_employees.json")
		require.NoError(t, err)
		require.JSONEq(t, string(actual), body)
	})
}
