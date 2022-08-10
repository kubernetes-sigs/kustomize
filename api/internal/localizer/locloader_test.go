// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer_test

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/ifc"
	lclzr "sigs.k8s.io/kustomize/api/internal/localizer"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const dstPrefix = "localized"

func makeMemoryFs(t *testing.T) filesys.FileSystem {
	t.Helper()
	req := require.New(t)

	fSys := filesys.MakeFsInMemory()
	req.NoError(fSys.MkdirAll("/a/b"))
	req.NoError(fSys.WriteFile("/a/kustomization.yaml", []byte("/a")))

	dirChain := "/alpha/beta/gamma/delta"
	req.NoError(fSys.MkdirAll(dirChain))
	req.NoError(fSys.WriteFile(dirChain+"/kustomization.yaml", []byte(dirChain)))
	req.NoError(fSys.Mkdir("/alpha/beta/c"))
	return fSys
}

func checkLocLoader(req *require.Assertions, ll *lclzr.LocLoader, root string, scope string) {
	checkLoader(req, ll, root)
	req.Equal(scope, ll.Scope())
}

func checkLoader(req *require.Assertions, ldr ifc.Loader, root string) {
	req.Equal(root, ldr.Root())
	repo, isRemote := ldr.Repo()
	req.Equal(false, isRemote)
	req.Equal("", repo)
}

func checkConfirmedDir(req *require.Assertions, dir filesys.ConfirmedDir, path string, fSys filesys.FileSystem) {
	req.Equal(path, dir.String())
	req.True(fSys.Exists(path))
}

func TestLocalLoadNewAndCleanup(t *testing.T) {
	req := require.New(t)
	fSys := makeMemoryFs(t)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	// typical setup
	ll, dst, err := lclzr.ValidateLocArgs("a", "/", "/newDir", fSys)
	req.NoError(err)
	checkLocLoader(req, ll, "/a", "/")

	req.Equal("/newDir", dst.String())
	fSysCopy := makeMemoryFs(t)
	req.NoError(fSysCopy.Mkdir("/newDir"))
	req.Equal(fSysCopy, fSys)

	// easy load directly in root
	content, err := ll.Load("kustomization.yaml")
	req.NoError(err)
	req.Equal([]byte("/a"), content)

	// typical sibling root reference
	sibLL, err := ll.New("../alpha")
	req.NoError(err)
	checkLoader(req, sibLL, "/alpha")

	// only need to test once, since don't need to call Cleanup() on local target
	req.NoError(ll.Cleanup())

	// file system checks are also one-time
	req.Equal(fSysCopy, fSys)
	req.Empty(buf.String())
}

func TestValidateLocArgsDefaultForRootTarget(t *testing.T) {
	cases := map[string]struct {
		target string
		scope  string
	}{
		"explicit": {
			"/",
			".",
		},
		"implicit": {
			".",
			"",
		},
	}
	for name, params := range cases {
		params := params
		t.Run(name, func(t *testing.T) {
			req := require.New(t)
			fSys := makeMemoryFs(t)

			ll, dst, err := lclzr.ValidateLocArgs(params.target, params.scope, "", fSys)
			req.NoError(err)
			checkLocLoader(req, ll, "/", "/")
			checkConfirmedDir(req, dst, "/"+dstPrefix, fSys)

			// file in root, but nested
			content, err := ll.Load("a/kustomization.yaml")
			req.NoError(err)
			req.Equal([]byte("/a"), content)

			childLL, err := ll.New("a")
			req.NoError(err)
			checkLoader(req, childLL, "/a")

			// messy, uncleaned path
			content, err = childLL.Load("./../a/kustomization.yaml")
			req.NoError(err)
			req.Equal([]byte("/a"), content)
		})
	}
}

func TestNewMultiple(t *testing.T) {
	req := require.New(t)
	fSys := makeMemoryFs(t)

	// default destination for non-file system root target
	// destination outside of scope
	ll, dst, err := lclzr.ValidateLocArgs("/alpha/beta", "/alpha", "", fSys)
	req.NoError(err)
	newDir := "/" + dstPrefix + "-beta"
	checkLocLoader(req, ll, "/alpha/beta", "/alpha")
	checkConfirmedDir(req, dst, newDir, fSys)

	// nested child root that isn't cleaned
	descLL, err := ll.New("../beta/gamma/delta")
	req.NoError(err)
	checkLoader(req, descLL, "/alpha/beta/gamma/delta")

	// upwards traversal
	higherLL, err := descLL.New("../../c")
	req.NoError(err)
	checkLoader(req, higherLL, "/alpha/beta/c")
}

func makeWdFs(t *testing.T) map[string]filesys.FileSystem {
	t.Helper()
	req := require.New(t)

	root := filesys.MakeEmptyDirInMemory()
	req.NoError(root.MkdirAll("a/b/c/d/e"))

	outer, err := root.Find("a")
	req.NoError(err)
	middle, err := root.Find("a/b/c")
	req.NoError(err)

	return map[string]filesys.FileSystem{
		"a":     outer,
		"a/b/c": middle,
	}
}

func TestValidateLocArgsCwdNotRoot(t *testing.T) {
	cases := map[string]struct {
		wd     string
		target string
		scope  string
		newDir string
	}{
		// target not immediate child of scope
		"outer dir": {
			"a",
			"b/c/d/e",
			"b/c",
			"b/newDir",
		},
		"scope": {
			"a/b/c",
			"d/e",
			".",
			"d/e/newDir",
		},
	}

	for name, test := range cases {
		test := test
		t.Run(name, func(t *testing.T) {
			req := require.New(t)
			fSys := makeWdFs(t)[test.wd]

			ll, dst, err := lclzr.ValidateLocArgs(test.target, test.scope, test.newDir, fSys)
			req.NoError(err)
			newDir := test.wd + "/" + test.newDir
			checkLocLoader(req, ll, "a/b/c/d/e", "a/b/c")

			req.Equal(newDir, dst.String())
			// memory file system can only find paths rooted at current node
			req.True(fSys.Exists(test.newDir))
		})
	}
}

func TestValidateLocArgsFails(t *testing.T) {
	cases := map[string]struct {
		target string
		scope  string
		dest   string
	}{
		"non-existent target": {
			"/b",
			"/",
			"/newDir",
		},
		"file target": {
			"/a/kustomization.yaml",
			"/",
			"/newDir",
		},
		"inner scope": {
			"/alpha",
			"/alpha/beta",
			"/newDir",
		},
		"side scope": {
			"/alpha",
			"/a",
			"/newDir",
		},
		"existing dst": {
			"/alpha",
			"/",
			"/a",
		},
	}
	for name, params := range cases {
		params := params
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			log.SetOutput(&buf)
			_, _, err := lclzr.ValidateLocArgs(params.target, params.scope, params.dest, makeMemoryFs(t))
			require.Error(t, err)
			require.Empty(t, buf.String())
		})
	}
}

func TestNewFails(t *testing.T) {
	t.Run("ValidateLocArgs", func(t *testing.T) {
		req := require.New(t)
		fSys := makeMemoryFs(t)

		ll, dst, err := lclzr.ValidateLocArgs("/alpha/beta/gamma", "alpha", "alpha/beta/gamma/newDir", fSys)
		req.NoError(err)
		checkLocLoader(req, ll, "/alpha/beta/gamma", "/alpha")
		checkConfirmedDir(req, dst, "/alpha/beta/gamma/newDir", fSys)
	})
	cases := map[string]string{
		"outside scope":     "../../../a",
		"at dst":            "newDir",
		"ancestor":          "../../beta",
		"non-existent root": "delt",
		"file":              "delta/kustomization.yaml",
	}
	for name, root := range cases {
		root := root
		t.Run(name, func(t *testing.T) {
			fSys := makeMemoryFs(t)

			ll, _, err := lclzr.ValidateLocArgs("/alpha/beta/gamma", "alpha", "alpha/beta/gamma/newDir", fSys)
			require.NoError(t, err)

			_, err = ll.New(root)
			require.Error(t, err)
		})
	}
}

func TestLoadFails(t *testing.T) {
	t.Run("ValidateLocArgs", func(t *testing.T) {
		req := require.New(t)
		fSys := makeMemoryFs(t)

		ll, dst, err := lclzr.ValidateLocArgs("a", "", "/a/b/newDir", fSys)
		req.NoError(err)
		checkLocLoader(req, ll, "/a", "/a")
		checkConfirmedDir(req, dst, "/a/b/newDir", fSys)
	})
	cases := map[string]string{
		"absolute path":     "/a/kustomization.yaml",
		"directory":         "b",
		"non-existent file": "kubectl.yaml",
		"file outside root": "../alpha/beta/gamma/delta/kustomization.yaml",
		"inside dst":        "newDir/kustomization.yaml",
	}
	for name, file := range cases {
		file := file
		t.Run(name, func(t *testing.T) {
			req := require.New(t)
			fSys := makeMemoryFs(t)

			ll, _, err := lclzr.ValidateLocArgs("./a/../a", "/a/../a", "/a/newDir", fSys)
			req.NoError(err)

			req.NoError(fSys.WriteFile("/a/newDir/kustomization.yaml", []byte("/a/newDir")))

			_, err = ll.Load(file)
			req.Error(err)
		})
	}
}
