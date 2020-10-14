package swagger

import (
	"context"
	"errors"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
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
	router        *mux.Router
	SwaggerSchema *openapi3.Swagger
	context       context.Context
}

// Options to be passed to create the new router and swagger
type Options struct {
	Context context.Context
	Openapi *openapi3.Swagger
}

// New generate new router with swagger. Default to OpenAPI 3.0.0
func New(router *mux.Router, options Options) (*Router, error) {
	swagger, err := generateNewValidSwagger(options.Openapi)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrValidatingSwagger, err)
	}

	var ctx = options.Context
	if options.Context == nil {
		ctx = context.Background()
	}

	return &Router{
		router:        router,
		SwaggerSchema: swagger,
		context:       ctx,
	}, nil
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
