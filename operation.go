package swagger

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// Operation type
type Operation struct {
	*openapi3.Operation
}

// NewOperation returns an OpenApi operation.
func NewOperation() Operation {
	return Operation{
		openapi3.NewOperation(),
	}
}

// AddRequestBody set request body into operation.
func (o *Operation) AddRequestBody(requestBody *openapi3.RequestBody) {
	o.RequestBody = &openapi3.RequestBodyRef{
		Value: requestBody,
	}
}

// AddResponse add response to operation. It check if the description is present
// (otherwise default to empty string). This method does not add the default response,
// but it is always possible to add it manually.
func (o *Operation) AddResponse(status int, response *openapi3.Response) {
	if o.Responses == nil {
		o.Responses = make(openapi3.Responses)
	}
	if response.Description == nil {
		// a description is required by kin openapi, so we set an empty description
		// if it is not given.
		response.WithDescription("")
	}
	o.Operation.AddResponse(status, response)
}

func (o *Operation) addSecurityRequirements(securityRequirements SecurityRequirements) {
	if securityRequirements != nil && o.Security == nil {
		o.Security = openapi3.NewSecurityRequirements()
	}
	for _, securityRequirement := range securityRequirements {
		o.Security.With(openapi3.SecurityRequirement(securityRequirement))
	}
}
