// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	. "sigs.k8s.io/kustomize/api/internal/target"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/krusty"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestTargetMustHaveKustomizationFile(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: aService
`)
	th.WriteF("deeper/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: anotherService
`)
	err := th.RunWithErr(".", th.MakeDefaultOptions())
	if err == nil {
		t.Fatalf("expected an error")
	}
	if !IsMissingKustomizationFileError(err) {
		t.Fatalf("unexpected error: %q", err)
	}
}

func TestTargetMustHaveOnlyOneKustomizationFile(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	for _, n := range konfig.RecognizedKustomizationFileNames() {
		th.WriteF(filepath.Join(".", n), `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`)
	}
	err := th.RunWithErr(".", th.MakeDefaultOptions())
	if err == nil {
		t.Fatalf("expected an error")
	}
	if !strings.Contains(err.Error(), "Found multiple kustomization files") {
		t.Fatalf("unexpected error: %q", err)
	}
}

func TestBaseMustHaveKustomizationFile(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- base
`)
	th.WriteF("base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  selector:
    backend: bungie
  ports:
    - port: 7002
`)
	err := th.RunWithErr(".", th.MakeDefaultOptions())
	if err == nil {
		t.Fatalf("expected an error")
	}
	if !strings.Contains(err.Error(), "accumulating resources") {
		t.Fatalf("unexpected error: %q", err)
	}
}

func TestResourceNotFound(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- deployment.yaml
`)
	err := th.RunWithErr(".", th.MakeDefaultOptions())
	if err == nil {
		t.Fatalf("expected an error")
	}
	if !strings.Contains(err.Error(), "accumulating resources") {
		t.Fatalf("unexpected error: %q", err)
	}
}

func TestResourceHasAnchor(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- ingress.yaml
`)
	th.WriteF("ingress.yaml", `
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: blog
spec:
  tls:
  - hosts:
    - xyz.me
    - www.xyz.me
    secretName: cert-tls
  rules:
  - host: xyz.me
    http: &xxx_rules
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: service
            port:
              number: 80
  - host: www.xyz.me
    http: *xxx_rules
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: blog
spec:
  rules:
  - host: xyz.me
    http:
      paths:
      - backend:
          service:
            name: service
            port:
              number: 80
        path: /
        pathType: Prefix
  - host: www.xyz.me
    http:
      paths:
      - backend:
          service:
            name: service
            port:
              number: 80
        path: /
        pathType: Prefix
  tls:
  - hosts:
    - xyz.me
    - www.xyz.me
    secretName: cert-tls
`)
}

func TestAccumulateResourcesErrors(t *testing.T) {
	type testcase struct {
		name     string
		resource string
		// regex error message that is output when kustomize tries to
		// accumulate resource as file and dir, respectively
		errFile, errDir string
	}
	buildError := func(tc testcase) string {
		const (
			prefix            = "accumulating resources"
			filePrefixf       = "accumulating resources from '%s'"
			fileWrapperIfDirf = "accumulation err='%s'"
			separator         = ": "
		)
		parts := []string{
			prefix,
			strings.Join([]string{
				fmt.Sprintf(filePrefixf, regexp.QuoteMeta(tc.resource)),
				tc.errFile,
			}, separator),
		}
		if tc.errDir != "" {
			parts[1] = fmt.Sprintf(fileWrapperIfDirf, parts[1])
			parts = append(parts, tc.errDir)
		}
		return strings.Join(parts, separator)
	}
	for _, test := range []testcase{
		{
			name: "remote file not considered repo",
			// This url is too short to be considered a remote repo.
			resource: "https://raw.githubusercontent.com/kustomize",
			// It is acceptable that the error for a remote file-like reference
			// (that is not long enough to be considered a repo) does not
			// indicate said reference is not a local directory.
			// Though it is possible for the remote file-like reference to be
			// a local directory, it is very unlikely.
			errFile: `HTTP Error: status code 400 \(Bad Request\)\z`,
		},
		{
			name:     "remote file qualifies as repo",
			resource: "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/kustomize/v5.0.0/examples/helloWorld/configMap",
			// TODO(4788): This error message is technically wrong. Just
			// because we fail to GET a reference does not mean the reference is
			// not a remote file. We should return the GET status code instead.
			errFile: "URL is a git repository",
			errDir:  `failed to run \S+/git fetch --depth=1 .+`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			// should use real file system to indicate that we are creating
			// new temporary directories on disk when we attempt to fetch repos
			fSys, tmpDir := kusttest_test.CreateKustDir(t, fmt.Sprintf(`
resources:
- %s
`, test.resource))
			b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
			_, err := b.Run(fSys, tmpDir.String())
			require.Regexp(t, buildError(test), err.Error())
		})
	}
	// TODO(annasong): add tests that check accumulateResources errors for
	// - local files
	// - repos
	// - local directories
	// - files that yield malformed yaml errors
}
