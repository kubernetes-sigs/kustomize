// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestRemoteLoad_LocalProtocol(t *testing.T) {
	type testRepos struct {
		root          string
		simple        string
		multiBaseDev  string
		withSubmodule string
	}

	// creates git repos under a root temporary directory with the following structure
	// root/
	//   simple.git/			- base with just a pod
	//   multibase.git/			- base with a dev overlay
	//   with-submodule.git/	- includes `simple` as a submodule
	//     submodule/			- the submodule referencing `simple
	createGitRepos := func(t *testing.T) testRepos {
		t.Helper()

		bash := func(script string) {
			cmd := exec.Command("sh", "-c", script)
			o, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("error running %v\nerr: %v\n%s", script, err, string(o))
			}
		}
		root := t.TempDir()
		bash(fmt.Sprintf(`
set -eux

export ROOT="%s"
export GIT_AUTHOR_EMAIL=nobody@kustomize.io
export GIT_AUTHOR_NAME=Nobody
export GIT_COMMITTER_EMAIL=nobody@kustomize.io
export GIT_COMMITTER_NAME=Nobody

cp -r testdata/remoteload/simple $ROOT/simple.git
(
	cd $ROOT/simple.git
	git init --initial-branch=main
	git add .
	git commit -m "import"
	git checkout -b change-image
	cat >>kustomization.yaml <<EOF

images:
- name: nginx
  newName: nginx
  newTag: "2"
EOF
	git commit -am "image change"
	git checkout main
)
cp -r testdata/remoteload/multibase $ROOT/multibase.git
(
	cd $ROOT/multibase.git
	git init --initial-branch=main
	git add .
	git commit -m "import"
)
(
	mkdir $ROOT/with-submodule.git
	cd $ROOT/with-submodule.git
	git init --initial-branch=main
	git submodule add $ROOT/simple.git submodule
	git commit -m "import"
)
`, root))
		return testRepos{
			root:          root,
			simple:        "simple.git",
			multiBaseDev:  "multibase.git",
			withSubmodule: "with-submodule.git",
		}
	}

	const simpleBuild = `apiVersion: v1
kind: Pod
metadata:
  labels:
    app: myapp
  name: myapp-pod
spec:
  containers:
  - image: nginx:1.7.9
    name: nginx
`
	var simpleBuildWithNginx2 = strings.ReplaceAll(simpleBuild, "nginx:1.7.9", "nginx:2")
	var multibaseDevExampleBuild = strings.ReplaceAll(simpleBuild, "myapp-pod", "dev-myapp-pod")

	repos := createGitRepos(t)
	tests := []struct {
		name          string
		kustomization string
		expected      string
		err           string
		skip          bool
	}{
		{
			name: "simple",
			kustomization: `
resources:
- file://$ROOT/simple.git
`,
			expected: simpleBuild,
		},
		{
			name: "has path",
			kustomization: `
resources:
- file://$ROOT/multibase.git/dev
`,
			expected: multibaseDevExampleBuild,
		},
		{
			name: "has ref",
			kustomization: `
resources: 
- "file://$ROOT/simple.git?ref=change-image"
`,

			expected: simpleBuildWithNginx2,
		},
		{
			// Version is the same as ref
			name: "has version",
			kustomization: `
resources:
- file://$ROOT/simple.git?version=change-image
`,
			expected: simpleBuildWithNginx2,
		},
		{
			name: "has submodule",
			kustomization: `
resources:
- file://$ROOT/with-submodule.git/submodule
`,
			expected: simpleBuild,
		},
		{
			name: "has timeout",
			kustomization: `
resources:
- file://$ROOT/simple.git?timeout=500
`,
			expected: simpleBuild,
		},
		{
			name: "triple slash absolute path",
			kustomization: `
resources:
- file:///$ROOT/simple.git
`,
			expected: simpleBuild,
		},
		{
			name: "has submodule but not initialized",
			kustomization: `
resources:
- file://$ROOT/with-submodule.git/submodule?submodules=0
`,
			err: "unable to find one of 'kustomization.yaml', 'kustomization.yml' or 'Kustomization' in directory",
		},
		{
			name: "has origin annotation",
			skip: true, // The annotated path should be "pod.yaml" but is "notCloned/pod.yaml"
			kustomization: `
resources:
- file://$ROOT/simple.git
buildMetadata: [originAnnotations]
`,
			expected: `apiVersion: v1
kind: Pod
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: pod.yaml
      repo: file://$ROOT/simple.git
  labels:
    app: myapp
  name: myapp-pod
spec:
  containers:
  - image: nginx:1.7.9
    name: nginx
`,
		},
		{
			name: "has ref path timeout and origin annotation",
			kustomization: `
resources:
- file://$ROOT/multibase.git/dev?version=main&timeout=500
buildMetadata: [originAnnotations]
`,
			expected: `apiVersion: v1
kind: Pod
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: base/pod.yaml
      repo: file://$ROOT/multibase.git
      ref: main
  labels:
    app: myapp
  name: dev-myapp-pod
spec:
  containers:
  - image: nginx:1.7.9
    name: nginx
`,
		},
		{
			name: "repo does not exist ",
			kustomization: `
resources:
- file:///not/a/real/repo
`,
			err: "fatal: '/not/a/real/repo' does not appear to be a git repository",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.skip {
				t.SkipNow()
			}

			kust := strings.ReplaceAll(test.kustomization, "$ROOT", repos.root)
			fSys, tmpDir := createKustDir(t, kust)

			b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
			m, err := b.Run(
				fSys,
				tmpDir.String())

			if test.err != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.err)
			} else {
				require.NoError(t, err)
				checkYaml(t, m, strings.ReplaceAll(test.expected, "$ROOT", repos.root))
			}
		})
	}
}

func TestRemoteLoad_RemoteProtocols(t *testing.T) {
	// Slow remote tests with long timeouts.
	// TODO: If these end up flaking, they should retry. If not, remove this TODO.
	tests := []struct {
		name          string
		kustomization string
		err           string
		errT          error
		beforeTest    func(t *testing.T, )
	}{
		{
			name: "https",
			kustomization: `
resources:
- https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev/?submodules=0&ref=kustomize%2Fv4.5.7&timeout=300
`,
		},
		{
			name: "git double-colon https",
			kustomization: `
resources:
- git::https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev/?submodules=0&ref=kustomize%2Fv4.5.7&timeout=300
`,
		},
		{
			name: "https raw remote file",
			kustomization: `
resources:
- https://raw.githubusercontent.com/kubernetes-sigs/kustomize/v3.1.0/examples/multibases/base/pod.yaml?timeout=300
namePrefix: dev-
`,
		},
		{
			name:       "ssh",
			beforeTest: configureGitSSHCommand,
			kustomization: `
resources:
- git@github.com/kubernetes-sigs/kustomize/examples/multibases/dev?submodules=0&ref=kustomize%2Fv4.5.7&timeout=300
`,
		},
		{
			name:       "ssh with colon",
			beforeTest: configureGitSSHCommand,
			kustomization: `
resources:
- git@github.com:kubernetes-sigs/kustomize/examples/multibases/dev?submodules=0&ref=kustomize%2Fv4.5.7&timeout=300
`,
		},
		{
			name:       "ssh without username",
			beforeTest: configureGitSSHCommand,
			kustomization: `
resources:
- github.com/kubernetes-sigs/kustomize/examples/multibases/dev?submodules=0&ref=kustomize%2Fv4.5.7&timeout=300
`,
		},
		{
			name:       "ssh scheme",
			beforeTest: configureGitSSHCommand,
			kustomization: `
resources:
- ssh://git@github.com/kubernetes-sigs/kustomize/examples/multibases/dev?submodules=0&ref=kustomize%2Fv4.5.7&timeout=300
`,
		},
		{
			name: "http error",
			kustomization: `
resources:
- https://github.com/thisisa404.yaml
`,
			err:  "accumulating resources: accumulating resources from 'https://github.com/thisisa404.yaml': HTTP Error: status code 404 (Not Found)",
			errT: loader.ErrHTTP,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.beforeTest != nil {
				test.beforeTest(t)
			}
			fSys, tmpDir := createKustDir(t, test.kustomization)

			b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
			m, err := b.Run(
				fSys,
				tmpDir.String())

			if test.err != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.err)
				if test.errT != nil {
					assert.ErrorIs(t, err, test.errT)
				}
			} else {
				require.NoError(t, err)
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
				checkYaml(t, m, multibaseDevExampleBuild)
			}
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

func createKustDir(t *testing.T, content string) (filesys.FileSystem, filesys.ConfirmedDir) {
	t.Helper()

	fSys := filesys.MakeFsOnDisk()
	tmpDir, err := filesys.NewTmpConfirmedDir()
	require.NoError(t, err)
	require.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "kustomization.yaml"), []byte(content)))
	return fSys, tmpDir
}

func checkYaml(t *testing.T, actual resmap.ResMap, expected string) {
	t.Helper()

	yml, err := actual.AsYaml()
	assert.NoError(t, err)
	assert.Equal(t, expected, string(yml))
}
