// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package target_test

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	"sigs.k8s.io/kustomize/api/types"
)

// KustTarget is primarily tested in the krusty package with
// high level tests.

func TestLoad(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	expectedTypeMeta := types.TypeMeta{
		APIVersion: "kustomize.config.k8s.io/v1beta1",
		Kind:       "Kustomization",
	}

	testCases := map[string]struct {
		errContains string
		content     string
		k           types.Kustomization
	}{
		"empty": {
			// no content
			k: types.Kustomization{
				TypeMeta: expectedTypeMeta,
			},
		},
		"nonsenseLatin": {
			errContains: "error converting YAML to JSON",
			content: `
		Lorem ipsum dolor sit amet, consectetur
		adipiscing elit, sed do eiusmod tempor
		incididunt ut labore et dolore magna aliqua.
		Ut enim ad minim veniam, quis nostrud
		exercitation ullamco laboris nisi ut
		aliquip ex ea commodo consequat.
		`,
		},
		"simple": {
			content: `
commonLabels:
  app: nginx
`,
			k: types.Kustomization{
				TypeMeta:     expectedTypeMeta,
				CommonLabels: map[string]string{"app": "nginx"},
			},
		},
		"commented": {
			content: `
# Licensed to the Blah Blah Software Foundation
# ...
# yada yada yada.

commonLabels:
 app: nginx
`,
			k: types.Kustomization{
				TypeMeta:     expectedTypeMeta,
				CommonLabels: map[string]string{"app": "nginx"},
			},
		},
	}

	kt := makeKustTargetWithRf(
		t, th.GetFSys(), "/",
		provider.NewDefaultDepProvider(), 1)
	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			th.WriteK("/", tc.content)
			err := kt.Load()
			if tc.errContains != "" {
				require.NotNilf(t, err, "expected error containing: `%s`", tc.errContains)
				require.Contains(t, err.Error(), tc.errContains)
			} else {
				require.Nilf(t, err, "got error: %v", err)
				k := kt.Kustomization()
				require.Condition(t, func() bool {
					return reflect.DeepEqual(tc.k, k)
				}, "expected %v, got %v", tc.k, k)
			}
		})
	}

	// Repeat test with parallel accumulation
	kt = makeKustTargetWithRf(
		t, th.GetFSys(), "/",
		provider.NewDefaultDepProvider(), 16)
	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			th.WriteK("/", tc.content)
			err := kt.Load()
			if tc.errContains != "" {
				require.NotNilf(t, err, "expected error containing: `%s`", tc.errContains)
				require.Contains(t, err.Error(), tc.errContains)
			} else {
				require.Nilf(t, err, "got error: %v", err)
				k := kt.Kustomization()
				require.Condition(t, func() bool {
					return reflect.DeepEqual(tc.k, k)
				}, "expected %v, got %v", tc.k, k)
			}
		})
	}
}

func TestMakeCustomizedResMap(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/whatever", `
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
generatorOptions:
  disableNameSuffixHash: false
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
`)
	th.WriteF("/whatever/deployment.yaml", `
apiVersion: apps/v1
metadata:
  name: dply1
kind: Deployment
`)
	th.WriteF("/whatever/namespace.yaml", `
apiVersion: v1
kind: Namespace
metadata:
  name: ns1
`)
	th.WriteF("/whatever/jsonpatch.json", `[
    {"op": "add", "path": "/spec/replica", "value": "3"}
]`)

	pvd := provider.NewDefaultDepProvider()
	resFactory := pvd.GetResourceFactory()

	resources := []*resource.Resource{
		resFactory.FromMapWithName("dply1", map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "foo-dply1-bar",
				"namespace": "ns1",
				"labels": map[string]interface{}{
					"app": "nginx",
				},
				"annotations": map[string]interface{}{
					"note": "This is a test annotation",
				},
			},
			"spec": map[string]interface{}{
				"replica": "3",
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app": "nginx",
					},
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"annotations": map[string]interface{}{
							"note": "This is a test annotation",
						},
						"labels": map[string]interface{}{
							"app": "nginx",
						},
					},
				},
			},
		}),
		resFactory.FromMapWithName("ns1", map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name": "ns1",
				"labels": map[string]interface{}{
					"app": "nginx",
				},
				"annotations": map[string]interface{}{
					"note": "This is a test annotation",
				},
			},
		}),
		resFactory.FromMapWithName("literalConfigMap",
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "foo-literalConfigMap-bar-g5f6t456f5",
					"namespace": "ns1",
					"labels": map[string]interface{}{
						"app": "nginx",
					},
					"annotations": map[string]interface{}{
						"note": "This is a test annotation",
					},
				},
				"data": map[string]interface{}{
					"DB_USERNAME": "admin",
					"DB_PASSWORD": "somepw",
				},
			}),
		resFactory.FromMapWithName("secret",
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"metadata": map[string]interface{}{
					"name":      "foo-secret-bar-82c2g5f8f6",
					"namespace": "ns1",
					"labels": map[string]interface{}{
						"app": "nginx",
					},
					"annotations": map[string]interface{}{
						"note": "This is a test annotation",
					},
				},
				"type": ifc.SecretTypeOpaque,
				"data": map[string]interface{}{
					"DB_USERNAME": base64.StdEncoding.EncodeToString([]byte("admin")),
					"DB_PASSWORD": base64.StdEncoding.EncodeToString([]byte("somepw")),
				},
			}),
	}

	expected := resmap.New()
	for _, r := range resources {
		if err := expected.Append(r); err != nil {
			t.Fatalf("unexpected error %v", err)
		}
	}
	expYaml, err := expected.AsYaml()
	assert.NoError(t, err)

	kt := makeKustTargetWithRf(
		t, th.GetFSys(), "/whatever", pvd, 1)
	assert.NoError(t, kt.Load())
	actual, err := kt.MakeCustomizedResMap()
	assert.NoError(t, err)
	actYaml, err := actual.AsYaml()
	assert.NoError(t, err)
	assert.Equal(t, expYaml, actYaml)

	// Repeat test with parallel accumulation
	kt = makeKustTargetWithRf(
		t, th.GetFSys(), "/whatever", pvd, 16)
	assert.NoError(t, kt.Load())
	actual, err = kt.MakeCustomizedResMap()
	assert.NoError(t, err)

	if err = expected.ErrorIfNotEqualSets(actual); err != nil {
		t.Fatalf("unexpected inequality: %v", err)
	}
}

func TestMaxParallelAccumulate(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("/app/serviceA.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: myServiceA
spec:
  ports:
  - port: 7002
`)
	th.WriteF("/app/serviceB.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: myServiceB
spec:
  ports:
  - port: 7002
`)
	th.WriteF("/app/serviceC.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: myServiceC
spec:
  ports:
  - port: 7002
`)
	th.WriteF("/app/serviceD.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: myServiceD
spec:
  ports:
  - port: 7002
`)
	th.WriteK("/app", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- serviceA.yaml
- serviceB.yaml
- serviceC.yaml
- serviceD.yaml
`)
	options := th.MakeDefaultOptions()
	options.MaxParallelAccumulate = 4
	//options.AddManagedbyLabel = true
	m := th.Run("/app", options)
	//th.AssertActualEqualsExpected(m, `
	serviceName := regexp.MustCompile("name: myService[A-D]")
	th.AssertActualEqualsExpectedWithTweak(m,
		func(x []byte) []byte {
			return serviceName.ReplaceAll(x, []byte("name: myServiceX"))
		}, `
apiVersion: v1
kind: Service
metadata:
  name: myServiceX
spec:
  ports:
  - port: 7002
---
apiVersion: v1
kind: Service
metadata:
  name: myServiceX
spec:
  ports:
  - port: 7002
---
apiVersion: v1
kind: Service
metadata:
  name: myServiceX
spec:
  ports:
  - port: 7002
---
apiVersion: v1
kind: Service
metadata:
  name: myServiceX
spec:
  ports:
  - port: 7002
`)
}

func BenchmarkMaxParallelAccumulate1(b *testing.B) {
	benchmarkMaxParallelAccumulate(1, b)
}

func BenchmarkMaxParallelAccumulate2(b *testing.B) {
	benchmarkMaxParallelAccumulate(2, b)
}
func BenchmarkMaxParallelAccumulate3(b *testing.B) {
	benchmarkMaxParallelAccumulate(3, b)
}

func benchmarkMaxParallelAccumulate(i int, b *testing.B) {

	dir := makeTmpDir(b)
	defer os.RemoveAll(dir)

	th := kusttest_test.MakeHarnessWithFs(b, filesys.MakeFsOnDisk())
	th.WriteK(dir, `
resources:
- git::https://github.com/kubernetes-sigs/kustomize.git//examples/helloWorld
- git::https://github.com/kubernetes-sigs/kustomize.git//examples/multibases
- git::https://github.com/kubernetes-sigs/kustomize.git//examples/loadHttp
`)
	options := th.MakeDefaultOptions()
	options.LoadRestrictions = types.LoadRestrictionsNone
	options.MaxParallelAccumulate = i
	th.Run(dir, options)
}

func makeTmpDir(t testing.TB) string {
	base, err := os.Getwd()
	if err != nil {
		t.Fatalf("err %v", err)
	}
	dir, err := ioutil.TempDir(base, "kustomize-tmp-test-")
	if err != nil {
		t.Fatalf("err %v", err)
	}
	return dir
}
