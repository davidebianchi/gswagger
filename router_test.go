package swagger

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestAddRoute(t *testing.T) {
	okHandler := func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`OK`))
	}

	t.Run("router works correctly - simple request body", func(t *testing.T) {
		context := context.Background()
		r := mux.NewRouter()
		swagger := openapi3.Swagger{
			Info: &openapi3.Info{
				Title:   swaggerOpenapiTitle,
				Version: swaggerOpenapiVersion,
			},
		}

		router, err := New(r, Options{
			Context: context,
			Openapi: &swagger,
		})
		require.NoError(t, err)
		require.NotNil(t, router)

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

		_, err = router.AddRoute(http.MethodPost, "/users", okHandler, Schema{
			RequestBody: &User{},
			Responses: map[int]Response{
				201: {
					Value: "",
				},
				401: {
					Value:       &errorResponse{},
					Description: "invalid request",
				},
			},
		})
		require.NoError(t, err)

		_, err = router.AddRoute(http.MethodGet, "/users", okHandler, Schema{
			Responses: map[int]Response{
				200: {
					Value: &Users{},
				},
			},
		})
		require.NoError(t, err)

		err = router.GenerateAndExposeSwagger()
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/users", nil)

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Result().StatusCode)

		body := readBody(t, w.Result().Body)
		require.Equal(t, "OK", body)

		t.Run("and generate swagger", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, JSONDocumentationPath, nil)

			router.ServeHTTP(w, r)

			require.Equal(t, http.StatusOK, w.Result().StatusCode)

			body := readBody(t, w.Result().Body)
			actual, err := ioutil.ReadFile("testdata/users.json")
			require.NoError(t, err)
			// require.JSONEq(t, string(actual), body)
			require.Equal(t, string(actual), body)
		})
	})
}
