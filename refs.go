package openapi3

import (
	"go/ast"

	"github.com/getkin/kin-openapi/openapi3"
	"goyave.dev/goyave/v3/validation"
)

// Refs cache structure associating validation rules pointers to OpenAPI refs
// to avoid generating them multiple times and allow the use of OpenAPI components.
type Refs struct {
	Schemas         map[*validation.Rules]*openapi3.SchemaRef
	ParamSchemas    map[string]*openapi3.SchemaRef
	Parameters      map[string]*openapi3.ParameterRef
	QueryParameters map[*validation.Rules][]*openapi3.ParameterRef
	RequestBodies   map[*validation.Rules]*openapi3.RequestBodyRef
	AST             map[string]*ast.File
	HandlerDocs     map[uintptr]*HandlerDoc
}

// NewRefs create a new Refs struct with initialized maps.
func NewRefs() *Refs {
	return &Refs{
		Schemas:         make(map[*validation.Rules]*openapi3.SchemaRef),
		ParamSchemas:    make(map[string]*openapi3.SchemaRef),
		Parameters:      make(map[string]*openapi3.ParameterRef),
		QueryParameters: make(map[*validation.Rules][]*openapi3.ParameterRef),
		RequestBodies:   make(map[*validation.Rules]*openapi3.RequestBodyRef),
		AST:             make(map[string]*ast.File),
		HandlerDocs:     make(map[uintptr]*HandlerDoc),
	}
}

// HandlerDoc info extracted from AST about a Handler.
type HandlerDoc struct {
	FuncName    string
	Description string
}
