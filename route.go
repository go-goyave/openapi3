package openapi3

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"goyave.dev/goyave/v3"
)

var (
	urlParamFormat        = regexp.MustCompile(`{\w+(:.+?)?}`)
	refInvalidCharsFormat = regexp.MustCompile(`[^A-Za-z0-9-._]`)
)

// RouteConverter converts goyave.Route to OpenAPI operations.
type RouteConverter struct {
	route       *goyave.Route
	refs        *Refs
	uri         string
	tag         string
	description string
	name        string
}

// NewRouteConverter create a new RouteConverter using the given Route as input.
// The converter will use and fill the given Refs.
func NewRouteConverter(route *goyave.Route, refs *Refs) *RouteConverter {
	return &RouteConverter{
		route: route,
		refs:  refs,
	}
}

// Convert route to OpenAPI operations and adds the results to the given spec.
func (c *RouteConverter) Convert(spec *openapi3.Swagger) {
	c.uri = c.cleanPath(c.route)
	c.tag = c.uriToTag(c.uri)
	c.name, c.description = c.readDescription()

	for _, m := range c.route.GetMethods() {
		if m == http.MethodHead || m == http.MethodOptions {
			continue
		}
		spec.AddOperation(c.uri, m, c.convertOperation(m, spec))
	}

	c.convertPathParameters(spec.Paths[c.uri], spec)
}

func (c *RouteConverter) convertOperation(method string, spec *openapi3.Swagger) *openapi3.Operation {
	op := openapi3.NewOperation()
	if c.tag != "" {
		op.Tags = []string{c.tag}
	}
	op.Description = c.description

	c.convertValidationRules(method, op, spec)

	op.Responses = openapi3.Responses{}
	// TODO annotations or something else for responses
	if len(op.Responses) == 0 {
		op.Responses["default"] = &openapi3.ResponseRef{Value: openapi3.NewResponse().WithDescription("")}
	}
	return op
}

func (c *RouteConverter) cleanPath(route *goyave.Route) string {
	// Regex are not allowed in URI, generate it without format definition
	_, params := route.GetFullURIAndParameters()
	bracedParams := make([]string, 0, len(params))
	for _, p := range params {
		bracedParams = append(bracedParams, "{"+p+"}")
	}

	return route.BuildURI(bracedParams...)
}

func (c *RouteConverter) uriToTag(uri string) string {
	// Take the first segment of the uri and use it as tag
	tag := ""
	if i := strings.Index(uri[1:], "/"); i != -1 {
		tag = uri[1 : i+1]
	} else {
		tag = uri[1:]
	}
	if len(tag) > 2 && tag[0] == '{' && tag[len(tag)-1] == '}' {
		// The first segment is a parameter
		return ""
	}

	return tag
}

func (c *RouteConverter) convertPathParameters(path *openapi3.PathItem, spec *openapi3.Swagger) {
	uri, params := c.route.GetFullURIAndParameters()
	formats := urlParamFormat.FindAllStringSubmatch(uri, -1)
	for i, p := range params {
		if c.parameterExists(path, p) {
			continue
		}
		param := openapi3.NewPathParameter(p)

		format := ""
		if len(formats[i]) == 2 {
			format = formats[i][1]
		}
		schemaRef := c.getParamSchema(p, format, spec)
		param.Schema = schemaRef
		ref := &openapi3.ParameterRef{Value: param}
		path.Parameters = append(path.Parameters, ref)
	}
}

func (c *RouteConverter) getParamSchema(paramName, format string, spec *openapi3.Swagger) *openapi3.SchemaRef {
	schema := openapi3.NewStringSchema()
	schema.Pattern = format
	schemaName := "param" + strings.Title(paramName)
	if format != "" {
		// Strip the colon
		schema.Pattern = schema.Pattern[1:]
		if schema.Pattern == "[0-9]+" {
			schema.Type = "integer"
			schemaName = "paramInteger"
		}
	} else {
		schemaName = "paramString"
	}
	spec.Components.Schemas[schemaName] = &openapi3.SchemaRef{Value: schema}
	schemaRef := &openapi3.SchemaRef{Ref: "#/components/schemas/" + schemaName}
	c.refs.ParamSchemas[schema.Pattern] = schemaRef
	return schemaRef
}

func (c *RouteConverter) parameterExists(path *openapi3.PathItem, param string) bool {
	for _, p := range path.Parameters { // TODO if refs used?
		if p.Value.Name == param {
			return true
		}
	}
	return false
}

func (c *RouteConverter) convertValidationRules(method string, op *openapi3.Operation, spec *openapi3.Swagger) {
	if rules := c.route.GetValidationRules(); rules != nil {
		if canHaveBody(method) {
			if cached, ok := c.refs.RequestBodies[rules]; ok {
				op.RequestBody = cached
				return
			}
			requestBody := ConvertToBody(rules)
			refName := c.rulesRefName()
			spec.Components.RequestBodies[refName] = requestBody
			requestBodyRef := &openapi3.RequestBodyRef{Ref: "#/components/requestBodies/" + refName}
			c.refs.RequestBodies[rules] = requestBodyRef
			op.RequestBody = requestBodyRef
		} else {
			if cached, ok := c.refs.Parameters[rules]; ok {
				op.Parameters = append(op.Parameters, cached...)
				return
			}
			refName := c.rulesRefName() + "-query-"
			query := ConvertToQuery(rules)
			c.refs.Parameters[rules] = make([]*openapi3.ParameterRef, 0, len(query))
			for _, p := range query {
				paramRefName := refName + p.Value.Name
				spec.Components.Parameters[paramRefName] = p

				ref := &openapi3.ParameterRef{Ref: "#/components/parameters/" + paramRefName}
				c.refs.Parameters[rules] = append(c.refs.Parameters[rules], ref)
				op.Parameters = append(op.Parameters, ref)
			}

		}
	}
}

func (c *RouteConverter) rulesRefName() string {
	// TODO this is using the name of the first route using a ref, which can be wrong sometimes
	return refInvalidCharsFormat.ReplaceAllString(c.name[strings.LastIndex(c.name, "/")+1:], "")
}

func (c *RouteConverter) readDescription() (string, string) {
	// TODO cache ast too
	pc := reflect.ValueOf(c.route.GetHandler()).Pointer()
	handlerValue := runtime.FuncForPC(pc)
	file, _ := handlerValue.FileLine(pc)
	funcName := handlerValue.Name()

	src, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}

	fset := token.NewFileSet() // positions are relative to fset

	f, err := parser.ParseFile(fset, file, src, parser.ParseComments)

	if err != nil {
		panic(err)
	}

	var doc *ast.CommentGroup

	// TODO optimize, this re-inspects the whole file for each route. Maybe cache already inspected files
	ast.Inspect(f, func(n ast.Node) bool { // TODO what would it do with closures and implementations?
		// Example output of "funcName" value for controller: goyave.dev/goyave/v3/auth.(*JWTController).Login-fm
		fn, ok := n.(*ast.FuncDecl)
		if ok {
			if fn.Name.IsExported() {
				if fn.Recv != nil {
					for _, f := range fn.Recv.List {
						if expr, ok := f.Type.(*ast.StarExpr); ok {
							if id, ok := expr.X.(*ast.Ident); ok {
								strct := fmt.Sprintf("(*%s)", id.Name) // TODO handle expr without star (no ptr)
								name := funcName[:len(funcName)-3]     // strip -fm
								expectedName := strct + "." + fn.Name.Name
								if name[len(name)-len(expectedName):] == expectedName {
									doc = fn.Doc
									return false
								}
							}
						}
					}
				}
				lastIndex := strings.LastIndex(funcName, ".")
				if funcName[lastIndex+1:] == fn.Name.Name {
					doc = fn.Doc
					return false
				}
			}
		}
		return true
	})

	if doc != nil {
		return funcName, strings.TrimSpace(doc.Text())
	}

	return "", ""
}
