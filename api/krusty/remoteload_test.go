// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/internal/utils"
	"sigs.k8s.io/kustomize/api/krusty"
)

func TestRemoteLoad(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()
	b := krusty.MakeKustomizer(fSys, krusty.MakeDefaultOptions())
	m, err := b.Run(
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
