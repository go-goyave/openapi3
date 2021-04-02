package validation

import (
	"github.com/getkin/kin-openapi/openapi3"
	"goyave.dev/goyave/v3/validation"
)

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
