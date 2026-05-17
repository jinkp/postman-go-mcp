package discovery

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/pb33f/libopenapi"
	v2 "github.com/pb33f/libopenapi/datamodel/high/v2"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// LibOpenAPIParser implements Parser using the pb33f/libopenapi library.
// It supports OpenAPI 3.0/3.1 and Swagger 2.0, from file paths or HTTP URLs.
type LibOpenAPIParser struct {
	HTTPClient *http.Client
}

// NewLibOpenAPIParser creates a LibOpenAPIParser with default HTTP client.
func NewLibOpenAPIParser() *LibOpenAPIParser {
	return &LibOpenAPIParser{
		HTTPClient: http.DefaultClient,
	}
}

// Parse reads and parses the spec at source (file path or HTTP URL).
func (p *LibOpenAPIParser) Parse(source string) (*APISpec, error) {
	data, err := p.readSource(source)
	if err != nil {
		return nil, fmt.Errorf("read spec source %q: %w", source, err)
	}

	doc, err := libopenapi.NewDocument(data)
	if err != nil {
		return nil, fmt.Errorf("parse spec document: %w", err)
	}

	info := doc.GetSpecInfo()
	if info == nil {
		return nil, fmt.Errorf("could not determine spec version for %q", source)
	}

	specType := strings.ToLower(info.SpecType)
	switch {
	case strings.HasPrefix(specType, "openapi"):
		return p.parseV3(doc)
	case specType == "swagger":
		return p.parseV2(doc)
	default:
		return nil, fmt.Errorf("unsupported spec type %q", info.SpecType)
	}
}

// readSource fetches spec data from a file path or HTTP URL.
func (p *LibOpenAPIParser) readSource(source string) ([]byte, error) {
	u, err := url.Parse(source)
	if err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		resp, err := p.HTTPClient.Get(source)
		if err != nil {
			return nil, fmt.Errorf("http get: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("http status %d", resp.StatusCode)
		}
		return io.ReadAll(resp.Body)
	}
	return os.ReadFile(source)
}

// parseV3 converts an OpenAPI 3.x document to APISpec.
func (p *LibOpenAPIParser) parseV3(doc libopenapi.Document) (*APISpec, error) {
	model, err := doc.BuildV3Model()
	if err != nil && model == nil {
		return nil, fmt.Errorf("build v3 model: %w", err)
	}

	spec := &APISpec{}
	if model.Model.Info != nil {
		spec.Title = model.Model.Info.Title
		spec.Version = model.Model.Info.Version
		if model.Model.Info.Description != "" {
			spec.Description = model.Model.Info.Description
		}
	}

	if len(model.Model.Servers) > 0 && model.Model.Servers[0] != nil {
		spec.BaseURL = model.Model.Servers[0].URL
	}

	if model.Model.Paths != nil {
		for pathPair := model.Model.Paths.PathItems.Oldest(); pathPair != nil; pathPair = pathPair.Next() {
			path := pathPair.Key
			pathItem := pathPair.Value
			if pathItem == nil {
				continue
			}

			ops := map[string]*v3.Operation{
				"GET":     pathItem.Get,
				"POST":    pathItem.Post,
				"PUT":     pathItem.Put,
				"PATCH":   pathItem.Patch,
				"DELETE":  pathItem.Delete,
				"HEAD":    pathItem.Head,
				"OPTIONS": pathItem.Options,
				"TRACE":   pathItem.Trace,
			}

			for method, op := range ops {
				if op == nil {
					continue
				}
				endpoint := Endpoint{
					Method:      method,
					Path:        path,
					OperationID: op.OperationId,
					Summary:     op.Summary,
					Description: op.Description,
				}

				for _, tag := range op.Tags {
					endpoint.Tags = append(endpoint.Tags, tag)
				}

				for _, param := range op.Parameters {
					if param == nil {
						continue
					}
					p2 := Parameter{
						Name:        param.Name,
						In:          param.In,
						Description: param.Description,
						Required:    param.Required != nil && *param.Required,
					}
					endpoint.Parameters = append(endpoint.Parameters, p2)
				}

				if op.RequestBody != nil {
					rb := &RequestBody{
						Description: op.RequestBody.Description,
						Required:    op.RequestBody.Required != nil && *op.RequestBody.Required,
						Content:     map[string]MediaType{},
					}
					if op.RequestBody.Content != nil {
						for ctPair := op.RequestBody.Content.Oldest(); ctPair != nil; ctPair = ctPair.Next() {
							rb.Content[ctPair.Key] = MediaType{}
						}
					}
					endpoint.RequestBody = rb
				}

				endpoint.Responses = map[string]Response{}
				if op.Responses != nil && op.Responses.Codes != nil {
					for rPair := op.Responses.Codes.Oldest(); rPair != nil; rPair = rPair.Next() {
						if rPair.Value == nil {
							continue
						}
						endpoint.Responses[rPair.Key] = Response{
							Description: rPair.Value.Description,
						}
					}
				}

				spec.Endpoints = append(spec.Endpoints, endpoint)
			}
		}
	}

	return spec, nil
}

// parseV2 converts a Swagger 2.0 document to APISpec.
func (p *LibOpenAPIParser) parseV2(doc libopenapi.Document) (*APISpec, error) {
	model, err := doc.BuildV2Model()
	if err != nil && model == nil {
		return nil, fmt.Errorf("build v2 model: %w", err)
	}

	spec := &APISpec{}
	if model.Model.Info != nil {
		spec.Title = model.Model.Info.Title
		spec.Version = model.Model.Info.Version
	}

	if model.Model.Host != "" {
		scheme := "https"
		if len(model.Model.Schemes) > 0 {
			scheme = model.Model.Schemes[0]
		}
		basePath := model.Model.BasePath
		if basePath == "" {
			basePath = "/"
		}
		spec.BaseURL = fmt.Sprintf("%s://%s%s", scheme, model.Model.Host, basePath)
	}

	if model.Model.Paths != nil {
		for pathPair := model.Model.Paths.PathItems.Oldest(); pathPair != nil; pathPair = pathPair.Next() {
			path := pathPair.Key
			pathItem := pathPair.Value
			if pathItem == nil {
				continue
			}

			ops := map[string]*v2.Operation{
				"GET":     pathItem.Get,
				"POST":    pathItem.Post,
				"PUT":     pathItem.Put,
				"PATCH":   pathItem.Patch,
				"DELETE":  pathItem.Delete,
				"HEAD":    pathItem.Head,
				"OPTIONS": pathItem.Options,
			}

			for method, op := range ops {
				if op == nil {
					continue
				}
				endpoint := Endpoint{
					Method:      method,
					Path:        path,
					OperationID: op.OperationId,
					Summary:     op.Summary,
					Description: op.Description,
				}
				for _, tag := range op.Tags {
					endpoint.Tags = append(endpoint.Tags, tag)
				}

				for _, param := range op.Parameters {
					if param == nil {
						continue
					}
					p2 := Parameter{
						Name:        param.Name,
						In:          param.In,
						Description: param.Description,
						Required:    param.Required != nil && *param.Required,
					}
					endpoint.Parameters = append(endpoint.Parameters, p2)
				}

				endpoint.Responses = map[string]Response{}
				if op.Responses != nil && op.Responses.Codes != nil {
					for rPair := op.Responses.Codes.Oldest(); rPair != nil; rPair = rPair.Next() {
						if rPair.Value == nil {
							continue
						}
						endpoint.Responses[rPair.Key] = Response{
							Description: rPair.Value.Description,
						}
					}
				}

				spec.Endpoints = append(spec.Endpoints, endpoint)
			}
		}
	}

	return spec, nil
}
