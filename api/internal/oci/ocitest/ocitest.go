// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package ocitest provides in-process OCI registries and helpers
// to build test artifacts and push them.
package ocitest

import (
	"archive/tar"
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"testing"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/registry"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
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

// Handler returns the handler of an in-process registry that does not log.
func Handler() http.Handler {
	return registry.New(registry.Logger(log.New(io.Discard, "", 0)))
}

// IgnoreLocalCredentials points DOCKER_CONFIG at an empty directory for the
// rest of the test, so that pulls through the default keychain are anonymous
// and do not run the local credential helpers, which may block.
func IgnoreLocalCredentials(t *testing.T) {
	t.Helper()
	t.Setenv("DOCKER_CONFIG", t.TempDir())
}

// NewRegistry starts an in-process registry on the in-memory network of
// httptest.NewTestServer, stopped when the test ends, and returns the client
// that reaches it. The client sends every request to the registry, whatever
// its host. The registry only serves HTTPS; the client does not verify its
// certificate. Local credentials are ignored for the rest of the test.
func NewRegistry(t *testing.T) *http.Client {
	t.Helper()
	IgnoreLocalCredentials(t)
	return httptest.NewTestServer(t, httpsOnly(Handler())).Client()
}

// httpsOnly rejects the requests that do not use TLS.
func httpsOnly(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.TLS == nil {
			http.Error(w, "plain HTTP is not served", http.StatusBadRequest)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// StartLoopbackRegistry starts an in-process registry serving plain HTTP on
// the loopback interface, stopped when the test ends, and returns its host
// (address and port). The default client reaches it, as go-containerregistry
// allows plain HTTP for loopback addresses. Local credentials are ignored for
// the rest of the test.
func StartLoopbackRegistry(t *testing.T) string {
	t.Helper()
	IgnoreLocalCredentials(t)
	srv := httptest.NewTestServer(t, Handler())
	srv.Start()
	u, err := url.Parse(srv.URL)
	require.NoError(t, err)
	return u.Host
}

// Push pushes the image to ref anonymously, through the given client, nil
// for the default one, and returns the digest of its manifest.
func Push(t *testing.T, client *http.Client, ref string, img v1.Image) string {
	t.Helper()
	options := []crane.Option{crane.WithAuth(authn.Anonymous)}
	if client != nil {
		options = append(options, crane.WithTransport(client.Transport))
	}
	require.NoError(t, crane.Push(img, ref, options...))
	digest, err := img.Digest()
	require.NoError(t, err)
	return digest.String()
}

// Image returns a container image holding the given regular files in a
// single layer, as crane builds it: gzip-compressed, with Docker media types.
func Image(t *testing.T, files map[string]string) v1.Image {
	t.Helper()
	img, err := crane.Image(fileMap(files))
	require.NoError(t, err)
	return img
}

// Files returns an OCI artifact holding the given regular files
// in a single uncompressed layer.
func Files(t *testing.T, files map[string]string) v1.Image {
	t.Helper()
	return Artifact(t, UncompressedLayer(t, files), types.OCIUncompressedLayer, types.OCIConfigJSON)
}

// FluxArtifact returns an artifact holding the given regular files,
// shaped like the output of `flux push artifact`.
func FluxArtifact(t *testing.T, files map[string]string) v1.Image {
	t.Helper()
	return Artifact(t, Layer(t, files), FluxLayerMediaType, FluxConfigMediaType)
}

// Artifact returns an OCI artifact holding the given layer, with the
// given media types for the layer and the config.
func Artifact(t *testing.T, layer v1.Layer, layerMediaType, configMediaType types.MediaType) v1.Image {
	t.Helper()
	img := mutate.MediaType(empty.Image, types.OCIManifestSchema1)
	img = mutate.ConfigMediaType(img, configMediaType)
	img, err := mutate.Append(img, mutate.Addendum{Layer: layer, MediaType: layerMediaType})
	require.NoError(t, err)
	return img
}

// Layer returns a gzip-compressed tar layer holding the given regular files.
func Layer(t *testing.T, files map[string]string) v1.Layer {
	t.Helper()
	layer, err := crane.Layer(fileMap(files))
	require.NoError(t, err)
	return layer
}

// UncompressedLayer returns an uncompressed tar layer holding the given
// regular files.
func UncompressedLayer(t *testing.T, files map[string]string) v1.Layer {
	t.Helper()
	content, err := Layer(t, files).Uncompressed()
	require.NoError(t, err)
	defer content.Close()
	archive, err := io.ReadAll(content)
	require.NoError(t, err)
	return static.NewLayer(archive, types.OCIUncompressedLayer)
}

// Entry is an entry of a layer built by TarLayer.
type Entry struct {
	// Typeflag is the tar entry type, tar.TypeReg if zero.
	Typeflag byte
	// Content of a regular file.
	Content string
	// Linkname of a link.
	Linkname string
}

// TarLayer returns an uncompressed tar layer holding the given entries by
// path, in path order. Unlike Layer, it can hold links and directories.
func TarLayer(t *testing.T, entries map[string]Entry) v1.Layer {
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
	return static.NewLayer(buf.Bytes(), types.OCIUncompressedLayer)
}

func fileMap(files map[string]string) map[string][]byte {
	contents := make(map[string][]byte, len(files))
	for path, content := range files {
		contents[path] = []byte(content)
	}
	return contents
}
