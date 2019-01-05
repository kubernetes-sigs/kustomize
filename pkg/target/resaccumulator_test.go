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

package target

import (
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/transformers/config"
	"sigs.k8s.io/kustomize/pkg/types"
)

func TestResolveVars(t *testing.T) {
	ra := MakeEmptyAccumulator()
	err := ra.MergeConfig(config.NewFactory(nil).DefaultConfig())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
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
										"--somebackendService $(FOO)",
										"--yetAnother $(BAR)",
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
	err = ra.MergeVars([]types.Var{
		{
			Name: "FOO",
			ObjRef: types.Target{
				Gvk:  gvk.Gvk{Version: "v1", Kind: "Service"},
				Name: "backendOne"},
		}, {
			Name: "BAR",
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
