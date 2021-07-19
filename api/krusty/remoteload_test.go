// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/internal/utils"
	"sigs.k8s.io/kustomize/api/krusty"
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
	fSys.Mkdir("tmp")
	fSys.WriteFile("tmp/kustomization.yaml", []byte(`
resources:
- github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6
`))
	m, err := b.Run(
		fSys,
		"tmp")
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
	assert.NoError(t, fSys.RemoveAll("tmp"))
}

func TestRemoteResourceAnnoOrigin(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()
	b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	fSys.Mkdir("tmp")
	fSys.WriteFile("tmp/kustomization.yaml", []byte(`
resources:
- github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6
buildOptions:
  addAnnoOrigin: true
`))
	m, err := b.Run(
		fSys,
		"tmp")
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
	assert.NoError(t, fSys.RemoveAll("tmp"))
}

func TestRemoteResourceAsBaseWithAnnoOrigin(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()
	b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	fSys.Mkdir("tmp")
	fSys.Mkdir("tmp/base")
	fSys.WriteFile("tmp/base/kustomization.yaml", []byte(`
resources:
- github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6
`))
	fSys.Mkdir("tmp/prod")
	fSys.WriteFile("tmp/prod/kustomization.yaml", []byte(`
resources:
- ../base
namePrefix: prefix-
buildOptions:
  addAnnoOrigin: true
`))
	m, err := b.Run(
		fSys,
		"tmp/prod")
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
	assert.NoError(t, fSys.RemoveAll("tmp"))
}
