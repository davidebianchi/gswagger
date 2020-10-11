package swagger

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gorilla/mux"
)

var (
	// ErrGenerateSwagger throws when fails the marshalling of the swagger struct.
	ErrGenerateSwagger = errors.New("fail to generate swagger")
	// ErrValidatingSwagger throws when given swagger params are not correct.
	ErrValidatingSwagger = errors.New("fails to validate swagger")
)

const (
	// JSONDocumentationPath is the path of the swagger documentation in json format.
	JSONDocumentationPath = "/documentation/json"

	defaultOpenapiVersion = "3.0.0"
)

// Router handle the gorilla mux router and the swagger schema
type Router struct {
	router                  *mux.Router
	SwaggerSchema           *openapi3.Swagger
	enableRequestValidation bool
	context                 context.Context
	swaggerRouter           *openapi3filter.Router
}

// Handler is the http type handler
type Handler func(w http.ResponseWriter, req *http.Request)

// GenerateAndExposeSwagger creates a /documentation/json route on router and
// expose the generated swagger
func (r Router) GenerateAndExposeSwagger() error {
	if err := r.SwaggerSchema.Validate(r.context); err != nil {
		return fmt.Errorf("%w: %s", ErrValidatingSwagger, err)
	}

	jsonSwagger, err := r.SwaggerSchema.MarshalJSON()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrGenerateSwagger, err)
	}

	r.router.HandleFunc(JSONDocumentationPath, func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("content-type", "application/json")
		w.Write(jsonSwagger)
	})
	// TODO: add yaml endpoint

	err = r.swaggerRouter.AddSwagger(r.SwaggerSchema)
	if err != nil {
		return err
	}

	return nil
}

// AddRoute add route to router with specific method, path and handler. Add the
// router also to the swagger schema, after validating it
func (r Router) AddRoute(method string, path string, handler Handler, operation Operation) (*mux.Route, error) {
	if operation.Operation != nil {
		err := operation.Validate(r.context)
		if err != nil {
			return nil, err
		}
	} else {
		operation.Operation = openapi3.NewOperation()
		operation.Responses = openapi3.NewResponses()
	}
	r.SwaggerSchema.AddOperation(path, method, operation.Operation)

	if operation.Operation != nil && r.enableRequestValidation {
		return r.router.HandleFunc(path, func(h http.ResponseWriter, req *http.Request) {
			err := validateRequest(r, req)
			if err != nil {
				// TODO: add response for validation response
				return
			}
			handler(h, req)

		}).Methods(method), nil
	}
	return r.router.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
		// Handle, when content-type is json, the request/response marshalling? Maybe with a specific option.
		handler(w, req)
	}).Methods(method), nil
}

func (r Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

// RouterOptions to be passed to create the new router and swagger
type RouterOptions struct {
	Context                 context.Context
	EnableRequestValidation bool
	Openapi                 *openapi3.Swagger
}

// New generate new router with swagger. Default to OpenAPI 3.0.0
func New(router *mux.Router, options RouterOptions) (*Router, error) {
	swagger, err := generateNewValidSwagger(options.Openapi)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrValidatingSwagger, err)
	}

	var ctx = options.Context
	if options.Context == nil {
		ctx = context.Background()
	}

	return &Router{
		router:                  router,
		enableRequestValidation: options.EnableRequestValidation,
		SwaggerSchema:           swagger,
		context:                 ctx,
		swaggerRouter:           openapi3filter.NewRouter(),
	}, nil
}

// Operation type
type Operation struct {
	*openapi3.Operation
	// TODO: handle request and response
}

func validateRequest(r Router, req *http.Request) error {
	// Find route
	route, pathParams, err := r.swaggerRouter.FindRoute(req.Method, req.URL)
	if err != nil {
		return err
	}

	// Validate request
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
		// TODO: add query params
	}

	return openapi3filter.ValidateRequest(req.Context(), requestValidationInput)
}

func generateNewValidSwagger(swagger *openapi3.Swagger) (*openapi3.Swagger, error) {
	if swagger == nil {
		swagger = &openapi3.Swagger{
			OpenAPI: defaultOpenapiVersion,
		}
	}
	if swagger.OpenAPI == "" {
		swagger.OpenAPI = defaultOpenapiVersion
	}

	if swagger.Paths == nil {
		swagger.Paths = openapi3.Paths{}
	}
	if swagger.Info == nil {
		return nil, fmt.Errorf("swagger info must not be empty")
	}
	if swagger.Info.Title == "" {
		return nil, fmt.Errorf("swagger info title is required")
	}
	if swagger.Info.Version == "" {
		return nil, fmt.Errorf("swagger info version is required")
	}

	return swagger, nil
}
