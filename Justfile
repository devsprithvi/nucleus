# Justfile at Root
# Variables

tools_path := ".config/tools"

# 1. Install all tools defined in .config/tools/go.mod
install-tools:
    @echo "Installing tools from {{ tools_path }}..."
    cd {{ tools_path }} && go install github.com/evilmartians/lefthook
    cd {{ tools_path }} && go install github.com/zricethezav/gitleaks/v8
    cd {{ tools_path }} && go install github.com/golangci/golangci-lint/cmd/golangci-lint
    @echo "Tools installed successfully!"

# 2. Setup the Hooks (runs lefthook install)
setup: install-tools
    lefthook install
