// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	loc "sigs.k8s.io/kustomize/api/internal/localizer"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const simpleHTTPS = "https://github.com/kubernetes-sigs/kustomize//api/krusty/testdata/localize/simple"
const fileURL = "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/kustomize/v4.5.7/api/krusty/testdata/localize/simple/deployment.yaml"
const query = "?submodules=0&ref=kustomize/v4.5.7&timeout=300"

const orgToRefPrefix = "kubernetes-sigs/kustomize/kustomize/v4.5.7"
const pathPrefix = "api/krusty/testdata/localize"

func orgrepoDirs() map[string]struct{} {
	return map[string]struct{}{
		"kubernetes-sigs":           {},
		"kubernetes-sigs/kustomize": {},
	}
}
func refDirs() map[string]struct{} {
	return map[string]struct{}{
		"kubernetes-sigs/kustomize/kustomize": {},
		orgToRefPrefix:                        {},
	}
}
func testDirs() map[string]struct{} {
	return map[string]struct{}{
		"api":                 {},
		"api/krusty":          {},
		"api/krusty/testdata": {},
		pathPrefix:            {},
	}
}

func simpleFiles() map[string]string {
	return map[string]string{
		"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: localize-
resources:
- deployment.yaml
- service.yaml
`,

		"deployment.yaml": `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment-simple
  labels:
    app: deployment-simple
spec:
  selector:
    matchLabels:
      app: simple
  template:
    metadata:
      labels:
        app: simple
    spec:
      containers:
      - name: nginx
        image: nginx:1.16
        ports:
        - containerPort: 8080
`,

		"service.yaml": `apiVersion: v1
kind: Service
metadata:
  name: test-service-simple
spec:
  selector:
    app: deployment-simple
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080 
`,
	}
}
func remoteFiles() map[string]string {
	return map[string]string{
		"kustomization.yaml": fmt.Sprintf(`apiVersion: kustomize.config.k8s.io/v1beta1
commonLabels:
  purpose: remoteReference
kind: Kustomization
resources:
- %s/github.com/kubernetes-sigs/kustomize/kustomize/v4.5.7/api/krusty/testdata/localize/simple
- hpa.yaml
`, loc.LocalizeDir),

		"hpa.yaml": `apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: hpa-deployment
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: localize-test-deployment-simple
  minReplicas: 1
  maxReplicas: 10`,
	}
}

func checkLogs(t *testing.T) {
	t.Helper()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() {
		require.Empty(t, buf.String())
	})
}

func addPrefixToPaths[T interface{}](t *testing.T, prefix string, entries map[string]T) map[string]T {
	t.Helper()

	prefixed := map[string]T{}
	for path, value := range entries {
		prefixed[filepath.Join(prefix, path)] = value
	}
	return prefixed
}

func mkDirsAndFiles(t *testing.T, fSys filesys.FileSystem, dirs map[string]struct{}, files map[string]string) {
	t.Helper()

	for path := range dirs {
		err := fSys.MkdirAll(path)
		require.NoError(t, err)
	}

	for file, content := range files {
		err := fSys.WriteFile(file, []byte(content))
		require.NoError(t, err)
	}
}

func mkSymlinks(t *testing.T, links map[string]string) {
	t.Helper()

	for linkFile, linkDst := range links {
		require.NoError(t, os.Symlink(linkDst, linkFile))
	}
}

func changeWd(t *testing.T, newWd string) {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(wd))
	})

	err = os.Chdir(newWd)
	require.NoError(t, err)
}

func combine[T interface{}](t *testing.T, srcs ...map[string]T) map[string]T {
	t.Helper()

	combination := map[string]T{}
	for _, src := range srcs {
		for key, value := range src {
			combination[key] = value
		}
	}
	return combination
}

func checkFileSystem(t *testing.T, fSys filesys.FileSystem, walkDir string, dirs map[string]struct{}, files map[string]string) {
	t.Helper()
	req := require.New(t)

	// does not follow symbolic links, so should be no repeats
	count := 0
	err := fSys.Walk(walkDir, func(path string, info fs.FileInfo, err error) error {
		req.NoError(err)
		count++

		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			req.Contains(files, path)
			return nil
		}
		if info.IsDir() {
			req.Contains(dirs, path)
		} else {
			req.Contains(files, path)
			content, readErr := fSys.ReadFile(path)
			req.NoError(readErr)
			req.Equal(files[path], string(content))
		}
		return nil
	})
	req.NoError(err)
	req.Equal(len(dirs)+len(files), count)
}

func TestSymlinks(t *testing.T) {
	req := require.New(t)
	fSys := filesys.MakeFsOnDisk()
	checkLogs(t)

	dir := t.TempDir()
	scope := filepath.Join(dir, "scope")
	require.NoError(t, fSys.Mkdir(scope))

	absoluteDirs := addPrefixToPaths(t, scope, map[string]struct{}{
		"target":            {},
		"root2":             {},
		"root2/nested-root": {},
	})
	actualFiles := map[string]string{
		// symlinks from outside to inside scope
		"target/kustomization.yaml": `
resources:
- ../../symlink-pod.yaml
- ../../scope/../root2-link
namePrefix: my-`,

		"target/pod.yaml": "pod configuration",

		"root2/openapi.yaml": "openapi schema",

		// symlink to kustomization whose paths are calculated relative to kustomization root
		// tests openapi field
		"root2/nested-root/kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
openapi:
  path: openapi.yaml
`,
	}
	absoluteFiles := addPrefixToPaths(t, scope, actualFiles)
	mkDirsAndFiles(t, fSys, absoluteDirs, absoluteFiles)

	links := map[string]string{
		filepath.Join(dir, "fake-target"):                   filepath.Join(scope, "target"),
		filepath.Join(dir, "symlink-pod.yaml"):              filepath.Join(scope, "target", "pod.yaml"),
		filepath.Join(dir, "root2-link"):                    filepath.Join(scope, "root2"),
		filepath.Join(scope, "root2", "kustomization.yaml"): filepath.Join(scope, "root2", "nested-root", "kustomization.yaml"),
	}
	mkSymlinks(t, links)

	// unable to reference parent directories in in-memory file system
	changeWd(t, filepath.Join(scope, "root2", "nested-root"))
	// target, destination arguments contain symlinks
	err := loc.Run(fSys, "../../../fake-target", "../..", "../../../root2-link/newDir")
	req.NoError(err)

	dst := filepath.Join(scope, "root2", "newDir")
	expectedDirs := combine(t, map[string]struct{}{
		dir:   {},
		scope: {},
		dst:   {},
	}, absoluteDirs, addPrefixToPaths(t, dst, map[string]struct{}{"target": {}, "root2": {}}))
	expectedFiles := combine(t, addPrefixToPaths(t, dst, map[string]string{
		"target/kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: my-
resources:
- pod.yaml
- ../root2
`,
		"target/pod.yaml":          actualFiles["target/pod.yaml"],
		"root2/kustomization.yaml": actualFiles["root2/nested-root/kustomization.yaml"],
		"root2/openapi.yaml":       actualFiles["root2/openapi.yaml"],
	}), absoluteFiles, links)
	checkFileSystem(t, fSys, dir, expectedDirs, expectedFiles)
}

func TestRemoteTarget(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()
	checkLogs(t)
	dir := t.TempDir()
	changeWd(t, dir)

	// slash in ref
	target := simpleHTTPS + query
	// folder in repo with default localize destination
	err := loc.Run(fSys, target, "", "")
	require.NoError(t, err)

	dst := filepath.Join(dir, "localized-simple-kustomize-v4.5.7")
	simpleDir := filepath.Join(pathPrefix, "simple")
	dirs := addPrefixToPaths(t, dst, combine(t, testDirs(), map[string]struct{}{simpleDir: {}}))
	dirs[dir] = struct{}{}
	dirs[dst] = struct{}{}
	files := addPrefixToPaths(t, filepath.Join(dst, simpleDir), simpleFiles())
	checkFileSystem(t, fSys, dir, dirs, files)
}

func TestRemoteFile(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()
	checkLogs(t)
	dir := t.TempDir()

	// unreferenced localize directory should not cause error
	extraDirs := map[string]struct{}{filepath.Join(dir, loc.LocalizeDir): {}}
	files := map[string]string{
		filepath.Join(dir, "kustomization.yaml"): fmt.Sprintf(`
resources:
- %s
namePrefix: my-`, fileURL),
	}
	mkDirsAndFiles(t, fSys, extraDirs, files)

	dst := filepath.Join(dir, "newDir")
	err := loc.Run(fSys, dir, "", dst)
	require.NoError(t, err)

	hostPrefix := filepath.Join(dst, loc.LocalizeDir, "raw.githubusercontent.com")
	simplePrefix := filepath.Join(hostPrefix, orgToRefPrefix, pathPrefix, "simple")
	checkFileSystem(t, fSys, dir,
		combine(t,
			extraDirs,
			addPrefixToPaths(t, hostPrefix, combine(t, orgrepoDirs(), refDirs(),
				addPrefixToPaths(t, orgToRefPrefix, testDirs()))),
			map[string]struct{}{
				dir:                                 {},
				dst:                                 {},
				filepath.Join(dst, loc.LocalizeDir): {},
				hostPrefix:                          {},
				simplePrefix:                        {},
			}),

		combine(t, files, map[string]string{
			filepath.Join(dst, "kustomization.yaml"): fmt.Sprintf(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: my-
resources:
- %s/raw.githubusercontent.com/kubernetes-sigs/kustomize/kustomize/v4.5.7/api/krusty/testdata/localize/simple/deployment.yaml
`, loc.LocalizeDir),

			filepath.Join(simplePrefix, "deployment.yaml"): simpleFiles()["deployment.yaml"],
		}))
}

func TestNestedRemoteRoots(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()
	checkLogs(t)
	dir := t.TempDir()

	// should not conflict with localize directory name
	falseLocalizeDir := loc.LocalizeDir + loc.LocalizeDir
	roots := map[string]struct{}{
		falseLocalizeDir: {},
	}
	absoluteRoots := addPrefixToPaths(t, dir, roots)

	files := map[string]string{
		// ref without slash
		// ssh url
		"kustomization.yaml": fmt.Sprintf(`
resources:
- git@github.com:kubernetes-sigs/kustomize.git//api/krusty/testdata/localize/remote?submodules=0&ref=master&timeout=300
- %s%s
- ./%s
namePrefix: my-`, simpleHTTPS, query, falseLocalizeDir),

		filepath.Join(falseLocalizeDir, "kustomization.yaml"): `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: very-own-
resources:
- configmap.yaml
`,

		filepath.Join(falseLocalizeDir, "configmap.yaml"): "configmap configuration",
	}
	absoluteFiles := addPrefixToPaths(t, dir, files)
	mkDirsAndFiles(t, fSys, absoluteRoots, absoluteFiles)

	configureGitSSHCommand(t) // must be called before changing working directory
	changeWd(t, dir)

	err := loc.Run(fSys, ".", ".", "")
	require.NoError(t, err)

	dst := filepath.Join(dir, "localized-"+filepath.Base(dir))

	masterDirs := addPrefixToPaths(t, filepath.Join(loc.LocalizeDir, "github.com"), combine(t,
		orgrepoDirs(),
		map[string]struct{}{"kubernetes-sigs/kustomize/master": {}},
		addPrefixToPaths(t, "kubernetes-sigs/kustomize/master", testDirs())))
	versionDirs := addPrefixToPaths(t, filepath.Join(loc.LocalizeDir, "github.com"), combine(t,
		orgrepoDirs(), refDirs(), addPrefixToPaths(t, orgToRefPrefix, testDirs())))

	remoteDir := filepath.Join(dst, loc.LocalizeDir, "github.com", "kubernetes-sigs", "kustomize", "master", pathPrefix,
		"remote")
	simpleRelDir := filepath.Join(loc.LocalizeDir, "github.com", orgToRefPrefix, pathPrefix, "simple")

	checkFileSystem(t, fSys, dir,
		combine(t, absoluteRoots, addPrefixToPaths(t, dst, combine(t, roots, masterDirs, versionDirs)),
			addPrefixToPaths(t, remoteDir, versionDirs), map[string]struct{}{
				dir:                                 {},
				dst:                                 {},
				filepath.Join(dst, loc.LocalizeDir): {},
				filepath.Join(dst, loc.LocalizeDir, "github.com"):       {},
				filepath.Join(dst, simpleRelDir):                        {},
				remoteDir:                                               {},
				filepath.Join(remoteDir, loc.LocalizeDir):               {},
				filepath.Join(remoteDir, loc.LocalizeDir, "github.com"): {},
				filepath.Join(remoteDir, simpleRelDir):                  {},
			}),
		combine(t, absoluteFiles, addPrefixToPaths(t, remoteDir, remoteFiles()),
			addPrefixToPaths(t, filepath.Join(dst, simpleRelDir), simpleFiles()),
			addPrefixToPaths(t, filepath.Join(remoteDir, simpleRelDir), simpleFiles()),
			addPrefixToPaths(t, dst, map[string]string{
				"kustomization.yaml": fmt.Sprintf(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: my-
resources:
- %s/github.com/kubernetes-sigs/kustomize/master/api/krusty/testdata/localize/remote
- %s/github.com/kubernetes-sigs/kustomize/kustomize/v4.5.7/api/krusty/testdata/localize/simple
- %s
`, loc.LocalizeDir, loc.LocalizeDir, falseLocalizeDir),
				filepath.Join(falseLocalizeDir, "kustomization.yaml"): files[filepath.Join(falseLocalizeDir, "kustomization.yaml")],
				filepath.Join(falseLocalizeDir, "configmap.yaml"):     files[filepath.Join(falseLocalizeDir, "configmap.yaml")],
			})))
}

func TestExistingLocalizeDir(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()
	checkLogs(t)
	dir := t.TempDir()
	locDir := filepath.Join(loc.LocalizeDir, loc.LocalizeDir)

	folders := addPrefixToPaths(t, dir, map[string]struct{}{
		loc.LocalizeDir: {},
		locDir:          {},
		"root2":         {},
	})
	files := addPrefixToPaths(t, dir, map[string]string{
		filepath.Join(loc.LocalizeDir, "kustomization.yaml"): fmt.Sprintf(`
resources:
- %s
- ../root2
namePrefix: my-`, fileURL),
		"root2/kustomization.yaml": fmt.Sprintf(`
resources:
- ../%s
nameSuffix: config`, filepath.Join(loc.LocalizeDir, loc.LocalizeDir)),
		filepath.Join(locDir, "kustomization.yaml"): `
resources: 
- pod.yaml
namePrefix: prod-`,
		filepath.Join(locDir, "pod.yaml"): "pod configuration",
	})
	mkDirsAndFiles(t, fSys, folders, files)

	err := loc.Run(fSys, filepath.Join(dir, loc.LocalizeDir), dir, filepath.Join(dir, "newDir"))
	require.ErrorIs(t, err, loc.ErrLocalizeDirExists)
	checkFileSystem(t, fSys, dir, combine(t, map[string]struct{}{dir: {}}, folders), files)
}

func TestRemoteResourceNoRef(t *testing.T) {
	req := require.New(t)
	fSys := filesys.MakeFsOnDisk()
	checkLogs(t)

	dir := t.TempDir()
	kustPath := filepath.Join(dir, "kustomization.yaml")
	files := map[string]string{
		kustPath: fmt.Sprintf(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- %s
`, simpleHTTPS),
	}
	err := fSys.WriteFile(kustPath, []byte(files[kustPath]))
	req.NoError(err)

	err = loc.Run(fSys, dir, "", filepath.Join(dir, "newDir"))
	req.ErrorIs(err, loc.ErrNoRef)
	checkFileSystem(t, fSys, dir, map[string]struct{}{
		dir: {},
	}, files)
}

func TestBadArgs(t *testing.T) {
	tests := map[string]*struct {
		target string
		scope  string
		newDir string
	}{
		"target missing ref": {
			simpleHTTPS,
			"",
			"",
		},
		"non-empty scope for remote target": {
			simpleHTTPS + query,
			"api",
			"",
		},
		// unable to test on in-memory file system, which treats mkdir and mkdirall the same
		"dst in non-existing dir": {
			".",
			"",
			"does-not-exist/does-not-exist",
		},
	}
	for name, args := range tests {
		t.Run(name, func(t *testing.T) {
			fSys := filesys.MakeFsOnDisk()
			checkLogs(t)

			dir := t.TempDir()
			kustPath := filepath.Join(dir, "kustomization.yaml")
			files := map[string]string{
				kustPath: `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`,
			}
			err := fSys.WriteFile(filepath.Join(dir, "kustomization.yaml"), []byte(files[kustPath]))
			require.NoError(t, err)

			changeWd(t, dir)

			err = loc.Run(fSys, args.target, args.scope, args.newDir)
			require.Error(t, err)
			// cleaned dst
			checkFileSystem(t, fSys, dir, map[string]struct{}{
				dir: {},
			}, files)
		})
	}
}
