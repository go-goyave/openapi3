package openapi3

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v4"
	"goyave.dev/goyave/v4/config"
)

type UITestSuite struct {
	goyave.TestSuite
}

const (
	expectedBody = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
	<title>Generator API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css" >
    <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist/favicon-32x32.png" sizes="32x32" />
    <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist/favicon-16x16.png" sizes="16x16" />
    <style>
      html
      {
        box-sizing: border-box;
        overflow: -moz-scrollbars-vertical;
        overflow-y: scroll;
      }
      *,
      *:before,
      *:after
      {
        box-sizing: inherit;
      }
      body
      {
        margin:0;
        background: #fafafa;
      }
    </style>
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"> </script>
    <script src="https://unpkg.com/swagger-ui-dist/swagger-ui-standalone-preset.js"> </script>
    <script>
    window.onload = function() {
      // Begin Swagger UI call region
      const ui = SwaggerUIBundle({
        spec: {"info":{"title":"Test","version":"0.0.0"},"openapi":"3.0.0","paths":{}},
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout"
      })
      // End Swagger UI call region
      window.ui = ui
    }
  </script>
  </body>
</html>
`
)

func (suite *UITestSuite) TestNewOptions() {
	spec := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   "Test",
			Version: "0.0.0",
		},
		Paths: openapi3.Paths{},
	}
	opts := NewUIOptions(spec)
	suite.NotNil(opts)
	suite.Equal(`{"info":{"title":"Test","version":"0.0.0"},"openapi":"3.0.0","paths":{}}`, opts.Spec)
	suite.NotEmpty(opts.Favicon16)
	suite.NotEmpty(opts.Favicon32)
	suite.NotEmpty(opts.BundleURL)
	suite.NotEmpty(opts.PresetURL)
	suite.NotEmpty(opts.StylesURL)
	suite.Equal("Generator API Documentation", opts.Title)
}

func (suite *UITestSuite) TestNewOptionsNilSpec() {
	opts := NewUIOptions(nil)
	suite.NotNil(opts)
	suite.Empty(opts.Spec)
}

func (suite *UITestSuite) TestServe() {
	spec := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   "Test",
			Version: "0.0.0",
		},
		Paths: openapi3.Paths{},
	}
	opts := NewUIOptions(spec)

	suite.RunServer(func(r *goyave.Router) {
		Serve(r, "/swaggerui", opts)
	}, func() {
		resp, err := suite.Get("/swaggerui", nil)
		suite.Nil(err)
		if err == nil {
			body := suite.GetBody(resp)
			if err := resp.Body.Close(); err != nil {
				suite.Fail(err.Error())
			}
			suite.Equal(expectedBody, string(body))

			suite.Equal("text/html; charset=utf-8", resp.Header.Get("Content-Type"))
		}
	})

}

func TestUISuite(t *testing.T) {
	if err := config.LoadJSON(`{"app":{"name": "Generator"}}`); err != nil {
		assert.FailNow(t, err.Error())
	}
	goyave.RunTest(t, new(UITestSuite))
}
