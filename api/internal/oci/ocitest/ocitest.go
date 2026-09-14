// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package ocitest provides an in-process OCI registry and helpers
// to push test artifacts into it.
package ocitest

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"net/http/httptest"
	"net/url"
	"sort"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/stretchr/testify/require"
)

const fileMode = 0o644

// Media types used by `flux push artifact`.
const (
	FluxConfigMediaType = types.MediaType("application/vnd.cncf.flux.config.v1+json")
	FluxLayerMediaType  = types.MediaType("application/vnd.cncf.flux.content.v1.tar+gzip")
)

// Entry is a file in a test artifact.
type Entry struct {
	// Typeflag is the tar entry type, tar.TypeReg if zero.
	Typeflag byte
	// Content of a regular file.
	Content string
	// Linkname of a link.
	Linkname string
}

// Artifact describes an artifact to push.
type Artifact struct {
	// Entries by path within the artifact.
	Entries map[string]Entry
	// Compress the layer with gzip.
	Compress bool
	// LayerMediaType of the single layer, types.OCIUncompressedLayer
	// (or types.OCILayer if compressed) when empty.
	LayerMediaType types.MediaType
	// ConfigMediaType of the image config, types.OCIConfigJSON when empty.
	ConfigMediaType types.MediaType
}

// Files returns an Artifact holding the given regular files.
func Files(files map[string]string) Artifact {
	entries := make(map[string]Entry, len(files))
	for path, content := range files {
		entries[path] = Entry{Content: content}
	}
	return Artifact{Entries: entries}
}

// FluxArtifact returns an Artifact holding the given regular files,
// shaped like the output of `flux push artifact`.
func FluxArtifact(files map[string]string) Artifact {
	a := Files(files)
	a.Compress = true
	a.LayerMediaType = FluxLayerMediaType
	a.ConfigMediaType = FluxConfigMediaType
	return a
}

// StartRegistry starts an in-process registry that is stopped
// when the test ends, and returns its host (address and port).
func StartRegistry(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(registry.New(registry.Logger(log.New(io.Discard, "", 0))))
	t.Cleanup(srv.Close)
	u, err := url.Parse(srv.URL)
	require.NoError(t, err)
	return u.Host
}

// Push pushes the artifact to the given reference,
// returning the digest of its manifest.
func Push(t *testing.T, ref string, a Artifact) string {
	t.Helper()
	reference, err := name.ParseReference(ref)
	require.NoError(t, err)
	img := Image(t, a)
	require.NoError(t, remote.Write(reference, img))
	digest, err := img.Digest()
	require.NoError(t, err)
	return digest.String()
}

// Image builds the image of the artifact.
func Image(t *testing.T, a Artifact) v1.Image {
	t.Helper()
	content := Tar(t, a.Entries)
	layerMediaType := a.LayerMediaType
	if a.Compress {
		content = gzipped(t, content)
		if layerMediaType == "" {
			layerMediaType = types.OCILayer
		}
	} else if layerMediaType == "" {
		layerMediaType = types.OCIUncompressedLayer
	}
	configMediaType := a.ConfigMediaType
	if configMediaType == "" {
		configMediaType = types.OCIConfigJSON
	}
	img := mutate.MediaType(empty.Image, types.OCIManifestSchema1)
	img = mutate.ConfigMediaType(img, configMediaType)
	img, err := mutate.AppendLayers(img, static.NewLayer(content, layerMediaType))
	require.NoError(t, err)
	return img
}

// Tar returns a tar archive of the entries, in path order.
func Tar(t *testing.T, entries map[string]Entry) []byte {
	t.Helper()
	paths := make([]string, 0, len(entries))
	for path := range entries {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, path := range paths {
		entry := entries[path]
		header := &tar.Header{
			Typeflag: entry.Typeflag,
			Name:     path,
			Linkname: entry.Linkname,
			Mode:     fileMode,
		}
		if header.Typeflag == 0 {
			header.Typeflag = tar.TypeReg
		}
		if header.Typeflag == tar.TypeReg {
			header.Size = int64(len(entry.Content))
		}
		require.NoError(t, tw.WriteHeader(header))
		if header.Typeflag == tar.TypeReg {
			_, err := tw.Write([]byte(entry.Content))
			require.NoError(t, err)
		}
	}
	require.NoError(t, tw.Close())
	return buf.Bytes()
}

func gzipped(t *testing.T, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write(content)
	require.NoError(t, err)
	require.NoError(t, gz.Close())
	return buf.Bytes()
}
