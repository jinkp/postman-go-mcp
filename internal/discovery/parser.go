package discovery

// Parser is the port for reading an API spec from any source.
// Implementations must handle file paths and HTTP/HTTPS URLs.
type Parser interface {
	// Parse reads the spec at source and returns a normalized APISpec.
	// source can be a local file path or an HTTP/HTTPS URL.
	Parse(source string) (*APISpec, error)
}
