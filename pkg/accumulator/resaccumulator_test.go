// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package accumulator_test

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/v3/k8sdeps/kunstruct"
	. "sigs.k8s.io/kustomize/v3/pkg/accumulator"
	"sigs.k8s.io/kustomize/v3/pkg/gvk"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	resmaptest_test "sigs.k8s.io/kustomize/v3/pkg/resmaptest"
	"sigs.k8s.io/kustomize/v3/pkg/resource"
	"sigs.k8s.io/kustomize/v3/pkg/transformers/config"
	"sigs.k8s.io/kustomize/v3/pkg/types"
)

func makeResAccumulator(t *testing.T) (*ResAccumulator, *resource.Factory) {
	ra := MakeEmptyAccumulator()
	err := ra.MergeConfig(config.MakeDefaultConfig())
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
				Gvk:  gvk.Gvk{Version: "v1", Kind: "Service"},
				Name: "backendOne"},
		},
		{
			Name: "SERVICE_TWO",
			ObjRef: types.Target{
				Gvk:  gvk.Gvk{Version: "v1", Kind: "Service"},
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
				Gvk:  gvk.Gvk{Version: "v1", Kind: "Service"},
				Name: "backendOne"},
		},
		{
			Name: "SERVICE_UNUSED",
			ObjRef: types.Target{
				Gvk:  gvk.Gvk{Version: "v1", Kind: "Service"},
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
				Gvk:  gvk.Gvk{Version: "v1", Kind: "Service"},
				Name: "backendOne",
			},
		},
	})
	if err != nil {
		t.Fatalf("expected error")
	}

	// Behavior has been modified.
	// Conflict detection moved to VarMap object
	// =============================================
	// if err == nil {
	// 	t.Fatalf("expected error")
	// }
	// if !strings.Contains(
	// 	err.Error(), "unable to disambiguate") {
	// 	t.Fatalf("unexpected err: %v", err)
	// }
}

func TestResolveVarsGoodResIdBadField(t *testing.T) {
	ra, _ := makeResAccumulator(t)
	err := ra.MergeVars([]types.Var{
		{
			Name: "SERVICE_ONE",
			ObjRef: types.Target{
				Gvk:  gvk.Gvk{Version: "v1", Kind: "Service"},
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
				Gvk:  gvk.Gvk{Version: "v1", Kind: "Service"},
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
				Gvk:  gvk.Gvk{Version: "v1", Kind: "Service"},
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
				Gvk:  gvk.Gvk{Version: "v1", Kind: "Service"},
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

func makeResAccumulatorWithPod(t *testing.T, podName, containerName string) *ResAccumulator {
	ra := MakeEmptyAccumulator()
	rf := resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())
	err := ra.AppendAll(
		resmaptest_test.NewRmBuilder(t, rf).
			Add(map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]interface{}{
					"name": podName,
				},
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"name": []interface{}{
								containerName,
							},
						},
					},
				}}).ResMap())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	return ra
}

func TestHandoverConflictingResourcesNoMatches(t *testing.T) {
	ra := makeResAccumulatorWithPod(t, "pod1", "container1")
	subRa := makeResAccumulatorWithPod(t, "pod2", "container2")
	subRa.HandoverConflictingResources(ra)
	if subRa.ResMap().Size() != 1 {
		t.Errorf("subRa's ResMap should have 1 item, got %d", subRa.ResMap().Size())
	}
}

func TestHandoverConflictingResourcesMatchesWithNoConflicts(t *testing.T) {
	ra := makeResAccumulatorWithPod(t, "pod1", "container1")
	subRa := makeResAccumulatorWithPod(t, "pod1", "container1")
	subRa.HandoverConflictingResources(ra)
	if subRa.ResMap().Size() != 0 {
		t.Errorf("subRa's ResMap should have 0 items, got %d", subRa.ResMap().Size())
	}
}

func TestHandoverConflictingResourcesMatchesWithConflicts(t *testing.T) {
	ra := makeResAccumulatorWithPod(t, "pod1", "container1")
	subRa := makeResAccumulatorWithPod(t, "pod1", "container2")
	subRa.HandoverConflictingResources(ra)
	if len(ra.PatchSet()) != 2 {
		t.Errorf("ra's PatchSet should have 2 item, got %d", len(ra.PatchSet()))
	}
	if subRa.ResMap().Size() != 0 {
		t.Errorf("subRa's ResMap should have 0 items, got %d", subRa.ResMap().Size())
	}
}

func makeServiceVar(varName, serviceName string) types.Var {
	return types.Var{
		Name: varName,
		ObjRef: types.Target{
			Gvk: gvk.Gvk{
				Version: "v1",
				Kind:    "Service",
			},
			Name: serviceName,
		},
	}
}

func TestAppendResolvedVarsNoConflicts(t *testing.T) {
	ra, _ := makeResAccumulator(t)
	vars := types.NewVarSet()
	vars.MergeSlice([]types.Var{makeServiceVar("SERVICE_ONE", "backendOne")})
	vars.MergeSlice([]types.Var{makeServiceVar("SERVICE_TWO", "backendTwo")})

	if err := ra.AppendResolvedVars(vars); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	expectedLen := len(vars.AsSlice())
	if len(ra.Vars()) != expectedLen {
		t.Errorf("ra should have %d vars, got %d", expectedLen, len(ra.Vars()))
	}
}

func TestAppendResolvedVarsWithConflicts(t *testing.T) {
	ra, _ := makeResAccumulator(t)
	vars1 := types.NewVarSet()
	vars1.MergeSlice([]types.Var{makeServiceVar("SERVICE_ONE", "backendOne")})
	vars2 := types.NewVarSet()
	vars2.MergeSlice([]types.Var{makeServiceVar("SERVICE_ONE", "backendTwo")})

	// The first call adds the variable
	if err := ra.AppendResolvedVars(vars1); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	// The second call adds a another variable with the same name, but a
	// different value - this should elicit an error
	expectedErrMsg := "var 'SERVICE_ONE' already encountered"
	if err := ra.AppendResolvedVars(vars2); err == nil {
		t.Fatalf("expected error, but none was received")
	} else if err.Error() != expectedErrMsg {
		t.Errorf("received the wrong error, expected `%s`, got `%s`",
			expectedErrMsg, err.Error())
	}
}

func TestAppendResolvedVarsDiamondNoConflicts(t *testing.T) {
	ra, _ := makeResAccumulator(t)
	vars := types.NewVarSet()
	vars.MergeSlice([]types.Var{makeServiceVar("SERVICE_ONE", "backendOne")})

	// The first call adds the variable
	if err := ra.AppendResolvedVars(vars); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	// The second call adds a another variable with the same name and value.
	// Since the variables are the same, this call does nothing
	if err := ra.AppendResolvedVars(vars); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	expectedLen := len(vars.AsSlice())
	if len(ra.Vars()) != expectedLen {
		t.Errorf("ra should have %d vars, got %d", expectedLen, len(ra.Vars()))
	}
}

func TestAppendResolvedVarsDiamondWithConflicts(t *testing.T) {
	ra := MakeEmptyAccumulator()
	rf := resource.NewFactory(
		kunstruct.NewKunstructuredFactoryImpl())
	// Add some services with "prefixes"
	err := ra.AppendAll(
		resmaptest_test.NewRmBuilder(t, rf).
			AddWithName("backend",
				map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Service",
					"metadata": map[string]interface{}{
						"name": "app1-backend",
					}}).
			AddWithName("backend",
				map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Service",
					"metadata": map[string]interface{}{
						"name": "app2-backend",
					}}).ResMap())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	serviceVar := types.Var{
		Name: "SERVICE_ONE",
		ObjRef: types.Target{
			Gvk: gvk.Gvk{
				Version: "v1",
				Kind:    "Service",
			},
			Name: "backend",
		},
		FieldRef: types.FieldSelector{
			FieldPath: "metadata.name",
		},
	}
	vars := types.NewVarSet()
	vars.MergeSlice([]types.Var{serviceVar})

	// The first call adds the variable without conflict
	if err := ra.AppendResolvedVars(vars); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	// Behavior has been modified.
	// Conflict detection moved to VarMap object
	// =============================================
	// // A second call will result in a conflicting variable name. Further,
	// // it will be unclear whether this variable refers to app1-backend or
	// // app2-backend. This is an error
	// expectedErrMsg := fmt.Sprintf(
	// 	"found %d resId matches for var %s "+
	// 		"(unable to disambiguate)",
	// 	ra.ResMap().Size(), serviceVar)
	// if err := ra.AppendResolvedVars(vars); err == nil {
	// 	t.Fatalf("expected error")
	// } else if err.Error() != expectedErrMsg {
	// 	t.Errorf("received the wrong error, expected `%s`, got `%s`",
	// 		expectedErrMsg, err.Error())
	// }
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
