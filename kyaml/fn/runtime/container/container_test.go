// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestFilter_setupExec(t *testing.T) {
	var tests = []struct {
		name           string
		functionConfig string
		expectedArgs   []string
		containerSpec  runtimeutil.ContainerSpec
		UIDGID         string
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
			containerSpec: runtimeutil.ContainerSpec{
				Image: "example.com:version",
			},
			UIDGID: "nobody",
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
			containerSpec: runtimeutil.ContainerSpec{
				Image:   "example.com:version",
				Network: true,
			},
			UIDGID: "nobody",
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
				// use filepath.Join for Windows filepath handling
				"--mount", fmt.Sprintf("type=%s,source=%s,target=%s,readonly", "bind", getAbsFilePath(string(filepath.Separator), "mount", "path"), "/local/"),
				"--mount", fmt.Sprintf("type=%s,source=%s,target=%s", "bind", getAbsFilePath(string(filepath.Separator), "mount", "pathrw"), "/localrw/"),
				"--mount", fmt.Sprintf("type=%s,source=%s,target=%s,readonly", "volume", getAbsFilePath(string(filepath.Separator), "myvol"), "/local/"),
				"--mount", fmt.Sprintf("type=%s,source=%s,target=%s,readonly", "tmpfs", getAbsFilePath(string(filepath.Separator)), "/local/"),
			},
			containerSpec: runtimeutil.ContainerSpec{
				Image: "example.com:version",
				StorageMounts: []runtimeutil.StorageMount{
					{MountType: "bind", Src: getAbsFilePath(string(filepath.Separator), "mount", "path"), DstPath: "/local/"},
					{MountType: "bind", Src: getAbsFilePath(string(filepath.Separator), "mount", "pathrw"), DstPath: "/localrw/", ReadWriteMode: true},
					{MountType: "volume", Src: getAbsFilePath(string(filepath.Separator), "myvol"), DstPath: "/local/"},
					{MountType: "tmpfs", Src: getAbsFilePath(string(filepath.Separator)), DstPath: "/local/"},
				},
			},
			UIDGID: "nobody",
		},
		{
			name: "as current user",
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
				"--user", "1:2",
				"--security-opt=no-new-privileges",
			},
			containerSpec: runtimeutil.ContainerSpec{
				Image: "example.com:version",
			},
			UIDGID: "1:2",
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := yaml.Parse(tt.functionConfig)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			instance := NewContainer(tt.containerSpec, tt.UIDGID)
			instance.Exec.FunctionConfig = cfg
			instance.Env = append(instance.Env, "KYAML_TEST=FOO")
			assert.NoError(t, instance.setupExec())

			tt.expectedArgs = append(tt.expectedArgs,
				runtimeutil.NewContainerEnvFromStringSlice(instance.Env).GetDockerFlags()...)
			tt.expectedArgs = append(tt.expectedArgs, instance.Image)

			if !assert.Equal(t, "docker", instance.Exec.Path) {
				t.FailNow()
			}
			if !assert.Equal(t, tt.expectedArgs, instance.Exec.Args) {
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
	instance.Exec.WorkingDir = getWorkingDir(t)

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
    internal.config.kubernetes.io/index: '0'
    internal.config.kubernetes.io/path: 'statefulset_deployment-foo.yaml'
    config.kubernetes.io/path: 'statefulset_deployment-foo.yaml'
---
apiVersion: v1
kind: Service
metadata:
  name: service-foo
  annotations:
    config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/index: '1'
    internal.config.kubernetes.io/path: 'service_service-foo.yaml'
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
	instance.Exec.WorkingDir = getWorkingDir(t)
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

func getWorkingDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	return wd
}

func getAbsFilePath(args ...string) string {
	path, _ := filepath.Abs(filepath.Join(args...))
	return path
}
