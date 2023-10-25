// Copyright 2023 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build_test

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	. "sigs.k8s.io/kustomize/kustomize/v5/commands/build"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// GenConfig configures the generation of a kustomization tree for benchmarking purposes
type GenConfig struct {
	// The number of plain file resources to generate
	fileResources int

	// The number of subdirectories to generate
	resources int

	// The number of patches to generate
	patches int

	// Whether to generate a namespace field
	namespaced bool

	// The name prefix to use (if any)
	namePrefix string

	// The name suffix to use (if any)
	nameSuffix string

	// Common labels to use (if any)
	labels map[string]string

	// Common annotations to use (if any)
	commonAnnotations map[string]string
}

// makeKustomization generates a kustomization tree with the given configuration (number of resources, patches, etc.)
// and writes it to the given filesystem at the given path. The depth is used to determine the current configuration
// as an index into the given configs slice.
//
// The function is recursive and will call itself for config as long as resources > 0.
func makeKustomization(configs []GenConfig, fSys filesys.FileSystem, path, id string, depth int) error {
	cfg := configs[depth]
	if err := fSys.MkdirAll(path); err != nil {
		return fmt.Errorf("failed to make directory %v: %w", path, err)
	}

	var buf bytes.Buffer
	if cfg.namespaced {
		fmt.Fprintf(&buf, "namespace: %s\n", id)
	}

	if cfg.namePrefix != "" {
		fmt.Fprintf(&buf, "namePrefix: %s\n", cfg.namePrefix)
	}

	if cfg.nameSuffix != "" {
		fmt.Fprintf(&buf, "nameSuffix: %s\n", cfg.nameSuffix)
	}

	if len(cfg.labels) > 0 {
		fmt.Fprintf(&buf, "labels:\n- includeSelectors: true\n  pairs:\n")
		for k, v := range cfg.labels {
			fmt.Fprintf(&buf, "    %s: %s\n", k, v)
		}
	}

	if len(cfg.commonAnnotations) > 0 {
		fmt.Fprintf(&buf, "commonAnnotations:\n")
		for k, v := range cfg.commonAnnotations {
			fmt.Fprintf(&buf, "  %s: %s\n", k, v)
		}
	}

	if cfg.fileResources > 0 || cfg.resources > 0 {
		fmt.Fprintf(&buf, "resources:\n")
		for res := 0; res < cfg.fileResources; res++ {
			fn := fmt.Sprintf("res%d.yaml", res)
			fmt.Fprintf(&buf, " - %v\n", fn)

			cm := fmt.Sprintf(`kind: ConfigMap
apiVersion: v1
metadata:
  name: %s-%d
  labels:
    foo: bar
  annotations:
    baz: blatti
data:
  k: v
`, id, res)
			if err := fSys.WriteFile(filepath.Join(path, fn), []byte(cm)); err != nil {
				return fmt.Errorf("failed to write file resource: %w", err)
			}
		}

		for res := 0; res < cfg.resources; res++ {
			fn := fmt.Sprintf("res%d", res)
			fmt.Fprintf(&buf, " - %v\n", fn)
			if err := makeKustomization(configs, fSys, path+"/"+fn, fmt.Sprintf("%s-%d", id, res), depth+1); err != nil {
				return fmt.Errorf("failed to make kustomization: %w", err)
			}
		}
	}

	for res := 0; res < cfg.patches; res++ {
		if res == 0 {
			fmt.Fprintf(&buf, "patches:\n")
		}

		// alternate between json and yaml patches to test both kinds
		if res%2 == 0 {
			fn := fmt.Sprintf("patch%d.yaml", res)
			fmt.Fprintf(&buf, " - path: %v\n", fn)
			cmPatch := fmt.Sprintf(`kind: ConfigMap
apiVersion: v1
metadata:
  name: %s-%d
data:
  k: v2
`, id, res)
			if err := fSys.WriteFile(filepath.Join(path, fn), []byte(cmPatch)); err != nil {
				return fmt.Errorf("failed to write patch: %w", err)
			}
		} else {
			fn := fmt.Sprintf("patch%d.json", res)
			fmt.Fprintf(&buf, ` - path: %v
   target:
    version: v1
    kind: ConfigMap
    name: %s-%d
`, fn, id, res-1)
			patch := `[{"op": "add", "path": "/data/k2", "value": "3"} ]`
			if err := fSys.WriteFile(filepath.Join(path, fn), []byte(patch)); err != nil {
				return fmt.Errorf("failed to write patch: %w", err)
			}
		}
	}

	if err := fSys.WriteFile(filepath.Join(path, "kustomization.yaml"), buf.Bytes()); err != nil {
		return fmt.Errorf("failed to write kustomization.yaml: %w", err)
	}
	return nil
}

func BenchmarkBuild(b *testing.B) {
	// This benchmark generates a kustomization tree with the following structure:
	genConfig := []GenConfig{
		{
			resources: 2, // two nested resources

			// these operations should be very fast, so lets perform them on a *lot* of resources
			namePrefix: "foo-",
			nameSuffix: "-bar",
			labels: map[string]string{
				"foo": "bar",
			},
			commonAnnotations: map[string]string{
				"baz": "blatti",
			},
		},
		{
			// test some more nesting (this could be `apps/` or `components/` directory with 50 apps or components)
			resources: 50,
		},
		{
			// this should be almost the same as using 150 above and skipping it, but it is currently not, so lets have some more nesting
			resources: 3,
		},
		{
			// here we have an actual component/app with lots or resources. Typically here we set namespace and have some patches
			resources:     1,
			namespaced:    true,
			fileResources: 30,
			patches:       10,
		},
		{
			// we can also have a base/ or shared resources included into the namespace
			fileResources: 2,
		},
	}

	fSys := filesys.MakeFsInMemory()
	if err := makeKustomization(genConfig, fSys, "testdata", "res", 0); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffy := new(bytes.Buffer)
		cmd := NewCmdBuild(fSys, MakeHelp("foo", "bar"), buffy)
		if err := cmd.RunE(cmd, []string{"./testdata"}); err != nil {
			b.Fatal(err)
		}
	}
}
