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
	"goyave.dev/goyave/v4"
)

var (
	urlParamFormat        = regexp.MustCompile(`{\w+(:.+?)?}`)
	refInvalidCharsFormat = regexp.MustCompile(`[^A-Za-z0-9-._]`)
	closureFormat         = regexp.MustCompile(`\.[a-zA-Z-]+\.func[0-9]+$`)
)

// RouteConverter converts goyave.Route to OpenAPI operations.
type RouteConverter struct {
	route       *goyave.Route
	refs        *Refs
	uri         string
	tag         string
	description string
	funcName    string
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
func (c *RouteConverter) Convert(spec *openapi3.T) {
	c.uri = c.cleanPath(c.route)
	c.tag = c.uriToTag(c.uri)
	c.funcName, c.description = c.readDescription()

	for _, m := range c.route.GetMethods() {
		if m == http.MethodHead || m == http.MethodOptions {
			continue
		}
		if !c.operationExists(spec, c.uri, m) {
			spec.AddOperation(c.uri, m, c.convertOperation(m, spec))
		}
	}

	c.convertPathParameters(spec.Paths[c.uri], spec)
}

func (c *RouteConverter) operationExists(spec *openapi3.T, path, method string) bool {
	if spec.Paths == nil {
		return false
	}
	pathItem := spec.Paths[path]
	if pathItem == nil {
		return false
	}

	switch method {
	case http.MethodConnect:
		return pathItem.Connect != nil
	case http.MethodDelete:
		return pathItem.Delete != nil
	case http.MethodGet:
		return pathItem.Get != nil
	case http.MethodHead:
		return pathItem.Head != nil
	case http.MethodOptions:
		return pathItem.Options != nil
	case http.MethodPatch:
		return pathItem.Patch != nil
	case http.MethodPost:
		return pathItem.Post != nil
	case http.MethodPut:
		return pathItem.Put != nil
	case http.MethodTrace:
		return pathItem.Trace != nil
	default:
		panic(fmt.Errorf("unsupported HTTP method %q", method))
	}
}

func (c *RouteConverter) convertOperation(method string, spec *openapi3.T) *openapi3.Operation {
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
	startIndex := 1
	if uri[0] != '/' {
		startIndex = 0
	}
	if i := strings.Index(uri[startIndex:], "/"); i != -1 {
		tag = uri[startIndex : i+startIndex]
	} else {
		tag = uri[startIndex:]
	}
	if len(tag) > 2 && tag[0] == '{' && tag[len(tag)-1] == '}' {
		// The first segment is a parameter
		return ""
	}

	return tag
}

func (c *RouteConverter) convertPathParameters(path *openapi3.PathItem, spec *openapi3.T) {
	uri, params := c.route.GetFullURIAndParameters()
	formats := urlParamFormat.FindAllStringSubmatch(uri, -1)
	for i, p := range params {
		format := ""
		if len(formats[i]) == 2 {
			format = formats[i][1]
			if format != "" {
				format = format[1:] // Strip the colon
			}
		}
		schemaRef := c.getParamSchema(p, format, spec)

		paramName := p
		i := 1
		for {
			if paramRef, ok := c.refs.Parameters[paramName]; ok {
				if param, exists := spec.Components.Parameters[paramName]; exists && param.Value.Schema.Ref != schemaRef.Ref {
					i++
					paramName = fmt.Sprintf("%s.%d", p, i)
					continue
				}
				if c.parameterExists(path, paramRef) {
					break
				}
				path.Parameters = append(path.Parameters, paramRef)
				break
			} else {
				param := openapi3.NewPathParameter(p)
				param.Schema = schemaRef
				spec.Components.Parameters[paramName] = &openapi3.ParameterRef{Value: param}
				paramRef := &openapi3.ParameterRef{Ref: "#/components/parameters/" + paramName}
				c.refs.Parameters[paramName] = paramRef
				path.Parameters = append(path.Parameters, paramRef)
				break
			}
		}
	}
}

func (c *RouteConverter) getParamSchema(paramName, format string, spec *openapi3.T) *openapi3.SchemaRef {
	schema := openapi3.NewStringSchema()
	schema.Pattern = format
	originalSchemaName := "param" + strings.Title(paramName)
	schemaName := originalSchemaName
	if format == "" {
		schemaName = "paramString"
	} else if format == "[0-9]+" {
		schema.Type = "integer"
		schemaName = "paramInteger"
	}

	i := 1
	for {
		if cached, ok := c.refs.ParamSchemas[schemaName]; ok {
			if s, exists := spec.Components.Schemas[schemaName]; exists && (s.Value.Pattern != format || s.Value.Type != schema.Type) {
				i++
				schemaName = fmt.Sprintf("%s.%d", originalSchemaName, i)
				continue
			} else {
				return cached
			}
		}
		break
	}

	spec.Components.Schemas[schemaName] = &openapi3.SchemaRef{Value: schema}
	schemaRef := &openapi3.SchemaRef{Ref: "#/components/schemas/" + schemaName}
	c.refs.ParamSchemas[schemaName] = schemaRef
	return schemaRef
}

func (c *RouteConverter) parameterExists(path *openapi3.PathItem, ref *openapi3.ParameterRef) bool {
	for _, p := range path.Parameters {
		if p.Ref == ref.Ref {
			return true
		}
	}
	return false
}

func (c *RouteConverter) convertValidationRules(method string, op *openapi3.Operation, spec *openapi3.T) {
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
			if cached, ok := c.refs.QueryParameters[rules]; ok {
				op.Parameters = append(op.Parameters, cached...)
				return
			}
			refName := c.rulesRefName() + "-query-"
			query := ConvertToQuery(rules)
			c.refs.QueryParameters[rules] = make([]*openapi3.ParameterRef, 0, len(query))
			for _, p := range query {
				paramRefName := refName + p.Value.Name
				spec.Components.Parameters[paramRefName] = p

				ref := &openapi3.ParameterRef{Ref: "#/components/parameters/" + paramRefName}
				c.refs.QueryParameters[rules] = append(c.refs.QueryParameters[rules], ref)
				op.Parameters = append(op.Parameters, ref)
			}

		}
	}
}

func (c *RouteConverter) rulesRefName() string {
	// TODO this is using the name of the first route using a ref, which can be wrong sometimes
	return refInvalidCharsFormat.ReplaceAllString(c.funcName[strings.LastIndex(c.funcName, "/")+1:], "")
}

func (c *RouteConverter) readDescription() (string, string) {
	pc := reflect.ValueOf(c.route.GetHandler()).Pointer()
	if cached, ok := c.refs.HandlerDocs[pc]; ok {
		return cached.FuncName, cached.Description
	}
	handlerValue := runtime.FuncForPC(pc)
	funcName := handlerValue.Name()

	if closureFormat.MatchString(funcName) {
		// Closures can't be documented, there's no need to parse AST
		c.refs.HandlerDocs[pc] = &HandlerDoc{funcName, ""}
		return funcName, ""
	}

	file, _ := handlerValue.FileLine(pc)
	astFile := c.getAST(file)

	var doc *ast.CommentGroup

	ast.Inspect(astFile, func(n ast.Node) bool {
		// Example output of "funcName" value for controller: goyave.dev/goyave/v4/auth.(*JWTController).Login-fm
		fn, ok := n.(*ast.FuncDecl)
		if ok {
			if fn.Recv != nil {
				for _, f := range fn.Recv.List {
					strct := ""
					switch expr := f.Type.(type) {
					case *ast.StarExpr:
						if id, ok := expr.X.(*ast.Ident); ok {
							strct = fmt.Sprintf("(*%s)", id.Name)
						} else {
							continue
						}
					case *ast.Ident:
						strct = expr.Name
					default:
						continue
					}
					name := funcName
					if strings.HasSuffix(name, "-fm") {
						// strip -fm suffix
						name = funcName[:len(funcName)-3]
					}
					expectedName := strct + "." + fn.Name.Name
					startIndex := len(name) - len(expectedName)
					if startIndex > 0 && name[startIndex:] == expectedName {
						doc = fn.Doc
						return false
					}
				}
				return true
			}
			lastIndex := strings.LastIndex(funcName, ".")
			if funcName[lastIndex+1:] == fn.Name.Name {
				doc = fn.Doc
				return false
			}
		}
		return true
	})

	docs := ""
	if doc != nil {
		docs = strings.TrimSpace(doc.Text())
	}

	c.refs.HandlerDocs[pc] = &HandlerDoc{funcName, docs}
	return funcName, docs
}

func (c *RouteConverter) getAST(file string) *ast.File {
	astFile := c.refs.AST[file]
	if astFile == nil {
		src, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}

		fset := token.NewFileSet() // positions are relative to fset

		astFile, err = parser.ParseFile(fset, file, src, parser.ParseComments)
		if err != nil {
			panic(err)
		}
		c.refs.AST[file] = astFile
	}
	return astFile
}
