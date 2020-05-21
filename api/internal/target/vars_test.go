// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package target_test

import (
	"fmt"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/resid"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	"sigs.k8s.io/kustomize/api/types"
)

// To simplify tests, these vars specified in alphabetical order.
var someVars = []types.Var{
	{
		Name: "AWARD",
		ObjRef: types.Target{
			APIVersion: "v7",
			Gvk:        resid.Gvk{Kind: "Service"},
			Name:       "nobelPrize"},
		FieldRef: types.FieldSelector{FieldPath: "some.arbitrary.path"},
	},
	{
		Name: "BIRD",
		ObjRef: types.Target{
			APIVersion: "v300",
			Gvk:        resid.Gvk{Kind: "Service"},
			Name:       "heron"},
		FieldRef: types.FieldSelector{FieldPath: "metadata.name"},
	},
	{
		Name: "FRUIT",
		ObjRef: types.Target{
			Gvk:  resid.Gvk{Kind: "Service"},
			Name: "apple"},
		FieldRef: types.FieldSelector{FieldPath: "metadata.name"},
	},
	{
		Name: "VEGETABLE",
		ObjRef: types.Target{
			Gvk:  resid.Gvk{Kind: "Leafy"},
			Name: "kale"},
		FieldRef: types.FieldSelector{FieldPath: "metadata.name"},
	},
}

func TestGetAllVarsSimple(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app", `
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
	ra, err := makeAndLoadKustTarget(
		t, th.GetFSys(), "/app").AccumulateTarget()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	vars := ra.Vars()
	if len(vars) != 2 {
		t.Fatalf("unexpected size %d", len(vars))
	}
	for i := range vars[:2] {
		// By using Var.DeepEqual, we are protecting the code
		// from a potential invocation of vars[i].ObjRef.GVK()
		// during accumulateTarget
		if !vars[i].DeepEqual(someVars[i]) {
			t.Fatalf("unexpected var[%d]:\n  %v\n  %v", i, vars[i], someVars[i])
		}
	}
}

func TestGetAllVarsNested(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app/base", `
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
	th.WriteK("/app/overlays/o1", `
vars:
  - name: FRUIT
    objref:
      kind: Service
      name: apple
resources:
- ../../base
`)
	th.WriteK("/app/overlays/o2", `
vars:
  - name: VEGETABLE
    objref:
      kind: Leafy
      name: kale
resources:
- ../o1
`)

	ra, err := makeAndLoadKustTarget(
		t, th.GetFSys(), "/app/overlays/o2").AccumulateTarget()
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
		// By using Var.DeepEqual, we are protecting the code
		// from a potential invocation of vars[i].ObjRef.GVK()
		// during accumulateTarget
		if !vars[i].DeepEqual(someVars[i]) {
			t.Fatalf("unexpected var[%d]:\n  %v\n  %v", i, vars[i], someVars[i])
		}
	}
}

func TestVarCollisionsForbidden(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app/base", `
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
	th.WriteK("/app/overlays/o1", `
vars:
  - name: AWARD
    objref:
      kind: Service
      name: academy
resources:
- ../../base
`)
	th.WriteK("/app/overlays/o2", `
vars:
  - name: VEGETABLE
    objref:
      kind: Leafy
      name: kale
resources:
- ../o1
`)
	_, err := makeAndLoadKustTarget(
		t, th.GetFSys(), "/app/overlays/o2").AccumulateTarget()
	if err == nil {
		t.Fatalf("expected var collision")
	}
	if !strings.Contains(err.Error(),
		"var 'AWARD' already encountered") {
		t.Fatalf("unexpected error: %v", err)
	}
}
