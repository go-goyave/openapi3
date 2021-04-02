package openapi3

import (
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"goyave.dev/goyave/v3"
	"goyave.dev/goyave/v3/config"
)

type Generator struct {
	spec *openapi3.Swagger
	refs *Refs
}

func NewGenerator() *Generator {
	return &Generator{
		refs: NewRefs(),
	}
}

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
