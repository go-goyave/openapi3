package openapi3

import (
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/suite"
	"goyave.dev/goyave/v4/validation"
)

type ValidationTestSuite struct {
	suite.Suite
}

func (suite *ValidationTestSuite) TestHasFile() {
	rules := (&validation.Rules{
		Fields: validation.FieldMap{
			"field1": &validation.Field{
				Rules: []*validation.Rule{
					{Name: "required"},
					{Name: "string"},
				},
			},
			"field2": &validation.Field{
				Rules: []*validation.Rule{
					{Name: "required"},
					{Name: "file"},
				},
			},
		},
	}).AsRules()

	suite.True(HasFile(rules))

	rules = (&validation.Rules{
		Fields: validation.FieldMap{
			"field1": &validation.Field{
				Rules: []*validation.Rule{
					{Name: "required"},
					{Name: "string"},
				},
			},
			"field2": &validation.Field{
				Rules: []*validation.Rule{
					{Name: "required"},
					{Name: "bool"},
				},
			},
		},
	}).AsRules()

	suite.False(HasFile(rules))

}

func (suite *ValidationTestSuite) TestHasRequired() {
	rules := (&validation.Rules{
		Fields: validation.FieldMap{
			"field1": &validation.Field{
				Rules: []*validation.Rule{
					{Name: "required"},
					{Name: "string"},
				},
			},
			"field2": &validation.Field{
				Rules: []*validation.Rule{
					{Name: "file"},
					{Name: "nullable"},
				},
			},
		},
	}).AsRules()

	suite.True(HasRequired(rules))

	rules = (&validation.Rules{
		Fields: validation.FieldMap{
			"field1": &validation.Field{
				Rules: []*validation.Rule{
					{Name: "nullable"},
					{Name: "string"},
				},
			},
			"field2": &validation.Field{
				Rules: []*validation.Rule{
					{Name: "nullable"},
					{Name: "bool"},
				},
			},
			"object.field": &validation.Field{
				Rules: []*validation.Rule{
					{Name: "required"},
					{Name: "numeric"},
				},
			},
		},
	}).AsRules()

	suite.False(HasRequired(rules))
}

func (suite *ValidationTestSuite) TestHasOnlyOptionalFiles() {
	rules := (&validation.Rules{
		Fields: validation.FieldMap{
			"field1": &validation.Field{
				Rules: []*validation.Rule{
					{Name: "nullable"},
					{Name: "file"},
				},
			},
			"field2": &validation.Field{
				Rules: []*validation.Rule{
					{Name: "file"},
				},
			},
		},
	}).AsRules()

	suite.True(HasOnlyOptionalFiles(rules))

	rules = (&validation.Rules{
		Fields: validation.FieldMap{
			"field1": &validation.Field{
				Rules: []*validation.Rule{
					{Name: "nullable"},
					{Name: "file"},
				},
			},
			"field2": &validation.Field{
				Rules: []*validation.Rule{
					{Name: "required"},
					{Name: "file"},
				},
			},
		},
	}).AsRules()

	suite.False(HasOnlyOptionalFiles(rules))
}

func (suite *ValidationTestSuite) TestFindFirstTypeRule() {
	rules := (&validation.Rules{
		Fields: validation.FieldMap{
			"fieldString": &validation.Field{
				Rules: []*validation.Rule{
					{Name: "required"},
					{Name: "string"},
				},
			},
			"fieldFile": &validation.Field{
				Rules: []*validation.Rule{
					{Name: "required"},
					{Name: "file"},
				},
			},
			"fieldArray": &validation.Field{
				Rules: []*validation.Rule{
					{Name: "required"},
					{Name: "array"},
				},
			},
			"fieldNoType": &validation.Field{
				Rules: []*validation.Rule{
					{Name: "required"},
				},
			},
		},
	}).AsRules()

	suite.Equal(rules.Fields["fieldString"].(*validation.Field).Rules[1], findFirstTypeRule(rules.Fields["fieldString"].(*validation.Field)))
	suite.Equal(rules.Fields["fieldFile"].(*validation.Field).Rules[1], findFirstTypeRule(rules.Fields["fieldFile"].(*validation.Field)))
	suite.Equal(rules.Fields["fieldArray"].(*validation.Field).Rules[1], findFirstTypeRule(rules.Fields["fieldArray"].(*validation.Field)))
	suite.Nil(findFirstTypeRule(rules.Fields["fieldNoType"].(*validation.Field)))
}

func (suite *ValidationTestSuite) TestRuleNameToType() {
	suite.Equal("number", ruleNameToType("numeric"))
	suite.Equal("boolean", ruleNameToType("bool"))
	suite.Equal("string", ruleNameToType("file"))
	suite.Equal("integer", ruleNameToType("integer"))
}

func (suite *ValidationTestSuite) TestRegisterRuleConverter() {
	RegisterRuleConverter("testrule", func(r *validation.Rule, s *openapi3.Schema, encoding *openapi3.Encoding) {})
	suite.Contains(ruleConverters, "testrule")
}

func (suite *ValidationTestSuite) TestMinRuleConverter() {
	f := ruleConverters["min"]
	schema := openapi3.NewStringSchema()
	f(&validation.Rule{Params: []string{"5"}}, schema, nil)
	suite.Equal(uint64(5), schema.MinLength)

	schema = openapi3.NewSchema()
	schema.Type = "number"
	f(&validation.Rule{Params: []string{"5"}}, schema, nil)
	suite.Equal(float64(5), *schema.Min)

	schema = openapi3.NewSchema()
	schema.Type = "integer"
	f(&validation.Rule{Params: []string{"5"}}, schema, nil)
	suite.Equal(float64(5), *schema.Min)

	schema = openapi3.NewSchema()
	schema.Type = "array"
	f(&validation.Rule{Params: []string{"5"}}, schema, nil)
	suite.Equal(uint64(5), schema.MinItems)
}

func (suite *ValidationTestSuite) TestMaxRuleConverter() {
	f := ruleConverters["max"]
	schema := openapi3.NewStringSchema()
	f(&validation.Rule{Params: []string{"5"}}, schema, nil)
	suite.Equal(uint64(5), *schema.MaxLength)

	schema = openapi3.NewSchema()
	schema.Type = "number"
	f(&validation.Rule{Params: []string{"5"}}, schema, nil)
	suite.Equal(float64(5), *schema.Max)

	schema = openapi3.NewSchema()
	schema.Type = "integer"
	f(&validation.Rule{Params: []string{"5"}}, schema, nil)
	suite.Equal(float64(5), *schema.Max)

	schema = openapi3.NewSchema()
	schema.Type = "array"
	f(&validation.Rule{Params: []string{"5"}}, schema, nil)
	suite.Equal(uint64(5), *schema.MaxItems)
}

func (suite *ValidationTestSuite) TestBetweenRuleConverter() {
	f := ruleConverters["between"]
	schema := openapi3.NewStringSchema()
	f(&validation.Rule{Params: []string{"5", "10"}}, schema, nil)
	suite.Equal(uint64(5), schema.MinLength)
	suite.Equal(uint64(10), *schema.MaxLength)

	schema = openapi3.NewSchema()
	schema.Type = "number"
	f(&validation.Rule{Params: []string{"5", "10"}}, schema, nil)
	suite.Equal(float64(5), *schema.Min)
	suite.Equal(float64(10), *schema.Max)

	schema = openapi3.NewSchema()
	schema.Type = "integer"
	f(&validation.Rule{Params: []string{"5", "10"}}, schema, nil)
	suite.Equal(float64(5), *schema.Min)
	suite.Equal(float64(10), *schema.Max)

	schema = openapi3.NewSchema()
	schema.Type = "array"
	f(&validation.Rule{Params: []string{"5", "10"}}, schema, nil)
	suite.Equal(uint64(5), schema.MinItems)
	suite.Equal(uint64(10), *schema.MaxItems)
}

func (suite *ValidationTestSuite) TestSizeRuleConverter() {
	f := ruleConverters["size"]
	schema := openapi3.NewStringSchema()
	f(&validation.Rule{Params: []string{"5"}}, schema, nil)
	suite.Equal(uint64(5), schema.MinLength)
	suite.Equal(uint64(5), *schema.MaxLength)

	schema = openapi3.NewSchema()
	schema.Type = "number"
	f(&validation.Rule{Params: []string{"5"}}, schema, nil)
	suite.Equal(float64(5), *schema.Min)
	suite.Equal(float64(5), *schema.Max)

	schema = openapi3.NewSchema()
	schema.Type = "integer"
	f(&validation.Rule{Params: []string{"5"}}, schema, nil)
	suite.Equal(float64(5), *schema.Min)
	suite.Equal(float64(5), *schema.Max)

	schema = openapi3.NewSchema()
	schema.Type = "array"
	f(&validation.Rule{Params: []string{"5"}}, schema, nil)
	suite.Equal(uint64(5), schema.MinItems)
	suite.Equal(uint64(5), *schema.MaxItems)
}

func (suite *ValidationTestSuite) TestDistinctRuleConverter() {
	f := ruleConverters["distinct"]
	schema := openapi3.NewArraySchema()
	f(&validation.Rule{}, schema, nil)
	suite.True(schema.UniqueItems)
}

func (suite *ValidationTestSuite) TestDigitsRuleConverter() {
	f := ruleConverters["digits"]
	schema := openapi3.NewStringSchema()
	f(&validation.Rule{}, schema, nil)
	suite.Equal("^[0-9]*$", schema.Pattern)
}

func (suite *ValidationTestSuite) TestRegexRuleConverter() {
	f := ruleConverters["regex"]
	schema := openapi3.NewStringSchema()
	f(&validation.Rule{Params: []string{"[0-9]+"}}, schema, nil)
	suite.Equal("[0-9]+", schema.Pattern)
}

func (suite *ValidationTestSuite) TestEmailRuleConverter() {
	f := ruleConverters["email"]
	schema := openapi3.NewStringSchema()
	f(&validation.Rule{}, schema, nil)
	suite.Equal("^[^@\\r\\n\\t]{1,64}@[^\\s]+$", schema.Pattern)
}

func (suite *ValidationTestSuite) TestAlphaRuleConverter() {
	f := ruleConverters["alpha"]
	schema := openapi3.NewStringSchema()
	f(&validation.Rule{}, schema, nil)
	suite.Equal("^[\\pL\\pM]+$", schema.Pattern)
}

func (suite *ValidationTestSuite) TestAlphaDashRuleConverter() {
	f := ruleConverters["alpha_dash"]
	schema := openapi3.NewStringSchema()
	f(&validation.Rule{}, schema, nil)
	suite.Equal("^[\\pL\\pM0-9_-]+$", schema.Pattern)
}

func (suite *ValidationTestSuite) TestAlphaNumRuleConverter() {
	f := ruleConverters["alpha_num"]
	schema := openapi3.NewStringSchema()
	f(&validation.Rule{}, schema, nil)
	suite.Equal("^[\\pL\\pM0-9]+$", schema.Pattern)
}

func (suite *ValidationTestSuite) TestStartsWithRuleConverter() {
	f := ruleConverters["starts_with"]
	schema := openapi3.NewStringSchema()
	f(&validation.Rule{Params: []string{"test"}}, schema, nil)
	suite.Equal("^test", schema.Pattern)
}

func (suite *ValidationTestSuite) TestEndsWithRuleConverter() {
	f := ruleConverters["ends_with"]
	schema := openapi3.NewStringSchema()
	f(&validation.Rule{Params: []string{"test"}}, schema, nil)
	suite.Equal("test$", schema.Pattern)
}

func (suite *ValidationTestSuite) TestIPv4RuleConverter() {
	f := ruleConverters["ipv4"]
	schema := openapi3.NewStringSchema()
	f(&validation.Rule{}, schema, nil)
	suite.Equal("ipv4", schema.Format)
}

func (suite *ValidationTestSuite) TestIPv6RuleConverter() {
	f := ruleConverters["ipv6"]
	schema := openapi3.NewStringSchema()
	f(&validation.Rule{}, schema, nil)
	suite.Equal("ipv6", schema.Format)
}

func (suite *ValidationTestSuite) TestURLRuleConverter() {
	f := ruleConverters["url"]
	schema := openapi3.NewStringSchema()
	f(&validation.Rule{}, schema, nil)
	suite.Equal("uri", schema.Format)
}

func (suite *ValidationTestSuite) TestUUIDRuleConverter() {
	f := ruleConverters["uuid"]
	schema := openapi3.NewStringSchema()
	f(&validation.Rule{}, schema, nil)
	suite.Equal("uuid", schema.Format)
}

func (suite *ValidationTestSuite) TestMimeRuleConverter() {
	f := ruleConverters["mime"]
	encoding := openapi3.NewEncoding()
	f(&validation.Rule{Params: []string{"application/json", "text/html"}}, nil, encoding)
	suite.Equal("application/json, text/html", encoding.ContentType)
}

func (suite *ValidationTestSuite) TestImageRuleConverter() {
	f := ruleConverters["image"]
	encoding := openapi3.NewEncoding()
	f(&validation.Rule{}, nil, encoding)
	suite.Equal("image/jpeg, image/png, image/gif, image/bmp, image/svg+xml, image/webp", encoding.ContentType)
}

func (suite *ValidationTestSuite) TestCountRuleConverter() {
	f := ruleConverters["count"]
	schema := openapi3.NewStringSchema()
	f(&validation.Rule{Params: []string{"1"}}, schema, nil)
	suite.Equal("string", schema.Type)

	f(&validation.Rule{Params: []string{"3"}}, schema, nil)
	suite.Equal("array", schema.Type)
	suite.Empty(schema.Format)
	suite.NotNil(schema.Items)
	suite.NotNil(schema.Items.Value)
	suite.Equal("string", schema.Items.Value.Type)
	suite.Equal("binary", schema.Items.Value.Format)
	suite.Equal(uint64(3), schema.MinItems)
	suite.Equal(uint64(3), *schema.MaxItems)
}

func (suite *ValidationTestSuite) TestCountMinRuleConverter() {
	f := ruleConverters["count_min"]
	schema := openapi3.NewStringSchema()
	f(&validation.Rule{Params: []string{"1"}}, schema, nil)
	suite.Equal("array", schema.Type)
	suite.Empty(schema.Format)
	suite.NotNil(schema.Items)
	suite.NotNil(schema.Items.Value)
	suite.Equal("string", schema.Items.Value.Type)
	suite.Equal("binary", schema.Items.Value.Format)
	suite.Equal(uint64(1), schema.MinItems)
}

func (suite *ValidationTestSuite) TestCountMaxRuleConverter() {
	f := ruleConverters["count_max"]
	schema := openapi3.NewStringSchema()
	f(&validation.Rule{Params: []string{"1"}}, schema, nil)
	suite.Equal("array", schema.Type)
	suite.Empty(schema.Format)
	suite.NotNil(schema.Items)
	suite.NotNil(schema.Items.Value)
	suite.Equal("string", schema.Items.Value.Type)
	suite.Equal("binary", schema.Items.Value.Format)
	suite.Equal(uint64(1), *schema.MaxItems)
}

func (suite *ValidationTestSuite) TestCountBetweenRuleConverter() {
	f := ruleConverters["count_between"]
	schema := openapi3.NewStringSchema()
	f(&validation.Rule{Params: []string{"3", "5"}}, schema, nil)
	suite.Equal("array", schema.Type)
	suite.Empty(schema.Format)
	suite.NotNil(schema.Items)
	suite.NotNil(schema.Items.Value)
	suite.Equal("string", schema.Items.Value.Type)
	suite.Equal("binary", schema.Items.Value.Format)
	suite.Equal(uint64(3), schema.MinItems)
	suite.Equal(uint64(5), *schema.MaxItems)
}

func (suite *ValidationTestSuite) TestDateRuleConverter() {
	f := ruleConverters["date"]
	schema := openapi3.NewStringSchema()
	f(&validation.Rule{}, schema, nil)
	suite.Equal("date", schema.Format)

	schema = openapi3.NewStringSchema()
	f(&validation.Rule{Params: []string{time.RFC3339}}, schema, nil)
	suite.Equal("date-time", schema.Format)
}

func checkField(field *validation.Field) {
	// This is required so the field can be checked and
	// isNullable and such can be cached
	(&validation.Rules{
		Fields: validation.FieldMap{
			"field": field,
		},
	}).AsRules()
}

func (suite *ValidationTestSuite) TestGenerateSchema() {
	field := &validation.Field{
		Rules: []*validation.Rule{
			{Name: "numeric"},
			{Name: "min", Params: []string{"5"}},
			{Name: "max", Params: []string{"10"}},
		},
	}
	checkField(field)

	schema, encoding := SchemaFromField(field)
	suite.Nil(encoding)
	suite.Equal("number", schema.Type)
	suite.Equal(float64(5), *schema.Min)
	suite.Equal(float64(10), *schema.Max)
	suite.False(schema.Nullable)

	field.Rules[0].Name = "integer"
	field.Rules = append(field.Rules, &validation.Rule{Name: "nullable"})
	checkField(field)

	schema, encoding = SchemaFromField(field)
	suite.Nil(encoding)
	suite.Equal("integer", schema.Type)
	suite.Equal(float64(5), *schema.Min)
	suite.Equal(float64(10), *schema.Max)
	suite.True(schema.Nullable)

	field = &validation.Field{Rules: []*validation.Rule{{Name: "bool"}}}
	checkField(field)
	schema, encoding = SchemaFromField(field)
	suite.Nil(encoding)
	suite.Equal("boolean", schema.Type)
}

func (suite *ValidationTestSuite) TestGenerateSchemaArray() {
	rules := (validation.RuleSet{
		"array":       validation.List{"array"},
		"array[]":     validation.List{"array", "max:3"},
		"array[][]":   validation.List{"array:numeric"},
		"array[][][]": validation.List{"max:4"},
	}).AsRules()

	schema, _ := SchemaFromField(rules.Fields["array"].(*validation.Field))
	suite.Equal("array", schema.Type)

	items := schema.Items
	suite.NotNil(items)
	suite.Equal("array", items.Value.Type)
	suite.Equal(uint64(3), *items.Value.MaxItems)

	items = items.Value.Items
	suite.NotNil(items)
	suite.Equal("array", items.Value.Type)

	items = items.Value.Items
	suite.NotNil(items)
	suite.Equal("number", items.Value.Type)
	suite.Equal(float64(4), *items.Value.Max)

	rules = (validation.RuleSet{
		"array":     validation.List{"array"},
		"array[]":   validation.List{"array", "max:3"},
		"array[][]": validation.List{"array:numeric"},
	}).AsRules()

	schema, _ = SchemaFromField(rules.Fields["array"].(*validation.Field))
	suite.Equal("array", schema.Type)
	items = schema.Items.Value.Items.Value.Items
	suite.NotNil(items)
	suite.Equal("number", items.Value.Type)
}

func (suite *ValidationTestSuite) TestGenerateSchemaArrayOfObject() {
	rules := (validation.RuleSet{
		"array":         validation.List{"array"},
		"array[].field": validation.List{"numeric", "max:3"},
	}).AsRules()

	body := ConvertToBody(rules)
	schema := body.Value.Content["application/json"].Schema.Value

	three := 3.0
	expected := &openapi3.Schema{
		Type: "object",
		Properties: openapi3.Schemas{
			"array": &openapi3.SchemaRef{Value: &openapi3.Schema{
				Type: "array",
				Items: &openapi3.SchemaRef{Value: &openapi3.Schema{
					Type: "object",
					Properties: openapi3.Schemas{
						"field": &openapi3.SchemaRef{Value: &openapi3.Schema{
							Type: "number",
							Max:  &three,
						}},
					},
				}},
			}},
		},
	}

	suite.Equal(expected, schema)
}

func (suite *ValidationTestSuite) TestGenerateSchemaFile() {
	field := &validation.Field{
		Rules: []*validation.Rule{
			{Name: "file"},
			{Name: "mime", Params: []string{"application/json", "text/html"}},
		},
	}
	checkField(field)

	schema, encoding := SchemaFromField(field)
	suite.NotNil(encoding)
	suite.Equal("application/json, text/html", encoding.ContentType)
	suite.Equal("string", schema.Type)
	suite.Equal("binary", schema.Format)
}

func (suite *ValidationTestSuite) TestNewContent() {
	rules := &validation.Rules{
		Fields: validation.FieldMap{
			"field": &validation.Field{Rules: []*validation.Rule{
				{Name: "required"},
				{Name: "string"},
			}},
		},
	}
	encodings := map[string]*openapi3.Encoding{}
	schema := openapi3.NewObjectSchema()

	content := newContent(rules, schema, encodings)
	suite.Contains(content, "application/json")
	mediaType := content["application/json"]
	suite.Same(schema, mediaType.Schema.Value)
}

func (suite *ValidationTestSuite) TestNewContentFile() {
	rules := (&validation.Rules{
		Fields: validation.FieldMap{
			"field": &validation.Field{Rules: []*validation.Rule{
				{Name: "required"},
				{Name: "string"},
			}},
			"file": &validation.Field{Rules: []*validation.Rule{
				{Name: "required"},
				{Name: "file"},
				{Name: "mime", Params: []string{"application/json", "text/html"}},
			}},
		},
	}).AsRules()

	encodings := map[string]*openapi3.Encoding{
		"file": {ContentType: "application/json, text/html"},
	}
	schema := openapi3.NewObjectSchema()

	content := newContent(rules, schema, encodings)
	suite.Contains(content, "multipart/form-data")
	mediaType := content["multipart/form-data"]
	suite.Same(schema, mediaType.Schema.Value)
	suite.Equal(encodings, mediaType.Encoding)
}

func (suite *ValidationTestSuite) TestNewContentFileOptional() {
	rules := (&validation.Rules{
		Fields: validation.FieldMap{
			"field": &validation.Field{Rules: []*validation.Rule{
				{Name: "required"},
				{Name: "string"},
			}},
			"file": &validation.Field{Rules: []*validation.Rule{
				{Name: "file"},
				{Name: "mime", Params: []string{"application/json", "text/html"}},
			}},
		},
	}).AsRules()

	encodings := map[string]*openapi3.Encoding{
		"file": {ContentType: "application/json, text/html"},
	}
	schema := openapi3.NewObjectSchema()
	schema.Properties["field"] = &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string"}}
	schema.Properties["file"] = &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", Format: "binary"}}

	content := newContent(rules, schema, encodings)
	suite.Contains(content, "multipart/form-data")
	mediaType := content["multipart/form-data"]
	suite.Same(schema, mediaType.Schema.Value)
	suite.Equal(encodings, mediaType.Encoding)

	suite.Contains(content, "application/json")
	mediaType = content["application/json"]
	suite.NotSame(schema, mediaType.Schema.Value)
	suite.Contains(mediaType.Schema.Value.Properties, "field")
	suite.NotContains(mediaType.Schema.Value.Properties, "file")
}

func (suite *ValidationTestSuite) TestConvertToBody() {
	suite.Nil(ConvertToBody(nil))

	rules := &validation.Rules{
		Fields: validation.FieldMap{
			"field1": &validation.Field{Rules: []*validation.Rule{
				{Name: "required"},
				{Name: "string"},
			}},
			"field2": &validation.Field{Rules: []*validation.Rule{
				{Name: "nullable"},
				{Name: "numeric"},
			}},
			"object": &validation.Field{Rules: []*validation.Rule{
				{Name: "object"},
			}},
			"object.prop": &validation.Field{Rules: []*validation.Rule{
				{Name: "required"},
				{Name: "string"},
			}},
			"object.subobject": &validation.Field{Rules: []*validation.Rule{
				{Name: "object"},
			}},
			"object.subobject.prop2": &validation.Field{Rules: []*validation.Rule{
				{Name: "numeric"},
			}},
			"object.subobject.prop3": &validation.Field{Rules: []*validation.Rule{
				{Name: "string"},
			}},
			"object.subobject.prop4": &validation.Field{Rules: []*validation.Rule{
				{Name: "bool"},
			}},
		},
	}

	bodyRef := ConvertToBody(rules)
	suite.NotNil(bodyRef.Value)

	suite.True(bodyRef.Value.Required)

	content := bodyRef.Value.Content["application/json"]
	suite.Contains(content.Schema.Value.Properties, "field1")
	suite.Contains(content.Schema.Value.Properties, "field2")
	suite.Contains(content.Schema.Value.Properties, "object")
	suite.Contains(content.Schema.Value.Required, "field1")

	object := content.Schema.Value.Properties["object"].Value
	suite.Contains(object.Properties, "prop")
	suite.Contains(object.Properties, "subobject")
	suite.Contains(object.Required, "prop")

	suite.Contains(object.Properties["subobject"].Value.Properties, "prop2")
	suite.Contains(object.Properties["subobject"].Value.Properties, "prop3")
	suite.Contains(object.Properties["subobject"].Value.Properties, "prop4")
}

func (suite *ValidationTestSuite) TestConvertToBodyEncoding() {
	rules := &validation.Rules{
		Fields: validation.FieldMap{
			"field1": &validation.Field{Rules: []*validation.Rule{
				{Name: "required"},
				{Name: "string"},
			}},
			"file": &validation.Field{Rules: []*validation.Rule{
				{Name: "required"},
				{Name: "file"},
				{Name: "mime", Params: []string{"application/json", "text/html"}},
			}},
		},
	}

	bodyRef := ConvertToBody(rules)
	content := bodyRef.Value.Content["multipart/form-data"]
	encoding := content.Encoding["file"]
	suite.NotNil(encoding)
	suite.Equal("application/json, text/html", encoding.ContentType)
}

func (suite *ValidationTestSuite) TestConvertToQuery() {
	suite.Nil(ConvertToQuery(nil))

	rules := &validation.Rules{
		Fields: validation.FieldMap{
			"field1": &validation.Field{Rules: []*validation.Rule{
				{Name: "required"},
				{Name: "string"},
			}},
			"field2": &validation.Field{Rules: []*validation.Rule{
				{Name: "nullable"},
				{Name: "numeric"},
			}},
			"object": &validation.Field{Rules: []*validation.Rule{
				{Name: "object"},
			}},
			"object.prop": &validation.Field{Rules: []*validation.Rule{
				{Name: "required"},
				{Name: "string"},
			}},
			"object.subobject": &validation.Field{Rules: []*validation.Rule{
				{Name: "object"},
			}},
			"object.subobject.prop2": &validation.Field{Rules: []*validation.Rule{
				{Name: "numeric"},
			}},
			"object.subobject.prop3": &validation.Field{Rules: []*validation.Rule{
				{Name: "string"},
			}},
			"object.subobject.prop4": &validation.Field{Rules: []*validation.Rule{
				{Name: "bool"},
			}},
			"file": &validation.Field{Rules: []*validation.Rule{
				{Name: "file"},
			}},
		},
	}

	query := ConvertToQuery(rules)

	field1 := findParam(query, "field1")
	field2 := findParam(query, "field2")
	object := findParam(query, "object")
	suite.Nil(findParam(query, "prop"))
	suite.Nil(findParam(query, "subobject"))
	suite.Nil(findParam(query, "prop2"))
	suite.Nil(findParam(query, "prop3"))
	suite.Nil(findParam(query, "prop4"))
	suite.Nil(findParam(query, "file"))

	suite.NotNil(field1)
	suite.NotNil(field2)
	suite.NotNil(object)

	suite.Contains(object.Value.Schema.Value.Required, "prop")
	suite.True(field1.Value.Required)
	suite.False(field2.Value.Required)
}

func findParam(query []*openapi3.ParameterRef, name string) *openapi3.ParameterRef {
	for _, v := range query {
		if v.Value.Name == name {
			return v
		}
	}
	return nil
}

func TestValidationSuite(t *testing.T) {
	suite.Run(t, new(ValidationTestSuite))
}
