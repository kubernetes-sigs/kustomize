// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"fmt"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const resourcesField = `resources:
- %s`
const resourceErrorFormat = "accumulating resources: accumulation err='accumulating resources from '%s': "
const fileError = "evalsymlink failure"
const repoFindError = "URL is a git repository"

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
	error         bool
	expected      string // if error, expected is error string
}

func createKustDir(content string, require *require.Assertions) (filesys.FileSystem, filesys.ConfirmedDir) {
	fSys := filesys.MakeFsOnDisk()
	tmpDir, err := filesys.NewTmpConfirmedDir()
	require.NoError(err)
	require.NoError(fSys.WriteFile(filepath.Join(tmpDir.String(), "kustomization.yaml"), []byte(content)))
	return fSys, tmpDir
}

func checkYaml(actual resmap.ResMap, expected string, require *require.Assertions) {
	yml, err := actual.AsYaml()
	require.NoError(err)
	require.Equal(expected, string(yml))
}

func testRemoteResource(require *require.Assertions, test *remoteResourceCase) {
	fSys, tmpDir := createKustDir(test.kustomization, require)

	b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	m, err := b.Run(
		fSys,
		tmpDir.String())

	if test.error {
		require.Error(err)
		require.Contains(err.Error(), test.expected)
	} else {
		require.NoError(err)
		checkYaml(m, test.expected, require)
	}

	require.NoError(fSys.RemoveAll(tmpDir.String()))
}

func TestRemoteLoad(t *testing.T) {
	require := require.New(t)

	fSys := filesys.MakeFsOnDisk()
	b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())

	m, err := b.Run(
		fSys,
		"github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6")
	require.NoError(err)
	checkYaml(m, multibaseDevExampleBuild, require)
}

func TestRemoteResourceHttps(t *testing.T) {
	tests := map[string]remoteResourceCase{
		"basic": {
			kustomization: `
resources:
- https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev/?ref=v1.0.6`,
			expected: multibaseDevExampleBuild,
		},
		".git repo suffix, no slash suffix": {
			kustomization: `
resources:
- https://github.com/kubernetes-sigs/kustomize.git//examples/multibases/dev?ref=v1.0.6`,
			expected: multibaseDevExampleBuild,
		},
		"repo": {
			kustomization: `
resources:
- https://github.com/annasong20/kustomize-test.git?ref=main`,
			expected: multibaseDevExampleBuild,
		},
		"raw remote file": {
			kustomization: `
resources:
- https://raw.githubusercontent.com/kubernetes-sigs/kustomize/v3.1.0/examples/multibases/base/pod.yaml
namePrefix: dev-`,
			expected: multibaseDevExampleBuild,
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			if test.skip {
				t.SkipNow()
			}
			testRemoteResource(require.New(t), &test)
		})
	}
}

func TestRemoteResourceSsh(t *testing.T) {
	// skip all tests until server has ssh keys
	tests := map[string]remoteResourceCase{
		"scp shorthand": {
			skip: true,
			kustomization: `
resources:
- git@github.com:kubernetes-sigs/kustomize//examples/multibases/dev/?ref=v1.0.6`,
			expected: multibaseDevExampleBuild,
		},
		"full ssh, no ending slash": {
			skip: true,
			kustomization: `
resources:
- ssh://git@github.com/kubernetes-sigs/kustomize//examples/multibases/dev?ref=v1.0.6`,
			expected: multibaseDevExampleBuild,
		},
		"repo": {
			skip: true,
			kustomization: `
resources:
- ssh://git@github.com/annasong20/kustomize-test.git?ref=main`,
			expected: multibaseDevExampleBuild,
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			if test.skip {
				t.SkipNow()
			}
			testRemoteResource(require.New(t), &test)
		})
	}
}

func TestRemoteResourcePort(t *testing.T) {
	sshURL := "ssh://git@github.com:22/kubernetes-sigs/kustomize//examples/multibases/dev/?ref=v1.0.6"
	httpsURL := "https://github.com:443/kubernetes-sigs/kustomize//examples/multibases/dev/?ref=v1.0.6"

	// ports not currently supported; should implement in future
	tests := map[string]remoteResourceCase{
		"ssh": {
			skip:          true,
			kustomization: fmt.Sprintf(resourcesField, sshURL),
			error:         true,
			expected:      fmt.Sprintf(resourceErrorFormat+fileError, sshURL),
		},
		"https": {
			kustomization: fmt.Sprintf(resourcesField, httpsURL),
			error:         true,
			expected:      fmt.Sprintf(resourceErrorFormat+repoFindError, httpsURL),
		},
	}
	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			if test.skip {
				t.SkipNow()
			}
			testRemoteResource(require.New(t), &test)
		})
	}
}

func TestRemoteResourceRepo(t *testing.T) {
	tests := map[string]remoteResourceCase{
		"https, no ref": {
			kustomization: `
resources:
- https://github.com/annasong20/kustomize-test.git`,
			expected: multibaseDevExampleBuild,
		},
		"ssh, no ref": {
			skip: true,
			kustomization: `
resources:
- git@github.com:annasong20/kustomize-test.git`,
			expected: multibaseDevExampleBuild,
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			if test.skip {
				t.SkipNow()
			}
			testRemoteResource(require.New(t), &test)
		})
	}
}

func TestRemoteResourceParameters(t *testing.T) {
	httpsNoParam := "https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev/"
	httpsMasterBranch := "https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev?ref=master"
	sshNoParams := "git@github.com:kubernetes-sigs/kustomize//examples/multibases/dev"

	// cases with expected errors are query parameter combinations that aren't supported yet; should implement in future
	tests := map[string]remoteResourceCase{
		"https no params": {
			skip:          true, // flaky: passes locally, but sometimes fails on server; fix in future
			kustomization: fmt.Sprintf(resourcesField, httpsNoParam),
			error:         true,
			expected:      fmt.Sprintf(resourceErrorFormat+repoFindError, httpsNoParam),
		},
		"https master": {
			skip:          true, // flaky: passes locally, but sometimes fails on server; fix in future
			kustomization: fmt.Sprintf(resourcesField, httpsMasterBranch),
			error:         true,
			expected:      fmt.Sprintf(resourceErrorFormat+repoFindError, httpsMasterBranch),
		},
		"https master and no submodules": {
			kustomization: `
resources:
- https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev?ref=master&submodules=false`,
			expected: multibaseDevExampleBuild,
		},
		"https all params": {
			kustomization: `
resources:
- https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev?ref=v1.0.6&timeout=10&submodules=true`,
			expected: multibaseDevExampleBuild,
		},
		"ssh no params": {
			skip:          true,
			kustomization: fmt.Sprintf(resourcesField, sshNoParams),
			error:         true,
			expected:      fmt.Sprintf(resourceErrorFormat+fileError, sshNoParams),
		},
		"ssh all params": {
			skip: true,
			kustomization: `
resources:
- ssh://git@github.com/annasong20/kustomize-test.git?ref=main&timeout=10&submodules=true`,
			expected: multibaseDevExampleBuild,
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			if test.skip {
				t.SkipNow()
			}
			testRemoteResource(require.New(t), &test)
		})
	}
}

func TestRemoteResourceGoGetter(t *testing.T) {
	tests := map[string]remoteResourceCase{
		"git detector with / subdirectory separator": {
			kustomization: `
resources:
- github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6`,
			expected: multibaseDevExampleBuild,
		},
		"git detector for repo": {
			kustomization: `
resources:
- github.com/annasong20/kustomize-test`,
			expected: multibaseDevExampleBuild,
		},
		"https with / subdirectory separator": {
			kustomization: `
resources:
- https://github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6`,
			expected: multibaseDevExampleBuild,
		},
		"git forced protocol": {
			kustomization: `
resources:
- git::https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev/?ref=v1.0.6`,
			expected: multibaseDevExampleBuild,
		},
		"git forced protocol with / subdirectory separator": {
			kustomization: `
resources:
- git::https://github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6`,
			expected: multibaseDevExampleBuild,
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			if test.skip {
				t.SkipNow()
			}
			testRemoteResource(require.New(t), &test)
		})
	}
}

func TestRemoteResourceWithHttpError(t *testing.T) {
	require := require.New(t)

	url404 := "https://github.com/thisisa404.yaml"
	fSys, tmpDir := createKustDir(fmt.Sprintf(resourcesField, url404), require)
	b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())

	_, err := b.Run(fSys, tmpDir.String())

	httpErr := fmt.Errorf("%w: status code %d (%s)", loader.ErrorHTTP, 404, http.StatusText(404))
	accuFromErr := fmt.Errorf("accumulating resources from '%s': %w", url404, httpErr)
	expectedErr := fmt.Errorf("accumulating resources: %w", accuFromErr)
	require.EqualErrorf(err, expectedErr.Error(), url404)

	require.NoError(fSys.RemoveAll(tmpDir.String()))
}

func TestRemoteResourceAnnoOrigin(t *testing.T) {
	test := remoteResourceCase{
		kustomization: `
resources:
- github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6
buildMetadata: [originAnnotations]
`,
		expected: `apiVersion: v1
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
`,
	}

	testRemoteResource(require.New(t), &test)
}

func TestRemoteResourceAsBaseWithAnnoOrigin(t *testing.T) {
	require := require.New(t)

	fSys := filesys.MakeFsOnDisk()
	b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	tmpDir, err := filesys.NewTmpConfirmedDir()
	require.NoError(err)
	base := filepath.Join(tmpDir.String(), "base")
	prod := filepath.Join(tmpDir.String(), "prod")
	require.NoError(fSys.Mkdir(base))
	require.NoError(fSys.Mkdir(prod))
	require.NoError(fSys.WriteFile(filepath.Join(base, "kustomization.yaml"), []byte(`
resources:
- github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6
`)))
	require.NoError(fSys.WriteFile(filepath.Join(prod, "kustomization.yaml"), []byte(`
resources:
- ../base
namePrefix: prefix-
buildMetadata: [originAnnotations]
`)))

	m, err := b.Run(
		fSys,
		prod)
	require.NoError(err)

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
	checkYaml(m, expected, require)

	require.NoError(fSys.RemoveAll(tmpDir.String()))
}
