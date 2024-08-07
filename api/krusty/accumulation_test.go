// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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

const validResource = `
apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  selector:
    backend: bungie
  ports:
    - port: 7002
`

const invalidResource = `apiVersion: v1
kind: Service
metadata:
  name: kapacitor
  labels:
    app.kubernetes.io/name: tick-kapacitor
spec:
  selector:
    app.kubernetes.io/name: tick-kapacitor
    - name: http
      port: 9092
      protocol: TCP
  type: ClusterIP`

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
	th.WriteF("base/service.yaml", validResource)
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
		// resourceFunc generates a resource string using the URL to the local
		// test server (optional).
		resourceFunc func(string) string
		// resourceServerSetup configures the local test server (optional).
		resourceServerSetup func(*http.ServeMux)
		isAbsolute          bool
		files               map[string]string
		// errFile, errDir are regex for the expected error message output
		// when kustomize tries to accumulate resource as file and dir,
		// respectively. The test substitutes occurrences of "%s" in the
		// error strings with the absolute path where kustomize looks for it.
		errFile, errDir string
	}
	populateAbsolutePaths := func(tc testcase, dir string) testcase {
		filePaths := make(map[string]string, len(tc.files)+1)
		for file, content := range tc.files {
			filePaths[filepath.Join(dir, file)] = content
		}
		resourcePath := filepath.Join(dir, tc.resource)
		if tc.isAbsolute {
			tc.resource = resourcePath
		}
		filePaths[filepath.Join(dir, "kustomization.yaml")] = fmt.Sprintf(`
resources:
- %s
`, tc.resource)
		tc.files = filePaths
		regPath := regexp.QuoteMeta(resourcePath)
		tc.errFile = strings.ReplaceAll(tc.errFile, "%s", regPath)
		tc.errDir = strings.ReplaceAll(tc.errDir, "%s", regPath)
		return tc
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
			resourceFunc: func(url string) string {
				return fmt.Sprintf("%s/segments-too-few-to-be-repo", url)
			},
			resourceServerSetup: func(server *http.ServeMux) {
				server.HandleFunc("/", func(out http.ResponseWriter, req *http.Request) {
					out.WriteHeader(http.StatusNotFound)
				})
			},
			// It's acceptable for the error output of a remote file-like
			// resource to not indicate the resource's status as a
			// local directory. Though it is possible for a remote file-like
			// resource to be a local directory, it is very unlikely.
			errFile: `HTTP Error: status code 404 \(Not Found\)\z`,
		},
		{
			name: "remote file qualifies as repo",
			resourceFunc: func(url string) string {
				return fmt.Sprintf("%s/long/enough/to/have/org/and/repo", url)
			},
			resourceServerSetup: func(server *http.ServeMux) {
				server.HandleFunc("/", func(out http.ResponseWriter, req *http.Request) {
					out.WriteHeader(http.StatusInternalServerError)
				})
			},
			// TODO(4788): This error message is technically wrong. Just
			// because we fail to GET a resource does not mean the resource is
			// not a remote file. We should return the GET status code as well.
			errFile: "URL is a git repository",
			errDir:  `failed to run \S+/git fetch --depth=1 .+`,
		},
		{
			name: "local file qualifies as repo",
			// The .example top level domain is reserved for example purposes,
			// see RFC 2606.
			resource: "package@v1.28.0.example/configs/base",
			errFile:  `evalsymlink failure on '%s' .+`,
			errDir:   `failed to run \S+/git fetch --depth=1 .+`,
		},
		{
			name:     "relative path does not exist",
			resource: "file-or-directory",
			errFile:  `evalsymlink failure on '%s' .+`,
			errDir:   `must build at directory: not a valid directory: evalsymlink failure .+`,
		},
		{
			name:       "absolute path does not exist",
			resource:   "file-or-directory",
			isAbsolute: true,
			errFile:    `evalsymlink failure on '%s' .+`,
			errDir:     `new root '%s' cannot be absolute`,
		},
		{
			name:     "relative file violates restrictions",
			resource: "../base/resource.yaml",
			files: map[string]string{
				"../base/resource.yaml": validResource,
			},
			errFile: "security; file '%s' is not in or below .+",
			// TODO(4348): Over-inclusion of directory error message when we
			// know resource is file.
			errDir: "must build at directory: '%s': file is not directory",
		},
		{
			name:       "absolute file violates restrictions",
			resource:   "../base/resource.yaml",
			isAbsolute: true,
			files: map[string]string{
				"../base/resource.yaml": validResource,
			},
			errFile: "security; file '%s' is not in or below .+",
			// TODO(4348): Over-inclusion of directory error message when we
			// know resource is file.
			errDir: `new root '%s' cannot be absolute`,
		},
		{
			name:     "malformed yaml yields an error",
			resource: "service.yaml",
			files: map[string]string{
				"service.yaml": invalidResource,
			},
			errFile: "MalformedYAMLError",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if test.resourceFunc != nil {
				// Configure test server handler
				handler := http.NewServeMux()
				if test.resourceServerSetup != nil {
					test.resourceServerSetup(handler)
				}
				// Start test server
				svr := httptest.NewServer(handler)
				defer svr.Close()
				// Generate resource with test server address
				test.resource = test.resourceFunc(svr.URL)
			}

			// Should use real file system to indicate that we are creating
			// new temporary directories on disk when we attempt to fetch repos.
			fs, tmpDir := kusttest_test.Setup(t)
			root := tmpDir.Join("root")
			require.NoError(t, fs.Mkdir(root))

			test = populateAbsolutePaths(test, root)
			for file, content := range test.files {
				dir := filepath.Dir(file)
				require.NoError(t, fs.MkdirAll(dir))
				require.NoError(t, fs.WriteFile(file, []byte(content)))
			}

			b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
			_, err := b.Run(fs, root)
			require.Regexp(t, buildError(test), err.Error())
		})
	}
	// TODO(annasong): add tests that check accumulateResources errors for
	// - repos
	// - local directories
}
