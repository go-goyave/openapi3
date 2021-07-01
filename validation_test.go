package openapi3

import (
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/suite"
	"goyave.dev/goyave/v3/validation"
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

func (suite *ValidationTestSuite) TestSortKeys() {
	rules := (&validation.RuleSet{
		"field1.field2":        []string{},
		"field1.field2.field3": []string{},
		"field1":               []string{},
	}).AsRules()

	keys := sortKeys(rules)
	suite.Equal([]string{"field1", "field1.field2", "field1.field2.field3"}, keys)
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
			"fieldArrayDim": &validation.Field{
				Rules: []*validation.Rule{
					{Name: "required"},
					{Name: "string", ArrayDimension: 1},
				},
			},
			"fieldNoType": &validation.Field{
				Rules: []*validation.Rule{
					{Name: "required"},
				},
			},
		},
	}).AsRules()

	suite.Equal(rules.Fields["fieldString"].Rules[1], findFirstTypeRule(rules.Fields["fieldString"], 0))
	suite.Equal(rules.Fields["fieldFile"].Rules[1], findFirstTypeRule(rules.Fields["fieldFile"], 0))
	suite.Equal(rules.Fields["fieldArray"].Rules[1], findFirstTypeRule(rules.Fields["fieldArray"], 0))
	suite.Equal(rules.Fields["fieldArrayDim"].Rules[1], findFirstTypeRule(rules.Fields["fieldArrayDim"], 1))
	suite.Nil(findFirstTypeRule(rules.Fields["fieldNoType"], 0))
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

func (suite *ValidationTestSuite) TestParentSchema() {
	schema := openapi3.NewObjectSchema()

	prop1 := openapi3.NewObjectSchema()
	prop2 := openapi3.NewObjectSchema()
	prop3 := openapi3.NewStringSchema()

	prop2.Properties["prop3"] = &openapi3.SchemaRef{Value: prop3}
	prop1.Properties["prop2"] = &openapi3.SchemaRef{Value: prop2}
	schema.Properties["prop1"] = &openapi3.SchemaRef{Value: prop1}

	parent, name := findParentSchema(schema, "prop1.prop2.prop3")
	suite.Same(prop2, parent)
	suite.Equal("prop3", name)

	parent, name = findParentSchema(schema, "prop1.prop2.prop4.prop5")
	suite.Same(prop2.Properties["prop4"].Value, parent)
	suite.Equal("prop5", name)
}

func (suite *ValidationTestSuite) TestParentSchemaQuery() {
	parameters := make(openapi3.Parameters, 0, 3)

	param1 := openapi3.NewObjectSchema()
	param2 := openapi3.NewStringSchema()
	prop1 := openapi3.NewObjectSchema()
	prop2 := openapi3.NewObjectSchema()
	prop3 := openapi3.NewStringSchema()

	prop2.Properties["prop3"] = &openapi3.SchemaRef{Value: prop3}
	prop1.Properties["prop2"] = &openapi3.SchemaRef{Value: prop2}
	param1.Properties["prop1"] = &openapi3.SchemaRef{Value: prop1}

	parameters = append(parameters, &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:   "param1",
			Schema: &openapi3.SchemaRef{Value: param1}},
	})
	parameters = append(parameters, &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			Name:   "param2",
			Schema: &openapi3.SchemaRef{Value: param2}},
	})

	p, target, name := findParentSchemaQuery(parameters, "param2")
	suite.Equal(parameters, p)
	suite.Same(param2, target)
	suite.Equal("param2", name)

	p, target, name = findParentSchemaQuery(parameters, "param3")
	suite.NotNil(target)
	suite.Len(p, 3)
	suite.Equal("param3", name)

	_, target, name = findParentSchemaQuery(parameters, "param1.prop1.prop2")
	suite.Same(prop1, target)
	suite.Equal("prop2", name)

	_, target, name = findParentSchemaQuery(parameters, "param1.prop1.newprop.prop")
	suite.Same(prop1.Properties["newprop"].Value, target)
	suite.Equal("prop", name)

	_, target, name = findParentSchemaQuery(parameters, "param1.prop1.prop2.prop3")
	suite.Same(prop2, target)
	suite.Equal("prop3", name)

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

func (suite *ValidationTestSuite) TestGenerateSchemaTypeFallback() {
	field := &validation.Field{
		Rules: []*validation.Rule{
			{Name: "min", Params: []string{"5"}},
			{Name: "max", Params: []string{"10"}},
		},
	}
	checkField(field)
	schema, encoding := generateSchema(field, "fallback", 0)
	suite.Nil(encoding)
	suite.Equal("fallback", schema.Type)
}

func (suite *ValidationTestSuite) TestGenerateSchemaArray() {
	field := &validation.Field{
		Rules: []*validation.Rule{
			{Name: "array"},
			{Name: "array", ArrayDimension: 1},
			{Name: "array", Params: []string{"numeric"}, ArrayDimension: 2},
			{Name: "max", Params: []string{"3"}, ArrayDimension: 1},
			{Name: "max", Params: []string{"4"}, ArrayDimension: 3},
		},
	}
	checkField(field)

	schema, _ := SchemaFromField(field)
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

	// No array type fallback
	field = &validation.Field{Rules: []*validation.Rule{{Name: "array"}}}
	checkField(field)
	schema, _ = SchemaFromField(field)
	suite.Equal("array", schema.Type)
	items = schema.Items
	suite.NotNil(items)
	suite.Equal("string", items.Value.Type)
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
			"field": {Rules: []*validation.Rule{
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
			"field": {Rules: []*validation.Rule{
				{Name: "required"},
				{Name: "string"},
			}},
			"file": {Rules: []*validation.Rule{
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
			"field": {Rules: []*validation.Rule{
				{Name: "required"},
				{Name: "string"},
			}},
			"file": {Rules: []*validation.Rule{
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

func TestValidationSuite(t *testing.T) {
	suite.Run(t, new(ValidationTestSuite))
}
