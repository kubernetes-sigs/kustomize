// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/provenance"
	. "sigs.k8s.io/kustomize/kustomize/v4/commands/build"
)

func loadFileSystem(fSys filesys.FileSystem) {
	fSys.WriteFile(konfig.DefaultKustomizationFileName(), []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namePrefix: foo-
nameSuffix: -bar
namespace: ns1
commonLabels:
  app: nginx
commonAnnotations:
  note: This is a test annotation
resources:
  - deployment.yaml
  - namespace.yaml
configMapGenerator:
- name: literalConfigMap
  literals:
  - DB_USERNAME=admin
  - DB_PASSWORD=somepw
secretGenerator:
- name: secret
  literals:
    - DB_USERNAME=admin
    - DB_PASSWORD=somepw
  type: Opaque
patchesJson6902:
- target:
    group: apps
    version: v1
    kind: Deployment
    name: dply1
  path: jsonpatch.json
`))
	fSys.WriteFile("deployment.yaml", []byte(`
apiVersion: apps/v1
metadata:
  name: dply1
kind: Deployment
`))
	fSys.WriteFile("namespace.yaml", []byte(`
apiVersion: v1
kind: Namespace
metadata:
  name: ns1
`))
	fSys.WriteFile("jsonpatch.json", []byte(`[
    {"op": "add", "path": "/spec/replica", "value": "3"}
]`))
}

const expectedContent = `apiVersion: v1
kind: Namespace
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: nginx
  name: ns1
---
apiVersion: v1
data:
  DB_PASSWORD: somepw
  DB_USERNAME: admin
kind: ConfigMap
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: nginx
  name: foo-literalConfigMap-bar-g5f6t456f5
  namespace: ns1
---
apiVersion: v1
data:
  DB_PASSWORD: c29tZXB3
  DB_USERNAME: YWRtaW4=
kind: Secret
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: nginx
  name: foo-secret-bar-82c2g5f8f6
  namespace: ns1
type: Opaque
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: nginx
  name: foo-dply1-bar
  namespace: ns1
spec:
  replica: "3"
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      annotations:
        note: This is a test annotation
      labels:
        app: nginx
`

func TestBuild(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	loadFileSystem(fSys)
	buffy := new(bytes.Buffer)
	cmd := NewCmdBuild(fSys, MakeHelp("foo", "bar"), buffy)
	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatal(err)
	}
	if buffy.String() != expectedContent {
		t.Fatalf("Expected output:\n%s\n But got output:\n%s", expectedContent, buffy)
	}
}

func TestBuildWithShardedOutput(t *testing.T) {
	var err error
	fSys := filesys.MakeFsInMemory()
	loadFileSystem(fSys)
	fSys.Mkdir("someDir")
	buffy := new(bytes.Buffer)
	cmd := NewCmdBuild(fSys, MakeHelp("foo", "bar"), buffy)
	cmd.Flags().Set("output", "someDir")
	cmd.Flags().Set("enable-managedby-label", "true")
	if err = cmd.RunE(cmd, []string{}); err != nil {
		t.Fatal(err)
	}
	if buffy.String() != "" {
		t.Fatalf("Expected:\n%s\nBut got:\n%s\n", expectedContent, buffy)
	}
	if !fSys.IsDir("someDir") {
		t.Fatal("expected dir")
	}
	var data []byte
	if data, err = fSys.ReadFile(
		"someDir/v1_namespace_ns1.yaml"); err != nil {
		t.Fatal(err)
	}
	version := provenance.GetProvenance().Version
	expected := fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: nginx
    app.kubernetes.io/managed-by: kustomize-%s
  name: ns1
`, version)
	if string(data) != expected {
		t.Fatalf("Expected:\n%s\nBut got:\n%s\n", expected, string(data))
	}
	if data, err = fSys.ReadFile(
		"someDir/v1_secret_foo-secret-bar-82c2g5f8f6.yaml"); err != nil {
		t.Fatal(err)
	}
	expected = fmt.Sprintf(`apiVersion: v1
data:
  DB_PASSWORD: c29tZXB3
  DB_USERNAME: YWRtaW4=
kind: Secret
metadata:
  annotations:
    note: This is a test annotation
  labels:
    app: nginx
    app.kubernetes.io/managed-by: kustomize-%s
  name: foo-secret-bar-82c2g5f8f6
  namespace: ns1
type: Opaque
`, version)
	if string(data) != expected {
		t.Fatalf("Expected:\n%s\nBut got:\n%s\n", expected, string(data))
	}
}

func TestHelp(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	buffy := new(bytes.Buffer)
	cmd := NewCmdBuild(fSys, MakeHelp("foo", "bar"), buffy)
	if cmd.Use != "bar DIR" {
		t.Fatalf("Unexpected usage: %s\n", cmd.Use)
	}
	if cmd.Short != "Build a kustomization target from a directory or URL." {
		t.Fatalf("Unexpected short help: %s\n", cmd.Short)
	}
	if !strings.Contains(cmd.Long, "If DIR is omitted, '.' is assumed.") {
		t.Fatalf("Unexpected long help: %s\n", cmd.Long)
	}
	if !strings.Contains(cmd.Example, "foo bar /home/config/production") {
		t.Fatalf("Unexpected example: %s\n", cmd.Example)
	}
}

func TestValidation(t *testing.T) {
	var cases = map[string]struct {
		args  []string
		erMsg string
	}{
		"noArgs":    {[]string{}, "unable to find one of "},
		"dotArg":    {[]string{"."}, "unable to find one of "},
		"file":      {[]string{"beans"}, "'beans' doesn't exist"},
		"directory": {[]string{"a/b/c"}, "'a/b/c' doesn't exist"},
		"tooManyArgs": {[]string{"too", "many"},
			"specify one path to " +
				konfig.DefaultKustomizationFileName()},
	}
	for n := range cases {
		tc := cases[n]
		t.Run(n, func(t *testing.T) {
			fSys := filesys.MakeFsInMemory()
			buffy := new(bytes.Buffer)
			cmd := NewCmdBuild(fSys, MakeHelp("foo", "bar"), buffy)
			err := cmd.RunE(cmd, tc.args)
			if len(tc.erMsg) > 0 {
				if err == nil {
					t.Errorf("%s: Expected an error %v", n, tc.erMsg)
				}
				if !strings.Contains(err.Error(), tc.erMsg) {
					t.Errorf("%s: Expected error %s, but got %v",
						n, tc.erMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("%s: unknown error: %v", n, err)
				}
			}
		})
	}
}
