package swagger

import (
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/alecthomas/jsonschema"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
)

var (
	// ErrResponses is thrown if error occurs generating responses schemas.
	ErrResponses = errors.New("errors generating responses schema")
	// ErrRequestBody is thrown if error occurs generating responses schemas.
	ErrRequestBody = errors.New("errors generating request body schema")
	// ErrPathParams is thrown if error occurs generating path params schemas.
	ErrPathParams = errors.New("errors generating path parameters schema")
	// ErrQuerystring is thrown if error occurs generating querystring params schemas.
	ErrQuerystring = errors.New("errors generating querystring schema")
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

// SchemaValue is the struct containing the schema information.
type SchemaValue struct {
	Content     interface{}
	Description string

	// ContentType is to be used only with RequestBody. Valid ContentType
	// are application/json or multipart/form-data.
	ContentType               string
	AllowAdditionalProperties bool
}

// Schema of the route.
type Schema struct {
	PathParams  map[string]SchemaValue
	Querystring map[string]SchemaValue
	Headers     map[string]SchemaValue
	Cookies     map[string]SchemaValue
	RequestBody *SchemaValue
	Responses   map[int]SchemaValue
}

const (
	pathParamsType   = "path"
	queryParamType   = "query"
	headersParamType = "headers"
	cookieParamType  = "cookie"
)

// AddRoute add a route with json schema inferted by passed schema.
func (r Router) AddRoute(method string, path string, handler Handler, schema Schema) (*mux.Route, error) {
	operation := openapi3.NewOperation()
	operation.Responses = make(openapi3.Responses)

	err := r.resolveRequestBodySchema(schema.RequestBody, operation)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrRequestBody, err)
	}

	err = r.resolveResponsesSchema(schema.Responses, operation)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrResponses, err)
	}

	err = r.resolveParameterSchema(pathParamsType, schema.PathParams, operation)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrPathParams, err)
	}

	err = r.resolveParameterSchema(queryParamType, schema.Querystring, operation)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrPathParams, err)
	}

	err = r.resolveParameterSchema(headersParamType, schema.Headers, operation)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrPathParams, err)
	}

	err = r.resolveParameterSchema(cookieParamType, schema.Cookies, operation)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrPathParams, err)
	}

	return r.AddRawRoute(method, path, handler, Operation{operation})
}

func (r Router) getSchemaFromInterface(v interface{}, allowAdditionalProperties bool) (*openapi3.Schema, error) {
	if v == nil {
		return &openapi3.Schema{}, nil
	}

	reflector := &jsonschema.Reflector{
		DoNotReference:            true,
		AllowAdditionalProperties: allowAdditionalProperties,
	}

	jsonSchema := reflector.Reflect(v)
	jsonschema.Version = ""
	// Empty definitions. Definitions are not valid in openapi3, which use components.
	// In the future, we could add an option to fill the components in openapi spec.
	jsonSchema.Definitions = nil

	data, err := jsonSchema.MarshalJSON()
	if err != nil {
		return nil, err
	}

	schema := openapi3.NewSchema()
	err = schema.UnmarshalJSON(data)
	if err != nil {
		return nil, err
	}

	return schema, nil
}

func (r Router) resolveRequestBodySchema(bodySchema *SchemaValue, operation *openapi3.Operation) error {
	if bodySchema == nil {
		return nil
	}
	requestBodySchema, err := r.getSchemaFromInterface(bodySchema.Content, bodySchema.AllowAdditionalProperties)
	if err != nil {
		return err
	}

	requestBody := openapi3.NewRequestBody()
	switch bodySchema.ContentType {
	case "multipart/form-data":
		requestBody = requestBody.WithFormDataSchema(requestBodySchema)
	case "application/json", "":
		requestBody = requestBody.WithJSONSchema(requestBodySchema)
	default:
		return fmt.Errorf("invalid content-type in request body")
	}

	if bodySchema.Description != "" {
		requestBody.WithDescription(bodySchema.Description)
	}

	operation.RequestBody = &openapi3.RequestBodyRef{
		Value: requestBody,
	}
	return nil
}

func (r Router) resolveResponsesSchema(responses map[int]SchemaValue, operation *openapi3.Operation) error {
	if responses == nil {
		operation.Responses = openapi3.NewResponses()
	}
	for statusCode, v := range responses {
		response := openapi3.NewResponse()

		responseSchema, err := r.getSchemaFromInterface(v.Content, v.AllowAdditionalProperties)
		if err != nil {
			return err
		}

		response = response.WithDescription(v.Description)
		response = response.WithJSONSchema(responseSchema)

		operation.AddResponse(statusCode, response)
	}

	return nil
}

func (r Router) resolveParameterSchema(paramType string, paramConfig map[string]SchemaValue, operation *openapi3.Operation) error {
	var keys = make([]string, 0, len(paramConfig))
	for k := range paramConfig {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		v := paramConfig[key]
		var param *openapi3.Parameter
		switch paramType {
		case "path":
			param = openapi3.NewPathParameter(key)
		case "query":
			param = openapi3.NewQueryParameter(key)
		case "headers":
			param = openapi3.NewHeaderParameter(key)
		case "cookie":
			param = openapi3.NewCookieParameter(key)
		}

		if v.Content != nil {
			schema, err := r.getSchemaFromInterface(v.Content, v.AllowAdditionalProperties)
			if err != nil {
				return err
			}
			param = param.WithSchema(schema)
		}
		if v.Description != "" {
			param = param.WithDescription(v.Description)
		}

		operation.AddParameter(param)
	}

	return nil
}
