// Package postmanfmt defines Go structs for the Postman Collection v2.1 format.
// These are shared types used across all domain packages.
package postmanfmt

const CollectionSchemaV21 = "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"

// Collection is the root Postman Collection v2.1 object.
type Collection struct {
	Info     Info       `json:"info"`
	Item     []Item     `json:"item"`
	Variable []Variable `json:"variable,omitempty"`
	Auth     *Auth      `json:"auth,omitempty"`
	Event    []Event    `json:"event,omitempty"`
}

// Info holds collection metadata.
type Info struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Schema      string `json:"schema"`
}

// Item represents either a folder (has Item children) or a request (has Request).
type Item struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Item        []Item   `json:"item,omitempty"`    // folder: non-nil
	Request     *Request `json:"request,omitempty"` // leaf: non-nil
	Response    []Response `json:"response,omitempty"`
	Event       []Event  `json:"event,omitempty"`
}

// Request holds the HTTP request definition.
type Request struct {
	Method      string   `json:"method"`
	URL         URL      `json:"url"`
	Header      []Header `json:"header,omitempty"`
	Body        *Body    `json:"body,omitempty"`
	Auth        *Auth    `json:"auth,omitempty"`
	Description string   `json:"description,omitempty"`
}

// URL holds the request URL, supporting path variables and query params.
type URL struct {
	Raw      string   `json:"raw"`
	Protocol string   `json:"protocol,omitempty"`
	Host     []string `json:"host,omitempty"`
	Path     []string `json:"path,omitempty"`
	Variable []Variable `json:"variable,omitempty"`
	Query    []QueryParam `json:"query,omitempty"`
}

// QueryParam represents a URL query parameter.
type QueryParam struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	Disabled bool   `json:"disabled,omitempty"`
}

// Header represents an HTTP header key-value pair.
type Header struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
	Disabled    bool   `json:"disabled,omitempty"`
}

// Body holds the request body definition.
type Body struct {
	Mode    string      `json:"mode"` // raw | formdata | urlencoded | file | graphql
	Raw     string      `json:"raw,omitempty"`
	Options *BodyOptions `json:"options,omitempty"`
}

// BodyOptions provides body-type-specific options (e.g. language for raw).
type BodyOptions struct {
	Raw *RawBodyOptions `json:"raw,omitempty"`
}

// RawBodyOptions specifies the language for raw body mode.
type RawBodyOptions struct {
	Language string `json:"language"` // json | text | xml | html | javascript
}

// Auth defines the authentication scheme for a collection or request.
type Auth struct {
	Type   string      `json:"type"`
	Bearer []AuthParam `json:"bearer,omitempty"`
	Basic  []AuthParam `json:"basic,omitempty"`
	APIKey []AuthParam `json:"apikey,omitempty"`
}

// AuthParam is a key-value pair used in auth configurations.
type AuthParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type,omitempty"`
}

// Variable is a Postman collection or path variable.
type Variable struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
}

// Event holds a Postman script event (pre-request or test).
type Event struct {
	Listen string `json:"listen"` // prerequest | test
	Script Script `json:"script"`
}

// Script is the code inside an Event.
type Script struct {
	Type string   `json:"type"` // text/javascript
	Exec []string `json:"exec"`
}

// Response is a saved/example response in a request item.
type Response struct {
	Name   string `json:"name"`
	Status string `json:"status,omitempty"`
	Code   int    `json:"code,omitempty"`
	Body   string `json:"body,omitempty"`
}

// Environment is a Postman environment object.
type Environment struct {
	Name   string             `json:"name"`
	Values []EnvironmentValue `json:"values"`
}

// EnvironmentValue is a single key-value pair in a Postman environment.
type EnvironmentValue struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Enabled bool   `json:"enabled"`
	Type    string `json:"type,omitempty"` // default | secret
}

// CountEndpoints returns the total number of leaf request items (recursive).
func (c *Collection) CountEndpoints() int {
	return countItems(c.Item)
}

func countItems(items []Item) int {
	count := 0
	for _, item := range items {
		if item.Request != nil {
			count++
		}
		if len(item.Item) > 0 {
			count += countItems(item.Item)
		}
	}
	return count
}
