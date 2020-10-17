package swagger

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestAddRoute(t *testing.T) {
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

	okHandler := func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
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
		{
			name: "schema defined without content",
			routes: func(t *testing.T, router *Router) {
				_, err := router.AddRoute(http.MethodPost, "/{id}", okHandler, Schema{
					RequestBody: &SchemaValue{
						Description: "request body without schema",
					},
					Responses: map[int]SchemaValue{
						204: {},
					},
					PathParams: map[string]SchemaValue{
						"id": {},
					},
					Querystring: map[string]SchemaValue{
						"q": {},
					},
					Headers: map[string]SchemaValue{
						"key": {},
					},
					Cookies: map[string]SchemaValue{
						"cookie1": {},
					},
				})
				require.NoError(t, err)
			},
			testPath:     "/foobar",
			testMethod:   http.MethodPost,
			fixturesPath: "testdata/schema-no-content.json",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context := context.Background()
			r := mux.NewRouter()

			router, err := New(r, Options{
				Context: context,
				Openapi: getBaseSwagger(t),
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
				require.JSONEq(t, string(actual), body, "actual json data: ", string(actual))
			})
		})
	}

	t.Run("", func(t *testing.T) {

	})
}

func TestResolveRequestBodySchema(t *testing.T) {
	type TestStruct struct {
		ID string `json:"id,omitempty"`
	}
	tests := []struct {
		name         string
		bodySchema   *SchemaValue
		expectedErr  error
		expectedJSON string
	}{
		{
			name:         "empty body schema",
			expectedErr:  nil,
			expectedJSON: `{"responses": null}`,
		},
		{
			name:        "schema multipart",
			expectedErr: nil,
			bodySchema: &SchemaValue{
				Content:     &TestStruct{},
				ContentType: "multipart/form-data",
			},
			expectedJSON: `{
				"requestBody": {
					"content": {
						"multipart/form-data": {
							"schema": {
								"type":"object",
								"additionalProperties":false,
								"properties": {
									"id": {"type":"string"}
								}
							}
						}
					}
				},
				"responses": null
			}`,
		},
		{
			name:        "content-type application/json",
			expectedErr: nil,
			bodySchema: &SchemaValue{
				Content:     &TestStruct{},
				ContentType: "application/json",
			},
			expectedJSON: `{
				"requestBody": {
					"content": {
						"application/json": {
							"schema": {
								"type":"object",
								"additionalProperties":false,
								"properties": {
									"id": {"type":"string"}
								}
							}
						}
					}
				},
				"responses": null
			}`,
		},
		{
			name:        "no content-type - default to json",
			expectedErr: nil,
			bodySchema: &SchemaValue{
				Content: &TestStruct{},
			},
			expectedJSON: `{
				"requestBody": {
					"content": {
						"application/json": {
							"schema": {
								"type":"object",
								"additionalProperties":false,
								"properties": {
									"id": {"type":"string"}
								}
							}
						}
					}
				},
				"responses": null
			}`,
		},
		{
			name:        "with description",
			expectedErr: nil,
			bodySchema: &SchemaValue{
				Content:     &TestStruct{},
				Description: "my custom description",
			},
			expectedJSON: `{
				"requestBody": {
					"description": "my custom description",
					"content": {
						"application/json": {
							"schema": {
								"type":"object",
								"additionalProperties":false,
								"properties": {
									"id": {"type":"string"}
								}
							}
						}
					}
				},
				"responses": null
			}`,
		},
		// FIXME: this test case exhibits a wrong behavior. It should be supported.
		{
			name:        "content type text/plain",
			expectedErr: fmt.Errorf("invalid content-type in request body"),
			bodySchema: &SchemaValue{
				Content:     &TestStruct{},
				ContentType: "text/plain",
			},
		},
		// FIXME: this test case exhibits a wrong behavior. It should be supported.
		{
			name:        "generic content type - it represent all types",
			expectedErr: fmt.Errorf("invalid content-type in request body"),
			bodySchema: &SchemaValue{
				Content:     &TestStruct{},
				ContentType: "*/*",
			},
		},
	}

	mux := mux.NewRouter()
	router, err := New(mux, Options{
		Openapi: getBaseSwagger(t),
	})
	require.NoError(t, err)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			operation := openapi3.NewOperation()

			err := router.resolveRequestBodySchema(test.bodySchema, operation)

			if err == nil {
				data, _ := operation.MarshalJSON()
				jsonData := string(data)
				require.JSONEq(t, test.expectedJSON, jsonData, "actual json data: ", jsonData)
				require.NoError(t, err)
			}
			require.Equal(t, test.expectedErr, err)
		})
	}
}

func TestResolveResponsesSchema(t *testing.T) {
	type TestStruct struct {
		Message string `json:"message,omitempty"`
	}
	tests := []struct {
		name            string
		responsesSchema map[int]SchemaValue
		expectedErr     error
		expectedJSON    string
	}{
		{
			name:         "empty responses schema",
			expectedErr:  nil,
			expectedJSON: `{"responses": {"default":{"description":""}}}`,
		},
		{
			name: "with 1 status code",
			responsesSchema: map[int]SchemaValue{
				200: {
					Content: &TestStruct{},
				},
			},
			expectedErr: nil,
			expectedJSON: `{
				"responses": {
					"200": {
						"description": "",
						"content": {
							"application/json": {
								"schema": {
									"type": "object",
									"properties": {
										"message": {
											"type": "string"
										}
									},
									"additionalProperties": false
								}
							}
						}
					}
				}
			}`,
		},
		{
			name: "with more status codes",
			responsesSchema: map[int]SchemaValue{
				200: {
					Content: &TestStruct{},
				},
				400: {
					Content: "",
				},
			},
			expectedErr: nil,
			expectedJSON: `{
				"responses": {
					"200": {
						"description": "",
						"content": {
							"application/json": {
								"schema": {
									"type": "object",
									"properties": {
										"message": {
											"type": "string"
										}
									},
									"additionalProperties": false
								}
							}
						}
					},
					"400": {
						"description": "",
						"content": {
							"application/json": {
								"schema": {
									"type": "string"
								}
							}
						}
					}
				}
			}`,
		},
		{
			name: "with custom description",
			responsesSchema: map[int]SchemaValue{
				400: {
					Content:     "",
					Description: "a description",
				},
			},
			expectedErr: nil,
			expectedJSON: `{
				"responses": {
					"400": {
						"description": "a description",
						"content": {
							"application/json": {
								"schema": {
									"type": "string"
								}
							}
						}
					}
				}
			}`,
		},
	}

	mux := mux.NewRouter()
	router, err := New(mux, Options{
		Openapi: getBaseSwagger(t),
	})
	require.NoError(t, err)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			operation := openapi3.NewOperation()
			operation.Responses = make(openapi3.Responses)

			err := router.resolveResponsesSchema(test.responsesSchema, operation)

			if err == nil {
				data, _ := operation.MarshalJSON()
				jsonData := string(data)
				require.JSONEq(t, test.expectedJSON, jsonData, "actual json data: ", jsonData)
				require.NoError(t, err)
			}
			require.Equal(t, test.expectedErr, err)
		})
	}
}

func TestResolveParametersSchema(t *testing.T) {
	type TestStruct struct {
		Message string `json:"message,omitempty"`
	}
	tests := []struct {
		name         string
		paramsSchema map[string]SchemaValue
		paramType    string
		expectedErr  error
		expectedJSON string
	}{
		{
			name:         "empty responses schema",
			paramType:    pathParamsType,
			expectedJSON: `{"responses": null}`,
		},
		{
			name:      "path param",
			paramType: pathParamsType,
			paramsSchema: map[string]SchemaValue{
				"foo": {
					Content: "",
				},
			},
			expectedJSON: `{
				"parameters": [{
					"in": "path",
					"name": "foo",
					"required": true,
					"schema": {
						"type": "string"
					}
				}],
				"responses": null
			}`,
		},
		{
			name:      "query param",
			paramType: queryParamType,
			paramsSchema: map[string]SchemaValue{
				"foo": {
					Content: "",
				},
			},
			expectedJSON: `{
				"parameters": [{
					"in": "query",
					"name": "foo",
					"schema": {
						"type": "string"
					}
				}],
				"responses": null
			}`,
		},
		{
			name:      "cookie param",
			paramType: cookieParamType,
			paramsSchema: map[string]SchemaValue{
				"foo": {
					Content: "",
				},
			},
			expectedJSON: `{
				"parameters": [{
					"in": "cookie",
					"name": "foo",
					"schema": {
						"type": "string"
					}
				}],
				"responses": null
			}`,
		},
		{
			name:      "header param",
			paramType: headerParamType,
			paramsSchema: map[string]SchemaValue{
				"foo": {
					Content: "",
				},
			},
			expectedJSON: `{
				"parameters": [{
					"in": "header",
					"name": "foo",
					"schema": {
						"type": "string"
					}
				}],
				"responses": null
			}`,
		},
		{
			name:      "wrong param type",
			paramType: "wrong",
			paramsSchema: map[string]SchemaValue{
				"foo": {
					Content: "",
				},
			},
			expectedErr: fmt.Errorf("invalid param type"),
		},
	}

	mux := mux.NewRouter()
	router, err := New(mux, Options{
		Openapi: getBaseSwagger(t),
	})
	require.NoError(t, err)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			operation := openapi3.NewOperation()

			err := router.resolveParameterSchema(test.paramType, test.paramsSchema, operation)

			if err == nil {
				data, _ := operation.MarshalJSON()
				jsonData := string(data)
				require.JSONEq(t, test.expectedJSON, jsonData, "actual json data: ", jsonData)
				require.NoError(t, err)
			}
			require.Equal(t, test.expectedErr, err)
		})
	}
}

func getBaseSwagger(t *testing.T) *openapi3.Swagger {
	t.Helper()

	return &openapi3.Swagger{
		Info: &openapi3.Info{
			Title:   swaggerOpenapiTitle,
			Version: swaggerOpenapiVersion,
		},
	}
}
