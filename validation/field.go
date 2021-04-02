package validation

import (
	"github.com/getkin/kin-openapi/openapi3"
	"goyave.dev/goyave/v3/validation"
)

func SchemaFromField(field *validation.Field) *openapi3.Schema {
	// TODO save schema ref to refs
	s := openapi3.NewSchema()
	if rule := findFirstTypeRule(field); rule != nil {
		switch rule.Name {
		case "numeric":
			s.Type = "number"
		case "bool":
			s.Type = "boolean"
		case "file":
			// TODO format "binary" (or "bytes" ?)
			s.Type = "string"
			s.Format = "binary"
		case "array": // TODO multidimensional arrays
			s.Type = "array"
			schema := openapi3.NewSchema()
			schema.Type = ruleNameToType(rule.Name)
			s.Items = &openapi3.SchemaRef{Value: schema}
		// TODO objects and string formats
		// TODO email, uuid, uri, ipv4, ipv6, date, date-time (not types but patterns)
		default:
			s.Type = rule.Name
		}
	}

	for _, r := range field.Rules {
		convertRule(r, s)
	}
	s.Nullable = field.IsNullable()
	return s
}
