// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"fmt"
	"net/http"
	"path/filepath"
	"testing"

	"sigs.k8s.io/kustomize/api/resmap"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/internal/utils"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const multibaseDevExampleBuild = `apiVersion: v1
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

type remoteResourceCase struct {
	skip          bool
	kustomization string
}

func createKusDir(content string, assert *assert.Assertions) (filesys.FileSystem, filesys.ConfirmedDir) {
	fSys := filesys.MakeFsOnDisk()
	tmpDir, err := filesys.NewTmpConfirmedDir()
	assert.NoError(err)
	assert.NoError(fSys.WriteFile(filepath.Join(tmpDir.String(), "kustomization.yaml"), []byte(content)))
	return fSys, tmpDir
}

func checkYaml(actual resmap.ResMap, expected string, assert *assert.Assertions) {
	yml, err := actual.AsYaml()
	assert.NoError(err)
	assert.Equal(expected, string(yml))
}

func testRemoteResource(assert *assert.Assertions, kustomization string, expected string) {
	fSys, tmpDir := createKusDir(kustomization, assert)

	b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	m, err := b.Run(
		fSys,
		tmpDir.String())

	assert.NoError(err)
	checkYaml(m, expected, assert)

	assert.NoError(fSys.RemoveAll(tmpDir.String()))
}

func TestRemoteLoad(t *testing.T) {
	assert := assert.New(t)

	fSys := filesys.MakeFsOnDisk()
	b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())

	m, err := b.Run(
		fSys,
		"github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6")
	assert.NoError(err)
	checkYaml(m, multibaseDevExampleBuild, assert)
}

func TestRemoteResourceHttps(t *testing.T) {
	tests := map[string]remoteResourceCase{
		"basic": {
			kustomization: `
resources:
- https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev/?ref=v1.0.6`,
		},
		".git repo suffix, no slash suffix": {
			kustomization: `
resources:
- https://github.com/kubernetes-sigs/kustomize.git//examples/multibases/dev?ref=v1.0.6`,
		},
		"repo": {
			kustomization: `
resources:
- https://github.com/annasong20/kustomize-test.git?ref=main`,
		},
		"raw remote file": {
			kustomization: `
resources:
- https://raw.githubusercontent.com/kubernetes-sigs/kustomize/v3.1.0/examples/multibases/base/pod.yaml
namePrefix: dev-`,
		},
	}

	for name, test := range tests {
		test := test
		if !test.skip {
			t.Run(name, func(t *testing.T) {
				testRemoteResource(assert.New(t), test.kustomization, multibaseDevExampleBuild)
			})
		}
	}
}

func TestRemoteResourceSsh(t *testing.T) {
	tests := map[string]remoteResourceCase{
		"scp shorthand": {
			kustomization: `
resources:
- git@github.com:kubernetes-sigs/kustomize//examples/multibases/dev/?ref=v1.0.6`,
		},
		"full ssh, no ending slash": {
			kustomization: `
resources:
- ssh://git@github.com/kubernetes-sigs/kustomize//examples/multibases/dev?ref=v1.0.6`,
		},
		"repo": {
			kustomization: `
resources:
- ssh://git@github.com/annasong20/kustomize-test.git?ref=main`,
		},
	}

	for name, test := range tests {
		test := test
		if !test.skip {
			t.Run(name, func(t *testing.T) {
				testRemoteResource(assert.New(t), test.kustomization, multibaseDevExampleBuild)
			})
		}
	}
}

func TestRemoteResourcePort(t *testing.T) {
	tests := map[string]remoteResourceCase{
		"https": {
			skip: true, // timeout
			kustomization: `
resources:
- https://github.com:443/kubernetes-sigs/kustomize//examples/multibases/dev/?ref=v1.0.6`,
		},
		"ssh": {
			skip: true, // error
			kustomization: `
resources:
- ssh://git@github.com:22/kubernetes-sigs/kustomize//examples/multibases/dev/?ref=v1.0.6`,
		},
	}
	for name, test := range tests {
		test := test
		if !test.skip {
			t.Run(name, func(t *testing.T) {
				testRemoteResource(assert.New(t), test.kustomization, multibaseDevExampleBuild)
			})
		}
	}
}

func TestRemoteResourceRepo(t *testing.T) {
	tests := map[string]remoteResourceCase{
		"https, no ref": {
			kustomization: `
resources:
- https://github.com/annasong20/kustomize-test.git`,
		},
		"ssh, no ref": {
			kustomization: `
resources:
- git@github.com:annasong20/kustomize-test.git`,
		},
	}

	for name, test := range tests {
		test := test
		if !test.skip {
			t.Run(name, func(t *testing.T) {
				testRemoteResource(assert.New(t), test.kustomization, multibaseDevExampleBuild)
			})
		}
	}
}

func TestRemoteResourceParameters(t *testing.T) {
	tests := map[string]remoteResourceCase{
		"https no params": {
			skip: true, // timeout
			kustomization: `
resources:
- https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev/`,
		},
		"https master": {
			skip: true, // timeout
			kustomization: `
resources:
- https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev?ref=master`,
		},
		"https master and no submodules": {
			kustomization: `
resources:
- https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev?ref=master&submodules=false`,
		},
		"https all params": {
			kustomization: `
resources:
- https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev?ref=v1.0.6&timeout=10&submodules=true`,
		},
		"ssh no params": {
			skip: true, // timeout
			kustomization: `
resources:
- git@github.com:kubernetes-sigs/kustomize//examples/multibases/dev`,
		},
		"ssh all params": {
			kustomization: `
resources:
- ssh://git@github.com/annasong20/kustomize-test.git?ref=main&timeout=10&submodules=true`,
		},
	}

	for name, test := range tests {
		test := test
		if !test.skip {
			t.Run(name, func(t *testing.T) {
				testRemoteResource(assert.New(t), test.kustomization, multibaseDevExampleBuild)
			})
		}
	}
}

func TestRemoteResourceGoGetter(t *testing.T) {
	tests := map[string]remoteResourceCase{
		"git detector with / subdirectory separator": {
			kustomization: `
resources:
- github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6`,
		},
		"git detector for repo": {
			kustomization: `
resources:
- github.com/annasong20/kustomize-test`,
		},
		"https with / subdirectory separator": {
			kustomization: `
resources:
- https://github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6`,
		},
		"git forced protocol": {
			kustomization: `
resources:
- git::https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev/?ref=v1.0.6`,
		},
		"git forced protocol with / subdirectory separator": {
			kustomization: `
resources:
- git::https://github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6`,
		},
	}

	for name, test := range tests {
		test := test
		if !test.skip {
			t.Run(name, func(t *testing.T) {
				testRemoteResource(assert.New(t), test.kustomization, multibaseDevExampleBuild)
			})
		}
	}
}

func TestRemoteResourceWithHttpError(t *testing.T) {
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
	kustomization := `
resources:
- github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6
buildMetadata: [originAnnotations]
`
	expected := `apiVersion: v1
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
`
	testRemoteResource(assert.New(t), kustomization, expected)
}

func TestRemoteResourceAsBaseWithAnnoOrigin(t *testing.T) {
	assert := assert.New(t)

	fSys := filesys.MakeFsOnDisk()
	b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	tmpDir, err := filesys.NewTmpConfirmedDir()
	assert.NoError(err)
	base := filepath.Join(tmpDir.String(), "base")
	prod := filepath.Join(tmpDir.String(), "prod")
	assert.NoError(fSys.Mkdir(base))
	assert.NoError(fSys.Mkdir(prod))
	assert.NoError(fSys.WriteFile(filepath.Join(base, "kustomization.yaml"), []byte(`
resources:
- github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6
`)))
	assert.NoError(fSys.WriteFile(filepath.Join(prod, "kustomization.yaml"), []byte(`
resources:
- ../base
namePrefix: prefix-
buildMetadata: [originAnnotations]
`)))

	m, err := b.Run(
		fSys,
		prod)
	assert.NoError(err)

	expected := `apiVersion: v1
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
`
	checkYaml(m, expected, assert)

	assert.NoError(fSys.RemoveAll(tmpDir.String()))
}
