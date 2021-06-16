package openapi3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRefs(t *testing.T) {
	refs := NewRefs()
	assert.NotNil(t, refs)
	assert.NotNil(t, refs.Schemas)
	assert.NotNil(t, refs.ParamSchemas)
	assert.NotNil(t, refs.Parameters)
	assert.NotNil(t, refs.QueryParameters)
	assert.NotNil(t, refs.RequestBodies)
	assert.NotNil(t, refs.AST)
	assert.NotNil(t, refs.HandlerDocs)
}
