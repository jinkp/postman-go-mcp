package postmanfmt_test

import (
	"encoding/json"
	"testing"

	"github.com/isai-salazar-enc/postman-go-mcp/pkg/postmanfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectionMarshal(t *testing.T) {
	col := postmanfmt.Collection{
		Info: postmanfmt.Info{
			Name:   "Test API",
			Schema: postmanfmt.CollectionSchemaV21,
		},
		Item: []postmanfmt.Item{
			{
				Name: "Users",
				Item: []postmanfmt.Item{
					{
						Name: "Get Users",
						Request: &postmanfmt.Request{
							Method: "GET",
							URL: postmanfmt.URL{
								Raw: "{{baseUrl}}/users",
							},
						},
					},
				},
			},
		},
		Variable: []postmanfmt.Variable{
			{Key: "baseUrl", Value: "https://api.example.com"},
		},
	}

	data, err := json.Marshal(col)
	require.NoError(t, err)

	var decoded postmanfmt.Collection
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "Test API", decoded.Info.Name)
	assert.Equal(t, postmanfmt.CollectionSchemaV21, decoded.Info.Schema)
	assert.Len(t, decoded.Item, 1)
	assert.Equal(t, "Users", decoded.Item[0].Name)
	assert.Len(t, decoded.Item[0].Item, 1)
	assert.Equal(t, "GET", decoded.Item[0].Item[0].Request.Method)
}

func TestCollectionCountEndpoints(t *testing.T) {
	col := postmanfmt.Collection{
		Info: postmanfmt.Info{Name: "Test", Schema: postmanfmt.CollectionSchemaV21},
		Item: []postmanfmt.Item{
			{
				Name: "Folder A",
				Item: []postmanfmt.Item{
					{Name: "Req 1", Request: &postmanfmt.Request{Method: "GET"}},
					{Name: "Req 2", Request: &postmanfmt.Request{Method: "POST"}},
				},
			},
			{
				Name: "Folder B",
				Item: []postmanfmt.Item{
					{Name: "Req 3", Request: &postmanfmt.Request{Method: "DELETE"}},
				},
			},
		},
	}

	assert.Equal(t, 3, col.CountEndpoints())
}

func TestBodyOmitsNilOptions(t *testing.T) {
	body := postmanfmt.Body{
		Mode: "raw",
		Raw:  `{"key":"value"}`,
	}
	data, err := json.Marshal(body)
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &m))
	_, hasOptions := m["options"]
	assert.False(t, hasOptions, "options should be omitted when nil")
}

func TestEnvironmentMarshal(t *testing.T) {
	env := postmanfmt.Environment{
		Name: "dev",
		Values: []postmanfmt.EnvironmentValue{
			{Key: "baseUrl", Value: "https://dev.api.example.com", Enabled: true},
		},
	}
	data, err := json.Marshal(env)
	require.NoError(t, err)

	var decoded postmanfmt.Environment
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, "dev", decoded.Name)
	assert.True(t, decoded.Values[0].Enabled)
}

func TestSchemaConstant(t *testing.T) {
	assert.Equal(t,
		"https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		postmanfmt.CollectionSchemaV21,
	)
}
