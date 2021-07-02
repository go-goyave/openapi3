package openapi3

import (
	"go/ast"
	"net/http"
	"reflect"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/suite"
	"goyave.dev/goyave/v3"
	"goyave.dev/goyave/v3/validation"
)

type RouteTestSuite struct {
	suite.Suite
}

func (suite *RouteTestSuite) TestNewRouteConverter() {
	refs := NewRefs()
	route := &goyave.Route{}
	converter := NewRouteConverter(route, refs)
	suite.Same(route, converter.route)
	suite.Same(refs, converter.refs)
}

func (suite *RouteTestSuite) TestOperationExists() {
	spec := &openapi3.T{}
	refs := NewRefs()
	route := &goyave.Route{}
	converter := NewRouteConverter(route, refs)
	suite.False(converter.operationExists(spec, "/test", http.MethodConnect))

	spec.Paths = make(openapi3.Paths)
	suite.False(converter.operationExists(spec, "/test", http.MethodConnect))

	suite.addAndTestOperationExists(converter, spec, http.MethodConnect)
	suite.addAndTestOperationExists(converter, spec, http.MethodDelete)
	suite.addAndTestOperationExists(converter, spec, http.MethodGet)
	suite.addAndTestOperationExists(converter, spec, http.MethodHead)
	suite.addAndTestOperationExists(converter, spec, http.MethodOptions)
	suite.addAndTestOperationExists(converter, spec, http.MethodPatch)
	suite.addAndTestOperationExists(converter, spec, http.MethodPost)
	suite.addAndTestOperationExists(converter, spec, http.MethodPut)
	suite.addAndTestOperationExists(converter, spec, http.MethodTrace)

	suite.Panics(func() {
		converter.operationExists(spec, "/test", "not an HTTP method")
	})
}

func (suite *RouteTestSuite) addAndTestOperationExists(converter *RouteConverter, spec *openapi3.T, method string) {
	suite.False(converter.operationExists(spec, "/test", method))
	spec.AddOperation("/test", method, openapi3.NewOperation())
	suite.True(converter.operationExists(spec, "/test", method))
}

func (suite *RouteTestSuite) TestCleanPath() {
	router := goyave.NewRouter()
	route := router.Get("/test/{param1}/{param2:[0-9]+}", func(r1 *goyave.Response, r2 *goyave.Request) {})
	converter := NewRouteConverter(route, NewRefs())
	suite.Equal("/test/{param1}/{param2}", converter.cleanPath(route))
}

func (suite *RouteTestSuite) TestURIToTag() {
	converter := NewRouteConverter(&goyave.Route{}, NewRefs())
	suite.Equal("products", converter.uriToTag("/products/{id:[0-9]+}"))
	suite.Equal("products", converter.uriToTag("products/{id:[0-9]+}"))
	suite.Empty(converter.uriToTag("/{id:[0-9]+}"))
	suite.Empty(converter.uriToTag("{id:[0-9]+}"))
}

func (suite *RouteTestSuite) TestParameterExists() {
	converter := NewRouteConverter(&goyave.Route{}, NewRefs())
	path := &openapi3.PathItem{
		Parameters: openapi3.Parameters{
			&openapi3.ParameterRef{Ref: "param1"},
		},
	}
	suite.True(converter.parameterExists(path, &openapi3.ParameterRef{Ref: "param1"}))
	suite.False(converter.parameterExists(path, &openapi3.ParameterRef{Ref: "param2"}))
}

func (suite *RouteTestSuite) TestRulesRefName() {
	converter := NewRouteConverter(&goyave.Route{}, NewRefs())
	converter.funcName = "goyave.dev/goyave/v3/auth.(*JWTController).Login-fm"

	suite.Equal("auth.JWTController.Login-fm", converter.rulesRefName())
}

func (suite *RouteTestSuite) TestGetAST() {
	refs := NewRefs()
	converter := NewRouteConverter(&goyave.Route{}, refs)
	ast := converter.getAST("route.go")
	suite.Contains(refs.AST, "route.go")
	suite.Same(refs.AST["route.go"], ast)

	suite.Panics(func() {
		converter.getAST("notafile")
	})
	suite.Panics(func() {
		// Not a go file
		converter.getAST("go.mod")
	})
}

func (suite *RouteTestSuite) TestGetASTCached() {
	refs := NewRefs()
	astFile := &ast.File{}
	refs.AST["route.go"] = astFile
	converter := NewRouteConverter(&goyave.Route{}, refs)
	suite.Same(astFile, converter.getAST("route.go"))
}

func (suite *RouteTestSuite) TestReadDescription() {
	refs := NewRefs()
	router := goyave.NewRouter()
	route := router.Get("/test", HandlerTest)
	converter := NewRouteConverter(route, refs)
	pc := reflect.ValueOf(HandlerTest).Pointer()

	funcName, description := converter.readDescription()
	suite.Equal("goyave.dev/openapi3.HandlerTest", funcName)
	suite.Equal("HandlerTest a test handler for AST reading", description)
	suite.Contains(refs.HandlerDocs, pc)
}

func (suite *RouteTestSuite) TestReadDescriptionClosure() {
	refs := NewRefs()
	router := goyave.NewRouter()
	closure := func(r1 *goyave.Response, r2 *goyave.Request) {}
	route := router.Get("/test", closure)
	converter := NewRouteConverter(route, refs)
	pc := reflect.ValueOf(closure).Pointer()

	funcName, description := converter.readDescription()
	suite.Equal("goyave.dev/openapi3.(*RouteTestSuite).TestReadDescriptionClosure.func1", funcName)
	suite.Empty(description)
	suite.Contains(refs.HandlerDocs, pc)
}

func (suite *RouteTestSuite) TestReadDescriptionStruct() {
	refs := NewRefs()
	router := goyave.NewRouter()
	ctrl := &testController{}

	route := router.Get("/test", ctrl.handlerStar)
	converter := NewRouteConverter(route, refs)
	pc := reflect.ValueOf(ctrl.handlerStar).Pointer()

	funcName, description := converter.readDescription()
	suite.Equal("goyave.dev/openapi3.(*testController).handlerStar-fm", funcName)
	suite.Empty(description)
	suite.Contains(refs.HandlerDocs, pc)

	ctrl2 := testController{}
	route = router.Get("/test", ctrl2.handler)
	converter = NewRouteConverter(route, refs)
	pc2 := reflect.ValueOf(ctrl2.handler).Pointer()
	funcName, description = converter.readDescription()
	suite.Equal("goyave.dev/openapi3.testController.handler-fm", funcName)
	suite.Empty(description)
	suite.Contains(refs.HandlerDocs, pc2)
}

func (suite *RouteTestSuite) TestReadDescriptionCached() {
	refs := NewRefs()
	router := goyave.NewRouter()
	route := router.Get("/test", HandlerTest)
	converter := NewRouteConverter(route, refs)
	pc := reflect.ValueOf(HandlerTest).Pointer()
	refs.HandlerDocs[pc] = &HandlerDoc{
		FuncName:    "HandlerTest",
		Description: "Handler description",
	}
	funcName, description := converter.readDescription()
	suite.Equal("HandlerTest", funcName)
	suite.Equal("Handler description", description)
}

// HandlerTest a test handler for AST reading
func HandlerTest(resp *goyave.Response, req *goyave.Request) {
	resp.Status(http.StatusOK)
}

type testController struct{}

func (c *testController) handlerStar(resp *goyave.Response, req *goyave.Request) {
	resp.Status(http.StatusOK)
}

func (c testController) handler(resp *goyave.Response, req *goyave.Request) {
	resp.Status(http.StatusOK)
}

func (suite *RouteTestSuite) TestGetParamSchema() {
	spec := &openapi3.T{Components: openapi3.Components{Schemas: openapi3.Schemas{}}}
	refs := NewRefs()
	router := goyave.NewRouter()
	route := router.Get("/test/{param}/{id:[0-9]+}/{notint:[a-z0-9]+}", HandlerTest)
	converter := NewRouteConverter(route, refs)

	// no pattern
	schema := converter.getParamSchema("param", "", spec)
	suite.Nil(schema.Value)
	suite.Equal("#/components/schemas/paramString", schema.Ref)
	suite.Contains(spec.Components.Schemas, "paramString")
	ref := spec.Components.Schemas["paramString"]
	suite.Empty(ref.Value.Pattern)
	suite.Equal(refs.ParamSchemas["paramString"], schema)

	// with int pattern
	schema = converter.getParamSchema("id", "[0-9]+", spec)
	suite.Nil(schema.Value)
	suite.Equal("#/components/schemas/paramInteger", schema.Ref)
	suite.Contains(spec.Components.Schemas, "paramInteger")
	ref = spec.Components.Schemas["paramInteger"]
	suite.Equal(ref.Value.Pattern, "[0-9]+")
	suite.Equal(refs.ParamSchemas["paramInteger"], schema)

	// with pattern
	schema = converter.getParamSchema("notint", "[a-z0-9]+", spec)
	suite.Nil(schema.Value)
	suite.Equal("#/components/schemas/paramNotint", schema.Ref)
	suite.Contains(spec.Components.Schemas, "paramNotint")
	ref = spec.Components.Schemas["paramNotint"]
	suite.Equal(ref.Value.Pattern, "[a-z0-9]+")
	suite.Equal(refs.ParamSchemas["paramNotint"], schema)
}

func (suite *RouteTestSuite) TestGetParamSchemaCacheAndNaming() {
	spec := &openapi3.T{Components: openapi3.Components{Schemas: openapi3.Schemas{}}}
	refs := NewRefs()
	router := goyave.NewRouter()
	route := router.Get("/{param1:[a-z0-9]+}/{param2:[a-z0-9]+}", HandlerTest)
	converter := NewRouteConverter(route, refs)

	ref := converter.getParamSchema("param1", "[a-z0-9]+", spec) // First pattern is now cached
	cached := converter.getParamSchema("param1", "[a-z0-9]+", spec)
	suite.Same(ref, cached)
	suite.Contains(spec.Components.Schemas, "paramParam1")
	suite.Contains(refs.ParamSchemas, "paramParam1")

	route2 := router.Get("/{param1:[A-Z0-9]+}", HandlerTest) // Not the same pattern
	converter2 := NewRouteConverter(route2, refs)
	ref2 := converter2.getParamSchema("param1", "[A-Z0-9]+", spec)
	suite.NotSame(ref, ref2)
	suite.Contains(spec.Components.Schemas, "paramParam1.2")
	suite.Contains(refs.ParamSchemas, "paramParam1.2")
}

func (suite *RouteTestSuite) TestConvertPathParameter() {
	spec := &openapi3.T{
		Components: openapi3.Components{
			Schemas:    openapi3.Schemas{},
			Parameters: openapi3.ParametersMap{},
		},
	}
	refs := NewRefs()
	router := goyave.NewRouter()
	route := router.Get("/{test:[a-z0-9]+}/{param}/{id:[0-9]+}", HandlerTest)
	converter := NewRouteConverter(route, refs)

	path := &openapi3.PathItem{
		Get: &openapi3.Operation{
			Responses: openapi3.NewResponses(),
		},
	}

	converter.convertPathParameters(path, spec)

	suite.Contains(spec.Components.Parameters, "test")
	suite.Contains(spec.Components.Parameters, "param")
	suite.Contains(spec.Components.Parameters, "id")
	suite.Contains(refs.Parameters, "test")
	suite.Contains(refs.Parameters, "param")
	suite.Contains(refs.Parameters, "id")
	suite.Equal(refs.Parameters["test"].Ref, "#/components/parameters/test")
	suite.Equal(refs.Parameters["param"].Ref, "#/components/parameters/param")
	suite.Equal(refs.Parameters["id"].Ref, "#/components/parameters/id")
	suite.Contains(path.Parameters, refs.Parameters["test"])
	suite.Contains(path.Parameters, refs.Parameters["param"])
	suite.Contains(path.Parameters, refs.Parameters["id"])

	suite.Contains(spec.Components.Schemas, "paramTest")
	suite.Contains(spec.Components.Schemas, "paramString")
	suite.Contains(spec.Components.Schemas, "paramInteger")

	// Cache and naming
	path2 := &openapi3.PathItem{
		Get: &openapi3.Operation{
			Responses: openapi3.NewResponses(),
		},
	}
	route = router.Get("/{test:[A-Z0-9]+}/{param}", HandlerTest) // Not the same pattern
	converter = NewRouteConverter(route, refs)
	converter.convertPathParameters(path2, spec)
	suite.Contains(path2.Parameters, refs.Parameters["param"])
	suite.Contains(spec.Components.Parameters, "test.2")
	suite.Contains(refs.Parameters, "test.2")
	suite.Equal(refs.Parameters["test.2"].Ref, "#/components/parameters/test.2")

	lenBefore := len(path.Parameters)
	// No parameter should be added because they are already present
	converter.convertPathParameters(path2, spec)
	suite.Equal(lenBefore, len(path.Parameters))
}

func (suite *RouteTestSuite) TestConvertValidationRules() {
	spec := &openapi3.T{
		Components: openapi3.Components{
			Schemas:    openapi3.Schemas{},
			Parameters: openapi3.ParametersMap{},
		},
	}
	rules := &validation.Rules{
		Fields: validation.FieldMap{
			"field1": {Rules: []*validation.Rule{
				{Name: "required"},
				{Name: "string"},
			}},
			"field2": {Rules: []*validation.Rule{
				{Name: "nullable"},
				{Name: "numeric"},
			}},
		},
	}

	refs := NewRefs()
	router := goyave.NewRouter()
	route := router.Get("/test", HandlerTest).Validate(rules)
	converter := NewRouteConverter(route, refs)
	converter.funcName = "HandlerTest"

	op := &openapi3.Operation{Parameters: openapi3.Parameters{}}
	converter.convertValidationRules(http.MethodGet, op, spec)

	suite.Contains(refs.QueryParameters, rules)
	suite.Contains(spec.Components.Parameters, "HandlerTest-query-field1")
	suite.Contains(spec.Components.Parameters, "HandlerTest-query-field2")
	suite.NotNil(findQueryParamRef(refs.QueryParameters[rules], "#/components/parameters/HandlerTest-query-field1"))
	suite.NotNil(findQueryParamRef(refs.QueryParameters[rules], "#/components/parameters/HandlerTest-query-field2"))

	op = &openapi3.Operation{Parameters: openapi3.Parameters{}}
	converter.convertValidationRules(http.MethodGet, op, spec)
	suite.Equal(openapi3.Parameters(refs.QueryParameters[rules]), op.Parameters)
}

func (suite *RouteTestSuite) TestConvertValidationRulesWithBody() {
	spec := &openapi3.T{
		Components: openapi3.Components{
			Schemas:       openapi3.Schemas{},
			Parameters:    openapi3.ParametersMap{},
			RequestBodies: openapi3.RequestBodies{},
		},
	}
	rules := &validation.Rules{
		Fields: validation.FieldMap{
			"field1": {Rules: []*validation.Rule{
				{Name: "required"},
				{Name: "string"},
			}},
			"field2": {Rules: []*validation.Rule{
				{Name: "nullable"},
				{Name: "numeric"},
			}},
		},
	}

	refs := NewRefs()
	router := goyave.NewRouter()
	route := router.Post("/test", HandlerTest).Validate(rules)
	converter := NewRouteConverter(route, refs)
	converter.funcName = "HandlerTest"

	op := &openapi3.Operation{Parameters: openapi3.Parameters{}}
	converter.convertValidationRules(http.MethodPost, op, spec)

	suite.Contains(spec.Components.RequestBodies, "HandlerTest")
	suite.Contains(refs.RequestBodies, rules)
	suite.Equal(refs.RequestBodies[rules].Ref, "#/components/requestBodies/HandlerTest")
	suite.Same(op.RequestBody, refs.RequestBodies[rules])

	op = &openapi3.Operation{Parameters: openapi3.Parameters{}}
	converter.convertValidationRules(http.MethodPost, op, spec)
	suite.Equal(refs.RequestBodies[rules], op.RequestBody)
}

func findQueryParamRef(query []*openapi3.ParameterRef, ref string) *openapi3.ParameterRef {
	for _, v := range query {
		if v.Ref == ref {
			return v
		}
	}
	return nil
}

func (suite *RouteTestSuite) TestConvertOperation() {
	spec := &openapi3.T{
		Components: openapi3.Components{
			Schemas:       openapi3.Schemas{},
			Parameters:    openapi3.ParametersMap{},
			RequestBodies: openapi3.RequestBodies{},
		},
	}
	rules := &validation.Rules{
		Fields: validation.FieldMap{
			"field1": {Rules: []*validation.Rule{
				{Name: "required"},
				{Name: "string"},
			}},
			"field2": {Rules: []*validation.Rule{
				{Name: "nullable"},
				{Name: "numeric"},
			}},
		},
	}

	refs := NewRefs()
	router := goyave.NewRouter()
	route := router.Post("/test", HandlerTest).Validate(rules)
	converter := NewRouteConverter(route, refs)
	converter.funcName = "HandlerTest"
	converter.tag = "TestTag"
	converter.description = "Test Description"

	op := converter.convertOperation(http.MethodPost, spec)
	suite.Equal(converter.tag, op.Tags[0])
	suite.Equal(converter.description, op.Description)
	suite.Contains(op.Responses, "default")
	suite.NotNil(op.RequestBody)
}

func (suite *RouteTestSuite) TestConvert() {
	spec := &openapi3.T{
		Components: openapi3.Components{
			Schemas:       openapi3.Schemas{},
			Parameters:    openapi3.ParametersMap{},
			RequestBodies: openapi3.RequestBodies{},
		},
	}

	refs := NewRefs()
	router := goyave.NewRouter()
	route := router.Route("POST|HEAD|OPTIONS", "/test/{id:[0-9]+}", HandlerTest)
	converter := NewRouteConverter(route, refs)
	converter.Convert(spec)
	suite.Equal("/test/{id}", converter.uri)
	suite.Equal("test", converter.tag)
	suite.Equal("goyave.dev/openapi3.HandlerTest", converter.funcName)
	suite.Equal("HandlerTest a test handler for AST reading", converter.description)
	suite.Nil(spec.Paths["/test/{id}"].Head)
	suite.Nil(spec.Paths["/test/{id}"].Options)
	suite.NotNil(spec.Paths["/test/{id}"].Post)
}

func TestRouteSuite(t *testing.T) {
	suite.Run(t, new(RouteTestSuite))
}
