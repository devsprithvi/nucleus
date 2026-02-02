//go:build tools
// +build tools

package tools

import (
	// We import the "main" packages of the tools we want
	_ "github.com/evilmartians/lefthook"                    // The Orchestrator
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint" // Linter
	_ "github.com/zricethezav/gitleaks/v8"                  // Secret Scanner
	_ "golang.org/x/tools/cmd/goimports"                    // Formatter
)
