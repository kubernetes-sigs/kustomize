// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"path/filepath"
	"strings"
	"testing"

	. "sigs.k8s.io/kustomize/api/internal/target"
	"sigs.k8s.io/kustomize/api/konfig"
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
