package swagger

import (
	"context"
	"fmt"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestNewRouter(t *testing.T) {
	mRouter := mux.NewRouter()

	info := &openapi3.Info{
		Title:   "my title",
		Version: "my version",
	}
	openapi := &openapi3.Swagger{
		Info:  info,
		Paths: openapi3.Paths{},
	}

	t.Run("not ok - invalid Openapi option", func(t *testing.T) {
		r, err := New(mRouter, Options{})

		require.Nil(t, r)
		require.EqualError(t, err, fmt.Sprintf("%s: swagger is required", ErrValidatingSwagger))
	})

	t.Run("ok - with default context", func(t *testing.T) {
		r, err := New(mRouter, Options{
			Openapi: openapi,
		})

		require.NoError(t, err)
		require.Equal(t, &Router{
			context:       context.Background(),
			router:        mRouter,
			swaggerSchema: openapi,
		}, r)
	})

	t.Run("ok - with custom context", func(t *testing.T) {
		type key struct{}
		ctx := context.WithValue(context.Background(), key{}, "value")
		r, err := New(mRouter, Options{
			Openapi: openapi,
			Context: ctx,
		})

		require.NoError(t, err)
		require.Equal(t, &Router{
			context:       ctx,
			router:        mRouter,
			swaggerSchema: openapi,
		}, r)
	})
}

func TestGenerateValidSwagger(t *testing.T) {
	t.Run("not ok - empty swagger info", func(t *testing.T) {
		swagger := &openapi3.Swagger{}

		swagger, err := generateNewValidSwagger(swagger)
		require.Nil(t, swagger)
		require.EqualError(t, err, "swagger info is required")
	})

	t.Run("not ok - empty info title", func(t *testing.T) {
		swagger := &openapi3.Swagger{
			Info: &openapi3.Info{},
		}

		swagger, err := generateNewValidSwagger(swagger)
		require.Nil(t, swagger)
		require.EqualError(t, err, "swagger info title is required")
	})

	t.Run("not ok - empty info version", func(t *testing.T) {
		swagger := &openapi3.Swagger{
			Info: &openapi3.Info{
				Title: "title",
			},
		}

		swagger, err := generateNewValidSwagger(swagger)
		require.Nil(t, swagger)
		require.EqualError(t, err, "swagger info version is required")
	})

	t.Run("ok - custom swagger", func(t *testing.T) {
		swagger := &openapi3.Swagger{
			Info: &openapi3.Info{},
		}

		swagger, err := generateNewValidSwagger(swagger)
		require.Nil(t, swagger)
		require.EqualError(t, err, "swagger info title is required")
	})

	t.Run("not ok - swagger is required", func(t *testing.T) {
		swagger, err := generateNewValidSwagger(nil)
		require.Nil(t, swagger)
		require.EqualError(t, err, "swagger is required")
	})

	t.Run("ok", func(t *testing.T) {
		info := &openapi3.Info{
			Title:   "my title",
			Version: "my version",
		}
		swagger := &openapi3.Swagger{
			Info: info,
		}

		swagger, err := generateNewValidSwagger(swagger)
		require.NoError(t, err)
		require.Equal(t, &openapi3.Swagger{
			OpenAPI: defaultOpenapiVersion,
			Info:    info,
			Paths:   openapi3.Paths{},
		}, swagger)
	})
}
