// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filters

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestFilter_command(t *testing.T) {
	cfg, err := yaml.Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
`)
	if !assert.NoError(t, err) {
		return
	}
	instance := &ContainerFilter{
		Image:  "example.com:version",
		Config: cfg,
	}
	os.Setenv("KYAML_TEST", "FOO")
	cmd, err := instance.getCommand()
	if !assert.NoError(t, err) {
		return
	}

	expected := []string{
		"docker", "run",
		"--rm",
		"-i", "-a", "STDIN", "-a", "STDOUT", "-a", "STDERR",
		"--network", "none",
		"--user", "nobody",
		"--security-opt=no-new-privileges",
	}
	for _, e := range os.Environ() {
		// the process env
		expected = append(expected, "-e", strings.Split(e, "=")[0])
	}
	expected = append(expected, "example.com:version")
	assert.Equal(t, expected, cmd.Args)

	foundKyaml := false
	for _, e := range cmd.Env {
		// verify the command has the right environment variables to pass to the container
		split := strings.Split(e, "=")
		if split[0] == "KYAML_TEST" {
			assert.Equal(t, "FOO", split[1])
			foundKyaml = true
		}
	}
	assert.True(t, foundKyaml)
}

func TestFilter_command_StorageMount(t *testing.T) {
	cfg, err := yaml.Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
`)
	if !assert.NoError(t, err) {
		return
	}
	bindMount := StorageMount{"bind", "/mount/path", "/local/"}
	localVol := StorageMount{"volume", "myvol", "/local/"}
	tmpfs := StorageMount{"tmpfs", "", "/local/"}
	instance := &ContainerFilter{
		Image:         "example.com:version",
		Config:        cfg,
		StorageMounts: []StorageMount{bindMount, localVol, tmpfs},
	}
	cmd, err := instance.getCommand()
	if !assert.NoError(t, err) {
		return
	}

	expected := []string{
		"docker", "run",
		"--rm",
		"-i", "-a", "STDIN", "-a", "STDOUT", "-a", "STDERR",
		"--network", "none",
		"--user", "nobody",
		"--security-opt=no-new-privileges",
		"--mount", fmt.Sprintf("type=%s,src=%s,dst=%s:ro", "bind", "/mount/path", "/local/"),
		"--mount", fmt.Sprintf("type=%s,src=%s,dst=%s:ro", "volume", "myvol", "/local/"),
		"--mount", fmt.Sprintf("type=%s,src=%s,dst=%s:ro", "tmpfs", "", "/local/"),
	}
	for _, e := range os.Environ() {
		// the process env
		expected = append(expected, "-e", strings.Split(e, "=")[0])
	}
	expected = append(expected, "example.com:version")
	assert.Equal(t, expected, cmd.Args)
}

func TestFilter_command_network(t *testing.T) {
	cfg, err := yaml.Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
`)
	if !assert.NoError(t, err) {
		return
	}
	instance := &ContainerFilter{
		Image:   "example.com:version",
		Network: "test-net",
		Config:  cfg,
	}
	cmd, err := instance.getCommand()
	if !assert.NoError(t, err) {
		return
	}

	expected := []string{
		"docker", "run",
		"--rm",
		"-i", "-a", "STDIN", "-a", "STDOUT", "-a", "STDERR",
		"--network", "test-net",
		"--user", "nobody",
		"--security-opt=no-new-privileges",
	}
	for _, e := range os.Environ() {
		// the process env
		tokens := strings.Split(e, "=")
		if tokens[0] == "" {
			continue
		}
		expected = append(expected, "-e", tokens[0])
	}
	expected = append(expected, "example.com:version")
	assert.Equal(t, expected, cmd.Args)
}

func TestFilter_Filter(t *testing.T) {
	cfg, err := yaml.Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
`)
	if !assert.NoError(t, err) {
		return
	}

	input, err := (&kio.ByteReader{Reader: bytes.NewBufferString(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
`)}).Read()
	if !assert.NoError(t, err) {
		return
	}

	called := false
	result, err := (&ContainerFilter{
		Image:  "example.com:version",
		Config: cfg,
		args:   []string{"sed", "s/Deployment/StatefulSet/g"},
		checkInput: func(s string) {
			called = true
			if !assert.Equal(t, `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
    annotations:
      config.kubernetes.io/index: '0'
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
    annotations:
      config.kubernetes.io/index: '1'
functionConfig: {apiVersion: apps/v1, kind: Deployment, metadata: {name: foo}}
`, s) {
				t.FailNow()
			}
		},
	}).Filter(input)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.True(t, called) {
		return
	}

	b := &bytes.Buffer{}
	err = kio.ByteWriter{Writer: b, KeepReaderAnnotations: true}.Write(result)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'statefulset_deployment-foo.yaml'
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'service_service-foo.yaml'
`, b.String())
}

func TestFilter_Filter_noChange(t *testing.T) {
	cfg, err := yaml.Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
`)
	if !assert.NoError(t, err) {
		return
	}

	input, err := (&kio.ByteReader{Reader: bytes.NewBufferString(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
`)}).Read()
	if !assert.NoError(t, err) {
		return
	}

	called := false
	result, err := (&ContainerFilter{
		Image:  "example.com:version",
		Config: cfg,
		args:   []string{"sh", "-c", "cat <&0"},
		checkInput: func(s string) {
			called = true
			if !assert.Equal(t, `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
    annotations:
      config.kubernetes.io/index: '0'
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
    annotations:
      config.kubernetes.io/index: '1'
functionConfig: {apiVersion: apps/v1, kind: Deployment, metadata: {name: foo}}
`, s) {
				t.FailNow()
			}
		},
	}).Filter(input)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.True(t, called) {
		return
	}

	b := &bytes.Buffer{}
	err = kio.ByteWriter{Writer: b, KeepReaderAnnotations: true}.Write(result)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, `apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'deployment_deployment-foo.yaml'
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'service_service-foo.yaml'
`, b.String())
}

func Test_GetContainerName(t *testing.T) {
	// make sure gcr.io works
	n, err := yaml.Parse(`apiVersion: v1beta1
kind: MyThing
metadata:
  configFn:
    container:
      image: gcr.io/foo/bar:something
`)
	if !assert.NoError(t, err) {
		return
	}
	c, _ := GetContainerName(n)
	assert.Equal(t, "gcr.io/foo/bar:something", c)

	// container from config.kubernetes.io/container annotation
	n, err = yaml.Parse(`apiVersion: v1
kind: MyThing
metadata:
  annotations:
    config.kubernetes.io/container: gcr.io/foo/bar:something
`)
	if !assert.NoError(t, err) {
		return
	}
	c, _ = GetContainerName(n)
	assert.Equal(t, "gcr.io/foo/bar:something", c)

	// container from config.kubernetes.io/function annotation
	n, err = yaml.Parse(`apiVersion: v1
kind: MyThing
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: gcr.io/foo/bar:something
`)
	if !assert.NoError(t, err) {
		return
	}
	c, _ = GetContainerName(n)
	assert.Equal(t, "gcr.io/foo/bar:something", c)

	// doesn't have a container
	n, err = yaml.Parse(`apiVersion: v1
kind: MyThing
metadata:
`)
	if !assert.NoError(t, err) {
		return
	}
	c, _ = GetContainerName(n)
	assert.Equal(t, "", c)
}

func TestFilter_Filter_defaultNaming(t *testing.T) {
	cfg, err := yaml.Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  annotations:
    config.kubernetes.io/path: 'foo/bar.yaml'
`)
	if !assert.NoError(t, err) {
		return
	}

	input, err := (&kio.ByteReader{Reader: bytes.NewBufferString(``)}).Read()
	if !assert.NoError(t, err) {
		return
	}

	called := false
	result, err := (&ContainerFilter{
		Image:  "example.com:version",
		Config: cfg,
		args: []string{"echo", `apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
`},
		checkInput: func(s string) {
			called = true
			if !assert.Equal(t, `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items: []
functionConfig: {apiVersion: apps/v1, kind: Deployment, metadata: {name: foo, annotations: {
      config.kubernetes.io/path: 'foo/bar.yaml'}}}
`, s) {
				t.FailNow()
			}
		},
	}).Filter(input)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.True(t, called) {
		return
	}

	b := &bytes.Buffer{}
	err = kio.ByteWriter{Writer: b, KeepReaderAnnotations: true}.Write(result)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, `apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'foo/deployment_deployment-foo.yaml'
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'foo/service_service-foo.yaml'
`, b.String())
}

func TestFilter_Filter_defaultNamingFunctions(t *testing.T) {
	cfg, err := yaml.Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  annotations:
    config.kubernetes.io/path: 'foo/functions/bar.yaml'
`)
	if !assert.NoError(t, err) {
		return
	}

	input, err := (&kio.ByteReader{Reader: bytes.NewBufferString(``)}).Read()
	if !assert.NoError(t, err) {
		return
	}

	called := false
	result, err := (&ContainerFilter{
		Image:  "example.com:version",
		Config: cfg,
		args: []string{"echo", `apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
`},
		checkInput: func(s string) {
			called = true
			if !assert.Equal(t, `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items: []
functionConfig: {apiVersion: apps/v1, kind: Deployment, metadata: {name: foo, annotations: {
      config.kubernetes.io/path: 'foo/functions/bar.yaml'}}}
`, s) {
				t.FailNow()
			}
		},
	}).Filter(input)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.True(t, called) {
		return
	}

	b := &bytes.Buffer{}
	err = kio.ByteWriter{Writer: b, KeepReaderAnnotations: true}.Write(result)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, `apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'foo/deployment_deployment-foo.yaml'
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'foo/service_service-foo.yaml'
`, b.String())
}

func TestFilter_Filter_scopeMissingFromResource(t *testing.T) {
	cfg, err := yaml.Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  annotations:
    config.kubernetes.io/path: 'foo/bar.yaml'
`)
	if !assert.NoError(t, err) {
		return
	}

	input, err := (&kio.ByteReader{Reader: bytes.NewBufferString(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
`)}).Read()
	if !assert.NoError(t, err) {
		return
	}

	// no resources match the scope
	called := false
	result, err := (&ContainerFilter{
		Image:  "example.com:version",
		Config: cfg,
		args:   []string{"sed", "s/Deployment/StatefulSet/g"},
		checkInput: func(s string) {
			called = true
			if !assert.Equal(t, `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items: []
functionConfig: {apiVersion: apps/v1, kind: Deployment, metadata: {name: foo, annotations: {
      config.kubernetes.io/path: 'foo/bar.yaml'}}}
`, s) {
				t.FailNow()
			}
		},
	}).Filter(input)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.True(t, called) {
		return
	}

	b := &bytes.Buffer{}
	err = kio.ByteWriter{Writer: b, KeepReaderAnnotations: true}.Write(result)
	if !assert.NoError(t, err) {
		return
	}

	// Resources should be preserved -- paths shouldn't be set by container
	assert.Equal(t, `apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/index: '0'
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/index: '1'
`, b.String())
}

func TestFilter_Filter_globalScope(t *testing.T) {
	cfg, err := yaml.Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  annotations:
    config.kubernetes.io/path: 'foo/bar.yaml'
`)
	if !assert.NoError(t, err) {
		return
	}

	input, err := (&kio.ByteReader{Reader: bytes.NewBufferString(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
`)}).Read()
	if !assert.NoError(t, err) {
		return
	}

	// no resources match the scope
	called := false
	result, err := (&ContainerFilter{
		GlobalScope: true,
		Image:       "example.com:version",
		Config:      cfg,
		args:        []string{"sed", "s/Deployment/StatefulSet/g"},
		checkInput: func(s string) {
			called = true
			if !assert.Equal(t, `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
    annotations:
      config.kubernetes.io/index: '0'
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
    annotations:
      config.kubernetes.io/index: '1'
functionConfig: {apiVersion: apps/v1, kind: Deployment, metadata: {name: foo, annotations: {
      config.kubernetes.io/path: 'foo/bar.yaml'}}}
`, s) {
				t.FailNow()
			}
		},
	}).Filter(input)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.True(t, called) {
		return
	}

	b := &bytes.Buffer{}
	err = kio.ByteWriter{Writer: b, KeepReaderAnnotations: true}.Write(result)
	if !assert.NoError(t, err) {
		return
	}

	// Resources should be preserved -- paths shouldn't be set by container
	assert.Equal(t, `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/index: '0'
    config.kubernetes.io/path: 'foo/statefulset_deployment-foo.yaml'
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/index: '1'
    config.kubernetes.io/path: 'foo/service_service-foo.yaml'
`, b.String())
}

func TestFilter_Filter_scopeFunctionsDir(t *testing.T) {
	// functions under "functions/" dir should be scoped to parent dir
	cfg, err := yaml.Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  annotations:
    config.kubernetes.io/path: 'foo/functions/bar.yaml'
`)
	if !assert.NoError(t, err) {
		return
	}

	input, err := (&kio.ByteReader{Reader: bytes.NewBufferString(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/d.yaml'
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/s.yaml'
`)}).Read()
	if !assert.NoError(t, err) {
		return
	}

	// no resources match the scope
	called := false
	result, err := (&ContainerFilter{
		Image:  "example.com:version",
		Config: cfg,
		args:   []string{"sed", "s/Deployment/StatefulSet/g"},
		checkInput: func(s string) {
			called = true
			if !assert.Equal(t, `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
    annotations:
      config.kubernetes.io/path: 'foo/bar/d.yaml'
      config.kubernetes.io/index: '0'
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
    annotations:
      config.kubernetes.io/path: 'foo/bar/s.yaml'
      config.kubernetes.io/index: '1'
functionConfig: {apiVersion: apps/v1, kind: Deployment, metadata: {name: foo, annotations: {
      config.kubernetes.io/path: 'foo/functions/bar.yaml'}}}
`, s) {
				t.FailNow()
			}
		},
	}).Filter(input)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.True(t, called) {
		return
	}

	b := &bytes.Buffer{}
	err = kio.ByteWriter{Writer: b, KeepReaderAnnotations: true}.Write(result)
	if !assert.NoError(t, err) {
		return
	}

	// Resources should be modified
	assert.Equal(t, `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/d.yaml'
    config.kubernetes.io/index: '0'
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/s.yaml'
    config.kubernetes.io/index: '1'
`, b.String())
}

func TestFilter_Filter_scope_nested_resource(t *testing.T) {
	// functions under "functions/" dir should be scoped to parent dir
	cfg, err := yaml.Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  annotations:
    config.kubernetes.io/path: 'baz.yaml'
`)
	if !assert.NoError(t, err) {
		return
	}

	input, err := (&kio.ByteReader{Reader: bytes.NewBufferString(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/d.yaml'
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/s.yaml'
`)}).Read()
	if !assert.NoError(t, err) {
		return
	}

	// no resources match the scope
	called := false
	result, err := (&ContainerFilter{
		Image:  "example.com:version",
		Config: cfg,
		args:   []string{"sed", "s/Deployment/StatefulSet/g"},
		checkInput: func(s string) {
			called = true
			if !assert.Equal(t, `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
    annotations:
      config.kubernetes.io/path: 'foo/bar/d.yaml'
      config.kubernetes.io/index: '0'
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
    annotations:
      config.kubernetes.io/path: 'foo/bar/s.yaml'
      config.kubernetes.io/index: '1'
functionConfig: {apiVersion: apps/v1, kind: Deployment, metadata: {name: foo, annotations: {
      config.kubernetes.io/path: 'baz.yaml'}}}
`, s) {
				t.FailNow()
			}
		},
	}).Filter(input)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.True(t, called) {
		return
	}

	b := &bytes.Buffer{}
	err = kio.ByteWriter{Writer: b, KeepReaderAnnotations: true}.Write(result)
	if !assert.NoError(t, err) {
		return
	}

	// Resources should be modified
	assert.Equal(t, `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/d.yaml'
    config.kubernetes.io/index: '0'
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/s.yaml'
    config.kubernetes.io/index: '1'
`, b.String())
}

func TestFilter_Filter_scopeDir(t *testing.T) {
	// functions under "functions/" dir should be scoped to parent dir
	cfg, err := yaml.Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  annotations:
    config.kubernetes.io/path: 'foo/bar.yaml'
`)
	if !assert.NoError(t, err) {
		return
	}

	input, err := (&kio.ByteReader{Reader: bytes.NewBufferString(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/d.yaml'
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/s.yaml'
`)}).Read()
	if !assert.NoError(t, err) {
		return
	}

	// no resources match the scope
	called := false
	result, err := (&ContainerFilter{
		Image:  "example.com:version",
		Config: cfg,
		args:   []string{"sed", "s/Deployment/StatefulSet/g"},
		checkInput: func(s string) {
			called = true
			if !assert.Equal(t, `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: deployment-foo
    annotations:
      config.kubernetes.io/path: 'foo/bar/d.yaml'
      config.kubernetes.io/index: '0'
- apiVersion: v1
  kind: Service
  metadata:
    name: service-foo
    annotations:
      config.kubernetes.io/path: 'foo/bar/s.yaml'
      config.kubernetes.io/index: '1'
functionConfig: {apiVersion: apps/v1, kind: Deployment, metadata: {name: foo, annotations: {
      config.kubernetes.io/path: 'foo/bar.yaml'}}}
`, s) {
				t.FailNow()
			}
		},
	}).Filter(input)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.True(t, called) {
		return
	}

	b := &bytes.Buffer{}
	err = kio.ByteWriter{Writer: b, KeepReaderAnnotations: true}.Write(result)
	if !assert.NoError(t, err) {
		return
	}

	// Resources should be preserved
	assert.Equal(t, `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: deployment-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/d.yaml'
    config.kubernetes.io/index: '0'
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/path: 'foo/bar/s.yaml'
    config.kubernetes.io/index: '1'
`, b.String())
}

func TestContainerFilter_scope(t *testing.T) {
	cf := &ContainerFilter{}

	fnR, err := yaml.Parse(`apiVersion: config.kubernetes.io/v1beta1
kind: ConfigFunction
metadata:
  name: config-function
  annotations:
    config.kubernetes.io/path: 'functions/bar.yaml'
`)
	if !assert.NoError(t, err) {
		return
	}

	inRs := []*yaml.RNode{fnR}
	inScopeRs, notInScopeRs, err := cf.scope(".", inRs)
	if !assert.NoError(t, err) {
		return
	}
	assert.Len(t, inScopeRs, 1, "Number of in-scope Resources")
	assert.Len(t, notInScopeRs, 0, "Number of out-of-scope Resources")
}
