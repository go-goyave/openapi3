package openapi3

import (
	"strconv"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/imdario/mergo"
	"goyave.dev/goyave/v4/helper"
	"goyave.dev/goyave/v4/helper/walk"
	"goyave.dev/goyave/v4/validation"
)

// ConvertToBody convert validation.Rules to OpenAPI RequestBody.
func ConvertToBody(rules *validation.Rules) *openapi3.RequestBodyRef {
	if rules == nil {
		return nil
	}

	rules = rules.AsRules() // Ensure rules are checked

	encodings := map[string]*openapi3.Encoding{}

	schema := openapi3.NewObjectSchema()
	for name, f := range rules.Fields {
		field := f.(*validation.Field)
		s, encoding := SchemaFromField(field)
		addSchema(field, field.Path, &openapi3.SchemaRef{Value: schema}, s)
		if encoding != nil {
			// TODO encoding should be ignored for objects
			encodings[name] = encoding
		}
	}

	content := newContent(rules, schema, encodings)
	body := openapi3.NewRequestBody().WithContent(content)
	body.Required = HasRequired(rules)
	return &openapi3.RequestBodyRef{
		Value: body,
	}
}

func newContent(rules *validation.Rules, schema *openapi3.Schema, encodings map[string]*openapi3.Encoding) openapi3.Content {
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
		if len(encodings) != 0 {
			content["multipart/form-data"].Encoding = encodings
		}
	} else {
		content = openapi3.NewContentWithJSONSchema(schema)
	}
	return content
}

// ConvertToQuery convert validation.Rules to OpenAPI query Parameters.
func ConvertToQuery(rules *validation.Rules) []*openapi3.ParameterRef {
	if rules == nil {
		return nil
	}

	rules = rules.AsRules() // Ensure rules are checked

	tmpSchema := openapi3.NewObjectSchema()
	parameters := make([]*openapi3.ParameterRef, 0, len(rules.Fields))
	for _, f := range rules.Fields {
		field := f.(*validation.Field)
		s, _ := SchemaFromField(field)
		addSchema(field, field.Path, &openapi3.SchemaRef{Value: tmpSchema}, s)
	}

	for name, s := range tmpSchema.Properties {
		param := openapi3.NewQueryParameter(name)
		param.Schema = s
		format := param.Schema.Value.Format
		if format != "binary" && format != "bytes" {
			param.Required = helper.ContainsStr(tmpSchema.Required, name)
			parameters = append(parameters, &openapi3.ParameterRef{Value: param})
		}
	}

	return parameters
}

// SchemaFromField convert a validation.Field to OpenAPI Schema.
func SchemaFromField(field *validation.Field) (*openapi3.Schema, *openapi3.Encoding) {
	return generateSchema(field, "")
}

func generateSchema(field *validation.Field, typeFallback string) (*openapi3.Schema, *openapi3.Encoding) {
	s := openapi3.NewSchema()
	if rule := findFirstTypeRule(field); rule != nil {
		switch rule.Name {
		case "numeric":
			s.Type = "number"
		case "bool":
			s.Type = "boolean"
		case "file":
			s.Type = "string"
			s.Format = "binary"
		case "array":
			s.Type = "array"
			itemsType := ""
			if len(rule.Params) > 0 {
				itemsType = ruleNameToType(rule.Params[0])
			}
			if field.Elements != nil {
				items, _ := generateSchema(field.Elements, itemsType)
				s.Items = &openapi3.SchemaRef{Value: items}
			} else {
				s.Items = &openapi3.SchemaRef{Value: &openapi3.Schema{Type: itemsType}}
			}
		default:
			s.Type = rule.Name
		}
	} else {
		s.Type = typeFallback
	}

	var encoding *openapi3.Encoding
	for _, r := range field.Rules {
		if (r.Name == "image" || r.Name == "mime") && encoding == nil {
			encoding = openapi3.NewEncoding()
		}
		if converter, ok := ruleConverters[r.Name]; ok {
			converter(r, s, encoding)
		}
	}

	s.Nullable = field.IsNullable()
	return s, encoding
}

// HasFile returns true if the given set of rules contains at least
// one "file" rule, ignoring fields inside objects.
func HasFile(rules *validation.Rules) bool {
	return Has(rules, "file")
}

// HasRequired returns true if the given set of rules contains at least
// one "required" rule, ignoring fields inside objects.
func HasRequired(rules *validation.Rules) bool {
	return Has(rules, "required")
}

// Has returns true if the given set of rules contains at least
// one rule having the given name, ignoring fields inside objects.
func Has(rules *validation.Rules, ruleName string) bool {
	for name, f := range rules.Fields {
		if strings.Contains(name, ".") {
			continue
		}
		for _, r := range f.(*validation.Field).Rules {
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
		field := f.(*validation.Field)
		for _, r := range field.Rules {
			if r.Name == "file" && field.IsRequired() {
				return false
			}
		}
	}
	return true
}

func findFirstTypeRule(field *validation.Field) *validation.Rule {
	for _, rule := range field.Rules {
		if rule.IsType() || rule.Name == "file" || rule.Name == "array" {
			return rule
		}
	}
	return nil
}

func addSchema(field *validation.Field, path *walk.Path, currentElement *openapi3.SchemaRef, schema *openapi3.Schema) {
	element := currentElement
	if path.Name != "" {
		if currentElement.Value.Properties == nil {
			currentElement.Value.Properties = make(map[string]*openapi3.SchemaRef)
		}
		ref, ok := currentElement.Value.Properties[path.Name]
		if !ok {
			ref = &openapi3.SchemaRef{Value: &openapi3.Schema{}}
			currentElement.Value.Properties[path.Name] = ref
		}
		element = ref
	}
	switch path.Type {
	case walk.PathTypeElement:
		mergo.Merge(element.Value, schema)
		if field.IsRequired() && path.Name != "" {
			currentElement.Value.Required = append(currentElement.Value.Required, path.Name)
		}
	case walk.PathTypeArray:
		element.Value.Type = "array"
		if element.Value.Items == nil {
			element.Value.Items = &openapi3.SchemaRef{Value: &openapi3.Schema{}}
		}
		addSchema(field, path.Next, element.Value.Items, schema)
	case walk.PathTypeObject:
		element.Value.Type = "object"
		addSchema(field, path.Next, element, schema)
	}
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
}

// RuleConverter sets a schema's fields to values matching the given validation
// rule, if supported.
type RuleConverter func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding)

// RegisterRuleConverter register a RuleConverter function for the rule identified by
// the given ruleName. Registering a rule converter allows to handle custom rules.
func RegisterRuleConverter(ruleName string, converter RuleConverter) {
	ruleConverters[ruleName] = converter
}

var (
	ruleConverters = map[string]RuleConverter{
		"min": func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding) {
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
		},
		"max": func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding) {
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
		},
		"between": func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding) {
			switch s.Type {
			case "string":
				min, _ := strconv.ParseUint(r.Params[0], 10, 64)
				max, _ := strconv.ParseUint(r.Params[1], 10, 64)
				s.MinLength = min
				s.MaxLength = &max
			case "number", "integer":
				min, _ := strconv.ParseFloat(r.Params[0], 64)
				max, _ := strconv.ParseFloat(r.Params[1], 64)
				s.Min = &min
				s.Max = &max
			case "array":
				min, _ := strconv.ParseUint(r.Params[0], 10, 64)
				max, _ := strconv.ParseUint(r.Params[1], 10, 64)
				s.MinItems = min
				s.MaxItems = &max
			}
		},
		"size": func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding) {
			switch s.Type {
			case "string":
				length, _ := strconv.ParseUint(r.Params[0], 10, 64)
				s.MinLength = length
				s.MaxLength = &length
			case "number", "integer":
				n, _ := strconv.ParseFloat(r.Params[0], 64)
				s.Min = &n
				s.Max = &n
			case "array":
				count, _ := strconv.ParseUint(r.Params[0], 10, 64)
				s.MinItems = count
				s.MaxItems = &count
			}
		},
		"distinct": func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding) {
			s.UniqueItems = true
		},
		"digits": func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding) {
			s.Pattern = "^[0-9]*$"
		},
		"regex": func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding) {
			s.Pattern = r.Params[0]
		},
		"email": func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding) {
			s.Pattern = "^[^@\\r\\n\\t]{1,64}@[^\\s]+$"
		},
		"alpha": func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding) {
			s.Pattern = "^[\\pL\\pM]+$"
		},
		"alpha_dash": func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding) {
			s.Pattern = "^[\\pL\\pM0-9_-]+$"
		},
		"alpha_num": func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding) {
			s.Pattern = "^[\\pL\\pM0-9]+$"
		},
		"starts_with": func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding) {
			s.Pattern = "^" + r.Params[0]
		},
		"ends_with": func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding) {
			s.Pattern = r.Params[0] + "$"
		},
		"ipv4": func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding) {
			s.Format = "ipv4"
		},
		"ipv6": func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding) {
			s.Format = "ipv6"
		},
		"url": func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding) {
			s.Format = "uri"
		},
		"uuid": func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding) {
			s.Format = "uuid"
		},
		"mime": func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding) {
			encoding.ContentType = strings.Join(r.Params, ", ")
		},
		"image": func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding) {
			mimes := []string{"image/jpeg", "image/png", "image/gif", "image/bmp", "image/svg+xml", "image/webp"}
			encoding.ContentType = strings.Join(mimes, ", ")
		},
		"count": func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding) {
			if r.Params[0] != "1" {
				s.Type = "array"
				s.Format = ""
				s.Items = &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:   "string",
						Format: "binary",
					},
				}
				count, _ := strconv.ParseUint(r.Params[0], 10, 64)
				s.MinItems = count
				s.MaxItems = &count
			}
		},
		"count_min": func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding) {
			s.Type = "array"
			s.Format = ""
			s.Items = &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type:   "string",
					Format: "binary",
				},
			}
			count, _ := strconv.ParseUint(r.Params[0], 10, 64)
			s.MinItems = count
		},
		"count_max": func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding) {
			s.Type = "array"
			s.Format = ""
			s.Items = &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type:   "string",
					Format: "binary",
				},
			}
			count, _ := strconv.ParseUint(r.Params[0], 10, 64)
			s.MaxItems = &count
		},
		"count_between": func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding) {
			s.Type = "array"
			s.Format = ""
			s.Items = &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type:   "string",
					Format: "binary",
				},
			}
			min, _ := strconv.ParseUint(r.Params[0], 10, 64)
			max, _ := strconv.ParseUint(r.Params[1], 10, 64)
			s.MinItems = min
			s.MaxItems = &max
		},
		"date": func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding) {
			if len(r.Params) != 0 {
				if r.Params[0] == time.RFC3339 {
					s.Format = "date-time"
				}
			} else {
				s.Format = "date"
			}
		},
	}
)
