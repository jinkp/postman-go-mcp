// Package audit implements a quality rules engine for Postman collections.
package audit

import (
	"fmt"
	"strings"

	"github.com/isai-salazar-enc/postman-go-mcp/pkg/postmanfmt"
)

// AuditIssue describes a single quality problem found in a collection.
type AuditIssue struct {
	Severity string `json:"severity"` // error | warning | info
	Rule     string `json:"rule"`
	Location string `json:"location"`
	Message  string `json:"message"`
}

// AuditSummary holds counts by severity.
type AuditSummary struct {
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
	Infos    int `json:"infos"`
}

// AuditReport is the output of running the audit rules engine.
type AuditReport struct {
	Score   int          `json:"score"`
	Issues  []AuditIssue `json:"issues"`
	Summary AuditSummary `json:"summary"`
}

// Auditor is the port for running quality audits on collections.
type Auditor interface {
	Run(col *postmanfmt.Collection) (*AuditReport, error)
}

// RulesAuditor is the default implementation of Auditor.
type RulesAuditor struct{}

// NewAuditor creates a new RulesAuditor.
func NewAuditor() *RulesAuditor {
	return &RulesAuditor{}
}

// Run evaluates all audit rules against the collection and returns a scored report.
func (a *RulesAuditor) Run(col *postmanfmt.Collection) (*AuditReport, error) {
	if col == nil {
		return nil, fmt.Errorf("collection is required")
	}

	var issues []AuditIssue
	issues = append(issues, ruleNoTests(col)...)
	issues = append(issues, ruleNoDescription(col)...)
	issues = append(issues, ruleInconsistentAuth(col)...)
	issues = append(issues, ruleDuplicateNames(col)...)
	issues = append(issues, ruleMissingErrorResponses(col)...)
	issues = append(issues, ruleNamingConvention(col)...)

	summary := AuditSummary{}
	for _, issue := range issues {
		switch issue.Severity {
		case "error":
			summary.Errors++
		case "warning":
			summary.Warnings++
		case "info":
			summary.Infos++
		}
	}

	score := 100 - (summary.Errors*10 + summary.Warnings*3 + summary.Infos*1)
	if score < 0 {
		score = 0
	}

	return &AuditReport{
		Score:   score,
		Issues:  issues,
		Summary: summary,
	}, nil
}

// ruleNoTests raises an ERROR for every leaf request with no test events.
func ruleNoTests(col *postmanfmt.Collection) []AuditIssue {
	var issues []AuditIssue
	walkLeaves(col.Item, "", func(item postmanfmt.Item, folder string) {
		hasTest := false
		for _, ev := range item.Event {
			if ev.Listen == "test" {
				hasTest = true
				break
			}
		}
		if !hasTest {
			issues = append(issues, AuditIssue{
				Severity: "error",
				Rule:     "NoTests",
				Location: location(folder, item.Name),
				Message:  fmt.Sprintf("Request %q has no test scripts", item.Name),
			})
		}
	})
	return issues
}

// ruleNoDescription raises an ERROR for every leaf request with no description.
func ruleNoDescription(col *postmanfmt.Collection) []AuditIssue {
	var issues []AuditIssue
	walkLeaves(col.Item, "", func(item postmanfmt.Item, folder string) {
		desc := strings.TrimSpace(item.Description)
		if desc == "" && item.Request != nil {
			desc = strings.TrimSpace(item.Request.Description)
		}
		if desc == "" {
			issues = append(issues, AuditIssue{
				Severity: "error",
				Rule:     "NoDescription",
				Location: location(folder, item.Name),
				Message:  fmt.Sprintf("Request %q has no description", item.Name),
			})
		}
	})
	return issues
}

// ruleInconsistentAuth raises a WARNING if some requests have auth and others don't.
func ruleInconsistentAuth(col *postmanfmt.Collection) []AuditIssue {
	var withAuth, withoutAuth int
	walkLeaves(col.Item, "", func(item postmanfmt.Item, _ string) {
		if item.Request != nil && item.Request.Auth != nil {
			withAuth++
		} else {
			withoutAuth++
		}
	})
	if withAuth > 0 && withoutAuth > 0 {
		return []AuditIssue{{
			Severity: "warning",
			Rule:     "InconsistentAuth",
			Location: col.Info.Name,
			Message:  fmt.Sprintf("%d requests have auth, %d do not — check if this is intentional", withAuth, withoutAuth),
		}}
	}
	return nil
}

// ruleDuplicateNames raises a WARNING for duplicate request names within a folder.
func ruleDuplicateNames(col *postmanfmt.Collection) []AuditIssue {
	var issues []AuditIssue
	checkDuplicatesInFolder(col.Item, "", &issues)
	return issues
}

func checkDuplicatesInFolder(items []postmanfmt.Item, folder string, issues *[]AuditIssue) {
	seen := map[string]int{}
	for _, item := range items {
		if item.Request != nil {
			seen[item.Name]++
		}
		if len(item.Item) > 0 {
			checkDuplicatesInFolder(item.Item, item.Name, issues)
		}
	}
	for name, count := range seen {
		if count > 1 {
			*issues = append(*issues, AuditIssue{
				Severity: "warning",
				Rule:     "DuplicateNames",
				Location: folder,
				Message:  fmt.Sprintf("Request name %q appears %d times in folder %q", name, count, folder),
			})
		}
	}
}

// ruleMissingErrorResponses raises a WARNING for requests with no saved error responses.
// (In the MVP, this checks for the presence of any saved response examples.)
func ruleMissingErrorResponses(col *postmanfmt.Collection) []AuditIssue {
	var issues []AuditIssue
	walkLeaves(col.Item, "", func(item postmanfmt.Item, folder string) {
		if len(item.Response) == 0 {
			issues = append(issues, AuditIssue{
				Severity: "warning",
				Rule:     "MissingErrorResponses",
				Location: location(folder, item.Name),
				Message:  fmt.Sprintf("Request %q has no saved response examples", item.Name),
			})
		}
	})
	return issues
}

// ruleNamingConvention raises an INFO for requests whose names don't follow verb-noun convention.
func ruleNamingConvention(col *postmanfmt.Collection) []AuditIssue {
	httpVerbs := []string{"get", "list", "create", "update", "delete", "patch", "post", "put",
		"fetch", "retrieve", "remove", "add", "edit", "search", "find"}

	var issues []AuditIssue
	walkLeaves(col.Item, "", func(item postmanfmt.Item, folder string) {
		name := strings.ToLower(item.Name)
		follows := false
		for _, verb := range httpVerbs {
			if strings.HasPrefix(name, verb) {
				follows = true
				break
			}
		}
		if !follows {
			issues = append(issues, AuditIssue{
				Severity: "info",
				Rule:     "NamingConvention",
				Location: location(folder, item.Name),
				Message:  fmt.Sprintf("Request name %q does not start with a verb (e.g. Get, List, Create)", item.Name),
			})
		}
	})
	return issues
}

// walkLeaves calls fn for every leaf request item (recursive).
func walkLeaves(items []postmanfmt.Item, folder string, fn func(postmanfmt.Item, string)) {
	for _, item := range items {
		if len(item.Item) > 0 {
			walkLeaves(item.Item, item.Name, fn)
		} else if item.Request != nil {
			fn(item, folder)
		}
	}
}

func location(folder, name string) string {
	if folder == "" {
		return name
	}
	return folder + " / " + name
}
