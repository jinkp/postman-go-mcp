// Package docs provides documentation generation and auditing for Postman collections.
package docs

import (
	"fmt"
	"strings"

	"github.com/isai-salazar-enc/postman-go-mcp/pkg/postmanfmt"
)

// DocStyle controls documentation verbosity.
type DocStyle string

const (
	DocStyleConcise  DocStyle = "concise"
	DocStyleDetailed DocStyle = "detailed"
)

// UndocumentedItem identifies an endpoint missing documentation.
type UndocumentedItem struct {
	Folder string
	Name   string
	Method string
	Path   string
}

// DocAuditReport summarizes documentation coverage.
type DocAuditReport struct {
	TotalEndpoints  int
	Documented      int
	Undocumented    []UndocumentedItem
	CoveragePercent float64
}

// DocGenerator is the port for documentation generation and auditing.
type DocGenerator interface {
	Generate(col *postmanfmt.Collection, style DocStyle) (*postmanfmt.Collection, error)
	Audit(col *postmanfmt.Collection) (*DocAuditReport, error)
}

// Generator is the default implementation of DocGenerator.
type Generator struct{}

// NewGenerator creates a new Generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate enriches all leaf requests in the collection with auto-generated descriptions.
func (g *Generator) Generate(col *postmanfmt.Collection, style DocStyle) (*postmanfmt.Collection, error) {
	if col == nil {
		return nil, fmt.Errorf("collection is required")
	}
	enriched := *col
	enriched.Item = enrichItems(col.Item, "", style)
	return &enriched, nil
}

func enrichItems(items []postmanfmt.Item, folder string, style DocStyle) []postmanfmt.Item {
	result := make([]postmanfmt.Item, len(items))
	for i, item := range items {
		cp := item
		if len(item.Item) > 0 {
			// folder
			cp.Item = enrichItems(item.Item, item.Name, style)
		} else if item.Request != nil {
			cp.Description = buildDescription(item, style)
			if cp.Request != nil {
				reqCp := *cp.Request
				reqCp.Description = cp.Description
				cp.Request = &reqCp
			}
		}
		result[i] = cp
	}
	return result
}

func buildDescription(item postmanfmt.Item, style DocStyle) string {
	if item.Request == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**%s** `%s`\n\n", item.Request.Method, item.Request.URL.Raw))

	if item.Description != "" {
		sb.WriteString(item.Description + "\n\n")
	}

	if style == DocStyleDetailed {
		if len(item.Request.URL.Variable) > 0 {
			sb.WriteString("**Path Variables**\n\n")
			for _, v := range item.Request.URL.Variable {
				sb.WriteString(fmt.Sprintf("- `%s`: %s\n", v.Key, v.Description))
			}
			sb.WriteString("\n")
		}

		if len(item.Request.Header) > 0 {
			sb.WriteString("**Headers**\n\n")
			for _, h := range item.Request.Header {
				sb.WriteString(fmt.Sprintf("- `%s`: %s\n", h.Key, h.Description))
			}
			sb.WriteString("\n")
		}
	}

	return strings.TrimSpace(sb.String())
}

// Audit returns documentation coverage statistics for the collection.
func (g *Generator) Audit(col *postmanfmt.Collection) (*DocAuditReport, error) {
	if col == nil {
		return nil, fmt.Errorf("collection is required")
	}

	report := &DocAuditReport{}
	auditItems(col.Item, "", report)

	if report.TotalEndpoints > 0 {
		report.CoveragePercent = float64(report.Documented) / float64(report.TotalEndpoints) * 100
	}
	return report, nil
}

func auditItems(items []postmanfmt.Item, folder string, report *DocAuditReport) {
	for _, item := range items {
		if len(item.Item) > 0 {
			auditItems(item.Item, item.Name, report)
		} else if item.Request != nil {
			report.TotalEndpoints++
			desc := strings.TrimSpace(item.Description)
			if desc == "" && item.Request != nil {
				desc = strings.TrimSpace(item.Request.Description)
			}
			if desc != "" {
				report.Documented++
			} else {
				report.Undocumented = append(report.Undocumented, UndocumentedItem{
					Folder: folder,
					Name:   item.Name,
					Method: item.Request.Method,
					Path:   item.Request.URL.Raw,
				})
			}
		}
	}
}
