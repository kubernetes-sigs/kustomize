// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krmfunction

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTransformerCode() []byte {
	// a simple namespace transformer
	return []byte(`
package main

import (
  "fmt"

  "sigs.k8s.io/kustomize/api/resmap"
  "sigs.k8s.io/yaml"
  "sigs.k8s.io/kustomize/api/filters/namespace"
  "sigs.k8s.io/kustomize/kyaml/filtersutil"
  "sigs.k8s.io/kustomize/api/types"
)

type plugin struct{
  Namespace string ` + "`json:\"namespace,omitempty\" yaml:\"namespace,omitempty\"`" + `
  FieldSpecs       []types.FieldSpec ` + "`json:\"fieldSpecs,omitempty\" yaml:\"fieldSpecs,omitempty\"`" + `
}

var KustomizePlugin plugin //nolint:gochecknoglobals

func (p *plugin) Config(
  _ *resmap.PluginHelpers, config []byte) (err error) {
  return yaml.Unmarshal(config, p)
}

func (p *plugin) Transform(rm resmap.ResMap) error {
  if len(p.Namespace) == 0 {
    return nil
  }
  for _, r := range rm.Resources() {
    if r.IsNilOrEmpty() {
      // Don't mutate empty objects?
      continue
    }
    err := filtersutil.ApplyToJSON(namespace.Filter{
      Namespace: p.Namespace,
      FsSlice:   p.FieldSpecs,
    }, r)
    if err != nil {
      return err
    }
    matches := rm.GetMatchingResourcesByCurrentId(r.CurId().Equals)
    if len(matches) != 1 {
      return fmt.Errorf(
        "namespace transformation produces ID conflict: %+v", matches)
    }
  }
  return nil
}
`)
}

func getTransformerInputResource() []byte {
	return []byte(`
apiVersion: config.kubernetes.io/v1
kind: ResourceList
functionConfig:
  apiVersion: foo-corp.com/v1
  kind: FulfillmentCenter
  metadata:
    name: staging
    annotations:
      config.kubernetes.io/function: |
        container:
          image: gcr.io/example/foo:v1.0.0
  data:
    namespace: foo
    fieldSpecs:
    - path: metadata/namespace
      create: true
items:
  - apiVersion: apps/v1
    kind: foobar
    metadata:
      name: whatever
`)
}

func runKrmFunction(t *testing.T, input []byte, dir string) []byte {
	t.Helper()
	prepareCmd := func(name string, arg ...string) (*exec.Cmd, *bytes.Buffer, *bytes.Buffer) {
		ob := bytes.NewBuffer([]byte{})
		eb := bytes.NewBuffer([]byte{})
		cmd := exec.Command(name, arg...)
		cmd.Stdout = ob
		cmd.Stderr = eb
		cmd.Dir = dir
		return cmd, ob, eb
	}
	cmd, ob, eb := prepareCmd("go", "mod", "tidy")
	require.NoErrorf(t, cmd.Run(), "Stdout:\n%s\nStderr:\n%s\n", ob.String(), eb.String())

	cmd, ob, eb = prepareCmd("go", "run", ".")
	cmd.Stdin = bytes.NewReader(input)
	require.NoErrorf(t, cmd.Run(), "Stdout:\n%s\nStderr:\n%s\n", ob.String(), eb.String())

	return ob.Bytes()
}

func TestTransformerConverter(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "Plugin.go"),
		getTransformerCode(), 0644)
	require.NoError(t, err)

	c := NewConverter(filepath.Join(dir, "output"),
		filepath.Join(dir, "Plugin.go"))

	err = c.Convert()
	assert.NoError(t, err)

	output := runKrmFunction(t, getTransformerInputResource(), filepath.Join(dir, "output"))
	assert.Equal(t, `apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: foobar
  metadata:
    name: whatever
    namespace: foo
functionConfig:
  apiVersion: foo-corp.com/v1
  kind: FulfillmentCenter
  metadata:
    name: staging
    annotations:
      config.kubernetes.io/function: |
        container:
          image: gcr.io/example/foo:v1.0.0
  data:
    namespace: foo
    fieldSpecs:
    - path: metadata/namespace
      create: true
`, string(output))
}

func getGeneratorCode() []byte {
	return []byte(`package main

import (
	"sigs.k8s.io/kustomize/api/kv"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

type plugin struct {
	h                *resmap.PluginHelpers
	types.ObjectMeta ` + "`json:\"metadata,omitempty\" yaml:\"metadata,omitempty\"`" + `
	types.ConfigMapArgs
}

var KustomizePlugin plugin //nolint:gochecknoglobals

func (p *plugin) Config(h *resmap.PluginHelpers, config []byte) (err error) {
	p.ConfigMapArgs = types.ConfigMapArgs{}
	err = yaml.Unmarshal(config, p)
	if p.ConfigMapArgs.Name == "" {
		p.ConfigMapArgs.Name = p.Name
	}
	if p.ConfigMapArgs.Namespace == "" {
		p.ConfigMapArgs.Namespace = p.Namespace
	}
	p.h = h
	return
}

func (p *plugin) Generate() (resmap.ResMap, error) {
	return p.h.ResmapFactory().FromConfigMapArgs(
		kv.NewLoader(p.h.Loader(), p.h.Validator()), p.ConfigMapArgs)
}`)
}

func getGeneratorInputResource() []byte {
	return []byte(`
apiVersion: config.kubernetes.io/v1
kind: ResourceList
functionConfig:
  apiVersion: foo-corp.com/v1
  kind: FulfillmentCenter
  metadata:
    name: staging
    annotations:
      config.kubernetes.io/function: |
        container:
          image: gcr.io/example/foo:v1.0.0
  data:
    metadata:
      name: staging
items: []
`)
}

func TestGeneratorConverter(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "Plugin.go"),
		getGeneratorCode(), 0644)
	require.NoError(t, err)

	c := NewConverter(filepath.Join(dir, "output"),
		filepath.Join(dir, "Plugin.go"))

	err = c.Convert()
	assert.NoError(t, err)
	output := runKrmFunction(t, getGeneratorInputResource(), filepath.Join(dir, "output"))
	assert.Equal(t, `apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: staging
    annotations:
      internal.config.kubernetes.io/generatorBehavior: unspecified
      internal.config.kubernetes.io/needsHashSuffix: enabled
functionConfig:
  apiVersion: foo-corp.com/v1
  kind: FulfillmentCenter
  metadata:
    name: staging
    annotations:
      config.kubernetes.io/function: |
        container:
          image: gcr.io/example/foo:v1.0.0
  data:
    metadata:
      name: staging
`, string(output))
}
