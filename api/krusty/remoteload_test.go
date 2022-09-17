// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/kyaml/yaml"
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
	local         bool // only run locally; doesn't behave as expected on server
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

func isLocalEnv(require *require.Assertions) bool {
	// make variable that determines whether to run local-only tests
	if value, exists := os.LookupEnv("IS_LOCAL"); exists {
		isLocal, err := strconv.ParseBool(strings.TrimSpace(value))
		require.NoError(err)
		return isLocal
	}
	return false
}

func runResourceTests(t *testing.T, cases map[string]*remoteResourceCase) {
	t.Helper()

	req := require.New(t)
	for name, test := range cases {
		savedTest := test // test assignment changes; need assignment in this scope (iteration) of range
		t.Run(name, func(t *testing.T) {
			if savedTest.local && !isLocalEnv(req) {
				t.SkipNow()
			}
			configureGitSSHCommand(t)
			testRemoteResource(req, test)
		})
	}
}

func configureGitSSHCommand(t *testing.T) {
	t.Helper()

	// This contains a read-only Deploy Key for the kustomize repo.
	node, err := yaml.ReadFile("testdata/repo_read_only_ssh_key.yaml")
	require.NoError(t, err)
	keyB64, err := node.GetString("key")
	require.NoError(t, err)
	key, err := base64.StdEncoding.DecodeString(keyB64)
	require.NoError(t, err)

	// Write the key to a temp file and use it in SSH
	f, err := os.CreateTemp("", "kustomize_ssh")
	require.NoError(t, err)
	_, err = io.Copy(f, bytes.NewReader(key))
	require.NoError(t, err)
	cmd := fmt.Sprintf("ssh -i %s", f.Name())
	const SSHCommandKey = "GIT_SSH_COMMAND"
	t.Setenv(SSHCommandKey, cmd)
	t.Cleanup(func() {
		_ = os.Remove(f.Name())
	})
}

func TestRemoteLoad(t *testing.T) {
	req := require.New(t)

	fSys := filesys.MakeFsOnDisk()
	b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())

	m, err := b.Run(
		fSys,
		"github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6")
	req.NoError(err)
	checkYaml(m, multibaseDevExampleBuild, req)
}

func TestRemoteResourceHttps(t *testing.T) {
	tests := map[string]*remoteResourceCase{
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

	runResourceTests(t, tests)
}

func TestRemoteResourceSsh(t *testing.T) {
	tests := map[string]*remoteResourceCase{
		"scp shorthand": {
			kustomization: `
resources:
- git@github.com:kubernetes-sigs/kustomize//examples/multibases/dev/?ref=v1.0.6`,
			expected: multibaseDevExampleBuild,
		},
		"full ssh, no ending slash": {
			kustomization: `
resources:
- ssh://git@github.com/kubernetes-sigs/kustomize//examples/multibases/dev?ref=v1.0.6`,
			expected: multibaseDevExampleBuild,
		},
		"repo": {
			local: true,
			kustomization: `
resources:
- ssh://git@github.com/annasong20/kustomize-test.git?ref=main`,
			expected: multibaseDevExampleBuild,
		},
	}

	runResourceTests(t, tests)
}

func TestRemoteResourcePort(t *testing.T) {
	sshURL := "ssh://git@github.com:22/kubernetes-sigs/kustomize//examples/multibases/dev/?ref=v1.0.6"
	httpsURL := "https://github.com:443/kubernetes-sigs/kustomize//examples/multibases/dev/?ref=v1.0.6"

	// TODO: ports not currently supported; implement in future
	tests := map[string]*remoteResourceCase{
		"ssh": {
			local:         true,
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

	runResourceTests(t, tests)
}

func TestRemoteResourceRepo(t *testing.T) {
	tests := map[string]*remoteResourceCase{
		"https, no ref": {
			// TODO: fix flaky test that sporadically throws errors on server
			local: true,
			kustomization: `
resources:
- https://github.com/annasong20/kustomize-test.git`,
			expected: multibaseDevExampleBuild,
		},
		"ssh, no ref": {
			local: true,
			kustomization: `
resources:
- git@github.com:annasong20/kustomize-test.git`,
			expected: multibaseDevExampleBuild,
		},
	}

	runResourceTests(t, tests)
}

func TestRemoteResourceParameters(t *testing.T) {
	httpsNoParam := "https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev/"
	httpsMasterBranch := "https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev?ref=master"
	sshNoParams := "git@github.com:kubernetes-sigs/kustomize//examples/multibases/dev"

	// TODO: cases with expected errors are query parameter combinations that aren't supported yet; implement in future
	// TODO: fix flaky tests (non-ssh tests that we skip) that sporadically fail on server
	tests := map[string]*remoteResourceCase{
		"https no params": {
			local:         true,
			kustomization: fmt.Sprintf(resourcesField, httpsNoParam),
			error:         true,
			expected:      fmt.Sprintf(resourceErrorFormat+repoFindError, httpsNoParam),
		},
		"https master": {
			local:         true,
			kustomization: fmt.Sprintf(resourcesField, httpsMasterBranch),
			error:         true,
			expected:      fmt.Sprintf(resourceErrorFormat+repoFindError, httpsMasterBranch),
		},
		"https master and no submodules": {
			local: true,
			kustomization: `
resources:
- https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev?ref=master&submodules=false`,
			expected: multibaseDevExampleBuild,
		},
		"https all params": {
			local: true,
			kustomization: `
resources:
- https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev?ref=v1.0.6&timeout=10&submodules=true`,
			expected: multibaseDevExampleBuild,
		},
		"ssh no params": {
			local:         true,
			kustomization: fmt.Sprintf(resourcesField, sshNoParams),
			error:         true,
			expected:      fmt.Sprintf(resourceErrorFormat+fileError, sshNoParams),
		},
		"ssh all params": {
			local: true,
			kustomization: `
resources:
- ssh://git@github.com/annasong20/kustomize-test.git?ref=main&timeout=10&submodules=true`,
			expected: multibaseDevExampleBuild,
		},
	}

	runResourceTests(t, tests)
}

func TestRemoteResourceGoGetter(t *testing.T) {
	// TODO: fix flaky tests (the ones that we skip) that fail sporadically on server
	tests := map[string]*remoteResourceCase{
		"git detector with / subdirectory separator": {
			local: true,
			kustomization: `
resources:
- github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6`,
			expected: multibaseDevExampleBuild,
		},
		"git detector for repo": {
			local: true,
			kustomization: `
resources:
- github.com/annasong20/kustomize-test`,
			expected: multibaseDevExampleBuild,
		},
		"https with / subdirectory separator": {
			local: true,
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
			local: true,
			kustomization: `
resources:
- git::https://github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6`,
			expected: multibaseDevExampleBuild,
		},
	}

	runResourceTests(t, tests)
}

func TestRemoteResourceWithHttpError(t *testing.T) {
	req := require.New(t)

	url404 := "https://github.com/thisisa404.yaml"
	fSys, tmpDir := createKustDir(fmt.Sprintf(resourcesField, url404), req)
	b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())

	_, err := b.Run(fSys, tmpDir.String())

	httpErr := fmt.Errorf("%w: status code %d (%s)", loader.ErrHTTP, 404, http.StatusText(404))
	accuFromErr := fmt.Errorf("accumulating resources from '%s': %w", url404, httpErr)
	expectedErr := fmt.Errorf("accumulating resources: %w", accuFromErr)
	req.EqualErrorf(err, expectedErr.Error(), url404)

	req.NoError(fSys.RemoveAll(tmpDir.String()))
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

func TestRemoteResourceAnnoOriginFailing(t *testing.T) {
	test := remoteResourceCase{
		kustomization: `
resources:
- https://github.com/mightyguava/kustomize-bug.git
buildMetadata: [originAnnotations]
`,
		expected: `apiVersion: v1
kind: Pod
metadata:
  annotations:
    config.kubernetes.io/origin: |
      repo: https://github.com/mightyguava/kustomize-bug.git
      path: pod.yaml
  labels:
    app: myapp
  name: myapp-pod
spec:
  containers:
  - image: nginx:1.7.9
    name: nginx
`,
	}

	testRemoteResource(require.New(t), &test)
}

func TestRemoteResourceAsBaseWithAnnoOrigin(t *testing.T) {
	req := require.New(t)

	fSys := filesys.MakeFsOnDisk()
	b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	tmpDir, err := filesys.NewTmpConfirmedDir()
	req.NoError(err)
	base := filepath.Join(tmpDir.String(), "base")
	prod := filepath.Join(tmpDir.String(), "prod")
	req.NoError(fSys.Mkdir(base))
	req.NoError(fSys.Mkdir(prod))
	req.NoError(fSys.WriteFile(filepath.Join(base, "kustomization.yaml"), []byte(`
resources:
- github.com/kubernetes-sigs/kustomize/examples/multibases/dev/?ref=v1.0.6
`)))
	req.NoError(fSys.WriteFile(filepath.Join(prod, "kustomization.yaml"), []byte(`
resources:
- ../base
namePrefix: prefix-
buildMetadata: [originAnnotations]
`)))

	m, err := b.Run(
		fSys,
		prod)
	req.NoError(err)

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
	checkYaml(m, expected, req)

	req.NoError(fSys.RemoveAll(tmpDir.String()))
}
