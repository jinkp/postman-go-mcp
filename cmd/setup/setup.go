// Package setup is the entrypoint for the `mcp-postman setup` subcommand.
package setup

import (
	"fmt"
	"os"

	"github.com/isai-salazar-enc/postman-go-mcp/internal/setup"
)

// Run launches the setup wizard. Call this from main when os.Args[1] == "setup".
func Run() {
	if err := setup.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
