// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package oci_test

import (
	"archive/tar"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/internal/oci"
	"sigs.k8s.io/kustomize/api/internal/oci/ocitest"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const someRef = "registry.example.com/some-org/some-repo:v1.0.0"

func artifactFiles() map[string]string {
	return map[string]string{
		"kustomization.yaml":      "resources:\n- deployment.yaml\n",
		"deployment.yaml":         "kind: Deployment\n",
		"overlays/dev/patch.yaml": "kind: Deployment\n",
	}
}

// newRepoSpec returns the repo spec of url and the file system to pull it
// into. The pull dir is removed when the test ends.
func newRepoSpec(t *testing.T, url string) (*oci.RepoSpec, filesys.FileSystem) {
	t.Helper()
	fSys := filesys.MakeFsOnDisk()
	rs, err := oci.NewRepoSpecFromURL(url)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, rs.Cleaner(fSys)()) })
	return rs, fSys
}

func requireFiles(t *testing.T, fSys filesys.FileSystem, dir filesys.ConfirmedDir, files map[string]string) {
	t.Helper()
	for path, content := range files {
		data, err := fSys.ReadFile(dir.Join(path))
		require.NoError(t, err, path)
		require.Equal(t, content, string(data), path)
	}
}

// dialingTo returns a client that connects to addr whatever the host of the
// request, and verifies certificates against the system roots.
func dialingTo(addr string) *http.Client {
	dialer := &net.Dialer{}
	return &http.Client{Transport: &http.Transport{
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, addr)
		},
	}}
}

func TestPullerUsingRegistry(t *testing.T) {
	testcases := map[string]v1.Image{
		"oci uncompressed layer": ocitest.Files(t, artifactFiles()),
		"oci gzip layer": ocitest.Artifact(t, ocitest.Layer(t, artifactFiles()),
			types.OCILayer, types.OCIConfigJSON),
		"container image": ocitest.Image(t, artifactFiles()),
		"flux artifact":   ocitest.FluxArtifact(t, artifactFiles()),
	}
	for name, img := range testcases {
		t.Run(name, func(t *testing.T) {
			client := ocitest.NewRegistry(t)
			digest := ocitest.Push(t, client, someRef, img)

			t.Run("by tag", func(t *testing.T) {
				rs, fSys := newRepoSpec(t, "oci://registry.example.com/some-org/some-repo:v1.0.0//overlays/dev")
				require.NoError(t, oci.PullerUsingRegistry(rs, fSys, client))
				require.NotEqual(t, "/notPulled", rs.PullDir().String())
				require.Equal(t, rs.PullDir().Join("overlays/dev"), rs.AbsPath())
				requireFiles(t, fSys, rs.PullDir(), artifactFiles())
			})
			t.Run("by digest", func(t *testing.T) {
				rs, fSys := newRepoSpec(t, "oci://registry.example.com/some-org/some-repo@"+digest)
				require.NoError(t, oci.PullerUsingRegistry(rs, fSys, client))
				requireFiles(t, fSys, rs.PullDir(), artifactFiles())
			})
		})
	}
}

func TestPullerUsingRegistry_CleanerRemovesPullDir(t *testing.T) {
	client := ocitest.NewRegistry(t)
	ocitest.Push(t, client, someRef, ocitest.Files(t, artifactFiles()))

	fSys := filesys.MakeFsOnDisk()
	rs, err := oci.NewRepoSpecFromURL("oci://registry.example.com/some-org/some-repo:v1.0.0")
	require.NoError(t, err)
	require.NoError(t, oci.PullerUsingRegistry(rs, fSys, client))
	require.True(t, fSys.Exists(rs.PullDir().Join("kustomization.yaml")))
	require.NoError(t, rs.Cleaner(fSys)())
	require.False(t, fSys.Exists(rs.PullDir().String()))
}

func TestPullerUsingRegistry_MissingArtifact(t *testing.T) {
	client := ocitest.NewRegistry(t)
	fSys := filesys.MakeFsOnDisk()
	rs, err := oci.NewRepoSpecFromURL("oci://registry.example.com/some-org/missing:v1.0.0")
	require.NoError(t, err)
	err = oci.PullerUsingRegistry(rs, fSys, client)
	require.ErrorContains(t, err, "pulling oci artifact registry.example.com/some-org/missing:v1.0.0")
	// The pull dir is created before the pull and must be cleaned by the caller.
	require.True(t, fSys.Exists(rs.PullDir().String()))
	require.NoError(t, rs.Cleaner(fSys)())
	require.False(t, fSys.Exists(rs.PullDir().String()))
}

func TestPullerUsingRegistry_PlainHTTP(t *testing.T) {
	// This registry only serves plain HTTP.
	host := ocitest.StartLoopbackRegistry(t)
	ocitest.Push(t, nil, host+"/some-org/some-repo:v1.0.0", ocitest.Files(t, artifactFiles()))

	// As in go-containerregistry, plain HTTP is allowed for loopback addresses,
	rs, fSys := newRepoSpec(t, "oci://"+host+"/some-org/some-repo:v1.0.0")
	require.NoError(t, oci.PullerUsingRegistry(rs, fSys, nil))
	requireFiles(t, fSys, rs.PullDir(), artifactFiles())

	// but other hosts are only reached over HTTPS.
	rs, fSys = newRepoSpec(t, "oci://registry.example.com/some-org/some-repo:v1.0.0")
	err := oci.PullerUsingRegistry(rs, fSys, dialingTo(host))
	require.ErrorContains(t, err, "server gave HTTP response to HTTPS client")
}

func TestPullerUsingRegistry_UntrustedCertificate(t *testing.T) {
	// The certificate of the server is valid for *.example.com,
	// but it is not signed by any of the system roots.
	ocitest.IgnoreLocalCredentials(t)
	srv := httptest.NewTestServer(t, ocitest.Handler())
	srv.StartTLS()
	rs, fSys := newRepoSpec(t, "oci://registry.example.com/some-org/some-repo:v1.0.0")
	err := oci.PullerUsingRegistry(rs, fSys, dialingTo(srv.Listener.Addr().String()))
	require.ErrorContains(t, err, "tls: failed to verify certificate")
}

func TestPullerUsingRegistry_MalformedLayers(t *testing.T) {
	compressed, err := ocitest.Layer(t, artifactFiles()).Compressed()
	require.NoError(t, err)
	archive, err := io.ReadAll(compressed)
	require.NoError(t, err)
	require.NoError(t, compressed.Close())

	for name, tc := range map[string]struct {
		content   []byte
		mediaType types.MediaType
	}{
		"not a tar archive": {[]byte("not a tar archive"), types.OCIUncompressedLayer},
		"truncated gzip":    {archive[:len(archive)/2], types.OCILayer},
	} {
		t.Run(name, func(t *testing.T) {
			client := ocitest.NewRegistry(t)
			ocitest.Push(t, client, someRef, ocitest.Artifact(t,
				static.NewLayer(tc.content, tc.mediaType), tc.mediaType, types.OCIConfigJSON))
			rs, fSys := newRepoSpec(t, "oci://registry.example.com/some-org/some-repo:v1.0.0")
			err := oci.PullerUsingRegistry(rs, fSys, client)
			require.ErrorContains(t, err, "reading contents of oci artifact")
		})
	}
}

func TestPullerUsingRegistry_TamperedLayer(t *testing.T) {
	// Same size, different content.
	tampered := artifactFiles()
	tampered["deployment.yaml"] = "kind: DaemonSet!\n"
	for name, tc := range map[string]struct {
		original, replacement v1.Image
	}{
		"uncompressed": {ocitest.Files(t, artifactFiles()), ocitest.Files(t, tampered)},
		"gzip":         {ocitest.FluxArtifact(t, artifactFiles()), ocitest.FluxArtifact(t, tampered)},
	} {
		t.Run(name, func(t *testing.T) {
			ocitest.IgnoreLocalCredentials(t)
			client := httptest.NewTestServer(t, replacedBlob(ocitest.Handler(),
				firstLayer(t, tc.original), firstLayer(t, tc.replacement))).Client()
			ocitest.Push(t, client, someRef, tc.original)
			rs, fSys := newRepoSpec(t, "oci://registry.example.com/some-org/some-repo:v1.0.0")
			err := oci.PullerUsingRegistry(rs, fSys, client)
			require.ErrorContains(t, err, "verifying contents of oci artifact")
		})
	}
}

func firstLayer(t *testing.T, img v1.Image) v1.Layer {
	t.Helper()
	layers, err := img.Layers()
	require.NoError(t, err)
	return layers[0]
}

// replacedBlob serves the content of the replacement layer
// in place of the blob of the original one.
func replacedBlob(h http.Handler, original, replacement v1.Layer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		digest, err := original.Digest()
		if err != nil || r.Method != http.MethodGet ||
			!strings.HasSuffix(r.URL.Path, "/blobs/"+digest.String()) {
			h.ServeHTTP(w, r)
			return
		}
		content, err := replacement.Compressed()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer content.Close()
		_, _ = io.Copy(w, content)
	})
}

func TestPullerUsingRegistry_SkipsLinks(t *testing.T) {
	entries := map[string]ocitest.Entry{
		"escape": {Typeflag: tar.TypeSymlink, Linkname: "../../../etc"},
		"link":   {Typeflag: tar.TypeLink, Linkname: "deployment.yaml"},
		"empty/": {Typeflag: tar.TypeDir},
	}
	for path, content := range artifactFiles() {
		entries[path] = ocitest.Entry{Content: content}
	}
	client := ocitest.NewRegistry(t)
	ocitest.Push(t, client, someRef, ocitest.Artifact(t, ocitest.TarLayer(t, entries),
		types.OCIUncompressedLayer, types.OCIConfigJSON))

	rs, fSys := newRepoSpec(t, "oci://registry.example.com/some-org/some-repo:v1.0.0")
	require.NoError(t, oci.PullerUsingRegistry(rs, fSys, client))
	requireFiles(t, fSys, rs.PullDir(), artifactFiles())
	require.False(t, fSys.Exists(rs.PullDir().Join("escape")))
	require.False(t, fSys.Exists(rs.PullDir().Join("link")))
	require.True(t, fSys.IsDir(rs.PullDir().Join("empty")))
}

func TestPullerUsingRegistry_RejectsNonLocalEntries(t *testing.T) {
	for name, tc := range map[string]struct {
		path string
		errs []string // any of these
	}{
		// Rejected by go-containerregistry while flattening the layers
		// (newer releases) or by the puller's own check (older ones).
		"parent": {"../escape.yaml", []string{
			"reading contents of oci artifact", "is not local to the artifact"}},
		// Passed through by go-containerregistry, rejected here.
		"absolute": {"/escape.yaml", []string{"is not local to the artifact"}},
	} {
		t.Run(name, func(t *testing.T) {
			client := ocitest.NewRegistry(t)
			ocitest.Push(t, client, someRef, ocitest.Files(t, map[string]string{tc.path: "kind: Deployment\n"}))

			rs, fSys := newRepoSpec(t, "oci://registry.example.com/some-org/some-repo:v1.0.0")
			err := oci.PullerUsingRegistry(rs, fSys, client)
			require.Error(t, err)
			require.Condition(t, func() bool {
				for _, e := range tc.errs {
					if strings.Contains(err.Error(), e) {
						return true
					}
				}
				return false
			}, "error %q matches none of %q", err.Error(), tc.errs)
			require.Contains(t, err.Error(), "escape.yaml")
			_, err = os.Stat(filepath.Join(filepath.Dir(rs.PullDir().String()), "escape.yaml"))
			require.True(t, os.IsNotExist(err))
		})
	}
}
