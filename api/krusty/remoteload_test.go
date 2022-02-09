// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"fmt"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/internal/utils"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestRemoteLoad(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()
	b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	m, err := b.Run(
		fSys,
		"github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6")
	if utils.IsErrTimeout(err) {
		// Don't fail on timeouts.
		t.SkipNow()
	}
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	yml, err := m.AsYaml()
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: v1
kind: Pod
metadata:
  labels:
    app: myapp
  name: dev-myapp-pod
spec:
  containers:
  - image: nginx:1.7.9
    name: nginx
`, string(yml))
}

func TestRemoteResource(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()
	b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	tmpDir, err := filesys.NewTmpConfirmedDir()
	assert.NoError(t, err)
	assert.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "kustomization.yaml"), []byte(`
resources:
- github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6
`)))
	m, err := b.Run(
		fSys,
		tmpDir.String())
	if utils.IsErrTimeout(err) {
		// Don't fail on timeouts.
		t.SkipNow()
	}
	assert.NoError(t, err)
	yml, err := m.AsYaml()
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: v1
kind: Pod
metadata:
  labels:
    app: myapp
  name: dev-myapp-pod
spec:
  containers:
  - image: nginx:1.7.9
    name: nginx
`, string(yml))
	assert.NoError(t, fSys.RemoveAll(tmpDir.String()))
}

func TestRemoteResourceGitHTTP(t *testing.T) {
	output := `apiVersion: v1
kind: Pod
metadata:
  labels:
    app: myapp
  name: dev-myapp-pod
spec:
  containers:
  - image: nginx:1.7.9
    name: nginx
`
	tests := []struct {
		input []byte
	}{
		{
			input: []byte(`
resources:
- https://github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6
`),
		},
		{
			input: []byte(`
resources:
- https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev/?ref=v1.0.6
`),
		},
		{
			input: []byte(`
resources:
- git::https://github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6
`),
		},
		{
			input: []byte(`
resources:
- git::https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev/?ref=v1.0.6
`),
		},
	}

	for _, test := range tests {
		fSys := filesys.MakeFsOnDisk()
		b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
		tmpDir, err := filesys.NewTmpConfirmedDir()
		assert.NoError(t, err)
		assert.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "kustomization.yaml"), test.input))
		m, err := b.Run(fSys, tmpDir.String())
		if utils.IsErrTimeout(err) {
			// Don't fail on timeouts.
			t.SkipNow()
		}
		assert.NoError(t, err)
		yml, err := m.AsYaml()
		assert.NoError(t, err)
		assert.Equal(t, output, string(yml))
		assert.NoError(t, fSys.RemoveAll(tmpDir.String()))
	}
}

func TestRemoteResourceWithHTTPError(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()
	b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	tmpDir, err := filesys.NewTmpConfirmedDir()
	assert.NoError(t, err)

	url404 := "https://github.com/thisisa404.yaml"
	kusto := filepath.Join(tmpDir.String(), "kustomization.yaml")
	assert.NoError(t, fSys.WriteFile(kusto, []byte(fmt.Sprintf(`
resources:
- %s
`, url404))))

	_, err = b.Run(fSys, tmpDir.String())
	if utils.IsErrTimeout(err) {
		// Don't fail on timeouts.
		t.SkipNow()
	}

	httpErr := fmt.Errorf("%w: status code %d (%s)", loader.ErrorHTTP, 404, http.StatusText(404))
	accuFromErr := fmt.Errorf("accumulating resources from '%s': %w", url404, httpErr)
	expectedErr := fmt.Errorf("accumulating resources: %w", accuFromErr)
	assert.EqualErrorf(t, err, expectedErr.Error(), url404)
}

func TestRemoteResourceAnnoOrigin(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()
	b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	tmpDir, err := filesys.NewTmpConfirmedDir()
	assert.NoError(t, err)
	assert.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "kustomization.yaml"), []byte(`
resources:
- github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6
buildMetadata: [originAnnotations]
`)))
	m, err := b.Run(
		fSys,
		tmpDir.String())
	if utils.IsErrTimeout(err) {
		// Don't fail on timeouts.
		t.SkipNow()
	}
	assert.NoError(t, err)
	yml, err := m.AsYaml()
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: v1
kind: Pod
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: examples/multibases/base/pod.yaml
      repo: https://github.com/kubernetes-sigs/kustomize
      ref: v1.0.6
  labels:
    app: myapp
  name: dev-myapp-pod
spec:
  containers:
  - image: nginx:1.7.9
    name: nginx
`, string(yml))
	assert.NoError(t, fSys.RemoveAll(tmpDir.String()))
}

func TestRemoteResourceAsBaseWithAnnoOrigin(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()
	b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	tmpDir, err := filesys.NewTmpConfirmedDir()
	assert.NoError(t, err)
	base := filepath.Join(tmpDir.String(), "base")
	prod := filepath.Join(tmpDir.String(), "prod")
	assert.NoError(t, fSys.Mkdir(base))
	assert.NoError(t, fSys.Mkdir(prod))
	assert.NoError(t, fSys.WriteFile(filepath.Join(base, "kustomization.yaml"), []byte(`
resources:
- github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6
`)))
	assert.NoError(t, fSys.WriteFile(filepath.Join(prod, "kustomization.yaml"), []byte(`
resources:
- ../base
namePrefix: prefix-
buildMetadata: [originAnnotations]
`)))

	m, err := b.Run(
		fSys,
		prod)
	if utils.IsErrTimeout(err) {
		// Don't fail on timeouts.
		t.SkipNow()
	}
	assert.NoError(t, err)
	yml, err := m.AsYaml()
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: v1
kind: Pod
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: examples/multibases/base/pod.yaml
      repo: https://github.com/kubernetes-sigs/kustomize
      ref: v1.0.6
  labels:
    app: myapp
  name: prefix-dev-myapp-pod
spec:
  containers:
  - image: nginx:1.7.9
    name: nginx
`, string(yml))
	assert.NoError(t, fSys.RemoveAll(tmpDir.String()))
}
