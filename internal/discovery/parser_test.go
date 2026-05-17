package discovery_test

import (
	"testing"

	"github.com/isai-salazar-enc/postman-go-mcp/internal/discovery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLibOpenAPIParser_OpenAPI3(t *testing.T) {
	parser := discovery.NewLibOpenAPIParser()
	spec, err := parser.Parse("../../examples/petstore.yaml")
	require.NoError(t, err)
	require.NotNil(t, spec)

	assert.NotEmpty(t, spec.Title)
	assert.NotEmpty(t, spec.Endpoints)
}

func TestLibOpenAPIParser_InvalidSource(t *testing.T) {
	parser := discovery.NewLibOpenAPIParser()
	_, err := parser.Parse("./nonexistent.yaml")
	assert.Error(t, err)
}

func TestLibOpenAPIParser_InvalidSpec(t *testing.T) {
	parser := discovery.NewLibOpenAPIParser()
	_, err := parser.Parse("../../examples/invalid.yaml")
	assert.Error(t, err)
}
