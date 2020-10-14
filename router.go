package swagger

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/alecthomas/jsonschema"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
)

var (
	// ErrResponses is thrown if error occurs generating responses schemas.
	ErrResponses = errors.New("errors generating responses")
	// ErrRequestBody is thrown if error occurs generating responses schemas.
	ErrRequestBody = errors.New("errors generating request body")
)

// Operation type
type Operation struct {
	*openapi3.Operation
	// TODO: handle request and response
}

// Handler is the http type handler
type Handler func(w http.ResponseWriter, req *http.Request)

// GenerateAndExposeSwagger creates a /documentation/json route on router and
// expose the generated swagger
func (r Router) GenerateAndExposeSwagger() error {
	if err := r.swaggerSchema.Validate(r.context); err != nil {
		return fmt.Errorf("%w: %s", ErrValidatingSwagger, err)
	}

	jsonSwagger, err := r.swaggerSchema.MarshalJSON()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGenerateSwagger, err)
	}
	r.router.HandleFunc(JSONDocumentationPath, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonSwagger)
	})
	// TODO: add yaml endpoint

	return nil
}

// AddRawRoute add route to router with specific method, path and handler. Add the
// router also to the swagger schema, after validating it
func (r Router) AddRawRoute(method string, path string, handler Handler, operation Operation) (*mux.Route, error) {
	if operation.Operation != nil {
		err := operation.Validate(r.context)
		if err != nil {
			return nil, err
		}
	} else {
		operation.Operation = openapi3.NewOperation()
		operation.Responses = openapi3.NewResponses()
	}
	r.swaggerSchema.AddOperation(path, method, operation.Operation)

	return r.router.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
		// Handle, when content-type is json, the request/response marshalling? Maybe with a specific option.
		handler(w, req)
	}).Methods(method), nil
}

// Response is the struct containing a single route response.
type Response struct {
	Value       interface{}
	Description string
}

// Schema of the route.
type Schema struct {
	// Parameters  interface{}
	// Querystring interface{}
	RequestBody interface{}
	Responses   map[int]Response
}

// AddRoute add a route with json schema inferted by passed schema.
func (r Router) AddRoute(method string, path string, handler Handler, schema Schema) (*mux.Route, error) {
	operation := openapi3.NewOperation()
	operation.Responses = make(openapi3.Responses)

	if schema.RequestBody != nil {
		requestBody := openapi3.NewRequestBody()

		requestBodySchema, _, err := r.getSchemaFromInterface(schema.RequestBody, nil)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrRequestBody, err)
		}
		requestBody = requestBody.WithJSONSchema(requestBodySchema)

		operation.RequestBody = &openapi3.RequestBodyRef{
			Value: requestBody,
		}
	}
	if schema.Responses != nil {
		for statusCode, v := range schema.Responses {
			response := openapi3.NewResponse()

			responseSchema, _, err := r.getSchemaFromInterface(v.Value, nil)
			if err != nil {
				return nil, fmt.Errorf("%w: %s", ErrResponses, err)
			}

			response = response.WithDescription(v.Description)
			response = response.WithJSONSchema(responseSchema)

			operation.AddResponse(statusCode, response)
		}
	}

	return r.AddRawRoute(method, path, handler, Operation{operation})
}

func (r Router) getSchemaFromInterface(v interface{}, components *openapi3.Components) (*openapi3.Schema, *openapi3.Components, error) {
	if v == nil {
		return &openapi3.Schema{}, components, nil
	}

	reflector := &jsonschema.Reflector{
		DoNotReference: true,
	}

	jsonSchema := reflector.Reflect(v)
	jsonschema.Version = ""
	// Empty definitions. Definitions are not valid in openapi3, which use components.
	// In the future, we could add an option to fill the components in openapi spec.
	jsonSchema.Definitions = nil

	// jsonSchema = cleanJSONSchemaVersion(jsonSchema)
	data, err := jsonSchema.MarshalJSON()
	if err != nil {
		return nil, nil, err
	}

	schema := openapi3.NewSchema()
	err = schema.UnmarshalJSON(data)
	if err != nil {
		return nil, nil, err
	}

	return schema, components, nil
}
