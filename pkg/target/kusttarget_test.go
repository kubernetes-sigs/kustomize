/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package target_test

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/internal/loadertest"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	. "sigs.k8s.io/kustomize/pkg/target"
	"sigs.k8s.io/kustomize/pkg/types"
)

const (
	kustomizationContent = `
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
`
	deploymentContent = `
apiVersion: apps/v1
metadata:
  name: dply1
kind: Deployment
`
	namespaceContent = `
apiVersion: v1
kind: Namespace
metadata:
  name: ns1
`
	jsonpatchContent = `[
    {"op": "add", "path": "/spec/replica", "value": "3"}
]`
)

func TestResources(t *testing.T) {
	th := NewKustTestHarness(t, "/whatever")
	th.writeK("/whatever/", kustomizationContent)
	th.writeF("/whatever/deployment.yaml", deploymentContent)
	th.writeF("/whatever/namespace.yaml", namespaceContent)
	th.writeF("/whatever/jsonpatch.json", jsonpatchContent)

	expected := resmap.ResMap{
		resid.NewResIdWithPrefixSuffixNamespace(
			gvk.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"},
			"dply1", "foo-", "-bar", "ns1"): th.fromMap(
			map[string]interface{}{
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
		resid.NewResIdWithPrefixSuffixNamespace(
			gvk.Gvk{Version: "v1", Kind: "ConfigMap"},
			"literalConfigMap", "foo-", "-bar", "ns1"): th.fromMapAndOption(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "foo-literalConfigMap-bar-8d2dkb8k24",
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
			},
			&types.GeneratorArgs{},
			&types.GeneratorOptions{}),
		resid.NewResIdWithPrefixSuffixNamespace(
			gvk.Gvk{Version: "v1", Kind: "Secret"},
			"secret", "foo-", "-bar", "ns1"): th.fromMapAndOption(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"metadata": map[string]interface{}{
					"name":      "foo-secret-bar-9btc7bt4kb",
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
			},
			&types.GeneratorArgs{},
			&types.GeneratorOptions{}),
		resid.NewResIdWithPrefixSuffixNamespace(
			gvk.Gvk{Version: "v1", Kind: "Namespace"},
			"ns1", "foo-", "-bar", ""): th.fromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]interface{}{
					"name": "foo-ns1-bar",
					"labels": map[string]interface{}{
						"app": "nginx",
					},
					"annotations": map[string]interface{}{
						"note": "This is a test annotation",
					},
				},
			}),
	}
	actual, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("unexpected Resources error %v", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		err = expected.ErrorIfNotEqual(actual)
		t.Fatalf("unexpected inequality: %v", err)
	}
}

func TestKustomizationNotFound(t *testing.T) {
	_, err := NewKustTarget(loadertest.NewFakeLoader("/foo"), nil, nil)
	if err == nil {
		t.Fatalf("expected an error")
	}
	if err.Error() !=
		`unable to find one of 'kustomization.yaml', 'kustomization.yml' or 'Kustomization' in directory '/foo'` {
		t.Fatalf("unexpected error: %q", err)
	}
}

func TestResourceNotFound(t *testing.T) {
	th := NewKustTestHarness(t, "/whatever")
	th.writeK("/whatever", kustomizationContent)
	_, err := th.makeKustTarget().MakeCustomizedResMap()
	if err == nil {
		t.Fatalf("Didn't get the expected error for an unknown resource")
	}
	if !strings.Contains(err.Error(), `cannot read file`) {
		t.Fatalf("unexpected error: %q", err)
	}
}

func findSecret(m resmap.ResMap) *resource.Resource {
	for id, res := range m {
		if id.Gvk().Kind == "Secret" {
			return res
		}
	}
	return nil
}

func TestDisableNameSuffixHash(t *testing.T) {
	th := NewKustTestHarness(t, "/whatever")
	th.writeK("/whatever/", kustomizationContent)
	th.writeF("/whatever/deployment.yaml", deploymentContent)
	th.writeF("/whatever/namespace.yaml", namespaceContent)
	th.writeF("/whatever/jsonpatch.json", jsonpatchContent)

	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("unexpected Resources error %v", err)
	}
	secret := findSecret(m)
	if secret == nil {
		t.Errorf("Expected to find a Secret")
	}
	if secret.GetName() != "foo-secret-bar-9btc7bt4kb" {
		t.Errorf("unexpected secret resource name: %s", secret.GetName())
	}

	th.writeK("/whatever/",
		strings.Replace(kustomizationContent,
			"disableNameSuffixHash: false",
			"disableNameSuffixHash: true", -1))
	m, err = th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("unexpected Resources error %v", err)
	}
	secret = findSecret(m)
	if secret == nil {
		t.Errorf("Expected to find a Secret")
	}
	if secret.GetName() != "foo-secret-bar" { // No hash at end.
		t.Errorf("unexpected secret resource name: %s", secret.GetName())
	}
}

func TestIssue596AllowDirectoriesThatAreSubstringsOfEachOther(t *testing.T) {
	th := NewKustTestHarness(t, "/app/overlays/aws-sandbox2.us-east-1")
	th.writeK("/app/base", "")
	th.writeK("/app/overlays/aws", `
bases:
- ../../base
`)
	th.writeK("/app/overlays/aws-nonprod", `
bases:
- ../aws
`)
	th.writeK("/app/overlays/aws-sandbox2.us-east-1", `
bases:
- ../aws-nonprod
`)
	m, err := th.makeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.assertActualEqualsExpected(m, "")
}

// To simplify tests, these vars specified in alphabetical order.
var someVars = []types.Var{
	{
		Name: "AWARD",
		ObjRef: types.Target{
			APIVersion: "v7",
			Gvk:        gvk.Gvk{Kind: "Service"},
			Name:       "nobelPrize"},
		FieldRef: types.FieldSelector{FieldPath: "some.arbitrary.path"},
	},
	{
		Name: "BIRD",
		ObjRef: types.Target{
			APIVersion: "v300",
			Gvk:        gvk.Gvk{Kind: "Service"},
			Name:       "heron"},
		FieldRef: types.FieldSelector{FieldPath: "metadata.name"},
	},
	{
		Name: "FRUIT",
		ObjRef: types.Target{
			Gvk:  gvk.Gvk{Kind: "Service"},
			Name: "apple"},
		FieldRef: types.FieldSelector{FieldPath: "metadata.name"},
	},
	{
		Name: "VEGETABLE",
		ObjRef: types.Target{
			Gvk:  gvk.Gvk{Kind: "Leafy"},
			Name: "kale"},
		FieldRef: types.FieldSelector{FieldPath: "metadata.name"},
	},
}

func TestGetAllVarsSimple(t *testing.T) {
	th := NewKustTestHarness(t, "/app")
	th.writeK("/app", `
vars:
  - name: AWARD
    objref:
      kind: Service
      name: nobelPrize
      apiVersion: v7
    fieldref:
      fieldpath: some.arbitrary.path
  - name: BIRD
    objref:
      kind: Service
      name: heron
      apiVersion: v300
`)
	ra, err := th.makeKustTarget().AccumulateTarget()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	vars := ra.Vars()
	if len(vars) != 2 {
		t.Fatalf("unexpected size %d", len(vars))
	}
	for i := range vars[:2] {
		if !reflect.DeepEqual(vars[i], someVars[i]) {
			t.Fatalf("unexpected var[%d]:\n  %v\n  %v", i, vars[i], someVars[i])
		}
	}
}

func TestGetAllVarsNested(t *testing.T) {
	th := NewKustTestHarness(t, "/app/overlays/o2")
	th.writeK("/app/base", `
vars:
  - name: AWARD
    objref:
      kind: Service
      name: nobelPrize
      apiVersion: v7
    fieldref:
      fieldpath: some.arbitrary.path
  - name: BIRD
    objref:
      kind: Service
      name: heron
      apiVersion: v300
`)
	th.writeK("/app/overlays/o1", `
vars:
  - name: FRUIT
    objref:
      kind: Service
      name: apple
bases:
- ../../base
`)
	th.writeK("/app/overlays/o2", `
vars:
  - name: VEGETABLE
    objref:
      kind: Leafy
      name: kale
bases:
- ../o1
`)
	ra, err := th.makeKustTarget().AccumulateTarget()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	vars := ra.Vars()
	if len(vars) != 4 {
		for i, v := range vars {
			fmt.Printf("%v: %v\n", i, v)
		}
		t.Fatalf("expected 4 vars, got %d", len(vars))
	}
	for i := range vars {
		if !reflect.DeepEqual(vars[i], someVars[i]) {
			t.Fatalf("unexpected var[%d]:\n  %v\n  %v", i, vars[i], someVars[i])
		}
	}
}

func TestVarCollisionsForbidden(t *testing.T) {
	th := NewKustTestHarness(t, "/app/overlays/o2")
	th.writeK("/app/base", `
vars:
  - name: AWARD
    objref:
      kind: Service
      name: nobelPrize
      apiVersion: v7
    fieldref:
      fieldpath: some.arbitrary.path
  - name: BIRD
    objref:
      kind: Service
      name: heron
      apiVersion: v300
`)
	th.writeK("/app/overlays/o1", `
vars:
  - name: AWARD
    objref:
      kind: Service
      name: academy
bases:
- ../../base
`)
	th.writeK("/app/overlays/o2", `
vars:
  - name: VEGETABLE
    objref:
      kind: Leafy
      name: kale
bases:
- ../o1
`)
	_, err := th.makeKustTarget().AccumulateTarget()
	if err == nil {
		t.Fatalf("expected var collision")
	}
	if !strings.Contains(err.Error(),
		"var AWARD already encountered") {
		t.Fatalf("unexpected error: %v", err)
	}
}
