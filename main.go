package swagger

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/davidebianchi/gswagger/apirouter"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
)

var (
	// ErrGenerateSwagger throws when fails the marshalling of the swagger struct.
	ErrGenerateSwagger = errors.New("fail to generate swagger")
	// ErrValidatingSwagger throws when given swagger params are not correct.
	ErrValidatingSwagger = errors.New("fails to validate swagger")
)

const (
	// DefaultJSONDocumentationPath is the path of the swagger documentation in json format.
	DefaultJSONDocumentationPath = "/documentation/json"
	// DefaultYAMLDocumentationPath is the path of the swagger documentation in yaml format.
	DefaultYAMLDocumentationPath = "/documentation/yaml"
	defaultOpenapiVersion        = "3.0.0"
)

// Router handle the api router and the swagger schema.
// api router supported out of the box are:
// - gorilla mux
type Router struct {
	router                apirouter.Router
	swaggerSchema         *openapi3.T
	context               context.Context
	jsonDocumentationPath string
	yamlDocumentationPath string
	pathPrefix            string
}

func (r Router) GetSwaggerSchema() *openapi3.T {
	return r.swaggerSchema
}

// Options to be passed to create the new router and swagger
type Options struct {
	Context context.Context
	Openapi *openapi3.T
	// JSONDocumentationPath is the path exposed by json endpoint. Default to /documentation/json.
	JSONDocumentationPath string
	// YAMLDocumentationPath is the path exposed by yaml endpoint. Default to /documentation/yaml.
	YAMLDocumentationPath string
	// Add path prefix to add to every router path.
	PathPrefix string
}

// NewRouter generate new router with swagger. Default to OpenAPI 3.0.0
func NewRouter(router apirouter.Router, options Options) (*Router, error) {
	swagger, err := generateNewValidSwagger(options.Openapi)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrValidatingSwagger, err)
	}

	var ctx = options.Context
	if options.Context == nil {
		ctx = context.Background()
	}

	yamlDocumentationPath := DefaultYAMLDocumentationPath
	if options.YAMLDocumentationPath != "" {
		if err := isValidDocumentationPath(options.YAMLDocumentationPath); err != nil {
			return nil, err
		}
		yamlDocumentationPath = options.YAMLDocumentationPath
	}

	jsonDocumentationPath := DefaultJSONDocumentationPath
	if options.JSONDocumentationPath != "" {
		if err := isValidDocumentationPath(options.JSONDocumentationPath); err != nil {
			return nil, err
		}
		jsonDocumentationPath = options.JSONDocumentationPath
	}

	return &Router{
		router:                router,
		swaggerSchema:         swagger,
		context:               ctx,
		yamlDocumentationPath: yamlDocumentationPath,
		jsonDocumentationPath: jsonDocumentationPath,
		pathPrefix:            options.PathPrefix,
	}, nil
}

type SubRouterOptions struct {
	PathPrefix string
}

func (r Router) SubRouter(router apirouter.Router, opts SubRouterOptions) (*Router, error) {
	return &Router{
		router:                router,
		swaggerSchema:         r.swaggerSchema,
		context:               r.context,
		jsonDocumentationPath: r.jsonDocumentationPath,
		yamlDocumentationPath: r.yamlDocumentationPath,
		pathPrefix:            opts.PathPrefix,
	}, nil
}

func generateNewValidSwagger(swagger *openapi3.T) (*openapi3.T, error) {
	if swagger == nil {
		return nil, fmt.Errorf("swagger is required")
	}
	if swagger.OpenAPI == "" {
		swagger.OpenAPI = defaultOpenapiVersion
	}
	if swagger.Paths == nil {
		swagger.Paths = openapi3.Paths{}
	}

	if swagger.Info == nil {
		return nil, fmt.Errorf("swagger info is required")
	}
	if swagger.Info.Title == "" {
		return nil, fmt.Errorf("swagger info title is required")
	}
	if swagger.Info.Version == "" {
		return nil, fmt.Errorf("swagger info version is required")
	}

	return swagger, nil
}

// GenerateAndExposeSwagger creates a /documentation/json route on router and
// expose the generated swagger
func (r Router) GenerateAndExposeSwagger() error {
	if err := r.swaggerSchema.Validate(r.context); err != nil {
		return fmt.Errorf("%w: %s", ErrValidatingSwagger, err)
	}

	jsonSwagger, err := r.swaggerSchema.MarshalJSON()
	if err != nil {
		return fmt.Errorf("%w json marshal: %s", ErrGenerateSwagger, err)
	}
	r.router.AddRoute(r.jsonDocumentationPath, http.MethodGet, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonSwagger)
	})

	yamlSwagger, err := yaml.JSONToYAML(jsonSwagger)
	if err != nil {
		return fmt.Errorf("%w yaml marshal: %s", ErrGenerateSwagger, err)
	}
	r.router.AddRoute(r.yamlDocumentationPath, http.MethodGet, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write(yamlSwagger)
	})

	return nil
}

func isValidDocumentationPath(path string) error {
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("invalid path %s. Path should start with '/'", path)
	}
	return nil
}
