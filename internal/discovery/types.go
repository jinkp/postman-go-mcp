// Package discovery defines types and interfaces for parsing API specs.
package discovery

// APISpec is the normalized representation of a parsed OpenAPI/Swagger spec.
type APISpec struct {
	Title       string
	Description string
	Version     string
	BaseURL     string
	Endpoints   []Endpoint
}

// Endpoint represents a single API operation.
type Endpoint struct {
	Method      string
	Path        string
	OperationID string
	Summary     string
	Description string
	Tags        []string
	Parameters  []Parameter
	RequestBody *RequestBody
	Responses   map[string]Response
	Security    []SecurityRequirement
}

// Parameter represents a path, query, header, or cookie parameter.
type Parameter struct {
	Name        string
	In          string // path | query | header | cookie
	Description string
	Required    bool
	Schema      *Schema
}

// RequestBody describes the body accepted by an operation.
type RequestBody struct {
	Description string
	Required    bool
	Content     map[string]MediaType // key: content-type e.g. "application/json"
}

// MediaType holds the schema for a specific content type.
type MediaType struct {
	Schema  *Schema
	Example interface{}
}

// Response describes a possible response from an operation.
type Response struct {
	Description string
	Content     map[string]MediaType
}

// Schema is a simplified JSON Schema used for parameter and body types.
type Schema struct {
	Type        string
	Format      string
	Description string
	Properties  map[string]*Schema
	Items       *Schema   // for array types
	Required    []string
	Enum        []interface{}
	Example     interface{}
}

// SecurityRequirement represents an OAuth2/API key/etc. requirement.
type SecurityRequirement struct {
	Name   string
	Scopes []string
}
