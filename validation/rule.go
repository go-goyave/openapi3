package validation

import (
	"strconv"

	"github.com/getkin/kin-openapi/openapi3"
	"goyave.dev/goyave/v3/validation"
)

func HasFile(rules *validation.Rules) bool {
	return Has(rules, "file")
}

func HasRequired(rules *validation.Rules) bool {
	return Has(rules, "required")
}

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
