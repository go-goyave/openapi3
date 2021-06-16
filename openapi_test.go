package openapi3

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v3"
	"goyave.dev/goyave/v3/config"
)

type OpenAPITestSuite struct {
	goyave.TestSuite
}

func (suite *OpenAPITestSuite) TestNewGenerator() {
	generator := NewGenerator()
	suite.NotNil(generator)
	suite.NotNil(generator.refs)
}

func (suite *OpenAPITestSuite) TestMakeServers() {
	servers := makeServers()
	suite.Len(servers, 1)

	s := servers[0]
	suite.Equal("http://goyave.dev", s.URL)
}

func TestOpenAPISuite(t *testing.T) {
	if err := config.LoadJSON(`{
			"app": {
				"name": "Generator"
			},
			"server": {
				"domain": "goyave.dev",
				"port": 80
			}
		}`); err != nil {
		assert.FailNow(t, err.Error())
	}
	goyave.RunTest(t, new(OpenAPITestSuite))
}
