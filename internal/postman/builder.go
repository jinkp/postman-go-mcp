// Package postman builds Postman Collection v2.1 objects from discovery.APISpec.
package postman

import (
	"fmt"
	"strings"

	"github.com/isai-salazar-enc/postman-go-mcp/internal/discovery"
	"github.com/isai-salazar-enc/postman-go-mcp/pkg/postmanfmt"
)

// BuildOptions controls how the collection is generated.
type BuildOptions struct {
	CollectionName string
	BaseURL        string
	IncludeAuth    bool
	GroupByTags    bool
}

// CollectionBuilder is the port for building Postman collections.
type CollectionBuilder interface {
	Build(spec *discovery.APISpec, opts BuildOptions) (*postmanfmt.Collection, error)
}

// Builder is the default implementation of CollectionBuilder.
type Builder struct{}

// NewBuilder creates a new Builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// Build converts an APISpec into a Postman Collection v2.1.
func (b *Builder) Build(spec *discovery.APISpec, opts BuildOptions) (*postmanfmt.Collection, error) {
	if spec == nil {
		return nil, fmt.Errorf("spec is required")
	}

	name := opts.CollectionName
	if name == "" {
		name = spec.Title
	}
	if name == "" {
		name = "Unnamed Collection"
	}

	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = spec.BaseURL
	}
	if baseURL == "" {
		baseURL = "{{baseUrl}}"
	}

	col := &postmanfmt.Collection{
		Info: postmanfmt.Info{
			Name:        name,
			Description: spec.Description,
			Schema:      postmanfmt.CollectionSchemaV21,
		},
		Variable: []postmanfmt.Variable{
			{Key: "baseUrl", Value: baseURL, Description: "Base URL for all requests"},
		},
	}

	if opts.GroupByTags {
		col.Item = b.buildFolders(spec.Endpoints, baseURL, opts)
	} else {
		for _, ep := range spec.Endpoints {
			col.Item = append(col.Item, b.buildRequest(ep, baseURL, opts))
		}
	}

	return col, nil
}

// buildFolders groups endpoints by their first tag into folders.
func (b *Builder) buildFolders(endpoints []discovery.Endpoint, baseURL string, opts BuildOptions) []postmanfmt.Item {
	folderMap := make(map[string][]postmanfmt.Item)
	folderOrder := []string{}

	for _, ep := range endpoints {
		tag := "Default"
		if len(ep.Tags) > 0 {
			tag = ep.Tags[0]
		}
		if _, exists := folderMap[tag]; !exists {
			folderOrder = append(folderOrder, tag)
		}
		folderMap[tag] = append(folderMap[tag], b.buildRequest(ep, baseURL, opts))
	}

	var folders []postmanfmt.Item
	for _, tag := range folderOrder {
		folders = append(folders, postmanfmt.Item{
			Name: tag,
			Item: folderMap[tag],
		})
	}
	return folders
}

// buildRequest converts a single Endpoint to a Postman Item (leaf request).
func (b *Builder) buildRequest(ep discovery.Endpoint, baseURL string, opts BuildOptions) postmanfmt.Item {
	name := ep.Summary
	if name == "" {
		name = ep.OperationID
	}
	if name == "" {
		name = fmt.Sprintf("%s %s", ep.Method, ep.Path)
	}

	rawURL, pathParts, pathVars := buildURL(baseURL, ep.Path)

	var queryParams []postmanfmt.QueryParam
	var headers []postmanfmt.Header

	for _, param := range ep.Parameters {
		switch param.In {
		case "query":
			queryParams = append(queryParams, postmanfmt.QueryParam{
				Key:   param.Name,
				Value: "",
			})
		case "header":
			headers = append(headers, postmanfmt.Header{
				Key:         param.Name,
				Value:       "",
				Description: param.Description,
			})
		}
	}

	// Default Content-Type header when there's a request body
	if ep.RequestBody != nil {
		if _, ok := ep.RequestBody.Content["application/json"]; ok {
			headers = append(headers, postmanfmt.Header{
				Key:   "Content-Type",
				Value: "application/json",
			})
		}
	}

	req := &postmanfmt.Request{
		Method: ep.Method,
		URL: postmanfmt.URL{
			Raw:      rawURL,
			Variable: pathVars,
			Path:     pathParts,
			Query:    queryParams,
		},
		Header:      headers,
		Description: ep.Description,
	}

	if ep.RequestBody != nil {
		req.Body = buildBody(ep.RequestBody)
	}

	return postmanfmt.Item{
		Name:    name,
		Request: req,
	}
}

// buildURL constructs the raw URL string, path parts, and path variables.
func buildURL(baseURL, path string) (raw string, parts []string, vars []postmanfmt.Variable) {
	// Replace {param} with :param for Postman path variables
	postmanPath := path
	segments := strings.Split(strings.TrimPrefix(path, "/"), "/")
	var postmanParts []string

	for _, seg := range segments {
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			varName := strings.TrimPrefix(strings.TrimSuffix(seg, "}"), "{")
			postmanParts = append(postmanParts, ":"+varName)
			vars = append(vars, postmanfmt.Variable{
				Key:   varName,
				Value: "",
			})
			postmanPath = strings.Replace(postmanPath, "{"+varName+"}", ":"+varName, 1)
		} else {
			postmanParts = append(postmanParts, seg)
		}
	}

	raw = "{{baseUrl}}" + postmanPath
	return raw, postmanParts, vars
}

// buildBody creates a Postman Body from a RequestBody definition.
func buildBody(rb *discovery.RequestBody) *postmanfmt.Body {
	if rb == nil {
		return nil
	}
	if _, ok := rb.Content["application/json"]; ok {
		return &postmanfmt.Body{
			Mode: "raw",
			Raw:  "{}",
			Options: &postmanfmt.BodyOptions{
				Raw: &postmanfmt.RawBodyOptions{Language: "json"},
			},
		}
	}
	return nil
}
