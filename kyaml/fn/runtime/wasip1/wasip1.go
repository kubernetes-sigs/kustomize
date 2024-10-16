package wasip1

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

var _ kio.Filter = &Filter{}

type Filter struct {
	runtimeutil.Wasip1Spec `json:",inline" yaml:",inline"`

	WorkingDir string `yaml:"workingDir,omitempty"`

	runtimeutil.FunctionFilter
}

func (c *Filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	c.FunctionFilter.Run = c.Run
	return c.FunctionFilter.Filter(nodes) //nolint:wrapcheck
}

func (c *Filter) Run(reader io.Reader, writer io.Writer) error {
	// Create a new runtime
	ctx := context.Background()
	r := wazero.NewRuntime(ctx)
	defer r.Close(ctx)

	// Import the WASI snapshot preview1 module
	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	// Load the WASI binary from the filesystem
	wasmBytes, err := os.ReadFile(filepath.Join(c.WorkingDir, c.File))
	if err != nil {
		return fmt.Errorf("failed to read wasi binary(path: '%s'): %w", c.File, err)
	}

	// Create a new module config with stdin and stdout
	config := wazero.NewModuleConfig().WithStdin(reader).WithStdout(writer)

	// It will not work unless program name (arg[0]) is set.
	module, err := r.InstantiateWithConfig(ctx, wasmBytes, config.WithArgs("krm"))
	if err != nil {
		return fmt.Errorf("failed to wasip1 instantiate module: %w", err)
	}
	defer module.Close(ctx)

	return nil
}

// NewWasmFilter returns a new wasm filter
func NewWasmFilter(spec runtimeutil.Wasip1Spec) *Filter {
	return &Filter{Wasip1Spec: spec}
}
