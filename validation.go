package openapi3

import (
	"strconv"

	"github.com/getkin/kin-openapi/openapi3"
	"goyave.dev/goyave/v3/validation"
)

// ConvertToBody convert validation.Rules to OpenAPI RequestBody.
func ConvertToBody(rules *validation.Rules) *openapi3.RequestBodyRef {
	if rules == nil {
		return nil
	}
	// TODO cache using a simple map[*validation.Rules]*openapi3.RequestBodyRef
	schema := openapi3.NewObjectSchema()
	for name, field := range rules.Fields {
		schema.Properties[name] = &openapi3.SchemaRef{Value: SchemaFromField(field)}
		if field.IsRequired() {
			schema.Required = append(schema.Required, name)
		}
	}

	var content openapi3.Content
	if HasFile(rules) {
		content = openapi3.NewContentWithFormDataSchema(schema)
		if HasOnlyOptionalFiles(rules) {
			jsonSchema := openapi3.NewObjectSchema()
			jsonSchema.Required = schema.Required
			for name, prop := range schema.Properties {
				if prop.Value.Format != "binary" && prop.Value.Format != "bytes" {
					jsonSchema.Properties[name] = prop
				}
			}
			content["application/json"] = openapi3.NewMediaType().WithSchema(jsonSchema)
		}
	} else {
		content = openapi3.NewContentWithJSONSchema(schema)
	}
	body := openapi3.NewRequestBody().WithContent(content)
	if HasRequired(rules) {
		body.Required = true
	}
	return &openapi3.RequestBodyRef{
		Value: body,
	}
}

// ConvertToQuery convert validation.Rules to OpenAPI query Parameters.
func ConvertToQuery(rules *validation.Rules) []*openapi3.ParameterRef {
	if rules == nil {
		return nil
	}

	parameters := make([]*openapi3.ParameterRef, 0, len(rules.Fields))
	for name, field := range rules.Fields {
		param := openapi3.NewQueryParameter(name)
		param.Schema = &openapi3.SchemaRef{Value: SchemaFromField(field)}
		format := param.Schema.Value.Format
		if format != "binary" && format != "bytes" {
			param.Required = field.IsRequired()
			parameters = append(parameters, &openapi3.ParameterRef{Value: param})
		}
	}

	return parameters
}

// SchemaFromField convert a validation.Field to OpenAPI Schema.
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

// HasFile returns true if the given set of rules contains at least
// one "file" rule.
func HasFile(rules *validation.Rules) bool {
	return Has(rules, "file")
}

// HasRequired returns true if the given set of rules contains at least
// one "required" rule.
func HasRequired(rules *validation.Rules) bool {
	return Has(rules, "required")
}

// Has returns true if the given set of rules contains at least
// one rule having the given name.
func Has(rules *validation.Rules, ruleName string) bool {
	for _, f := range rules.Fields {
		for _, r := range f.Rules {
			if r.Name == ruleName {
				return true
			}
		}
	}
	return false
}

// HasOnlyOptionalFiles returns true if the given set of rules doesn't contain
// any required "file" rule.
func HasOnlyOptionalFiles(rules *validation.Rules) bool {
	for _, f := range rules.Fields {
		for _, r := range f.Rules {
			if r.Name == "file" && f.IsRequired() {
				return false
			}
		}
	}
	return true
}

func findFirstTypeRule(field *validation.Field) *validation.Rule {
	for _, rule := range field.Rules {
		if rule.IsType() || rule.Name == "file" {
			return rule
		}
	}
	return nil
}

func ruleNameToType(name string) string {
	switch name {
	case "numeric":
		return "number"
	case "bool":
		return "boolean"
	case "file":
		return "string"
	default:
		return name
	}
	// TODO match type rules with correct openapi types defined in spec
}

func convertRule(r *validation.Rule, s *openapi3.Schema) {
	// TODO minimum, maximum, string formats, arrays, uniqueItems (distinct)
	// TODO better architecture
	switch r.Name {
	case "min":
		switch s.Type {
		case "string":
			min, _ := strconv.ParseUint(r.Params[0], 10, 64)
			s.MinLength = min
		case "number", "integer":
			min, _ := strconv.ParseFloat(r.Params[0], 64)
			s.Min = &min
		case "array":
			min, _ := strconv.ParseUint(r.Params[0], 10, 64)
			s.MinItems = min
		}
	case "max":
		switch s.Type {
		case "string":
			max, _ := strconv.ParseUint(r.Params[0], 10, 64)
			s.MaxLength = &max
		case "number", "integer":
			max, _ := strconv.ParseFloat(r.Params[0], 64)
			s.Max = &max
		case "array":
			max, _ := strconv.ParseUint(r.Params[0], 10, 64)
			s.MaxItems = &max
		}
	}
}
