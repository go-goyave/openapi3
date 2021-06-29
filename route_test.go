package openapi3

import (
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/suite"
	"goyave.dev/goyave/v3"
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

func (suite *RouteTestSuite) addAndTestOperationExists(converter *RouteConverter, spec *openapi3.T, method string) {
	suite.False(converter.operationExists(spec, "/test", method))
	spec.AddOperation("/test", method, openapi3.NewOperation())
	suite.True(converter.operationExists(spec, "/test", method))
}

func TestRouteSuite(t *testing.T) {
	suite.Run(t, new(RouteTestSuite))
}
