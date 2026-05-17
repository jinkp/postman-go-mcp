// Package environments generates Postman environment objects for multiple stages.
package environments

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/isai-salazar-enc/postman-go-mcp/pkg/postmanfmt"
)

// EnvironmentGenerator is the port for generating Postman environments.
type EnvironmentGenerator interface {
	Generate(baseURL string, envNames []string, extras map[string]string) ([]postmanfmt.Environment, error)
}

// DefaultEnvironments are the standard environment names used when none are specified.
var DefaultEnvironments = []string{"dev", "qa", "stage", "prod"}

// Generator is the default implementation of EnvironmentGenerator.
type Generator struct{}

// NewGenerator creates a new Generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate creates a Postman environment for each name in envNames.
// For non-prod environments, it prepends the env name as a subdomain prefix.
// extras are additional key-value pairs added to all environments.
func (g *Generator) Generate(baseURL string, envNames []string, extras map[string]string) ([]postmanfmt.Environment, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("base_url is required")
	}
	if len(envNames) == 0 {
		envNames = DefaultEnvironments
	}

	var envs []postmanfmt.Environment
	for _, name := range envNames {
		derived, err := deriveURL(baseURL, name)
		if err != nil {
			return nil, fmt.Errorf("derive url for %q: %w", name, err)
		}

		values := []postmanfmt.EnvironmentValue{
			{Key: "baseUrl", Value: derived, Enabled: true},
		}
		for k, v := range extras {
			values = append(values, postmanfmt.EnvironmentValue{
				Key:     k,
				Value:   v,
				Enabled: true,
			})
		}

		envs = append(envs, postmanfmt.Environment{
			Name:   name,
			Values: values,
		})
	}
	return envs, nil
}

// deriveURL generates the environment-specific URL.
// For "prod", the original URL is returned unchanged.
// For other environments, the env name is prepended as a subdomain.
// Example: base="https://api.example.com", env="dev" → "https://dev.api.example.com"
func deriveURL(baseURL, envName string) (string, error) {
	if strings.EqualFold(envName, "prod") {
		return baseURL, nil
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}

	host := u.Hostname()
	port := u.Port()

	// Prepend the env name as subdomain
	newHost := envName + "." + host
	if port != "" {
		newHost = newHost + ":" + port
	}

	u.Host = newHost
	return u.String(), nil
}
