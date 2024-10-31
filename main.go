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
	// ErrGenerateOAS throws when fails the marshalling of the swagger struct.
	ErrGenerateOAS = errors.New("fail to generate openapi")
	// ErrValidatingOAS throws when given openapi params are not correct.
	ErrValidatingOAS = errors.New("fails to validate openapi")
)

const (
	// DefaultJSONDocumentationPath is the path of the openapi documentation in json format.
	DefaultJSONDocumentationPath = "/documentation/json"
	// DefaultYAMLDocumentationPath is the path of the openapi documentation in yaml format.
	DefaultYAMLDocumentationPath = "/documentation/yaml"
	defaultOpenapiVersion        = "3.0.0"
)

// Router handle the api router and the openapi schema.
// api router supported out of the box are:
// - gorilla mux
type Router[HandlerFunc, Route any] struct {
	router                apirouter.Router[HandlerFunc, Route]
	swaggerSchema         *openapi3.T
	context               context.Context
	jsonDocumentationPath string
	yamlDocumentationPath string
	pathPrefix            string
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

// NewRouter generate new router with openapi. Default to OpenAPI 3.0.0
func NewRouter[HandlerFunc, Route any](router apirouter.Router[HandlerFunc, Route], options Options) (*Router[HandlerFunc, Route], error) {
	openapi, err := generateNewValidOpenapi(options.Openapi)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrValidatingOAS, err)
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

	return &Router[HandlerFunc, Route]{
		router:                router,
		swaggerSchema:         openapi,
		context:               ctx,
		yamlDocumentationPath: yamlDocumentationPath,
		jsonDocumentationPath: jsonDocumentationPath,
		pathPrefix:            options.PathPrefix,
	}, nil
}

type SubRouterOptions struct {
	PathPrefix string
}

func (r Router[HandlerFunc, Route]) SubRouter(router apirouter.Router[HandlerFunc, Route], opts SubRouterOptions) (*Router[HandlerFunc, Route], error) {
	return &Router[HandlerFunc, Route]{
		router:                router,
		swaggerSchema:         r.swaggerSchema,
		context:               r.context,
		jsonDocumentationPath: r.jsonDocumentationPath,
		yamlDocumentationPath: r.yamlDocumentationPath,
		pathPrefix:            opts.PathPrefix,
	}, nil
}

func generateNewValidOpenapi(openapi *openapi3.T) (*openapi3.T, error) {
	if openapi == nil {
		return nil, fmt.Errorf("openapi is required")
	}
	if openapi.OpenAPI == "" {
		openapi.OpenAPI = defaultOpenapiVersion
	}
	if openapi.Paths == nil {
		openapi.Paths = &openapi3.Paths{}
	}

	if openapi.Info == nil {
		return nil, fmt.Errorf("openapi info is required")
	}
	if openapi.Info.Title == "" {
		return nil, fmt.Errorf("openapi info title is required")
	}
	if openapi.Info.Version == "" {
		return nil, fmt.Errorf("openapi info version is required")
	}

	return openapi, nil
}

// GenerateAndExposeOpenapi creates a /documentation/json route on router and
// expose the generated swagger
func (r Router[_, _]) GenerateAndExposeOpenapi() error {
	if err := r.swaggerSchema.Validate(r.context); err != nil {
		return fmt.Errorf("%w: %s", ErrValidatingOAS, err)
	}

	jsonSwagger, err := r.swaggerSchema.MarshalJSON()
	if err != nil {
		return fmt.Errorf("%w json marshal: %s", ErrGenerateOAS, err)
	}
	r.router.AddRoute(http.MethodGet, r.jsonDocumentationPath, r.router.SwaggerHandler("application/json", jsonSwagger))

	yamlSwagger, err := yaml.JSONToYAML(jsonSwagger)
	if err != nil {
		return fmt.Errorf("%w yaml marshal: %s", ErrGenerateOAS, err)
	}
	r.router.AddRoute(http.MethodGet, r.yamlDocumentationPath, r.router.SwaggerHandler("text/plain", yamlSwagger))

	return nil
}

func isValidDocumentationPath(path string) error {
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("invalid path %s. Path should start with '/'", path)
	}
	return nil
}
