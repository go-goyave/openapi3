package openapi3

import (
	"bytes"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
	"goyave.dev/goyave/v4"
	"goyave.dev/goyave/v4/config"
)

// UIOptions options for the SwaggerUI Handler.
type UIOptions struct {

	// Title the title of the SwaggerUI HTML document
	Title string

	// Favicon32 URL to a 32x32 PNG favicon
	Favicon32 string
	// Favicon32 URL to a 16x16 PNG favicon
	Favicon16 string

	// BundleURL URL to the SwaggerUI js bundle
	BundleURL string
	// BundleURL URL to the SwaggerUI standalone preset js bundle
	PresetURL string
	// StylesURL URL to the SwaggerUI CSS
	StylesURL string

	// Spec JSON object
	Spec string
}

const (
	uiTemplate = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
	<title>{{ .Title }}</title>
    <link rel="stylesheet" type="text/css" href="{{ .StylesURL }}" >
    <link rel="icon" type="image/png" href="{{ .Favicon32 }}" sizes="32x32" />
    <link rel="icon" type="image/png" href="{{ .Favicon16 }}" sizes="16x16" />
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
    <script src="{{ .BundleURL }}"> </script>
    <script src="{{ .PresetURL }}"> </script>
    <script>
    window.onload = function() {
      // Begin Swagger UI call region
      const ui = SwaggerUIBundle({
        spec: {{ .Spec }},
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

// NewUIOptions create a new UIOption struct with default values.
//
// By default, favicons, swagger-ui js and css use official latest available
// versions from the unpkg.com CDN.
//
// The given spec can be `nil`, in which case, you'll have to set the returned
// struct's `Spec` field to a valid JSON string.
func NewUIOptions(spec *openapi3.T) *UIOptions {
	var json []byte
	if spec == nil {
		json = []byte{}
	} else {
		json, _ = spec.MarshalJSON()
	}
	return &UIOptions{
		Title:     config.GetString("app.name") + " API Documentation",
		Favicon16: "https://unpkg.com/swagger-ui-dist/favicon-16x16.png",
		Favicon32: "https://unpkg.com/swagger-ui-dist/favicon-32x32.png",
		BundleURL: "https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js",
		PresetURL: "https://unpkg.com/swagger-ui-dist/swagger-ui-standalone-preset.js",
		StylesURL: "https://unpkg.com/swagger-ui-dist/swagger-ui.css",
		Spec:      string(json),
	}
}

// Serve register the SwaggerUI route on the given router, with the given uri, and using
// the given UIOptions.
func Serve(router *goyave.Router, uri string, opts *UIOptions) {
	r := router.Subrouter(uri)

	tmpl := template.Must(template.New("swaggerui").Parse(uiTemplate))

	buf := bytes.NewBuffer(nil)
	tmpl.Execute(buf, opts)
	b := buf.Bytes()

	r.Get("/", func(resp *goyave.Response, req *goyave.Request) {
		resp.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := resp.Write(b); err != nil {
			panic(err)
		}
	})
}
