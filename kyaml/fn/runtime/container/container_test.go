// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestFilter_setupExec(t *testing.T) {
	var tests = []struct {
		name           string
		functionConfig string
		expectedArgs   []string
		instance       Filter
	}{
		{
			name: "command",
			functionConfig: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
`,
			expectedArgs: []string{
				"run",
				"--rm",
				"-i", "-a", "STDIN", "-a", "STDOUT", "-a", "STDERR",
				"--network", "none",
				"--user", "nobody",
				"--security-opt=no-new-privileges",
			},
			instance: NewContainer(
				runtimeutil.ContainerSpec{
					Image: "example.com:version",
					User:  "nobody",
				},
			),
		},

		{
			name: "network",
			functionConfig: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
`,
			expectedArgs: []string{
				"run",
				"--rm",
				"-i", "-a", "STDIN", "-a", "STDOUT", "-a", "STDERR",
				"--network", "host",
				"--user", "nobody",
				"--security-opt=no-new-privileges",
			},
			instance: NewContainer(
				runtimeutil.ContainerSpec{
					Image:   "example.com:version",
					Network: true,
					User:    "nobody",
				},
			),
		},

		{
			name: "storage_mounts",
			functionConfig: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
`,
			expectedArgs: []string{
				"run",
				"--rm",
				"-i", "-a", "STDIN", "-a", "STDOUT", "-a", "STDERR",
				"--network", "none",
				"--user", "nobody",
				"--security-opt=no-new-privileges",
				"--mount", fmt.Sprintf("type=%s,source=%s,target=%s,readonly", "bind", "/mount/path", "/local/"),
				"--mount", fmt.Sprintf("type=%s,source=%s,target=%s", "bind", "/mount/pathrw", "/localrw/"),
				"--mount", fmt.Sprintf("type=%s,source=%s,target=%s,readonly", "volume", "myvol", "/local/"),
				"--mount", fmt.Sprintf("type=%s,source=%s,target=%s,readonly", "tmpfs", "", "/local/"),
			},
			instance: NewContainer(
				runtimeutil.ContainerSpec{
					Image: "example.com:version",
					StorageMounts: []runtimeutil.StorageMount{
						{MountType: "bind", Src: "/mount/path", DstPath: "/local/"},
						{MountType: "bind", Src: "/mount/pathrw", DstPath: "/localrw/", ReadWriteMode: true},
						{MountType: "volume", Src: "myvol", DstPath: "/local/"},
						{MountType: "tmpfs", Src: "", DstPath: "/local/"},
					},
					User: "nobody",
				},
			),
		},
		{
			name: "root user",
			functionConfig: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
`,
			expectedArgs: []string{
				"run",
				"--rm",
				"-i", "-a", "STDIN", "-a", "STDOUT", "-a", "STDERR",
				"--network", "none",
				"--user", "root",
				"--security-opt=no-new-privileges",
			},
			instance: NewContainer(
				runtimeutil.ContainerSpec{
					Image: "example.com:version",
					User:  "root",
				},
			),
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := yaml.Parse(tt.functionConfig)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			tt.instance.Exec.FunctionConfig = cfg
			tt.instance.Env = append(tt.instance.Env, "KYAML_TEST=FOO")
			tt.instance.setupExec()

			tt.expectedArgs = append(tt.expectedArgs,
				runtimeutil.NewContainerEnvFromStringSlice(tt.instance.Env).GetDockerFlags()...)
			tt.expectedArgs = append(tt.expectedArgs, tt.instance.Image)

			if !assert.Equal(t, "docker", tt.instance.Exec.Path) {
				t.FailNow()
			}
			if !assert.Equal(t, tt.expectedArgs, tt.instance.Exec.Args) {
				t.FailNow()
			}
		})
	}
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

	instance := Filter{}
	instance.Exec.FunctionConfig = cfg
	instance.Exec.Path = "sed"
	instance.Exec.Args = []string{"s/Deployment/StatefulSet/g"}
	output, err := instance.Filter(input)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	b := &bytes.Buffer{}
	err = kio.ByteWriter{Writer: b, KeepReaderAnnotations: true}.Write(output)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.Equal(t, `apiVersion: apps/v1
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
`, b.String()) {
		t.FailNow()
	}
}
func TestFilter_String(t *testing.T) {
	instance := Filter{ContainerSpec: runtimeutil.ContainerSpec{Image: "foo"}}
	if !assert.Equal(t, "foo", instance.String()) {
		t.FailNow()
	}

	instance.Exec.DeferFailure = true
	if !assert.Equal(t, "foo deferFailure: true", instance.String()) {
		t.FailNow()
	}
}

func TestFilter_ExitCode(t *testing.T) {
	instance := Filter{}
	instance.Exec.Path = "/not/real/command"
	instance.Exec.DeferFailure = true
	_, err := instance.Filter(nil)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.Error(t, instance.GetExit()) {
		t.FailNow()
	}
	if !assert.Contains(t, instance.GetExit().Error(), "/not/real/command") {
		t.FailNow()
	}
}
