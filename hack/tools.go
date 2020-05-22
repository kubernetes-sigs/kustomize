// +build tools

// This package imports things required by build scripts, to force `go mod` to see them as dependencies
// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
package hack

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
)
