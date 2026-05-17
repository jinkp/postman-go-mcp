package docs_test

import (
	"testing"

	"github.com/isai-salazar-enc/postman-go-mcp/internal/docs"
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
						Name: "List Users",
						Request: &postmanfmt.Request{
							Method: "GET",
							URL:    postmanfmt.URL{Raw: "{{baseUrl}}/users"},
						},
					},
					{
						Name:        "Get User",
						Description: "Returns a single user",
						Request: &postmanfmt.Request{
							Method: "GET",
							URL:    postmanfmt.URL{Raw: "{{baseUrl}}/users/:id"},
						},
					},
				},
			},
		},
	}
}

func TestGenerate_AddsMissingDescriptions(t *testing.T) {
	g := docs.NewGenerator()
	col, err := g.Generate(sampleCollection(), docs.DocStyleConcise)
	require.NoError(t, err)
	require.NotNil(t, col)

	folder := col.Item[0]
	listUsers := folder.Item[0]
	assert.NotEmpty(t, listUsers.Description, "description should be generated")
}

func TestGenerate_NilCollection(t *testing.T) {
	g := docs.NewGenerator()
	_, err := g.Generate(nil, docs.DocStyleConcise)
	assert.Error(t, err)
}

func TestAudit_ReportsCoverage(t *testing.T) {
	g := docs.NewGenerator()
	report, err := g.Audit(sampleCollection())
	require.NoError(t, err)

	assert.Equal(t, 2, report.TotalEndpoints)
	assert.Equal(t, 1, report.Documented)
	assert.Len(t, report.Undocumented, 1)
	assert.Equal(t, 50.0, report.CoveragePercent)
}

func TestAudit_NilCollection(t *testing.T) {
	g := docs.NewGenerator()
	_, err := g.Audit(nil)
	assert.Error(t, err)
}

func TestAudit_FullCoverage(t *testing.T) {
	col := &postmanfmt.Collection{
		Info: postmanfmt.Info{Name: "Test", Schema: postmanfmt.CollectionSchemaV21},
		Item: []postmanfmt.Item{
			{
				Name:        "Get Items",
				Description: "Returns all items",
				Request:     &postmanfmt.Request{Method: "GET", URL: postmanfmt.URL{Raw: "{{baseUrl}}/items"}},
			},
		},
	}

	g := docs.NewGenerator()
	report, err := g.Audit(col)
	require.NoError(t, err)
	assert.Equal(t, 100.0, report.CoveragePercent)
	assert.Empty(t, report.Undocumented)
}
