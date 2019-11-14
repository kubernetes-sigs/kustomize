// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package accumulator_test

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	. "sigs.k8s.io/kustomize/api/internal/accumulator"
	"sigs.k8s.io/kustomize/api/internal/plugins/builtinconfig"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	resmaptest_test "sigs.k8s.io/kustomize/api/testutils/resmaptest"
	"sigs.k8s.io/kustomize/api/types"
)

func makeResAccumulator(t *testing.T) (*ResAccumulator, *resource.Factory) {
	ra := MakeEmptyAccumulator()
	err := ra.MergeConfig(builtinconfig.MakeDefaultConfig())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	rf := resource.NewFactory(
		kunstruct.NewKunstructuredFactoryImpl())
	err = ra.AppendAll(
		resmaptest_test.NewRmBuilder(t, rf).
			Add(map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "deploy1",
				},
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"command": []interface{}{
										"myserver",
										"--somebackendService $(SERVICE_ONE)",
										"--yetAnother $(SERVICE_TWO)",
									},
								},
							},
						},
					},
				}}).
			Add(map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name": "backendOne",
				}}).
			Add(map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name": "backendTwo",
				}}).ResMap())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	return ra, rf
}

func TestResolveVarsHappy(t *testing.T) {
	ra, _ := makeResAccumulator(t)
	err := ra.MergeVars([]types.Var{
		{
			Name: "SERVICE_ONE",
			ObjRef: types.Target{
				Gvk:  resid.Gvk{Version: "v1", Kind: "Service"},
				Name: "backendOne"},
		},
		{
			Name: "SERVICE_TWO",
			ObjRef: types.Target{
				Gvk:  resid.Gvk{Version: "v1", Kind: "Service"},
				Name: "backendTwo"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	err = ra.ResolveVars()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	c := getCommand(find("deploy1", ra.ResMap()))
	if c != "myserver --somebackendService backendOne --yetAnother backendTwo" {
		t.Fatalf("unexpected command: %s", c)
	}
}

func TestResolveVarsOneUnused(t *testing.T) {
	ra, _ := makeResAccumulator(t)
	err := ra.MergeVars([]types.Var{
		{
			Name: "SERVICE_ONE",
			ObjRef: types.Target{
				Gvk:  resid.Gvk{Version: "v1", Kind: "Service"},
				Name: "backendOne"},
		},
		{
			Name: "SERVICE_UNUSED",
			ObjRef: types.Target{
				Gvk:  resid.Gvk{Version: "v1", Kind: "Service"},
				Name: "backendTwo"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	err = ra.ResolveVars()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	expectLog(t, buf, "well-defined vars that were never replaced: SERVICE_UNUSED")
	c := getCommand(find("deploy1", ra.ResMap()))
	if c != "myserver --somebackendService backendOne --yetAnother $(SERVICE_TWO)" {
		t.Fatalf("unexpected command: %s", c)
	}
}

func expectLog(t *testing.T, log bytes.Buffer, expect string) {
	if !strings.Contains(log.String(), expect) {
		t.Fatalf("expected log containing '%s', got '%s'", expect, log.String())
	}
}

func TestResolveVarsVarNeedsDisambiguation(t *testing.T) {
	ra, rf := makeResAccumulator(t)

	rm0 := resmap.New()
	err := rm0.Append(
		rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name":      "backendOne",
					"namespace": "fooNamespace",
				},
			}))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	err = ra.AppendAll(rm0)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	err = ra.MergeVars([]types.Var{
		{
			Name: "SERVICE_ONE",
			ObjRef: types.Target{
				Gvk:  resid.Gvk{Version: "v1", Kind: "Service"},
				Name: "backendOne",
			},
		},
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(
		err.Error(), "unable to disambiguate") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func makeNamespacedConfigMapWithDataProviderValue(
	namespace string,
	value string,
) map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name":      "environment",
			"namespace": namespace,
		},
		"data": map[string]interface{}{
			"provider": value,
		},
	}
}

func makeVarToNamepaceAndPath(
	name string,
	namespace string,
	path string,
) types.Var {
	return types.Var{
		Name: name,
		ObjRef: types.Target{
			Gvk:       resid.Gvk{Version: "v1", Kind: "ConfigMap"},
			Name:      "environment",
			Namespace: namespace,
		},
		FieldRef: types.FieldSelector{FieldPath: path},
	}
}

func TestResolveVarConflicts(t *testing.T) {
	rf := resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())

	// create configmaps in foo and bar namespaces with `data.provider` values.
	fooAws := makeNamespacedConfigMapWithDataProviderValue("foo", "aws")
	barAws := makeNamespacedConfigMapWithDataProviderValue("bar", "aws")
	barGcp := makeNamespacedConfigMapWithDataProviderValue("bar", "gcp")

	// create two variables with (apparently) conflicting names that point to
	// fieldpaths that could be generalized.
	varFoo := makeVarToNamepaceAndPath("PROVIDER", "foo", "data.provider")
	varBar := makeVarToNamepaceAndPath("PROVIDER", "bar", "data.provider")

	// create accumulators holding apparently conflicting vars that are not
	// actually in conflict because they point to the same concrete value.
	rm0 := resmap.New()
	rm0.Append(rf.FromMap(fooAws))
	ac0 := MakeEmptyAccumulator()
	ac0.AppendAll(rm0)
	ac0.MergeVars([]types.Var{varFoo})

	rm1 := resmap.New()
	rm1.Append(rf.FromMap(barAws))
	ac1 := MakeEmptyAccumulator()
	ac1.AppendAll(rm1)
	ac1.MergeVars([]types.Var{varBar})

	// validate that two vars of the same name which reference the same concrete
	// value do not produce a conflict.
	err := ac0.MergeAccumulator(ac1)
	if err == nil {
		t.Fatalf("see bug gh-1600")
	}

	// create an accumulator will have an actually conflicting value with the
	// two above (because it contains a variable whose name is used in the other
	// accumulators AND whose concrete values are different).
	rm2 := resmap.New()
	rm2.Append(rf.FromMap(barGcp))
	ac2 := MakeEmptyAccumulator()
	ac2.AppendAll(rm2)
	ac2.MergeVars([]types.Var{varBar})
	err = ac1.MergeAccumulator(ac2)
	if err == nil {
		t.Fatalf("dupe vars w/ different concrete values should conflict")
	}
}

func TestResolveVarsGoodResIdBadField(t *testing.T) {
	ra, _ := makeResAccumulator(t)
	err := ra.MergeVars([]types.Var{
		{
			Name: "SERVICE_ONE",
			ObjRef: types.Target{
				Gvk:  resid.Gvk{Version: "v1", Kind: "Service"},
				Name: "backendOne"},
			FieldRef: types.FieldSelector{FieldPath: "nope_nope_nope"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	err = ra.ResolveVars()
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(
		err.Error(),
		"not found in corresponding resource") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestResolveVarsUnmappableVar(t *testing.T) {
	ra, _ := makeResAccumulator(t)
	err := ra.MergeVars([]types.Var{
		{
			Name: "SERVICE_THREE",
			ObjRef: types.Target{
				Gvk:  resid.Gvk{Version: "v1", Kind: "Service"},
				Name: "doesNotExist"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	err = ra.ResolveVars()
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(
		err.Error(),
		"cannot be mapped to a field in the set of known resources") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestResolveVarsWithNoambiguation(t *testing.T) {
	ra1, rf := makeResAccumulator(t)
	err := ra1.MergeVars([]types.Var{
		{
			Name: "SERVICE_ONE",
			ObjRef: types.Target{
				Gvk:  resid.Gvk{Version: "v1", Kind: "Service"},
				Name: "backendOne",
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	// Create another accumulator having a resource with different prefix
	ra2 := MakeEmptyAccumulator()

	m := resmaptest_test.NewRmBuilder(t, rf).
		Add(map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "deploy2",
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"command": []interface{}{
									"myserver",
									"--somebackendService $(SUB_SERVICE_ONE)",
								},
							},
						},
					},
				},
			}}).
		Add(map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name": "backendOne",
			}}).ResMap()

	// Make it seem like this resource
	// went through a prefix transformer.
	r := m.GetByIndex(1)
	r.AddNamePrefix("sub-")
	r.SetName("sub-backendOne") // original name remains "backendOne"

	err = ra2.AppendAll(m)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	err = ra2.MergeVars([]types.Var{
		{
			Name: "SUB_SERVICE_ONE",
			ObjRef: types.Target{
				Gvk:  resid.Gvk{Version: "v1", Kind: "Service"},
				Name: "backendOne",
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	err = ra1.MergeAccumulator(ra2)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	err = ra1.ResolveVars()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func find(name string, resMap resmap.ResMap) *resource.Resource {
	for _, r := range resMap.Resources() {
		if r.GetName() == name {
			return r
		}
	}
	return nil
}

// Assumes arg is a deployment, returns the command of first container.
func getCommand(r *resource.Resource) string {
	var m map[string]interface{}
	var c []interface{}
	m, _ = r.Map()["spec"].(map[string]interface{})
	m, _ = m["template"].(map[string]interface{})
	m, _ = m["spec"].(map[string]interface{})
	c, _ = m["containers"].([]interface{})
	m, _ = c[0].(map[string]interface{})

	cmd, _ := m["command"].([]interface{})
	n := make([]string, len(cmd))
	for i, v := range cmd {
		n[i] = v.(string)
	}
	return strings.Join(n, " ")
}
