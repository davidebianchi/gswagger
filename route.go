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
}

// Handler is the http type handler
type Handler func(w http.ResponseWriter, req *http.Request)

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

// Content is the type of a content.
// The key of the map define the content-type.
type Content map[string]Schema

// Schema contains the value and if properties allow additional properties.
type Schema struct {
	Value                     interface{}
	AllowAdditionalProperties bool
}

// ParameterValue is the struct containing the schema or the content information.
// If content is specified, it takes precedence.
type ParameterValue map[string]struct {
	Content     Content
	Schema      *Schema
	Description string
}

// ContentValue is the struct containing the content information.
type ContentValue struct {
	Content     Content
	Description string
}

// Definitions of the route.
type Definitions struct {
	PathParams  ParameterValue
	Querystring ParameterValue
	Headers     ParameterValue
	Cookies     ParameterValue
	RequestBody *ContentValue
	Responses   map[int]ContentValue
}

const (
	pathParamsType  = "path"
	queryParamType  = "query"
	headerParamType = "header"
	cookieParamType = "cookie"
)

// AddRoute add a route with json schema inferted by passed schema.
func (r Router) AddRoute(method string, path string, handler Handler, schema Definitions) (*mux.Route, error) {
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

	err = r.resolveParameterSchema(headerParamType, schema.Headers, operation)
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

func (r Router) resolveRequestBodySchema(bodySchema *ContentValue, operation *openapi3.Operation) error {
	if bodySchema == nil {
		return nil
	}
	requestBody := openapi3.NewRequestBody()

	content, err := r.addContentToOASSchema(bodySchema.Content)
	if err != nil {
		return err
	}
	requestBody = requestBody.WithContent(content)

	if bodySchema.Description != "" {
		requestBody.WithDescription(bodySchema.Description)
	}

	operation.RequestBody = &openapi3.RequestBodyRef{
		Value: requestBody,
	}
	return nil
}

func (r Router) resolveResponsesSchema(responses map[int]ContentValue, operation *openapi3.Operation) error {
	if responses == nil {
		operation.Responses = openapi3.NewResponses()
	}
	for statusCode, v := range responses {
		response := openapi3.NewResponse()
		content, err := r.addContentToOASSchema(v.Content)
		if err != nil {
			return err
		}
		response = response.WithContent(content)

		response = response.WithDescription(v.Description)

		operation.AddResponse(statusCode, response)
	}

	return nil
}

func (r Router) resolveParameterSchema(paramType string, paramConfig ParameterValue, operation *openapi3.Operation) error {
	var keys = make([]string, 0, len(paramConfig))
	for k := range paramConfig {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		v := paramConfig[key]
		var param *openapi3.Parameter
		switch paramType {
		case pathParamsType:
			param = openapi3.NewPathParameter(key)
		case queryParamType:
			param = openapi3.NewQueryParameter(key)
		case headerParamType:
			param = openapi3.NewHeaderParameter(key)
		case cookieParamType:
			param = openapi3.NewCookieParameter(key)
		default:
			return fmt.Errorf("invalid param type")
		}

		if v.Description != "" {
			param = param.WithDescription(v.Description)
		}

		if v.Content != nil {
			content, err := r.addContentToOASSchema(v.Content)
			if err != nil {
				return err
			}
			param.Content = content
		} else {
			schema := openapi3.NewSchema()
			if v.Schema != nil {
				var err error
				schema, err = r.getSchemaFromInterface(v.Schema.Value, v.Schema.AllowAdditionalProperties)
				if err != nil {
					return err
				}
			}
			param.WithSchema(schema)
		}

		operation.AddParameter(param)
	}

	return nil
}

func (r Router) addContentToOASSchema(content Content) (openapi3.Content, error) {
	oasContent := openapi3.NewContent()
	for k, v := range content {
		var err error
		schema, err := r.getSchemaFromInterface(v.Value, v.AllowAdditionalProperties)
		if err != nil {
			return nil, err
		}
		oasContent[k] = openapi3.NewMediaType().WithSchema(schema)
	}
	return oasContent, nil
}
