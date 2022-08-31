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

			expected: strings.ReplaceAll(simpleBuild, "nginx:1.7.9", "nginx:2"),
		},
		{
			// Version is the same as ref
			name: "has version",
			kustomization: `
resources:
- file://$ROOT/simple.git?version=change-image
`,
			expected: strings.ReplaceAll(simpleBuild, "nginx:1.7.9", "nginx:2"),
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
			name: "has submodule but not initialized",
			kustomization: `
resources:
- file://$ROOT/with-submodule.git/submodule?submodules=0
`,
			err: "unable to find",
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
			name: "has ref path and origin annotation",
			kustomization: `
resources:
- file://$ROOT/multibase.git/dev?version=main
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
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), test.err)
				}
			} else {
				assert.NoError(t, err)
				if err == nil {
					checkYaml(t, m, strings.ReplaceAll(test.expected, "$ROOT", repos.root))
				}
			}
		})
	}
}

func TestRemoteLoad_RemoteProtocols(t *testing.T) {
	// Slow remote tests with long timeouts.
	// TODO: Maybe they should retry too.
	tests := []struct {
		name          string
		kustomization string
		err           string
	}{
		{
			name: "https",
			kustomization: `
resources:
- https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev/?submodules=0&ref=v1.0.6&timeout=300
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
			name: "ssh",
			kustomization: `
resources:
- ssh://git@github.com/kubernetes-sigs/kustomize/examples/multibases/dev?submodules=0&ref=v1.0.6&timeout=300
`,
		},
	}

	configureGitSSHCommand(t)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fSys, tmpDir := createKustDir(t, test.kustomization)

			b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
			m, err := b.Run(
				fSys,
				tmpDir.String())

			if test.err != "" {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), test.err)
				}
			} else {
				assert.NoError(t, err)
				if err == nil {
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
