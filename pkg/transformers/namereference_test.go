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

package transformers

import (
	"reflect"
	"testing"

	"github.com/kubernetes-sigs/kustomize/pkg/resmap"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestNameReferenceRun(t *testing.T) {
	m := resmap.ResMap{
		resource.NewResId(cmap, "cm1"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "someprefix-cm1-somehash",
				},
			}),
		resource.NewResId(cmap, "cm2"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "someprefix-cm2-somehash",
				},
			}),
		resource.NewResId(secret, "secret1"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"metadata": map[string]interface{}{
					"name": "someprefix-secret1-somehash",
				},
			}),
		resource.NewResId(pvc, "claim1"): resource.NewResourceFromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "PersistentVolumeClaim",
				"metadata": map[string]interface{}{
					"name": "someprefix-claim1",
				},
			}),
		resource.NewResId(deploy, "deploy1"): resource.NewResourceFromMap(
			map[string]interface{}{
				"group":      "apps",
				"apiVersion": "v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "deploy1",
				},
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "nginx",
									"image": "nginx:1.7.9",
									"env": []interface{}{
										map[string]interface{}{
											"name": "CM_FOO",
											"valueFrom": map[string]interface{}{
												"configMapKeyRef": map[string]interface{}{
													"name": "cm1",
													"key":  "somekey",
												},
											},
										},
										map[string]interface{}{
											"name": "SECRET_FOO",
											"valueFrom": map[string]interface{}{
												"secretKeyRef": map[string]interface{}{
													"name": "secret1",
													"key":  "somekey",
												},
											},
										},
									},
									"envFrom": []interface{}{
										map[string]interface{}{
											"configMapRef": map[string]interface{}{
												"name": "cm1",
												"key":  "somekey",
											},
										},
										map[string]interface{}{
											"secretRef": map[string]interface{}{
												"name": "secret1",
												"key":  "somekey",
											},
										},
									},
								},
							},
							"imagePullSecrets": []interface{}{
								map[string]interface{}{
									"name": "secret1",
								},
							},
							"volumes": map[string]interface{}{
								"configMap": map[string]interface{}{
									"name": "cm1",
								},
								"projected": map[string]interface{}{
									"sources": map[string]interface{}{
										"configMap": map[string]interface{}{
											"name": "cm2",
										},
									},
								},
								"secret": map[string]interface{}{
									"secretName": "secret1",
								},
								"persistentVolumeClaim": map[string]interface{}{
									"claimName": "claim1",
								},
							},
						},
					},
				},
			}),
		resource.NewResId(statefulset, "statefulset1"): resource.NewResourceFromMap(
			map[string]interface{}{
				"group":      "apps",
				"apiVersion": "v1",
				"kind":       "StatefulSet",
				"metadata": map[string]interface{}{
					"name": "statefulset1",
				},
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "nginx",
									"image": "nginx:1.7.9",
								},
							},
							"volumes": map[string]interface{}{
								"projected": map[string]interface{}{
									"sources": map[string]interface{}{
										"configMap": map[string]interface{}{
											"name": "cm2",
										},
									},
								},
							},
						},
					},
				},
			}),
	}

	expected := resmap.ResMap{}
	for k, v := range m {
		expected[k] = v
	}

	expected[resource.NewResId(deploy, "deploy1")] = resource.NewResourceFromMap(
		map[string]interface{}{
			"group":      "apps",
			"apiVersion": "v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "deploy1",
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"name":  "nginx",
								"image": "nginx:1.7.9",
								"env": []interface{}{
									map[string]interface{}{
										"name": "CM_FOO",
										"valueFrom": map[string]interface{}{
											"configMapKeyRef": map[string]interface{}{
												"name": "someprefix-cm1-somehash",
												"key":  "somekey",
											},
										},
									},
									map[string]interface{}{
										"name": "SECRET_FOO",
										"valueFrom": map[string]interface{}{
											"secretKeyRef": map[string]interface{}{
												"name": "someprefix-secret1-somehash",
												"key":  "somekey",
											},
										},
									},
								},
								"envFrom": []interface{}{
									map[string]interface{}{
										"configMapRef": map[string]interface{}{
											"name": "someprefix-cm1-somehash",
											"key":  "somekey",
										},
									},
									map[string]interface{}{
										"secretRef": map[string]interface{}{
											"name": "someprefix-secret1-somehash",
											"key":  "somekey",
										},
									},
								},
							},
						},
						"imagePullSecrets": []interface{}{
							map[string]interface{}{
								"name": "someprefix-secret1-somehash",
							},
						},
						"volumes": map[string]interface{}{
							"configMap": map[string]interface{}{
								"name": "someprefix-cm1-somehash",
							},
							"projected": map[string]interface{}{
								"sources": map[string]interface{}{
									"configMap": map[string]interface{}{
										"name": "someprefix-cm2-somehash",
									},
								},
							},
							"secret": map[string]interface{}{
								"secretName": "someprefix-secret1-somehash",
							},
							"persistentVolumeClaim": map[string]interface{}{
								"claimName": "someprefix-claim1",
							},
						},
					},
				},
			},
		})
	expected[resource.NewResId(statefulset, "statefulset1")] = resource.NewResourceFromMap(
		map[string]interface{}{
			"group":      "apps",
			"apiVersion": "v1",
			"kind":       "StatefulSet",
			"metadata": map[string]interface{}{
				"name": "statefulset1",
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"name":  "nginx",
								"image": "nginx:1.7.9",
							},
						},
						"volumes": map[string]interface{}{
							"projected": map[string]interface{}{
								"sources": map[string]interface{}{
									"configMap": map[string]interface{}{
										"name": "someprefix-cm2-somehash",
									},
								},
							},
						},
					},
				},
			},
		})

	nrt, err := NewDefaultingNameReferenceTransformer()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = nrt.Transform(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(m, expected) {
		err = expected.ErrorIfNotEqual(m)
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}

func TestAddNameReferencePathConfigs(t *testing.T) {
	expected := len(defaultNameReferencePathConfigs) + 1

	pathConfigs := []ReferencePathConfig{
		{
			referencedGVK: schema.GroupVersionKind{
				Kind: "KindA",
			},
			pathConfigs: []PathConfig{
				{
					GroupVersionKind: &schema.GroupVersionKind{
						Kind: "KindB",
					},
					Path:               []string{"path", "to", "a", "field"},
					CreateIfNotPresent: false,
				},
			},
		},
	}

	AddNameReferencePathConfigs(pathConfigs)
	if len(defaultNameReferencePathConfigs) != expected {
		t.Fatalf("actual %v doesn't match expected: %v", len(defaultAnnotationsPathConfigs), expected)
	}
}
