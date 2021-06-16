package openapi3

import (
	"testing"

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

func TestValidationSuite(t *testing.T) {
	suite.Run(t, new(ValidationTestSuite))
}
