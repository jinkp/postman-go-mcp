package tests_test

import (
	"testing"

	"github.com/isai-salazar-enc/postman-go-mcp/internal/tests"
	"github.com/isai-salazar-enc/postman-go-mcp/pkg/postmanfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleCollection() *postmanfmt.Collection {
	return &postmanfmt.Collection{
		Info: postmanfmt.Info{Name: "Test API", Schema: postmanfmt.CollectionSchemaV21},
		Item: []postmanfmt.Item{
			{
				Name: "Users",
				Item: []postmanfmt.Item{
					{
						Name:    "List Users",
						Request: &postmanfmt.Request{Method: "GET", URL: postmanfmt.URL{Raw: "{{baseUrl}}/users"}},
					},
				},
			},
		},
	}
}

func TestGenerate_AddsTestEvents(t *testing.T) {
	g := tests.NewGenerator()
	col, err := g.Generate(sampleCollection(), nil)
	require.NoError(t, err)

	item := col.Item[0].Item[0]
	require.NotEmpty(t, item.Event, "expected test events to be added")
	assert.Equal(t, "test", item.Event[0].Listen)
	assert.NotEmpty(t, item.Event[0].Script.Exec)
}

func TestGenerate_StatusCodeScript(t *testing.T) {
	g := tests.NewGenerator()
	col, err := g.Generate(sampleCollection(), []tests.TestType{tests.TestStatusCode})
	require.NoError(t, err)

	exec := col.Item[0].Item[0].Event[0].Script.Exec
	found := false
	for _, line := range exec {
		if line == `pm.test("Status code is 2xx", function () {` {
			found = true
		}
	}
	assert.True(t, found, "status code test not found in script")
}

func TestGenerate_AllTypes(t *testing.T) {
	g := tests.NewGenerator()
	col, err := g.Generate(sampleCollection(), tests.AllTestTypes)
	require.NoError(t, err)

	exec := col.Item[0].Item[0].Event[0].Script.Exec
	joined := ""
	for _, line := range exec {
		joined += line + "\n"
	}

	assert.Contains(t, joined, "Status code is 2xx")
	assert.Contains(t, joined, "Response time is less than 2000ms")
	assert.Contains(t, joined, "Response has valid JSON")
	assert.Contains(t, joined, "Not unauthorized")
	assert.Contains(t, joined, "Response is not empty")
}

func TestGenerate_NilCollection(t *testing.T) {
	g := tests.NewGenerator()
	_, err := g.Generate(nil, nil)
	assert.Error(t, err)
}
