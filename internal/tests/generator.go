// Package tests generates Postman JS test scripts for collection requests.
package tests

import (
	"fmt"

	"github.com/isai-salazar-enc/postman-go-mcp/pkg/postmanfmt"
)

// TestType identifies a category of test script to generate.
type TestType string

const (
	TestStatusCode     TestType = "status_code"
	TestResponseTime   TestType = "response_time"
	TestSchema         TestType = "schema"
	TestAuth           TestType = "auth"
	TestRequiredFields TestType = "required_fields"
)

// AllTestTypes lists all supported test types.
var AllTestTypes = []TestType{
	TestStatusCode,
	TestResponseTime,
	TestSchema,
	TestAuth,
	TestRequiredFields,
}

// TestGenerator is the port for generating Postman test scripts.
type TestGenerator interface {
	Generate(col *postmanfmt.Collection, types []TestType) (*postmanfmt.Collection, error)
}

// Generator is the default implementation of TestGenerator.
type Generator struct{}

// NewGenerator creates a new Generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate adds Postman JS test scripts to every leaf request in the collection.
func (g *Generator) Generate(col *postmanfmt.Collection, types []TestType) (*postmanfmt.Collection, error) {
	if col == nil {
		return nil, fmt.Errorf("collection is required")
	}
	if len(types) == 0 {
		types = AllTestTypes
	}

	enriched := *col
	enriched.Item = addTestsToItems(col.Item, types)
	return &enriched, nil
}

func addTestsToItems(items []postmanfmt.Item, types []TestType) []postmanfmt.Item {
	result := make([]postmanfmt.Item, len(items))
	for i, item := range items {
		cp := item
		if len(item.Item) > 0 {
			cp.Item = addTestsToItems(item.Item, types)
		} else if item.Request != nil {
			cp.Event = buildTestEvents(types)
		}
		result[i] = cp
	}
	return result
}

func buildTestEvents(types []TestType) []postmanfmt.Event {
	scripts := []string{}
	for _, t := range types {
		scripts = append(scripts, testScript(t)...)
	}
	if len(scripts) == 0 {
		return nil
	}
	return []postmanfmt.Event{
		{
			Listen: "test",
			Script: postmanfmt.Script{
				Type: "text/javascript",
				Exec: scripts,
			},
		},
	}
}

func testScript(t TestType) []string {
	switch t {
	case TestStatusCode:
		return []string{
			`pm.test("Status code is 2xx", function () {`,
			`    pm.expect(pm.response.code).to.be.within(200, 299);`,
			`});`,
			``,
		}
	case TestResponseTime:
		return []string{
			`pm.test("Response time is less than 2000ms", function () {`,
			`    pm.expect(pm.response.responseTime).to.be.below(2000);`,
			`});`,
			``,
		}
	case TestSchema:
		return []string{
			`pm.test("Response has valid JSON", function () {`,
			`    pm.response.to.have.jsonBody();`,
			`});`,
			``,
		}
	case TestAuth:
		return []string{
			`pm.test("Not unauthorized", function () {`,
			`    pm.expect(pm.response.code).to.not.equal(401);`,
			`});`,
			``,
		}
	case TestRequiredFields:
		return []string{
			`pm.test("Response is not empty", function () {`,
			`    var jsonData = pm.response.json();`,
			`    pm.expect(jsonData).to.not.be.null;`,
			`});`,
			``,
		}
	default:
		return nil
	}
}
