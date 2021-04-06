package openapi3

import (
	"github.com/getkin/kin-openapi/openapi3"
	"goyave.dev/goyave/v3/validation"
)

// Refs cache structure associating validation rules pointers to OpenAPI refs
// to avoid generating them multiple times and allow the use of OpenAPI components.
type Refs struct {
	Schemas       map[*validation.Rules]*openapi3.SchemaRef
	FieldSchemas  map[*validation.Field]*openapi3.SchemaRef
	Parameters    map[*validation.Rules][]*openapi3.ParameterRef
	RequestBodies map[*validation.Rules]*openapi3.RequestBodyRef
}

// NewRefs create a new Refs struct with initialized maps.
func NewRefs() *Refs {
	return &Refs{
		Schemas:       make(map[*validation.Rules]*openapi3.SchemaRef),
		FieldSchemas:  make(map[*validation.Field]*openapi3.SchemaRef),
		Parameters:    make(map[*validation.Rules][]*openapi3.ParameterRef),
		RequestBodies: make(map[*validation.Rules]*openapi3.RequestBodyRef),
	}
}
