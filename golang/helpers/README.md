## setup instructions

### Go (requires Go 1.21+)

```bash
# Install tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install github.com/gordonklaus/ineffassign@latest
go install mvdan.cc/unparam@latest
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest

# Run the scanner
go run scan_code.go /path/to/your/code
```

**Optional: Create `.golangci.yml`:**

```yaml
linters:
  enable-all: true
  disable:
    - gomnd
    - testpackage

linters-settings:
  gocyclo:
    min-complexity: 15
  gosec:
    excludes:
      - G101  # Hardcoded credentials (if you have test data)

run:
  timeout: 5m
  tests: false
```
