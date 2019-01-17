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
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
)

func TestNameReferenceHappyRun(t *testing.T) {
	rf := resource.NewFactory(
		kunstruct.NewKunstructuredFactoryImpl())
	m := resmap.ResMap{
		resid.NewResId(cmap, "cm1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "someprefix-cm1-somehash",
				},
			}),
		resid.NewResId(cmap, "cm2"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "someprefix-cm2-somehash",
				},
			}),
		resid.NewResId(secret, "secret1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"metadata": map[string]interface{}{
					"name": "someprefix-secret1-somehash",
				},
			}),
		resid.NewResId(pvc, "claim1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "PersistentVolumeClaim",
				"metadata": map[string]interface{}{
					"name": "someprefix-claim1",
				},
			}),
		resid.NewResId(ingress, "ingress1"): rf.FromMap(
			map[string]interface{}{
				"group":      "extensions",
				"apiVersion": "v1beta1",
				"kind":       "Ingress",
				"metadata": map[string]interface{}{
					"name": "ingress1",
					"annotations": map[string]interface{}{
						"ingress.kubernetes.io/auth-secret":       "secret1",
						"nginx.ingress.kubernetes.io/auth-secret": "secret1",
					},
				},
				"spec": map[string]interface{}{
					"backend": map[string]interface{}{
						"serviceName": "testsvc",
						"servicePort": "80",
					},
				},
			},
		),
		resid.NewResId(deploy, "deploy1"): rf.FromMap(
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
										"secret": map[string]interface{}{
											"name": "secret1",
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
		resid.NewResId(statefulset, "statefulset1"): rf.FromMap(
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
										"secret": map[string]interface{}{
											"name": "secret1",
										},
									},
								},
							},
						},
					},
				},
			}),
		resid.NewResIdWithPrefixNamespace(sa, "sa", "", "test"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ServiceAccount",
				"metadata": map[string]interface{}{
					"name":      "someprefix-sa",
					"namespace": "test",
				},
			}),
		resid.NewResId(crb, "crb"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "rbac.authorization.k8s.io/v1",
				"kind":       "ClusterRoleBinding",
				"metadata": map[string]interface{}{
					"name": "crb",
				},
				"subjects": []interface{}{
					map[string]interface{}{
						"kind":      "ServiceAccount",
						"name":      "sa",
						"namespace": "test",
					},
				},
			}),
		resid.NewResId(cr, "cr"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "rbac.authorization.k8s.io/v1",
				"kind":       "ClusterRole",
				"metadata": map[string]interface{}{
					"name": "cr",
				},
				"rules": []interface{}{
					map[string]interface{}{
						"resources": []interface{}{
							"secrets",
						},
						"resourceNames": []interface{}{
							"secret1",
							"secret1",
							"secret2",
						},
					},
				},
			}),
	}

	expected := resmap.ResMap{}
	for k, v := range m {
		expected[k] = v
	}

	expected[resid.NewResId(deploy, "deploy1")] = rf.FromMap(
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
									"secret": map[string]interface{}{
										"name": "someprefix-secret1-somehash",
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
	expected[resid.NewResId(statefulset, "statefulset1")] = rf.FromMap(
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
									"secret": map[string]interface{}{
										"name": "someprefix-secret1-somehash",
									},
								},
							},
						},
					},
				},
			},
		})
	expected[resid.NewResId(ingress, "ingress1")] = rf.FromMap(
		map[string]interface{}{
			"group":      "extensions",
			"apiVersion": "v1beta1",
			"kind":       "Ingress",
			"metadata": map[string]interface{}{
				"name": "ingress1",
				"annotations": map[string]interface{}{
					"ingress.kubernetes.io/auth-secret":       "someprefix-secret1-somehash",
					"nginx.ingress.kubernetes.io/auth-secret": "someprefix-secret1-somehash",
				},
			},
			"spec": map[string]interface{}{
				"backend": map[string]interface{}{
					"serviceName": "testsvc",
					"servicePort": "80",
				},
			},
		},
	)
	expected[resid.NewResId(crb, "crb")] = rf.FromMap(
		map[string]interface{}{
			"apiVersion": "rbac.authorization.k8s.io/v1",
			"kind":       "ClusterRoleBinding",
			"metadata": map[string]interface{}{
				"name": "crb",
			},
			"subjects": []interface{}{
				map[string]interface{}{
					"kind":      "ServiceAccount",
					"name":      "someprefix-sa",
					"namespace": "test",
				},
			},
		})
	expected[resid.NewResId(cr, "cr")] = rf.FromMap(
		map[string]interface{}{
			"apiVersion": "rbac.authorization.k8s.io/v1",
			"kind":       "ClusterRole",
			"metadata": map[string]interface{}{
				"name": "cr",
			},
			"rules": []interface{}{
				map[string]interface{}{
					"resources": []interface{}{
						"secrets",
					},
					"resourceNames": []interface{}{
						"someprefix-secret1-somehash",
						"someprefix-secret1-somehash",
						"secret2",
					},
				},
			},
		})
	nrt := NewNameReferenceTransformer(defaultTransformerConfig.NameReference)
	err := nrt.Transform(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(m, expected) {
		err = expected.ErrorIfNotEqual(m)
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}

func TestNameReferenceUnhappyRun(t *testing.T) {
	rf := resource.NewFactory(
		kunstruct.NewKunstructuredFactoryImpl())
	tests := []struct {
		resMap      resmap.ResMap
		expectedErr string
	}{
		{
			resMap: resmap.ResMap{
				resid.NewResId(cr, "cr"): rf.FromMap(
					map[string]interface{}{
						"apiVersion": "rbac.authorization.k8s.io/v1",
						"kind":       "ClusterRole",
						"metadata": map[string]interface{}{
							"name": "cr",
						},
						"rules": []interface{}{
							map[string]interface{}{
								"resources": []interface{}{
									"secrets",
								},
								"resourceNames": []interface{}{
									[]interface{}{},
								},
							},
						},
					}),
			},
			expectedErr: "is expected to be string"},
		{resMap: resmap.ResMap{
			resid.NewResId(cr, "cr"): rf.FromMap(
				map[string]interface{}{
					"apiVersion": "rbac.authorization.k8s.io/v1",
					"kind":       "ClusterRole",
					"metadata": map[string]interface{}{
						"name": "cr",
					},
					"rules": []interface{}{
						map[string]interface{}{
							"resources": []interface{}{
								"secrets",
							},
							"resourceNames": map[string]interface{}{
								"foo": "bar",
							},
						},
					},
				}),
		},
			expectedErr: "is expected to be either a string or a []interface{}"},
	}

	nrt := NewNameReferenceTransformer(defaultTransformerConfig.NameReference)
	for _, test := range tests {
		err := nrt.Transform(test.resMap)
		if err == nil {
			t.Fatalf("expected error to happen")
		}

		if !strings.Contains(err.Error(), test.expectedErr) {
			t.Fatalf("Incorrect error.\nExpected: %s, but got %v",
				test.expectedErr, err)
		}
	}
}

func TestNameReferencePersistentVolumeHappyRun(t *testing.T) {
	rf := resource.NewFactory(
		kunstruct.NewKunstructuredFactoryImpl())
	m := resmap.ResMap{
		resid.NewResId(pv, "volume1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "PersistentVolume",
				"metadata": map[string]interface{}{
					"name": "someprefix-volume1",
				},
			}),

		resid.NewResId(pvc, "claim1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "PersistentVolumeClaim",
				"metadata": map[string]interface{}{
					"name":      "someprefix-claim1",
					"namespace": "some-namespace",
				},
				"spec": map[string]interface{}{
					"volumeName": "volume1",
				},
			}),
	}

	expected := resmap.ResMap{
		resid.NewResId(pv, "volume1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "PersistentVolume",
				"metadata": map[string]interface{}{
					"name": "someprefix-volume1",
				},
			}),

		resid.NewResId(pvc, "claim1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "PersistentVolumeClaim",
				"metadata": map[string]interface{}{
					"name":      "someprefix-claim1",
					"namespace": "some-namespace",
				},
				"spec": map[string]interface{}{
					"volumeName": "someprefix-volume1",
				},
			}),
	}
	nrt := NewNameReferenceTransformer(defaultTransformerConfig.NameReference)
	err := nrt.Transform(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(m, expected) {
		err = expected.ErrorIfNotEqual(m)
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}
