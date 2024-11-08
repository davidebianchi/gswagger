package swagger

import (
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestNewOperation(t *testing.T) {
	schema := openapi3.NewObjectSchema().WithProperties(map[string]*openapi3.Schema{
		"foo": openapi3.NewStringSchema(),
		"bar": openapi3.NewIntegerSchema().WithMax(15).WithMin(5),
	})

	tests := []struct {
		name         string
		getOperation func(t *testing.T, operation Operation) Operation
		expectedJSON string
	}{
		{
			name: "add request body",
			getOperation: func(t *testing.T, operation Operation) Operation {
				requestBody := openapi3.NewRequestBody().WithJSONSchema(schema)
				operation.AddRequestBody(requestBody)
				operation.Responses = openapi3.NewResponses()
				return operation
			},
			expectedJSON: `{"info":{"title":"test openapi title","version":"test openapi version"},"openapi":"3.0.0","paths":{"/":{"post":{"requestBody":{"content":{"application/json":{"schema":{"properties":{"bar":{"maximum":15,"minimum":5,"type":"integer"},"foo":{"type":"string"}},"type":"object"}}}},"responses":{"default":{"description":""}}}}}}`,
		},
		{
			name: "add response",
			getOperation: func(t *testing.T, operation Operation) Operation {
				response := openapi3.NewResponse().WithJSONSchema(schema)
				operation.AddResponse(200, response)
				return operation
			},
			expectedJSON: `{"info":{"title":"test openapi title","version":"test openapi version"},"openapi":"3.0.0","paths":{"/":{"post":{"responses":{"200":{"content":{"application/json":{"schema":{"properties":{"bar":{"maximum":15,"minimum":5,"type":"integer"},"foo":{"type":"string"}},"type":"object"}}},"description":""}}}}}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			openapi := getBaseSwagger(t)
			openapi.OpenAPI = "3.0.0"
			operation := test.getOperation(t, NewOperation())

			openapi.AddOperation("/", http.MethodPost, operation.Operation)

			data, _ := openapi.MarshalJSON()
			jsonData := string(data)
			require.JSONEq(t, test.expectedJSON, jsonData, "actual json data: %s", jsonData)
		})
	}
}
