// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package runfn

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"sigs.k8s.io/kustomize/kyaml/copyutil"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/container"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	ValueReplacerYAMLData = `apiVersion: v1
kind: ValueReplacer
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: gcr.io/example.com/image:version
    config.kubernetes.io/local-config: "true"
stringMatch: Deployment
replace: StatefulSet
`
)

func currentUser() (*user.User, error) {
	return &user.User{
		Uid: "1",
		Gid: "2",
	}, nil
}

func TestRunFns_init(t *testing.T) {
	instance := RunFns{}
	instance.init()
	if !assert.Equal(t, instance.Input, os.Stdin) {
		t.FailNow()
	}
	if !assert.Equal(t, instance.Output, os.Stdout) {
		t.FailNow()
	}

	api, err := yaml.Parse(`apiVersion: apps/v1
kind: 
`)
	spec := runtimeutil.FunctionSpec{
		Container: runtimeutil.ContainerSpec{
			Image: "example.com:version",
		},
	}
	if !assert.NoError(t, err) {
		return
	}
	filter, _ := instance.functionFilterProvider(spec, api, currentUser)
	c := container.NewContainer(runtimeutil.ContainerSpec{Image: "example.com:version"}, "nobody")
	cf := &c
	cf.Exec.FunctionConfig = api
	assert.Equal(t, cf, filter)
}

func TestRunFns_initAsCurrentUser(t *testing.T) {
	instance := RunFns{
		AsCurrentUser: true,
	}
	instance.init()
	if !assert.Equal(t, instance.Input, os.Stdin) {
		t.FailNow()
	}
	if !assert.Equal(t, instance.Output, os.Stdout) {
		t.FailNow()
	}

	api, err := yaml.Parse(`apiVersion: apps/v1
kind: 
`)
	spec := runtimeutil.FunctionSpec{
		Container: runtimeutil.ContainerSpec{
			Image: "example.com:version",
		},
	}
	if !assert.NoError(t, err) {
		return
	}
	filter, _ := instance.functionFilterProvider(spec, api, currentUser)
	c := container.NewContainer(runtimeutil.ContainerSpec{Image: "example.com:version"}, "1:2")
	cf := &c
	cf.Exec.FunctionConfig = api
	assert.Equal(t, cf, filter)
}

func TestRunFns_Execute__initGlobalScope(t *testing.T) {
	instance := RunFns{GlobalScope: true}
	instance.init()
	if !assert.Equal(t, instance.Input, os.Stdin) {
		t.FailNow()
	}
	if !assert.Equal(t, instance.Output, os.Stdout) {
		t.FailNow()
	}
	api, err := yaml.Parse(`apiVersion: apps/v1
kind: 
`)
	if !assert.NoError(t, err) {
		return
	}

	spec := runtimeutil.FunctionSpec{
		Container: runtimeutil.ContainerSpec{
			Image: "example.com:version",
		},
	}
	if !assert.NoError(t, err) {
		return
	}
	filter, _ := instance.functionFilterProvider(spec, api, currentUser)
	c := container.NewContainer(runtimeutil.ContainerSpec{Image: "example.com:version"}, "nobody")
	cf := &c
	cf.Exec.FunctionConfig = api
	cf.Exec.GlobalScope = true
	assert.Equal(t, cf, filter)
}

func TestRunFns_Execute__initDefault(t *testing.T) {
	b := &bytes.Buffer{}
	var tests = []struct {
		instance RunFns
		expected RunFns
		name     string
	}{
		{
			instance: RunFns{},
			name:     "empty",
			expected: RunFns{Output: os.Stdout, Input: os.Stdin, NoFunctionsFromInput: getFalse()},
		},
		{
			name:     "explicit output",
			instance: RunFns{Output: b},
			expected: RunFns{Output: b, Input: os.Stdin, NoFunctionsFromInput: getFalse()},
		},
		{
			name:     "explicit input",
			instance: RunFns{Input: b},
			expected: RunFns{Output: os.Stdout, Input: b, NoFunctionsFromInput: getFalse()},
		},
		{
			name:     "explicit functions -- no functions from input",
			instance: RunFns{Functions: []*yaml.RNode{{}}},
			expected: RunFns{Output: os.Stdout, Input: os.Stdin, NoFunctionsFromInput: getTrue(), Functions: []*yaml.RNode{{}}},
		},
		{
			name:     "explicit functions -- yes functions from input",
			instance: RunFns{Functions: []*yaml.RNode{{}}, NoFunctionsFromInput: getFalse()},
			expected: RunFns{Output: os.Stdout, Input: os.Stdin, NoFunctionsFromInput: getFalse(), Functions: []*yaml.RNode{{}}},
		},
		{
			name:     "explicit functions in paths -- no functions from input",
			instance: RunFns{FunctionPaths: []string{"foo"}},
			expected: RunFns{
				Output:               os.Stdout,
				Input:                os.Stdin,
				NoFunctionsFromInput: getTrue(),
				FunctionPaths:        []string{"foo"},
			},
		},
		{
			name:     "functions in paths -- yes functions from input",
			instance: RunFns{FunctionPaths: []string{"foo"}, NoFunctionsFromInput: getFalse()},
			expected: RunFns{
				Output:               os.Stdout,
				Input:                os.Stdin,
				NoFunctionsFromInput: getFalse(),
				FunctionPaths:        []string{"foo"},
			},
		},
		{
			name:     "explicit directories in mounts",
			instance: RunFns{StorageMounts: []runtimeutil.StorageMount{{MountType: "volume", Src: "myvol", DstPath: "/local/"}}},
			expected: RunFns{
				Output:               os.Stdout,
				Input:                os.Stdin,
				NoFunctionsFromInput: getFalse(),
				StorageMounts:        []runtimeutil.StorageMount{{MountType: "volume", Src: "myvol", DstPath: "/local/"}},
			},
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			(&tt.instance).init()
			(&tt.instance).functionFilterProvider = nil
			if !assert.Equal(t, tt.expected, tt.instance) {
				t.FailNow()
			}
		})
	}
}

func getTrue() *bool {
	t := true
	return &t
}

func getFalse() *bool {
	f := false
	return &f
}

// TestRunFns_getFilters tests how filters are found and sorted
func TestRunFns_getFilters(t *testing.T) {
	type f struct {
		// path to function file and string value to write
		path, value string
		// if true, create the function in a separate directory from
		// the config, and provide it through FunctionPaths
		outOfPackage bool

		// if true, create the function as an explicit Functions input
		explicitFunction bool

		// if true and outOfPackage is true, create a new directory
		// for this function separate from the previous one.  If
		// false and outOfPackage is true, create the function in
		// the directory created for the last outOfPackage function.
		newFnPath bool
	}
	var tests = []struct {
		// function files to write
		in []f
		// images to be run in a specific order
		out []string

		// images to be run in a specific order -- computed from directory path
		outFn func(string) []string

		// expected Error
		error string

		// name of the test
		name string
		// value to set for NoFunctionsFromInput
		noFunctionsFromInput *bool

		enableStarlark bool

		disableContainers bool
	}{
		// Test
		//
		//
		{name: "single implicit function",
			in: []f{
				{
					path: filepath.Join("foo", "bar.yaml"),
					value: `
apiVersion: example.com/v1alpha1
kind: ExampleFunction
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: gcr.io/example.com/image:v1.0.0
    config.kubernetes.io/local-config: "true"
`,
				},
			},
			out: []string{"gcr.io/example.com/image:v1.0.0"},
		},

		{name: "no function spec",
			in: []f{
				{
					explicitFunction: true,
					value: `
foo: bar
`,
				},
			},
		},

		// Test
		//
		//
		{name: "defer_failure",
			in: []f{
				{
					path: filepath.Join("foo", "bar.yaml"),
					value: `
apiVersion: example.com/v1alpha1
kind: ExampleFunction
metadata:
  annotations:
    config.kubernetes.io/function: |
      deferFailure: true
      container:
        image: gcr.io/example.com/image:v1.0.0
    config.kubernetes.io/local-config: "true"
`,
				},
			},
			out: []string{"gcr.io/example.com/image:v1.0.0 deferFailure: true"},
		},

		{name: "disable containers",
			in: []f{
				{
					path: filepath.Join("foo", "bar.yaml"),
					value: `
apiVersion: example.com/v1alpha1
kind: ExampleFunction
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: gcr.io/example.com/image:v1.0.0
    config.kubernetes.io/local-config: "true"
`,
				},
			},
			out:               nil,
			disableContainers: true,
		},

		// Test
		//
		//
		{name: "sort functions -- deepest first",
			in: []f{
				{
					path: filepath.Join("a.yaml"),
					value: `
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: a
`,
				},
				{
					path: filepath.Join("foo", "b.yaml"),
					value: `
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: b
`,
				},
			},
			out: []string{"b", "a"},
		},

		// Test
		//
		//
		{name: "sort functions -- skip implicit with output of package",
			in: []f{
				{
					path:         filepath.Join("foo", "a.yaml"),
					outOfPackage: true, // out of package is run last
					value: `
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: a
`,
				},
				{
					path: filepath.Join("b.yaml"),
					value: `
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: b
`,
				},
			},
			out: []string{"a"},
		},

		// Test
		//
		//
		{name: "sort functions -- skip implicit",
			noFunctionsFromInput: getTrue(),
			in: []f{
				{
					path: filepath.Join("foo", "a.yaml"),
					value: `
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: a
`,
				},
				{
					path: filepath.Join("b.yaml"),
					value: `
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: b
`,
				},
			},
			out: nil,
		},

		// Test
		//
		//
		{name: "sort functions -- include implicit",
			noFunctionsFromInput: getFalse(),
			in: []f{
				{
					path: filepath.Join("foo", "a.yaml"),
					value: `
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: a
`,
				},
				{
					path: filepath.Join("b.yaml"),
					value: `
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: b
`,
				},
			},
			out: []string{"a", "b"},
		},

		// Test
		//
		//
		{name: "sort functions -- implicit first",
			noFunctionsFromInput: getFalse(),
			in: []f{
				{
					path:         filepath.Join("foo", "a.yaml"),
					outOfPackage: true, // out of package is run last
					value: `
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: a
`,
				},
				{
					path: filepath.Join("b.yaml"),
					value: `
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: b
`,
				},
			},
			out: []string{"b", "a"},
		},

		// Test
		//
		//
		{name: "explicit functions",
			in: []f{
				{
					explicitFunction: true,
					value: `
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: c
`,
				},
				{
					path: filepath.Join("b.yaml"),
					value: `
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: b
`,
				},
			},
			out: []string{"c"},
		},

		// Test
		//
		//
		{name: "sort functions -- implicit first",
			noFunctionsFromInput: getFalse(),
			in: []f{
				{
					explicitFunction: true,
					value: `
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: c
`,
				},
				{
					path:         filepath.Join("foo", "a.yaml"),
					outOfPackage: true, // out of package is run last
					value: `
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: a
`,
				},
				{
					path: filepath.Join("b.yaml"),
					value: `
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: b
`,
				},
			},
			out: []string{"b", "a", "c"},
		},

		// Test
		//
		//
		{name: "starlark-function",
			in: []f{
				{
					path: filepath.Join("foo", "bar.yaml"),
					value: `
apiVersion: example.com/v1alpha1
kind: ExampleFunction
metadata:
  annotations:
    config.kubernetes.io/function: |
      starlark:
        path: a/b/c
`,
				},
			},
			enableStarlark: true,
			outFn: func(path string) []string {
				return []string{
					fmt.Sprintf("name:  path: %s/foo/a/b/c url:  program:", filepath.ToSlash(path))}
			},
		},

		// Test
		//
		//
		{name: "starlark-function-absolute",
			in: []f{
				{
					path: filepath.Join("foo", "bar.yaml"),
					value: `
apiVersion: example.com/v1alpha1
kind: ExampleFunction
metadata:
  annotations:
    config.kubernetes.io/function: |
      starlark:
        path: /a/b/c
`,
				},
			},
			enableStarlark: true,
			error:          "absolute function path /a/b/c not allowed",
		},

		// Test
		//
		//
		{name: "starlark-function-escape-parent",
			in: []f{
				{
					path: filepath.Join("foo", "bar.yaml"),
					value: `
apiVersion: example.com/v1alpha1
kind: ExampleFunction
metadata:
  annotations:
    config.kubernetes.io/function: |
      starlark:
        path: ../a/b/c
`,
				},
			},
			enableStarlark: true,
			error:          "function path ../a/b/c not allowed to start with ../",
		},

		{name: "starlark-function-disabled",
			in: []f{
				{
					path: filepath.Join("foo", "bar.yaml"),
					value: `
apiVersion: example.com/v1alpha1
kind: ExampleFunction
metadata:
  annotations:
    config.kubernetes.io/function: |
      starlark:
        path: a/b/c
`,
				},
			},
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			// setup the test directory
			d := setupTest(t)
			defer os.RemoveAll(d)

			// write the functions to files
			var fnPaths []string
			var parsedFns []*yaml.RNode
			var fnPath string
			var err error
			for _, f := range tt.in {
				// get the location for the file
				var dir string
				switch {
				case f.outOfPackage:
					// if out of package, write to a separate temp directory
					if f.newFnPath || fnPath == "" {
						// create a new fn directory
						fnPath, err = ioutil.TempDir("", "kustomize-test")
						if !assert.NoError(t, err) {
							t.FailNow()
						}
						defer os.RemoveAll(fnPath)
						fnPaths = append(fnPaths, fnPath)
					}
					dir = fnPath
				case f.explicitFunction:
					parsedFns = append(parsedFns, yaml.MustParse(f.value))
				default:
					// if in package, write to the dir containing the configs
					dir = d
				}

				if !f.explicitFunction {
					// create the parent dir and write the file
					err = os.MkdirAll(filepath.Join(dir, filepath.Dir(f.path)), 0700)
					if !assert.NoError(t, err) {
						t.FailNow()
					}
					err := ioutil.WriteFile(filepath.Join(dir, f.path), []byte(f.value), 0600)
					if !assert.NoError(t, err) {
						t.FailNow()
					}
				}
			}

			// init the instance
			r := &RunFns{
				EnableStarlark:       tt.enableStarlark,
				DisableContainers:    tt.disableContainers,
				FunctionPaths:        fnPaths,
				Functions:            parsedFns,
				Path:                 d,
				NoFunctionsFromInput: tt.noFunctionsFromInput,
			}
			r.init()

			// get the filters which would be run
			var results []string
			_, fltrs, _, err := r.getNodesAndFilters()

			if tt.error != "" {
				if !assert.EqualError(t, err, tt.error) {
					t.FailNow()
				}
				return
			}

			if !assert.NoError(t, err) {
				t.FailNow()
			}
			for _, f := range fltrs {
				results = append(results, strings.TrimSpace(fmt.Sprintf("%v", f)))
			}

			// compare the actual ordering to the expected ordering
			if tt.outFn != nil {
				if !assert.Equal(t, tt.outFn(d), results) {
					t.FailNow()
				}
			} else {
				if !assert.Equal(t, tt.out, results) {
					t.FailNow()
				}
			}
		})
	}
}

func TestRunFns_network(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		network       bool
		expectNetwork bool
		error         string
	}{
		{
			name: "imperative false, declarative false",
			input: `
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: a
        network: false
`,
			network:       false,
			expectNetwork: false,
		},
		{
			name: "imperative true, declarative false",
			input: `
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: a
        network: false
`,
			network:       true,
			expectNetwork: false,
		},
		{
			name: "imperative true, declarative true",
			input: `
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: a
        network: true
`,
			network:       true,
			expectNetwork: true,
		},
		{
			name: "imperative false, declarative true",
			input: `
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: a
        network: true
`,
			network: false,
			error:   "network required but not enabled with --network",
		},
	}

	for i := range tests {
		tt := tests[i]
		fn := yaml.MustParse(tt.input)
		t.Run(tt.name, func(t *testing.T) {
			// init the instance
			r := &RunFns{
				Functions: []*yaml.RNode{fn},
				Network:   tt.network,
			}
			r.init()

			_, fltrs, _, err := r.getNodesAndFilters()
			if tt.error != "" {
				if !assert.EqualError(t, err, tt.error) {
					t.FailNow()
				}
				return
			}
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			fltr := fltrs[0].(*container.Filter)
			if !assert.Equal(t, tt.expectNetwork, fltr.Network) {
				t.FailNow()
			}
		})
	}
}

func TestCmd_Execute(t *testing.T) {
	dir := setupTest(t)
	defer os.RemoveAll(dir)

	// write a test filter to the directory of configuration
	if !assert.NoError(t, ioutil.WriteFile(
		filepath.Join(dir, "filter.yaml"), []byte(ValueReplacerYAMLData), 0600)) {
		return
	}

	instance := RunFns{Path: dir, functionFilterProvider: getFilterProvider(t)}
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

type TestFilter struct {
	invoked bool
	Exit    error
}

func (f *TestFilter) Filter(input []*yaml.RNode) ([]*yaml.RNode, error) {
	f.invoked = true
	return input, nil
}

func (f *TestFilter) GetExit() error {
	return f.Exit
}

func TestCmd_Execute_deferFailure(t *testing.T) {
	dir := setupTest(t)
	defer os.RemoveAll(dir)

	// write a test filter to the directory of configuration
	if !assert.NoError(t, ioutil.WriteFile(
		filepath.Join(dir, "filter1.yaml"), []byte(`apiVersion: v1
kind: ValueReplacer
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: 1
    config.kubernetes.io/local-config: "true"
stringMatch: Deployment
replace: StatefulSet
`), 0600)) {
		t.FailNow()
	}

	// write a test filter to the directory of configuration
	if !assert.NoError(t, ioutil.WriteFile(
		filepath.Join(dir, "filter2.yaml"), []byte(`apiVersion: v1
kind: ValueReplacer
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: 2
    config.kubernetes.io/local-config: "true"
stringMatch: Deployment
replace: StatefulSet
`), 0600)) {
		t.FailNow()
	}

	var fltrs []*TestFilter
	instance := RunFns{
		Path: dir,
		functionFilterProvider: func(f runtimeutil.FunctionSpec, node *yaml.RNode, currentUser currentUserFunc) (kio.Filter, error) {
			tf := &TestFilter{
				Exit: errors.Errorf("message: %s", f.Container.Image),
			}
			fltrs = append(fltrs, tf)
			return tf, nil
		},
	}
	instance.init()

	err := instance.Execute()

	// make sure all filters were run
	if !assert.Equal(t, 2, len(fltrs)) {
		t.FailNow()
	}
	for i := range fltrs {
		if !assert.True(t, fltrs[i].invoked) {
			t.FailNow()
		}
	}

	if !assert.EqualError(t, err, "message: 1\n---\nmessage: 2") {
		t.FailNow()
	}
	b, err := ioutil.ReadFile(
		filepath.Join(dir, "java", "java-deployment.resource.yaml"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	// files weren't changed because there was an error
	assert.Contains(t, string(b), "kind: Deployment")
}

// TestCmd_Execute_setOutput tests the execution of a filter reading and writing to a dir
func TestCmd_Execute_setFunctionPaths(t *testing.T) {
	dir := setupTest(t)
	defer os.RemoveAll(dir)

	// write a test filter to a separate directory
	tmpF, err := ioutil.TempFile("", "filter*.yaml")
	if !assert.NoError(t, err) {
		return
	}
	os.RemoveAll(tmpF.Name())
	if !assert.NoError(t, ioutil.WriteFile(tmpF.Name(), []byte(ValueReplacerYAMLData), 0600)) {
		return
	}

	// run the functions, providing the path to the directory of filters
	instance := RunFns{
		FunctionPaths:          []string{tmpF.Name()},
		Path:                   dir,
		functionFilterProvider: getFilterProvider(t),
	}
	// initialize the defaults
	instance.init()

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

// TestCmd_Execute_setOutput tests the execution of a filter using an io.Writer as output
func TestCmd_Execute_setOutput(t *testing.T) {
	dir := setupTest(t)
	defer os.RemoveAll(dir)

	// write a test filter
	if !assert.NoError(t, ioutil.WriteFile(
		filepath.Join(dir, "filter.yaml"), []byte(ValueReplacerYAMLData), 0600)) {
		return
	}

	out := &bytes.Buffer{}
	instance := RunFns{
		Output:                 out, // write to out
		Path:                   dir,
		functionFilterProvider: getFilterProvider(t),
	}
	// initialize the defaults
	instance.init()

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

// TestCmd_Execute_setInput tests the execution of a filter using an io.Reader as input
func TestCmd_Execute_setInput(t *testing.T) {
	dir := setupTest(t)
	defer os.RemoveAll(dir)
	if !assert.NoError(t, ioutil.WriteFile(
		filepath.Join(dir, "filter.yaml"), []byte(ValueReplacerYAMLData), 0600)) {
		return
	}

	read, err := kio.LocalPackageReader{PackagePath: dir}.Read()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	input := &bytes.Buffer{}
	if !assert.NoError(t, kio.ByteWriter{Writer: input}.Write(read)) {
		t.FailNow()
	}

	outDir, err := ioutil.TempDir("", "kustomize-test")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.NoError(t, ioutil.WriteFile(
		filepath.Join(dir, "filter.yaml"), []byte(ValueReplacerYAMLData), 0600)) {
		return
	}

	instance := RunFns{
		Input:                  input, // read from input
		Path:                   outDir,
		functionFilterProvider: getFilterProvider(t),
	}
	// initialize the defaults
	instance.init()

	if !assert.NoError(t, instance.Execute()) {
		return
	}
	b, err := ioutil.ReadFile(
		filepath.Join(outDir, "java", "java-deployment.resource.yaml"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Contains(t, string(b), "kind: StatefulSet")
}

// TestCmd_Execute_enableLogSteps tests the execution of a filter with LogSteps enabled.
func TestCmd_Execute_enableLogSteps(t *testing.T) {
	dir := setupTest(t)
	defer os.RemoveAll(dir)

	// write a test filter to the directory of configuration
	if !assert.NoError(t, ioutil.WriteFile(
		filepath.Join(dir, "filter.yaml"), []byte(ValueReplacerYAMLData), 0600)) {
		return
	}

	logs := &bytes.Buffer{}
	instance := RunFns{
		Path:                   dir,
		functionFilterProvider: getFilterProvider(t),
		LogSteps:               true,
		LogWriter:              logs,
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
	assert.Equal(t, "Running unknown-type function\n", logs.String())
}

// setupTest initializes a temp test directory containing test data
func setupTest(t *testing.T) string {
	dir, err := ioutil.TempDir("", "kustomize-kyaml-test")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

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
		t.FailNow()
	}
	return dir
}

// getFilterProvider fakes the creation of a filter, replacing the ContainerFiler with
// a filter to s/kind: Deployment/kind: StatefulSet/g.
// this can be used to simulate running a filter.
func getFilterProvider(t *testing.T) func(runtimeutil.FunctionSpec, *yaml.RNode, currentUserFunc) (kio.Filter, error) {
	return func(f runtimeutil.FunctionSpec, node *yaml.RNode, currentUser currentUserFunc) (kio.Filter, error) {
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
		}, nil
	}
}

func TestRunFns_mergeContainerEnv(t *testing.T) {
	testcases := []struct {
		name      string
		instance  RunFns
		inputEnvs []string
		expect    runtimeutil.ContainerEnv
	}{
		{
			name:     "all empty",
			instance: RunFns{},
			expect:   *runtimeutil.NewContainerEnv(),
		},
		{
			name:      "empty command line envs",
			instance:  RunFns{},
			inputEnvs: []string{"foo=bar"},
			expect:    *runtimeutil.NewContainerEnvFromStringSlice([]string{"foo=bar"}),
		},
		{
			name: "empty declarative envs",
			instance: RunFns{
				Env: []string{"foo=bar"},
			},
			expect: *runtimeutil.NewContainerEnvFromStringSlice([]string{"foo=bar"}),
		},
		{
			name: "same key",
			instance: RunFns{
				Env: []string{"foo=bar", "foo"},
			},
			inputEnvs: []string{"foo=bar1", "bar"},
			expect:    *runtimeutil.NewContainerEnvFromStringSlice([]string{"foo=bar", "bar", "foo"}),
		},
		{
			name: "same exported key",
			instance: RunFns{
				Env: []string{"foo=bar", "foo"},
			},
			inputEnvs: []string{"foo1=bar1", "foo"},
			expect:    *runtimeutil.NewContainerEnvFromStringSlice([]string{"foo=bar", "foo1=bar1", "foo"}),
		},
	}

	for i := range testcases {
		tc := testcases[i]
		t.Run(tc.name, func(t *testing.T) {
			envs := tc.instance.mergeContainerEnv(tc.inputEnvs)
			assert.Equal(t, tc.expect.GetDockerFlags(), runtimeutil.NewContainerEnvFromStringSlice(envs).GetDockerFlags())
		})
	}
}
