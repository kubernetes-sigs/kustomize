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
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	. "sigs.k8s.io/kustomize/pkg/target"
	"sigs.k8s.io/kustomize/pkg/transformers/config"
	"sigs.k8s.io/kustomize/pkg/types"
)

func makeResAccumulator() (*ResAccumulator, *resource.Factory, error) {
	ra := MakeEmptyAccumulator()
	err := ra.MergeConfig(config.MakeDefaultConfig())
	if err != nil {
		return nil, nil, err
	}
	rf := resource.NewFactory(
		kunstruct.NewKunstructuredFactoryImpl())
	ra.MergeResourcesWithErrorOnIdCollision(resmap.ResMap{
		resid.NewResId(
			gvk.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"},
			"deploy1"): rf.FromMap(
			map[string]interface{}{
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
				},
			}),
		resid.NewResId(
			gvk.Gvk{Version: "v1", Kind: "Service"},
			"backendOne"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name": "backendOne",
				},
			}),
		resid.NewResId(
			gvk.Gvk{Version: "v1", Kind: "Service"},
			"backendTwo"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name": "backendTwo",
				},
			}),
	})
	return ra, rf, nil
}

func TestResolveVarsHappy(t *testing.T) {
	ra, _, err := makeResAccumulator()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	err = ra.MergeVars([]types.Var{
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
	ra, _, err := makeResAccumulator()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	err = ra.MergeVars([]types.Var{
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
	ra, rf, err := makeResAccumulator()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	ra.MergeResourcesWithErrorOnIdCollision(resmap.ResMap{
		resid.NewResIdWithPrefixNamespace(
			gvk.Gvk{Version: "v1", Kind: "Service"},
			"backendOne", "", "fooNamespace"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name": "backendOne",
				},
			}),
	})

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
		t.Fatalf("unexpected err: %v", err)
	}
	err = ra.ResolveVars()
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(
		err.Error(), "unable to disambiguate") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestResolveVarsGoodResIdBadField(t *testing.T) {
	ra, _, err := makeResAccumulator()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	err = ra.MergeVars([]types.Var{
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
	ra, _, err := makeResAccumulator()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	err = ra.MergeVars([]types.Var{
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

func find(name string, resMap resmap.ResMap) *resource.Resource {
	for k, v := range resMap {
		if k.Name() == name {
			return v
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
	return strings.Join(m["command"].([]string), " ")
}
