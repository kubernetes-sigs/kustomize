// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package runfn

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/copyutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	ValueReplacerYAMLData = `apiVersion: v1
kind: ValueReplacer
metadata:
  configFn:
    container:
      image: gcr.io/example.com/image:version
stringMatch: Deployment
replace: StatefulSet
`
)

func TestRunFns_Execute(t *testing.T) {
	instance := RunFns{}
	instance.init()
	api, err := yaml.Parse(`apiVersion: apps/v1
kind: 
`)
	if !assert.NoError(t, err) {
		return
	}
	filter := instance.containerFilterProvider("example.com:version", "", api)
	assert.Equal(t, &filters.ContainerFilter{Image: "example.com:version", Config: api}, filter)
}

func TestCmd_Execute(t *testing.T) {
	dir, err := ioutil.TempDir("", "kustomize-kyaml-test")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer os.RemoveAll(dir)

	_, filename, _, ok := runtime.Caller(0)
	if !assert.True(t, ok) {
		t.FailNow()
	}
	ds, err := filepath.Abs(filepath.Join(filepath.Dir(filename), "test", "testdata"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.NoError(t, copyutil.CopyDir(ds, dir)) {
		t.FailNow()
	}
	if !assert.NoError(t, os.Chdir(filepath.Dir(dir))) {
		return
	}

	// write a test filter
	if !assert.NoError(t, ioutil.WriteFile(
		filepath.Join(dir, "filter.yaml"), []byte(ValueReplacerYAMLData), 0600)) {
		return
	}

	instance := RunFns{
		Path: dir,
		containerFilterProvider: func(s, _ string, node *yaml.RNode) kio.Filter {
			// parse the filter from the input
			filter := yaml.YFilter{}
			b := &bytes.Buffer{}
			e := yaml.NewEncoder(b)
			if !assert.NoError(t, e.Encode(node.YNode())) {
				t.FailNow()
			}
			e.Close()
			d := yaml.NewDecoder(b)
			if !assert.NoError(t, d.Decode(&filter)) {
				t.FailNow()
			}

			return filters.Modifier{
				Filters: []yaml.YFilter{{Filter: yaml.Lookup("kind")}, filter},
			}
		},
	}
	if !assert.NoError(t, instance.Execute()) {
		t.FailNow()
	}
	b, err := ioutil.ReadFile(
		filepath.Join(dir, "java", "java-deployment.resource.yaml"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Contains(t, string(b), "kind: StatefulSet")
}

func TestCmd_Execute_APIs(t *testing.T) {
	dir, err := ioutil.TempDir("", "kustomize-kyaml-test")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer os.RemoveAll(dir)

	_, filename, _, ok := runtime.Caller(0)
	if !assert.True(t, ok) {
		t.FailNow()
	}
	ds, err := filepath.Abs(filepath.Join(filepath.Dir(filename), "test", "testdata"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.NoError(t, copyutil.CopyDir(ds, dir)) {
		t.FailNow()
	}
	if !assert.NoError(t, os.Chdir(filepath.Dir(dir))) {
		return
	}

	// write a test filter
	tmpF, err := ioutil.TempFile("", "filter*.yaml")
	if !assert.NoError(t, err) {
		return
	}
	os.RemoveAll(tmpF.Name())
	if !assert.NoError(t, ioutil.WriteFile(tmpF.Name(), []byte(ValueReplacerYAMLData), 0600)) {
		return
	}

	instance := RunFns{
		FunctionPaths: []string{tmpF.Name()},
		Path:          dir,
		containerFilterProvider: func(s, _ string, node *yaml.RNode) kio.Filter {
			// parse the filter from the input
			filter := yaml.YFilter{}
			b := &bytes.Buffer{}
			e := yaml.NewEncoder(b)
			if !assert.NoError(t, e.Encode(node.YNode())) {
				t.FailNow()
			}
			e.Close()
			d := yaml.NewDecoder(b)
			if !assert.NoError(t, d.Decode(&filter)) {
				t.FailNow()
			}

			return filters.Modifier{
				Filters: []yaml.YFilter{{Filter: yaml.Lookup("kind")}, filter},
			}
		},
	}
	err = instance.Execute()
	if !assert.NoError(t, err) {
		return
	}
	b, err := ioutil.ReadFile(
		filepath.Join(dir, "java", "java-deployment.resource.yaml"))
	if !assert.NoError(t, err) {
		return
	}
	assert.Contains(t, string(b), "kind: StatefulSet")
}

func TestCmd_Execute_Stdout(t *testing.T) {
	dir, err := ioutil.TempDir("", "kustomize-kyaml-test")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer os.RemoveAll(dir)

	_, filename, _, ok := runtime.Caller(0)
	if !assert.True(t, ok) {
		t.FailNow()
	}
	ds, err := filepath.Abs(filepath.Join(filepath.Dir(filename), "test", "testdata"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.NoError(t, copyutil.CopyDir(ds, dir)) {
		t.FailNow()
	}
	if !assert.NoError(t, os.Chdir(filepath.Dir(dir))) {
		return
	}

	// write a test filter
	if !assert.NoError(t, ioutil.WriteFile(
		filepath.Join(dir, "filter.yaml"), []byte(ValueReplacerYAMLData), 0600)) {
		return
	}

	out := &bytes.Buffer{}
	instance := RunFns{
		Output: out,
		Path:   dir,
		containerFilterProvider: func(s, _ string, node *yaml.RNode) kio.Filter {
			// parse the filter from the input
			filter := yaml.YFilter{}
			b := &bytes.Buffer{}
			e := yaml.NewEncoder(b)
			if !assert.NoError(t, e.Encode(node.YNode())) {
				t.FailNow()
			}
			e.Close()
			d := yaml.NewDecoder(b)
			if !assert.NoError(t, d.Decode(&filter)) {
				t.FailNow()
			}

			return filters.Modifier{
				Filters: []yaml.YFilter{{Filter: yaml.Lookup("kind")}, filter},
			}
		},
	}
	if !assert.NoError(t, instance.Execute()) {
		return
	}
	b, err := ioutil.ReadFile(
		filepath.Join(dir, "java", "java-deployment.resource.yaml"))
	if !assert.NoError(t, err) {
		return
	}
	assert.NotContains(t, string(b), "kind: StatefulSet")
	assert.Contains(t, out.String(), "kind: StatefulSet")
}
