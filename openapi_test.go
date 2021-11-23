package openapi3

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v4"
	"goyave.dev/goyave/v4/config"
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

func (suite *OpenAPITestSuite) TestCanHaveBody() {
	suite.False(canHaveBody(http.MethodGet))
	suite.False(canHaveBody(http.MethodHead))
	suite.False(canHaveBody(http.MethodOptions))
	suite.False(canHaveBody(http.MethodConnect))
	suite.False(canHaveBody(http.MethodTrace))
	suite.True(canHaveBody(http.MethodDelete))
	suite.True(canHaveBody(http.MethodPatch))
	suite.True(canHaveBody(http.MethodPut))
	suite.True(canHaveBody(http.MethodPost))
}

func TestOpenAPISuite(t *testing.T) {
	if err := config.LoadJSON(`{
			"app": {
				"name": "Generator"
			},
			"server": {
				"protocol": "http",
				"domain": "goyave.dev",
				"port": 80
			}
		}`); err != nil {
		assert.FailNow(t, err.Error())
	}
	goyave.RunTest(t, new(OpenAPITestSuite))
}
