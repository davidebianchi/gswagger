package swagger

import (
	"encoding/json"
	"testing"

	"github.com/alecthomas/jsonschema"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestConvertJSONSchema2OpenAPI(t *testing.T) {
	type User struct {
		Name        string   `json:"name" jsonschema:"title=The user name,required" jsonschema_extras:"example=Jane"`
		PhoneNumber int      `json:"phone" jsonschema:"title=mobile number of user"`
		Groups      []string `json:"groups,omitempty" jsonschema:"title=groups of the user,default=users"`
		Address     string   `json:"address" jsonschema:"title=user address"`
	}
	type Users []User

	type NestedDefinitions struct {
		Name  string `json:"name" jsonschema:"title=The user name,required" jsonschema_extras:"example=Jane"`
		Users Users  `json:"users,omitempty" jsonschema:"title=List of game users"`
	}
	reflector := &jsonschema.Reflector{
		DoNotReference: true,
	}

	t.Run("add definitions to components", func(t *testing.T) {
		jsonSchema := reflector.Reflect(&[]User{})
		jsonSchema.Version = ""

		components := &openapi3.Components{}

		definitionsToComponents(components, jsonSchema)

		output, err := json.Marshal(&components.Schemas)
		require.NoError(t, err)

		expected, err := json.Marshal(jsonSchema.Definitions)
		require.NoError(t, err)

		require.JSONEq(t, string(output), string(expected))
	})

	t.Run("add nested definitions to components", func(t *testing.T) {
		jsonSchema := reflector.Reflect(&NestedDefinitions{})
		jsonSchema.Version = ""

		components := &openapi3.Components{}

		definitionsToComponents(components, jsonSchema)

		output, err := json.Marshal(&components.Schemas)
		require.NoError(t, err)

		expected, err := json.Marshal(jsonSchema.Definitions)
		require.NoError(t, err)

		require.JSONEq(t, string(output), string(expected))
	})
}
