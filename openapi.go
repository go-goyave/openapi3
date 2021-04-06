package openapi3

import (
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"goyave.dev/goyave/v3"
	"goyave.dev/goyave/v3/config"
)

// Generator for OpenAPI 3 specification based on Router.
type Generator struct {
	spec *openapi3.Swagger
	refs *Refs
}

// NewGenerator create a new OpenAPI 3 specification Generator.
func NewGenerator() *Generator {
	return &Generator{
		refs: NewRefs(),
	}
}

// Generate an OpenAPI 3 specification based on the given Router.
//
// Goyave config will be loaded (if not already).
//
// The Info section is pre-filled with version 0.0.0 and the app name, fetched
// from the config.
// Servers section will be filled using the configuration as well, thanks to the
// goyave.BaseURL() function.
func (g *Generator) Generate(router *goyave.Router) *openapi3.Swagger {
	if err := loadConfig(); err != nil {
		fmt.Println(err)
		return nil
	}
	g.spec = &openapi3.Swagger{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   config.GetString("app.name"),
			Version: "0.0.0",
		},
		Paths:   make(openapi3.Paths),
		Servers: makeServers(),
		Components: openapi3.Components{
			Schemas:       make(openapi3.Schemas),
			RequestBodies: make(openapi3.RequestBodies),
			Responses:     make(openapi3.Responses),
			Parameters:    make(openapi3.ParametersMap),
		},
	}

	g.convertRouter(router)

	return g.spec
}

func (g *Generator) convertRouter(router *goyave.Router) {
	for _, route := range router.GetRoutes() {
		NewRouteConverter(route, g.refs).Convert(g.spec)
	}

	for _, subrouter := range router.GetSubrouters() {
		g.convertRouter(subrouter)
	}
}

func loadConfig() error {
	if !config.IsLoaded() {
		return config.Load()
	}
	return nil
}

func makeServers() openapi3.Servers {
	return openapi3.Servers{
		&openapi3.Server{
			URL: goyave.BaseURL(),
		},
	}
}

func canHaveBody(method string) bool {
	return method == http.MethodDelete ||
		method == http.MethodPatch ||
		method == http.MethodPost ||
		method == http.MethodPut
}
