package openapi3

import (
	"github.com/getkin/kin-openapi/openapi3"
	"goyave.dev/goyave/v3/validation"
)

type Refs struct {
	Schemas       map[*validation.Rules]*openapi3.SchemaRef // TODO associate a name with schemas
	FieldSchemas  map[*validation.Field]*openapi3.SchemaRef
	Parameters    map[*validation.Rules][]*openapi3.ParameterRef
	RequestBodies map[*validation.Rules]*openapi3.RequestBodyRef
}

func NewRefs() *Refs {
	return &Refs{
		Schemas:       make(map[*validation.Rules]*openapi3.SchemaRef),
		FieldSchemas:  make(map[*validation.Field]*openapi3.SchemaRef),
		Parameters:    make(map[*validation.Rules][]*openapi3.ParameterRef),
		RequestBodies: make(map[*validation.Rules]*openapi3.RequestBodyRef),
	}
}
