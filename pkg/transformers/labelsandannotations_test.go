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

	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/transformers/config"
)

var service = gvk.Gvk{Version: "v1", Kind: "Service"}
var secret = gvk.Gvk{Version: "v1", Kind: "Secret"}
var cmap = gvk.Gvk{Version: "v1", Kind: "ConfigMap"}
var ns = gvk.Gvk{Version: "v1", Kind: "Namespace"}
var deploy = gvk.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"}
var statefulset = gvk.Gvk{Group: "apps", Version: "v1", Kind: "StatefulSet"}
var crd = gvk.Gvk{Group: "apiwctensions.k8s.io", Version: "v1beta1", Kind: "CustomResourceDefinition"}
var job = gvk.Gvk{Group: "batch", Version: "v1", Kind: "Job"}
var cronjob = gvk.Gvk{Group: "batch", Version: "v1beta1", Kind: "CronJob"}
var pv = gvk.Gvk{Version: "v1", Kind: "PersistentVolume"}
var pvc = gvk.Gvk{Version: "v1", Kind: "PersistentVolumeClaim"}
var cr = gvk.Gvk{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole"}
var crb = gvk.Gvk{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRoleBinding"}
var sa = gvk.Gvk{Version: "v1", Kind: "ServiceAccount"}
var ingress = gvk.Gvk{Kind: "Ingress"}
var rf = resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())
var defaultTransformerConfig = config.MakeDefaultConfig()

func TestLabelsRun(t *testing.T) {
	m := resmap.ResMap{
		resid.NewResId(cmap, "cm1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cm1",
				},
			}),
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
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"old-label": "old-value",
							},
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "nginx",
									"image": "nginx:1.7.9",
								},
							},
						},
					},
				},
			}),
		resid.NewResId(service, "svc1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name": "svc1",
				},
				"spec": map[string]interface{}{
					"ports": []interface{}{
						map[string]interface{}{
							"name": "port1",
							"port": "12345",
						},
					},
				},
			}),
		resid.NewResId(job, "job1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "batch/v1",
				"kind":       "Job",
				"metadata": map[string]interface{}{
					"name": "job1",
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
						},
					},
				},
			}),
		resid.NewResId(job, "job2"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "batch/v1",
				"kind":       "Job",
				"metadata": map[string]interface{}{
					"name": "job2",
				},
				"spec": map[string]interface{}{
					"selector": map[string]interface{}{
						"matchLabels": map[string]interface{}{
							"old-label": "old-value",
						},
					},
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "nginx",
									"image": "nginx:1.7.9",
								},
							},
						},
					},
				},
			}),
		resid.NewResId(cronjob, "cronjob1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "batch/v1beta1",
				"kind":       "CronJob",
				"metadata": map[string]interface{}{
					"name": "cronjob1",
				},
				"spec": map[string]interface{}{
					"schedule": "* 23 * * *",
					"jobTemplate": map[string]interface{}{
						"spec": map[string]interface{}{
							"template": map[string]interface{}{
								"spec": map[string]interface{}{
									"containers": []interface{}{
										map[string]interface{}{
											"name":  "nginx",
											"image": "nginx:1.7.9",
										},
									},
								},
							},
						},
					},
				},
			}),
		resid.NewResId(cronjob, "cronjob2"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "batch/v1beta1",
				"kind":       "CronJob",
				"metadata": map[string]interface{}{
					"name": "cronjob2",
				},
				"spec": map[string]interface{}{
					"schedule": "* 23 * * *",
					"jobTemplate": map[string]interface{}{
						"spec": map[string]interface{}{
							"selector": map[string]interface{}{
								"matchLabels": map[string]interface{}{
									"old-label": "old-value",
								},
							},
							"template": map[string]interface{}{
								"spec": map[string]interface{}{
									"containers": []interface{}{
										map[string]interface{}{
											"name":  "nginx",
											"image": "nginx:1.7.9",
										},
									},
								},
							},
						},
					},
				},
			}),
	}
	expected := resmap.ResMap{
		resid.NewResId(cmap, "cm1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cm1",
					"labels": map[string]interface{}{
						"label-key1": "label-value1",
						"label-key2": "label-value2",
					},
				},
			}),
		resid.NewResId(deploy, "deploy1"): rf.FromMap(
			map[string]interface{}{
				"group":      "apps",
				"apiVersion": "v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "deploy1",
					"labels": map[string]interface{}{
						"label-key1": "label-value1",
						"label-key2": "label-value2",
					},
				},
				"spec": map[string]interface{}{
					"selector": map[string]interface{}{
						"matchLabels": map[string]interface{}{
							"label-key1": "label-value1",
							"label-key2": "label-value2",
						},
					},
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"old-label":  "old-value",
								"label-key1": "label-value1",
								"label-key2": "label-value2",
							},
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "nginx",
									"image": "nginx:1.7.9",
								},
							},
						},
					},
				},
			}),
		resid.NewResId(service, "svc1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name": "svc1",
					"labels": map[string]interface{}{
						"label-key1": "label-value1",
						"label-key2": "label-value2",
					},
				},
				"spec": map[string]interface{}{
					"ports": []interface{}{
						map[string]interface{}{
							"name": "port1",
							"port": "12345",
						},
					},
					"selector": map[string]interface{}{
						"label-key1": "label-value1",
						"label-key2": "label-value2",
					},
				},
			}),
		resid.NewResId(job, "job1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "batch/v1",
				"kind":       "Job",
				"metadata": map[string]interface{}{
					"name": "job1",
					"labels": map[string]interface{}{
						"label-key1": "label-value1",
						"label-key2": "label-value2",
					},
				},
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"label-key1": "label-value1",
								"label-key2": "label-value2",
							},
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "nginx",
									"image": "nginx:1.7.9",
								},
							},
						},
					},
				},
			}),
		resid.NewResId(job, "job2"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "batch/v1",
				"kind":       "Job",
				"metadata": map[string]interface{}{
					"name": "job2",
					"labels": map[string]interface{}{
						"label-key1": "label-value1",
						"label-key2": "label-value2",
					},
				},
				"spec": map[string]interface{}{
					"selector": map[string]interface{}{
						"matchLabels": map[string]interface{}{
							"label-key1": "label-value1",
							"label-key2": "label-value2",
							"old-label":  "old-value",
						},
					},
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"label-key1": "label-value1",
								"label-key2": "label-value2",
							},
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "nginx",
									"image": "nginx:1.7.9",
								},
							},
						},
					},
				},
			}),
		resid.NewResId(cronjob, "cronjob1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "batch/v1beta1",
				"kind":       "CronJob",
				"metadata": map[string]interface{}{
					"name": "cronjob1",
					"labels": map[string]interface{}{
						"label-key1": "label-value1",
						"label-key2": "label-value2",
					},
				},
				"spec": map[string]interface{}{
					"schedule": "* 23 * * *",
					"jobTemplate": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"label-key1": "label-value1",
								"label-key2": "label-value2",
							},
						},
						"spec": map[string]interface{}{
							"template": map[string]interface{}{
								"metadata": map[string]interface{}{
									"labels": map[string]interface{}{
										"label-key1": "label-value1",
										"label-key2": "label-value2",
									},
								},
								"spec": map[string]interface{}{
									"containers": []interface{}{
										map[string]interface{}{
											"name":  "nginx",
											"image": "nginx:1.7.9",
										},
									},
								},
							},
						},
					},
				},
			}),
		resid.NewResId(cronjob, "cronjob2"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "batch/v1beta1",
				"kind":       "CronJob",
				"metadata": map[string]interface{}{
					"name": "cronjob2",
					"labels": map[string]interface{}{
						"label-key1": "label-value1",
						"label-key2": "label-value2",
					},
				},
				"spec": map[string]interface{}{
					"schedule": "* 23 * * *",
					"jobTemplate": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"label-key1": "label-value1",
								"label-key2": "label-value2",
							},
						},
						"spec": map[string]interface{}{
							"selector": map[string]interface{}{
								"matchLabels": map[string]interface{}{
									"old-label":  "old-value",
									"label-key1": "label-value1",
									"label-key2": "label-value2",
								},
							},
							"template": map[string]interface{}{
								"metadata": map[string]interface{}{
									"labels": map[string]interface{}{
										"label-key1": "label-value1",
										"label-key2": "label-value2",
									},
								},
								"spec": map[string]interface{}{
									"containers": []interface{}{
										map[string]interface{}{
											"name":  "nginx",
											"image": "nginx:1.7.9",
										},
									},
								},
							},
						},
					},
				},
			}),
	}
	lt, err := NewLabelsMapTransformer(
		map[string]string{"label-key1": "label-value1", "label-key2": "label-value2"},
		defaultTransformerConfig.CommonLabels)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = lt.Transform(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(m, expected) {
		err = expected.ErrorIfNotEqual(m)
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}

func TestAnnotationsRun(t *testing.T) {
	m := resmap.ResMap{
		resid.NewResId(cmap, "cm1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cm1",
				},
			}),
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
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"old-label": "old-value",
							},
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "nginx",
									"image": "nginx:1.7.9",
								},
							},
						},
					},
				},
			}),
		resid.NewResId(service, "svc1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name": "svc1",
				},
				"spec": map[string]interface{}{
					"ports": []interface{}{
						map[string]interface{}{
							"name": "port1",
							"port": "12345",
						},
					},
				},
			}),
	}
	expected := resmap.ResMap{
		resid.NewResId(cmap, "cm1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name": "cm1",
					"annotations": map[string]interface{}{
						"anno-key1": "anno-value1",
						"anno-key2": "anno-value2",
					},
				},
			}),
		resid.NewResId(deploy, "deploy1"): rf.FromMap(
			map[string]interface{}{
				"group":      "apps",
				"apiVersion": "v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "deploy1",
					"annotations": map[string]interface{}{
						"anno-key1": "anno-value1",
						"anno-key2": "anno-value2",
					},
				},
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"annotations": map[string]interface{}{
								"anno-key1": "anno-value1",
								"anno-key2": "anno-value2",
							},
							"labels": map[string]interface{}{
								"old-label": "old-value",
							},
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "nginx",
									"image": "nginx:1.7.9",
								},
							},
						},
					},
				},
			}),
		resid.NewResId(service, "svc1"): rf.FromMap(
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name": "svc1",
					"annotations": map[string]interface{}{
						"anno-key1": "anno-value1",
						"anno-key2": "anno-value2",
					},
				},
				"spec": map[string]interface{}{
					"ports": []interface{}{
						map[string]interface{}{
							"name": "port1",
							"port": "12345",
						},
					},
				},
			}),
	}
	at, err := NewAnnotationsMapTransformer(
		map[string]string{"anno-key1": "anno-value1", "anno-key2": "anno-value2"},
		defaultTransformerConfig.CommonAnnotations)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = at.Transform(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(m, expected) {
		err = expected.ErrorIfNotEqual(m)
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}
