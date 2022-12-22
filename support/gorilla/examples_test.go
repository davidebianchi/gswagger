package gorilla_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/davidebianchi/gswagger/support/gorilla"
	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
)

func TestExample(t *testing.T) {
	context := context.Background()
	muxRouter := mux.NewRouter()

	router, _ := swagger.NewRouter(gorilla.NewRouter(muxRouter), swagger.Options{
		Context: context,
		Openapi: &openapi3.T{
			Info: &openapi3.Info{
				Title:   "my title",
				Version: "1.0.0",
			},
		},
	})

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
	type errorResponse struct {
		Message string `json:"message"`
	}

	router.AddRoute(http.MethodPost, "/users", okHandler, swagger.Definitions{
		RequestBody: &swagger.ContentValue{
			Content: swagger.Content{
				"application/json": {Value: User{}},
			},
		},
		Responses: map[int]swagger.ContentValue{
			201: {
				Content: swagger.Content{
					"text/html": {Value: ""},
				},
			},
			401: {
				Content: swagger.Content{
					"application/json": {Value: &errorResponse{}},
				},
				Description: "invalid request",
			},
		},
	})

	router.AddRoute(http.MethodGet, "/users", okHandler, swagger.Definitions{
		Responses: map[int]swagger.ContentValue{
			200: {
				Content: swagger.Content{
					"application/json": {Value: &[]User{}},
				},
			},
		},
	})

	carSchema := openapi3.NewObjectSchema().WithProperties(map[string]*openapi3.Schema{
		"foo": openapi3.NewStringSchema(),
		"bar": openapi3.NewIntegerSchema().WithMax(15).WithMin(5),
	})
	requestBody := openapi3.NewRequestBody().WithJSONSchema(carSchema)
	operation := swagger.NewOperation()
	operation.AddRequestBody(requestBody)

	router.AddRawRoute(http.MethodPost, "/cars", okHandler, operation)

	_, err := router.AddRoute(http.MethodGet, "/users/{userId}", okHandler, swagger.Definitions{
		Querystring: swagger.ParameterValue{
			"query": {
				Schema: &swagger.Schema{Value: ""},
			},
		},
	})
	require.NoError(t, err)

	_, err = router.AddRoute(http.MethodGet, "/cars/{carId}/drivers/{driverId}", okHandler, swagger.Definitions{})
	require.NoError(t, err)

	router.GenerateAndExposeOpenapi()

	t.Run("correctly exposes documentation", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, swagger.DefaultJSONDocumentationPath, nil)

		muxRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)
		require.Equal(t, "application/json", w.Result().Header.Get("Content-Type"))

		body, err := io.ReadAll(w.Result().Body)
		require.NoError(t, err)
		expected, err := os.ReadFile("./testdata/examples-users.json")
		require.NoError(t, err)
		require.JSONEq(t, string(expected), string(body), "actual json data: %s", body)
	})
}
