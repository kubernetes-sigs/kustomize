// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer_test

import (
	"bytes"
	"log"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	lclzr "sigs.k8s.io/kustomize/api/internal/localizer"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func makeMemoryFs(t *testing.T) filesys.FileSystem {
	t.Helper()
	req := require.New(t)

	fSys := filesys.MakeFsInMemory()
	req.NoError(fSys.Mkdir("/a"))
	req.NoError(fSys.WriteFile("/a/kustomization.yaml", []byte("file")))
	req.NoError(fSys.MkdirAll("/alpha/beta/gamma"))
	return fSys
}

func TestNewLocLoaderLocalTarget(t *testing.T) {
	cases := map[string]string{
		"absolute": "/a",
		"relative": "a",
	}
	for name, target := range cases {
		target := target
		t.Run(name, func(t *testing.T) {
			req := require.New(t)

			fSys := makeMemoryFs(t)

			var buf bytes.Buffer
			log.SetOutput(&buf)
			locLdr, err := lclzr.NewLocLoader(target, "/", "/newDir", fSys)
			req.NoError(err)

			fSysCopy := makeMemoryFs(t)
			req.Equal(fSysCopy, fSys)

			locLdr.Cleanup()
			req.Equal(fSysCopy, fSys)
			req.Empty(buf.String())
		})
	}
}

func TestNewLocLoaderDefaultScope(t *testing.T) {
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

			_, err := lclzr.NewLocLoader(params.target, params.scope, "/a/newDir", fSys)
			req.NoError(err)
		})
	}
}

func TestNewLocLoaderDefaultDst(t *testing.T) {
	_, err := lclzr.NewLocLoader("/alpha/beta", "/alpha", "", makeMemoryFs(t))
	require.NoError(t, err)
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

func TestNewLocLoaderCwdNotRoot(t *testing.T) {
	cases := map[string]struct {
		wd     string
		target string
		scope  string
		dest   string
	}{
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
			dest := filepath.Join(test.wd, test.dest)

			_, err := lclzr.NewLocLoader(test.target, test.scope, dest, fSys)
			req.NoError(err)
		})
	}
}

func TestNewLocLoaderFails(t *testing.T) {
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
	}
	for name, params := range cases {
		params := params
		t.Run(name, func(t *testing.T) {
			fSys := makeMemoryFs(t)

			_, err := lclzr.NewLocLoader(params.target, params.scope, params.dest, fSys)
			require.Error(t, err)
		})
	}
}
