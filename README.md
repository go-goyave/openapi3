# `openapi3` - Automatic spec generator for Goyave

[![Version](https://img.shields.io/github/v/release/go-goyave/openapi3?include_prereleases)](https://github.com/go-goyave/openapi3/releases)
[![Build Status](https://github.com/go-goyave/openapi3/workflows/Test/badge.svg)](https://github.com/go-goyave/gyv/actions)
[![Coverage Status](https://coveralls.io/repos/github/go-goyave/openapi3/badge.svg)](https://coveralls.io/github/go-goyave/openapi3)

An automated [OpenAPI 3](https://swagger.io/) specification generator for the [Goyave](https://github.com/go-goyave/goyave) REST API framework.

## Usage

```
go get -u goyave.dev/openapi3
```

Add the following at the end of your main route registrer:
```go
spec := openapi3.NewGenerator().Generate(router)
json, err := spec.MarshalJSON()
if err != nil {
    panic(err)
}
fmt.Println(string(json))
```

## TODO

- [ ] Proper tooling
- [ ] Handler to serve the openapi3 UI
- [ ] Tests

## License

This package is MIT Licensed. Copyright (c) 2021 Jérémy LAMBERT (SystemGlitch)