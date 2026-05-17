package environments_test

import (
	"testing"

	"github.com/isai-salazar-enc/postman-go-mcp/internal/environments"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate_DefaultEnvironments(t *testing.T) {
	g := environments.NewGenerator()
	envs, err := g.Generate("https://api.example.com", nil, nil)
	require.NoError(t, err)
	assert.Len(t, envs, 4)

	names := make(map[string]bool)
	for _, e := range envs {
		names[e.Name] = true
	}
	assert.True(t, names["dev"])
	assert.True(t, names["qa"])
	assert.True(t, names["stage"])
	assert.True(t, names["prod"])
}

func TestGenerate_DevURLPrepended(t *testing.T) {
	g := environments.NewGenerator()
	envs, err := g.Generate("https://api.example.com", []string{"dev"}, nil)
	require.NoError(t, err)
	require.Len(t, envs, 1)

	assert.Equal(t, "https://dev.api.example.com", envs[0].Values[0].Value)
}

func TestGenerate_ProdURLUnchanged(t *testing.T) {
	g := environments.NewGenerator()
	envs, err := g.Generate("https://api.example.com", []string{"prod"}, nil)
	require.NoError(t, err)
	require.Len(t, envs, 1)

	assert.Equal(t, "https://api.example.com", envs[0].Values[0].Value)
}

func TestGenerate_ExtraVariables(t *testing.T) {
	g := environments.NewGenerator()
	extras := map[string]string{"apiKey": "test-key"}
	envs, err := g.Generate("https://api.example.com", []string{"dev"}, extras)
	require.NoError(t, err)

	keys := make(map[string]string)
	for _, v := range envs[0].Values {
		keys[v.Key] = v.Value
	}
	assert.Equal(t, "test-key", keys["apiKey"])
}

func TestGenerate_EmptyBaseURL(t *testing.T) {
	g := environments.NewGenerator()
	_, err := g.Generate("", nil, nil)
	assert.Error(t, err)
}
