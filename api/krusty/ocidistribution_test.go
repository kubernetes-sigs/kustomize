// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"errors"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/internal/oci/ocitest"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// distributionImage is the reference implementation of a registry, pinned
// by digest.
const distributionImage = "ghcr.io/distribution/distribution:3.1.1" +
	"@sha256:bca24727f4002e51f959c18c42e816e4d1078198081a9837e16b8b7d7e43ebf8"

// docker runs the docker CLI and returns what it prints.
func docker(t *testing.T, args ...string) string {
	t.Helper()
	out, err := exec.Command("docker", args...).Output()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		t.Fatalf("docker %s: %v\n%s", strings.Join(args, " "), err, exitErr.Stderr)
	}
	require.NoError(t, err)
	return strings.TrimSpace(string(out))
}

// startDistributionRegistry runs a distribution registry in a container, on
// a loopback port, until the test ends, and returns its host. Like the docker
// e2e tests of cmd/config, it skips the test without docker or when
// KUSTOMIZE_DOCKER_E2E is false.
func startDistributionRegistry(t *testing.T) string {
	t.Helper()
	skipIfNoDocker(t)
	if os.Getenv("KUSTOMIZE_DOCKER_E2E") == "false" {
		t.Skip("skipping because KUSTOMIZE_DOCKER_E2E is false")
	}
	id := docker(t, "run", "--detach", "--rm", "--publish", "127.0.0.1::5000", distributionImage)
	t.Cleanup(func() {
		_ = exec.Command("docker", "rm", "--force", id).Run()
	})
	host, _, _ := strings.Cut(docker(t, "port", id, "5000/tcp"), "\n")
	// The registry takes a moment to listen.
	require.Eventually(t, func() bool {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+host+"/v2/", nil)
		if err != nil {
			return false
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return false
		}
		_ = resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, time.Minute, 100*time.Millisecond, "no answer from the registry at %s", host)
	return host
}

func TestOciArtifactFromDistributionRegistry(t *testing.T) {
	// The build pulls a flux artifact, by tag and by digest, from the
	// reference implementation of a registry rather than the in-process one.
	host := startDistributionRegistry(t)
	ocitest.IgnoreLocalCredentials(t)
	digest := ocitest.Push(t, nil, host+ociRepo, ocitest.FluxArtifact(t, ociArtifactFiles()))

	th := kusttest_test.MakeHarnessWithFs(t, filesys.MakeFsOnDisk())
	dir := t.TempDir()
	th.WriteK(dir, resourcesPrefix+ociScheme+host+ociRepo+ociOverlay+"\n- "+
		ociScheme+host+"/some-org/some-repo@"+digest+"//base\n")

	m := th.Run(dir, th.MakeDefaultOptions())
	yml, err := m.AsYaml()
	require.NoError(t, err)
	require.Equal(t, ociDevBuild+"---\n"+ociBaseBuild, string(yml))
}
