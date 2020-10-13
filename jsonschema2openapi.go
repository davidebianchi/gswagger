package swagger

import (
	"github.com/alecthomas/jsonschema"
	"github.com/getkin/kin-openapi/openapi3"
)

func (r Router) getSchemaFromInterface(v interface{}, components *openapi3.Components) (*openapi3.Schema, *openapi3.Components, error) {
	if v == nil {
		return &openapi3.Schema{}, components, nil
	}

	reflector := &jsonschema.Reflector{
		DoNotReference: true,
	}

	jsonschema.Version = ""
	jsonSchema := reflector.Reflect(v)
	// TODO: integrate to have components filled by option
	err := definitionsToComponents(components, jsonSchema)
	if err != nil {
		return nil, nil, err
	}

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

func definitionsToComponents(components *openapi3.Components, schema *jsonschema.Schema) error {
	if components == nil {
		schema.Definitions = nil
		return nil
	}

	if components.Schemas == nil {
		components.Schemas = map[string]*openapi3.SchemaRef{}
	}
	// Rename definitions to components
	for k, definition := range schema.Definitions {

		marshalledDefinition, err := definition.MarshalJSON()
		if err != nil {
			return err
		}

		components.Schemas[k] = &openapi3.SchemaRef{}
		err = components.Schemas[k].UnmarshalJSON(marshalledDefinition)
		if err != nil {
			return err
		}
	}
	return nil
}
