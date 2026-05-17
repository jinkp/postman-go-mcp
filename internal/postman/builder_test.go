package postman_test

import (
	"testing"

	"github.com/isai-salazar-enc/postman-go-mcp/internal/discovery"
	"github.com/isai-salazar-enc/postman-go-mcp/internal/postman"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleSpec() *discovery.APISpec {
	return &discovery.APISpec{
		Title:   "Pet Store",
		Version: "1.0.0",
		BaseURL: "https://api.example.com",
		Endpoints: []discovery.Endpoint{
			{
				Method:      "GET",
				Path:        "/pets",
				OperationID: "listPets",
				Summary:     "List all pets",
				Tags:        []string{"pets"},
			},
			{
				Method:      "POST",
				Path:        "/pets",
				OperationID: "createPet",
				Summary:     "Create a pet",
				Tags:        []string{"pets"},
				RequestBody: &discovery.RequestBody{
					Content: map[string]discovery.MediaType{
						"application/json": {},
					},
				},
			},
			{
				Method:      "GET",
				Path:        "/pets/{petId}",
				OperationID: "getPet",
				Summary:     "Get a pet by ID",
				Tags:        []string{"pets"},
				Parameters: []discovery.Parameter{
					{Name: "petId", In: "path", Required: true},
				},
			},
			{
				Method:      "GET",
				Path:        "/users",
				OperationID: "listUsers",
				Tags:        []string{},
			},
		},
	}
}

func TestBuild_GroupsByTags(t *testing.T) {
	b := postman.NewBuilder()
	col, err := b.Build(sampleSpec(), postman.BuildOptions{GroupByTags: true})
	require.NoError(t, err)
	require.NotNil(t, col)

	// Should have folders: "pets" and "Default"
	folderNames := make(map[string]bool)
	for _, item := range col.Item {
		folderNames[item.Name] = true
	}
	assert.True(t, folderNames["pets"])
	assert.True(t, folderNames["Default"])
}

func TestBuild_PathVariables(t *testing.T) {
	b := postman.NewBuilder()
	col, err := b.Build(sampleSpec(), postman.BuildOptions{GroupByTags: true})
	require.NoError(t, err)

	var getPet *postman.ItemFinder
	_ = getPet
	// Find getPet request in pets folder
	for _, folder := range col.Item {
		if folder.Name == "pets" {
			for _, item := range folder.Item {
				if item.Request != nil && item.Request.Method == "GET" && len(item.Request.URL.Variable) > 0 {
					assert.Contains(t, item.Request.URL.Raw, ":petId")
					assert.Equal(t, "petId", item.Request.URL.Variable[0].Key)
					return
				}
			}
		}
	}
	t.Error("expected to find GET /pets/{petId} with path variable")
}

func TestBuild_RequestBody(t *testing.T) {
	b := postman.NewBuilder()
	col, err := b.Build(sampleSpec(), postman.BuildOptions{GroupByTags: true})
	require.NoError(t, err)

	for _, folder := range col.Item {
		if folder.Name == "pets" {
			for _, item := range folder.Item {
				if item.Request != nil && item.Request.Method == "POST" {
					require.NotNil(t, item.Request.Body)
					assert.Equal(t, "raw", item.Request.Body.Mode)
					return
				}
			}
		}
	}
	t.Error("expected to find POST /pets with body")
}

func TestBuild_CollectionMetadata(t *testing.T) {
	b := postman.NewBuilder()
	col, err := b.Build(sampleSpec(), postman.BuildOptions{
		CollectionName: "My Custom Name",
		GroupByTags:    true,
	})
	require.NoError(t, err)
	assert.Equal(t, "My Custom Name", col.Info.Name)
	assert.NotEmpty(t, col.Variable)
	assert.Equal(t, "baseUrl", col.Variable[0].Key)
}

func TestBuild_NilSpec(t *testing.T) {
	b := postman.NewBuilder()
	_, err := b.Build(nil, postman.BuildOptions{})
	assert.Error(t, err)
}
