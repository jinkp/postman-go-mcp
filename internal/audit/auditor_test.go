package audit_test

import (
	"testing"

	"github.com/isai-salazar-enc/postman-go-mcp/internal/audit"
	"github.com/isai-salazar-enc/postman-go-mcp/pkg/postmanfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeRequest(name, method string) postmanfmt.Item {
	return postmanfmt.Item{
		Name: name,
		Request: &postmanfmt.Request{
			Method: method,
			URL:    postmanfmt.URL{Raw: "{{baseUrl}}/items"},
		},
	}
}

func makeRequestWithAll(name string) postmanfmt.Item {
	return postmanfmt.Item{
		Name:        name,
		Description: "A documented request",
		Response: []postmanfmt.Response{
			{Name: "200 OK", Code: 200},
		},
		Event: []postmanfmt.Event{
			{Listen: "test", Script: postmanfmt.Script{Type: "text/javascript", Exec: []string{`pm.test("ok", function(){});`}}},
		},
		Request: &postmanfmt.Request{
			Method: "GET",
			URL:    postmanfmt.URL{Raw: "{{baseUrl}}/items"},
		},
	}
}

func TestAudit_NilCollection(t *testing.T) {
	a := audit.NewAuditor()
	_, err := a.Run(nil)
	assert.Error(t, err)
}

func TestAudit_NoTests(t *testing.T) {
	col := &postmanfmt.Collection{
		Info: postmanfmt.Info{Name: "Test", Schema: postmanfmt.CollectionSchemaV21},
		Item: []postmanfmt.Item{makeRequest("Get Users", "GET")},
	}
	a := audit.NewAuditor()
	report, err := a.Run(col)
	require.NoError(t, err)

	hasNoTestsError := false
	for _, issue := range report.Issues {
		if issue.Rule == "NoTests" && issue.Severity == "error" {
			hasNoTestsError = true
		}
	}
	assert.True(t, hasNoTestsError)
}

func TestAudit_NoDescription(t *testing.T) {
	col := &postmanfmt.Collection{
		Info: postmanfmt.Info{Name: "Test", Schema: postmanfmt.CollectionSchemaV21},
		Item: []postmanfmt.Item{makeRequest("Get Users", "GET")},
	}
	a := audit.NewAuditor()
	report, err := a.Run(col)
	require.NoError(t, err)

	found := false
	for _, issue := range report.Issues {
		if issue.Rule == "NoDescription" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestAudit_DuplicateNames(t *testing.T) {
	col := &postmanfmt.Collection{
		Info: postmanfmt.Info{Name: "Test", Schema: postmanfmt.CollectionSchemaV21},
		Item: []postmanfmt.Item{
			{
				Name: "Users",
				Item: []postmanfmt.Item{
					makeRequest("Get Users", "GET"),
					makeRequest("Get Users", "GET"),
				},
			},
		},
	}
	a := audit.NewAuditor()
	report, err := a.Run(col)
	require.NoError(t, err)

	found := false
	for _, issue := range report.Issues {
		if issue.Rule == "DuplicateNames" && issue.Severity == "warning" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestAudit_ScoreDecreases(t *testing.T) {
	col := &postmanfmt.Collection{
		Info: postmanfmt.Info{Name: "Test", Schema: postmanfmt.CollectionSchemaV21},
		Item: []postmanfmt.Item{makeRequest("Get Users", "GET")},
	}
	a := audit.NewAuditor()
	report, err := a.Run(col)
	require.NoError(t, err)
	assert.Less(t, report.Score, 100)
	assert.GreaterOrEqual(t, report.Score, 0)
}

func TestAudit_PerfectScore(t *testing.T) {
	col := &postmanfmt.Collection{
		Info: postmanfmt.Info{Name: "Perfect API", Schema: postmanfmt.CollectionSchemaV21},
		Item: []postmanfmt.Item{makeRequestWithAll("Get Items")},
	}
	a := audit.NewAuditor()
	report, err := a.Run(col)
	require.NoError(t, err)

	// Even a "perfect" request may get info-level naming issues, score should be high
	assert.GreaterOrEqual(t, report.Score, 70)
}
