// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package patch

import (
	"reflect"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/v3/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/resmaptest"
	"sigs.k8s.io/kustomize/v3/pkg/resource"
)

// simple utility function to add an namespace in a resource
// map used as base, patch or expected result.
func addNamespace(namespace string, base map[string]interface{}) map[string]interface{} {
	metadata := base["metadata"].(map[string]interface{})
	metadata["namespace"] = namespace
	return base
}

// unExpectedError function handles unexpected error
func unExpectedError(t *testing.T, name string, err error) {
	t.Fatalf("%q; - unexpected error %v", name, err)
}

// compareExpectedError compares the expectedError and the actualError return by GetFieldValue
func compareExpectedError(t *testing.T, name string, err error, errorMsg string) {
	if err == nil {
		t.Fatalf("%q; - should return error, but no error returned", name)
	}

	if !strings.Contains(err.Error(), errorMsg) {
		t.Fatalf("%q; - expected error: \"%s\", got error: \"%v\"",
			name, errorMsg, err.Error())
	}
}

// compareValues compares the expectedValue and actualValue returned by Transform
func compareValues(t *testing.T, name string, expectedValue resmap.ResMap, actualValue resmap.ResMap) {
	if !reflect.DeepEqual(expectedValue, actualValue) {
		err := expectedValue.ErrorIfNotEqualLists(actualValue)
		t.Logf("%q; actual doesn't match expected: %v", name, err)
	}
}

var rf = resource.NewFactory(
	kunstruct.NewKunstructuredFactoryImpl())

const Deployment string = "Deployment"
const MyCRD string = "MyCRD"

// baseResource produces a base object which used to test
// patch transformation
// Also the structure is matching the Deployment syntax
// the kind can be replaced to allow testing using CRD
// without access to the schema
func baseResource(kind string) map[string]interface{} {

	return map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       kind,
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
							"image": "nginx",
						},
					},
				},
			},
		},
	}

}

// addContainerAndEnvPatch produces a patch object which adds
// an entry in the env slice of the first/nginx container
// as well as adding a label in the metadata
// Note that for SMP/WithSchema merge, the "name:nginx" entry
// is mandatory
func addLabelAndEnvPatch(kind string) map[string]interface{} {

	return map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       kind,
		"metadata": map[string]interface{}{
			"name": "deploy1",
		},
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"some-label": "some-value",
					},
				},
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"name": "nginx",
							"env": []interface{}{
								map[string]interface{}{
									"name":  "SOMEENV",
									"value": "SOMEVALUE",
								},
							},
						},
					},
				},
			},
		},
	}
}

// addContainerAndEnvPatch produces a patch object which adds
// an entry in the env slice of the first/nginx container
// as well as adding a second container in the container list
// Note that for SMP/WithSchema merge, the "name:nginx" entry
// is mandatory
func addContainerAndEnvPatch(kind string) map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       kind,
		"metadata": map[string]interface{}{
			"name": "deploy1",
		},
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"name": "nginx",
							"env": []interface{}{
								map[string]interface{}{
									"name":  "ANOTHERENV",
									"value": "ANOTHERVALUE",
								},
							},
						},
						map[string]interface{}{
							"name":  "anothercontainer",
							"image": "anotherimage",
						},
					},
				},
			},
		},
	}
}

// addContainerAndEnvPatch produces a patch object which replaces
// the value of the image field in the first/nginx container
// Note that for SMP/WithSchema merge, the "name:nginx" entry
// is mandatory
func changeImagePatch(kind string, newImage string) map[string]interface{} {

	return map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       kind,
		"metadata": map[string]interface{}{
			"name": "deploy1",
		},
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"name":  "nginx",
							"image": newImage,
						},
					},
				},
			},
		},
	}
}

// utility method building the expected output of a SMP
// TODO(jeb): nginxContainer part of the tree is still singled out
// to highlight the difference in behavior with JSONPatch
func expectedResultSMP() map[string]interface{} {

	// This will have been merged using StrategicMergePatch
	nginxContainer := map[string]interface{}{
		"name":  "nginx",
		"image": "nginx", // main difference with JSONPatch
		"env": []interface{}{
			map[string]interface{}{
				"name":  "SOMEENV",
				"value": "SOMEVALUE",
			},
		},
	}

	return map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       Deployment,
		"metadata": map[string]interface{}{
			"name": "deploy1",
		},
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"old-label":  "old-value",
						"some-label": "some-value",
					},
				},
				"spec": map[string]interface{}{
					"containers": []interface{}{
						nginxContainer,
					},
				},
			},
		},
	}
}

// utility method building the expected output of a JsonPatch.
// TODO(jeb): imagename parameter is only here to deal with what
// seems to be a bug in conflictdetector and JSON patch
func expectedResultJsonPatch(imagename string) map[string]interface{} {

	// This will have been merged using JSON Patch
	nginxContainer := map[string]interface{}{
		"name": "nginx",
		"env": []interface{}{
			map[string]interface{}{
				"name":  "SOMEENV",
				"value": "SOMEVALUE",
			},
		},
	}

	// This piece of code is only here to deal
	// a bug in conflictdetector
	if imagename != "" {
		nginxContainer = map[string]interface{}{
			"name":  "nginx",
			"image": imagename,
		}
	}

	return map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       MyCRD,
		"metadata": map[string]interface{}{
			"name": "deploy1",
		},
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"old-label":  "old-value",
						"some-label": "some-value",
					},
				},
				"spec": map[string]interface{}{
					"containers": []interface{}{
						nginxContainer,
					},
				},
			},
		},
	}
}

// utility method to build the expected result of a multipatch
// the order of the patches still have influence especially
// in the insertion location within arrays.
func expectedResultMultiPatch(kind string, reversed bool) map[string]interface{} {
	envs := []interface{}{
		map[string]interface{}{
			"name":  "ANOTHERENV",
			"value": "ANOTHERVALUE",
		},
		map[string]interface{}{
			"name":  "SOMEENV",
			"value": "SOMEVALUE",
		},
	}

	if reversed {
		envs = []interface{}{
			map[string]interface{}{
				"name":  "SOMEENV",
				"value": "SOMEVALUE",
			},
			map[string]interface{}{
				"name":  "ANOTHERENV",
				"value": "ANOTHERVALUE",
			},
		}

	}

	return map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       kind,
		"metadata": map[string]interface{}{
			"name": "deploy1",
		},
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"old-label":  "old-value",
						"some-label": "some-value",
					},
				},
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"name":  "nginx",
							"image": "nginx:latest",
							"env":   envs,
						},
						map[string]interface{}{
							"name":  "anothercontainer",
							"image": "anotherimage",
						},
					},
				},
			},
		},
	}
}

// TestOverlayRun validates the single patch use cases
// regarless of the schema availibility, which in turns
// relies on StrategicMergePatch or simple JSON Patch.
func TestOverlayRun(t *testing.T) {
	tests := []struct {
		name          string
		base          resmap.ResMap
		patch         []*resource.Resource
		expected      resmap.ResMap
		errorExpected bool
		errorMsg      string
	}{
		{
			name: "withschema",
			base: resmaptest_test.NewRmBuilder(t, rf).
				Add(baseResource(Deployment)).ResMap(),
			patch: []*resource.Resource{
				rf.FromMap(addLabelAndEnvPatch(Deployment)),
			},
			errorExpected: false,
			expected: resmaptest_test.NewRmBuilder(t, rf).
				Add(expectedResultSMP()).ResMap(),
		},
		{
			name: "noschema",
			base: resmaptest_test.NewRmBuilder(t, rf).
				Add(baseResource(MyCRD)).ResMap(),
			patch: []*resource.Resource{
				rf.FromMap(addLabelAndEnvPatch(MyCRD)),
			},
			errorExpected: false,
			expected: resmaptest_test.NewRmBuilder(t, rf).
				Add(expectedResultJsonPatch("")).ResMap(),
		},
	}
	for _, test := range tests {
		lt, err := NewTransformer(test.patch, rf)
		if err != nil {
			unExpectedError(t, test.name, err)
		}

		err = lt.Transform(test.base)
		if test.errorExpected {
			compareExpectedError(t, test.name, err, test.errorMsg)
			continue
		}
		if err != nil {
			unExpectedError(t, test.name, err)
		}
		compareValues(t, test.name, test.base, test.expected)
	}
}

// TestMultiplePatches checks that the patches are applied
// properly, that the same result is obtained,
// regardless of the order of the patches and regardless
// of the schema availibility (SMP vs JSON)
func TestMultiplePatches(t *testing.T) {
	tests := []struct {
		name          string
		base          resmap.ResMap
		patch         []*resource.Resource
		expected      resmap.ResMap
		errorExpected bool
		errorMsg      string
	}{
		{
			name: "withschema-label-image-container",
			base: resmaptest_test.NewRmBuilder(t, rf).
				Add(baseResource(Deployment)).ResMap(),
			patch: []*resource.Resource{
				rf.FromMap(addLabelAndEnvPatch(Deployment)),
				rf.FromMap(changeImagePatch(Deployment, "nginx:latest")),
				rf.FromMap(addContainerAndEnvPatch(Deployment)),
			},
			errorExpected: false,
			expected: resmaptest_test.NewRmBuilder(t, rf).
				Add(expectedResultMultiPatch(Deployment, false)).ResMap(),
		},
		{
			name: "withschema-image-container-label",
			base: resmaptest_test.NewRmBuilder(t, rf).
				Add(baseResource(Deployment)).ResMap(),
			patch: []*resource.Resource{
				rf.FromMap(changeImagePatch(Deployment, "nginx:latest")),
				rf.FromMap(addContainerAndEnvPatch(Deployment)),
				rf.FromMap(addLabelAndEnvPatch(Deployment)),
			},
			errorExpected: false,
			expected: resmaptest_test.NewRmBuilder(t, rf).
				Add(expectedResultMultiPatch(Deployment, true)).ResMap(),
		},
		{
			name: "withschema-container-label-image",
			base: resmaptest_test.NewRmBuilder(t, rf).
				Add(baseResource(Deployment)).ResMap(),
			patch: []*resource.Resource{
				rf.FromMap(addContainerAndEnvPatch(Deployment)),
				rf.FromMap(addLabelAndEnvPatch(Deployment)),
				rf.FromMap(changeImagePatch(Deployment, "nginx:latest")),
			},
			errorExpected: false,
			expected: resmaptest_test.NewRmBuilder(t, rf).
				Add(expectedResultMultiPatch(Deployment, true)).ResMap(),
		},
		{
			name: "noschema-label-image-container",
			base: resmaptest_test.NewRmBuilder(t, rf).
				Add(baseResource(MyCRD)).ResMap(),
			patch: []*resource.Resource{
				rf.FromMap(addLabelAndEnvPatch(MyCRD)),
				rf.FromMap(changeImagePatch(MyCRD, "nginx:latest")),
				rf.FromMap(addContainerAndEnvPatch(MyCRD)),
			},
			// This should work
			errorExpected: true,
			errorMsg:      "conflict",
		},
		{
			name: "noschema-image-container-label",
			base: resmaptest_test.NewRmBuilder(t, rf).
				Add(baseResource(MyCRD)).ResMap(),
			patch: []*resource.Resource{
				rf.FromMap(changeImagePatch(MyCRD, "nginx:latest")),
				rf.FromMap(addContainerAndEnvPatch(MyCRD)),
				rf.FromMap(addLabelAndEnvPatch(MyCRD)),
			},
			// This should work
			errorExpected: true,
			errorMsg:      "conflict",
		},
		{
			name: "noschema-container-label-image",
			base: resmaptest_test.NewRmBuilder(t, rf).
				Add(baseResource(MyCRD)).ResMap(),
			patch: []*resource.Resource{
				rf.FromMap(addContainerAndEnvPatch(MyCRD)),
				rf.FromMap(addLabelAndEnvPatch(MyCRD)),
				rf.FromMap(changeImagePatch(MyCRD, "nginx:latest")),
			},
			// This should work
			errorExpected: true,
			errorMsg:      "conflict",
		},
	}
	for _, test := range tests {
		lt, err := NewTransformer(test.patch, rf)
		if err != nil {
			unExpectedError(t, test.name, err)
		}

		err = lt.Transform(test.base)
		if test.errorExpected {
			compareExpectedError(t, test.name, err, test.errorMsg)
			continue
		}
		if err != nil {
			unExpectedError(t, test.name, err)
		}
		compareValues(t, test.name, test.base, test.expected)
	}

}

// TestMultiplePatchesWithConflict checks that the conflict are
// detected regardless of the order of the patches and regardless
// of the schema availibility (SMP vs JSON)
func TestMultiplePatchesWithConflict(t *testing.T) {
	tests := []struct {
		name          string
		base          resmap.ResMap
		patch         []*resource.Resource
		expected      resmap.ResMap
		errorExpected bool
		errorMsg      string
	}{
		{
			name: "withschema-label-latest-1.7.9",
			base: resmaptest_test.NewRmBuilder(t, rf).
				Add(baseResource(Deployment)).ResMap(),
			patch: []*resource.Resource{
				rf.FromMap(addLabelAndEnvPatch(Deployment)),
				rf.FromMap(changeImagePatch(Deployment, "nginx:latest")),
				rf.FromMap(changeImagePatch(Deployment, "nginx:1.7.9")),
			},
			errorExpected: true,
			errorMsg:      "conflict",
		},
		{
			name: "withschema-latest-label-1.7.9",
			base: resmaptest_test.NewRmBuilder(t, rf).
				Add(baseResource(Deployment)).ResMap(),
			patch: []*resource.Resource{
				rf.FromMap(changeImagePatch(Deployment, "nginx:latest")),
				rf.FromMap(addLabelAndEnvPatch(Deployment)),
				rf.FromMap(changeImagePatch(Deployment, "nginx:1.7.9")),
			},
			errorExpected: true,
			errorMsg:      "conflict",
		},
		{
			name: "withschema-1.7.9-label-latest",
			base: resmaptest_test.NewRmBuilder(t, rf).
				Add(baseResource(Deployment)).ResMap(),
			patch: []*resource.Resource{
				rf.FromMap(changeImagePatch(Deployment, "nginx:1.7.9")),
				rf.FromMap(addLabelAndEnvPatch(Deployment)),
				rf.FromMap(changeImagePatch(Deployment, "nginx:latest")),
			},
			errorExpected: true,
			errorMsg:      "conflict",
		},
		{
			name: "withschema-1.7.9-latest-label",
			base: resmaptest_test.NewRmBuilder(t, rf).
				Add(baseResource(Deployment)).ResMap(),
			patch: []*resource.Resource{
				rf.FromMap(changeImagePatch(Deployment, "nginx:1.7.9")),
				rf.FromMap(changeImagePatch(Deployment, "nginx:latest")),
				rf.FromMap(addLabelAndEnvPatch(Deployment)),
				rf.FromMap(changeImagePatch(Deployment, "nginx:nginx")),
			},
			errorExpected: true,
			errorMsg:      "conflict",
		},
		{
			name: "noschema-label-latest-1.7.9",
			base: resmaptest_test.NewRmBuilder(t, rf).
				Add(baseResource(MyCRD)).ResMap(),
			patch: []*resource.Resource{
				rf.FromMap(addLabelAndEnvPatch(MyCRD)),
				rf.FromMap(changeImagePatch(MyCRD, "nginx:latest")),
				rf.FromMap(changeImagePatch(MyCRD, "nginx:1.7.9")),
			},
			errorExpected: true,
			errorMsg:      "conflict",
		},
		{
			name: "noschema-latest-label-1.7.9",
			base: resmaptest_test.NewRmBuilder(t, rf).
				Add(baseResource(MyCRD)).ResMap(),
			patch: []*resource.Resource{
				rf.FromMap(changeImagePatch(MyCRD, "nginx:latest")),
				rf.FromMap(addLabelAndEnvPatch(MyCRD)),
				rf.FromMap(changeImagePatch(MyCRD, "nginx:1.7.9")),
			},
			errorExpected: false,
			//TODO(jeb): Why is there no conflict detected ?
			expected: resmaptest_test.NewRmBuilder(t, rf).
				Add(expectedResultJsonPatch("nginx:1.7.9")).ResMap(),
		},
		{
			name: "noschema-1.7.9-label-latest",
			base: resmaptest_test.NewRmBuilder(t, rf).
				Add(baseResource(MyCRD)).ResMap(),
			patch: []*resource.Resource{
				rf.FromMap(changeImagePatch(MyCRD, "nginx:1.7.9")),
				rf.FromMap(addLabelAndEnvPatch(MyCRD)),
				rf.FromMap(changeImagePatch(MyCRD, "nginx:latest")),
			},
			errorExpected: false,
			//TODO(jeb): Why is there no conflict detected ?
			expected: resmaptest_test.NewRmBuilder(t, rf).
				Add(expectedResultJsonPatch("nginx:latest")).ResMap(),
		},
		{
			name: "noschema-1.7.9-latest-label",
			base: resmaptest_test.NewRmBuilder(t, rf).
				Add(baseResource(MyCRD)).ResMap(),
			patch: []*resource.Resource{
				rf.FromMap(changeImagePatch(MyCRD, "nginx:1.7.9")),
				rf.FromMap(changeImagePatch(MyCRD, "nginx:latest")),
				rf.FromMap(addLabelAndEnvPatch(MyCRD)),
				rf.FromMap(changeImagePatch(MyCRD, "nginx:nginx")),
			},
			errorExpected: true,
		},
	}
	for _, test := range tests {
		lt, err := NewTransformer(test.patch, rf)
		if err != nil {
			unExpectedError(t, test.name, err)
		}

		err = lt.Transform(test.base)
		if test.errorExpected {
			compareExpectedError(t, test.name, err, test.errorMsg)
			continue
		}
		if err != nil {
			unExpectedError(t, test.name, err)
		}
		compareValues(t, test.name, test.base, test.expected)
	}
}

// TestMultipleNamespaces before the same patch
// on two objects have the same name but in a different namespaces
func TestMultipleNamespaces(t *testing.T) {
	tests := []struct {
		name          string
		base          resmap.ResMap
		patch         []*resource.Resource
		expected      resmap.ResMap
		errorExpected bool
		errorMsg      string
	}{
		{
			name: "withschema-ns1-ns2",
			base: resmaptest_test.NewRmBuilder(t, rf).
				Add(addNamespace("ns1", baseResource(Deployment))).
				Add(addNamespace("ns2", baseResource(Deployment))).
				ResMap(),
			patch: []*resource.Resource{
				rf.FromMap(addNamespace("ns1", addLabelAndEnvPatch(Deployment))),
				rf.FromMap(addNamespace("ns2", addLabelAndEnvPatch(Deployment))),
			},
			errorExpected: false,
			expected: resmaptest_test.NewRmBuilder(t, rf).
				Add(addNamespace("ns1", expectedResultSMP())).
				Add(addNamespace("ns2", expectedResultSMP())).
				ResMap(),
		},
		{
			name: "noschema-ns1-ns2",
			base: resmaptest_test.NewRmBuilder(t, rf).
				Add(addNamespace("ns1", baseResource(MyCRD))).
				Add(addNamespace("ns2", baseResource(MyCRD))).
				ResMap(),
			patch: []*resource.Resource{
				rf.FromMap(addNamespace("ns1", addLabelAndEnvPatch(MyCRD))),
				rf.FromMap(addNamespace("ns2", addLabelAndEnvPatch(MyCRD))),
			},
			errorExpected: false,
			expected: resmaptest_test.NewRmBuilder(t, rf).
				Add(addNamespace("ns1", expectedResultJsonPatch(""))).
				Add(addNamespace("ns2", expectedResultJsonPatch(""))).
				ResMap(),
		},
		{
			name: "withschema-ns1-ns2",
			base: resmaptest_test.NewRmBuilder(t, rf).
				Add(addNamespace("ns1", baseResource(Deployment))).
				ResMap(),
			patch: []*resource.Resource{
				rf.FromMap(addNamespace("ns2", changeImagePatch(Deployment, "nginx:1.7.9"))),
			},
			errorExpected: true,
			errorMsg:      "failed to find unique target for patch",
		},
		{
			name: "withschema-nil-ns2",
			base: resmaptest_test.NewRmBuilder(t, rf).
				Add(baseResource(Deployment)).
				ResMap(),
			patch: []*resource.Resource{
				rf.FromMap(addNamespace("ns2", changeImagePatch(Deployment, "nginx:1.7.9"))),
			},
			errorExpected: true,
			errorMsg:      "failed to find unique target for patch",
		},
		{
			name: "withschema-ns1-nil",
			base: resmaptest_test.NewRmBuilder(t, rf).
				Add(addNamespace("ns1", baseResource(Deployment))).
				ResMap(),
			patch: []*resource.Resource{
				rf.FromMap(changeImagePatch(Deployment, "nginx:1.7.9")),
			},
			errorExpected: true,
			errorMsg:      "failed to find unique target for patch",
		},
		{
			name: "noschema-ns1-ns2",
			base: resmaptest_test.NewRmBuilder(t, rf).
				Add(addNamespace("ns1", baseResource(MyCRD))).
				ResMap(),
			patch: []*resource.Resource{
				rf.FromMap(addNamespace("ns2", changeImagePatch(MyCRD, "nginx:1.7.9"))),
			},
			errorExpected: true,
			errorMsg:      "failed to find unique target for patch",
		},
		{
			name: "noschema-nil-ns2",
			base: resmaptest_test.NewRmBuilder(t, rf).
				Add(baseResource(MyCRD)).
				ResMap(),
			patch: []*resource.Resource{
				rf.FromMap(addNamespace("ns2", changeImagePatch(MyCRD, "nginx:1.7.9"))),
			},
			errorExpected: true,
			errorMsg:      "failed to find unique target for patch",
		},
		{
			name: "noschema-ns1-nil",
			base: resmaptest_test.NewRmBuilder(t, rf).
				Add(addNamespace("ns1", baseResource(MyCRD))).
				ResMap(),
			patch: []*resource.Resource{
				rf.FromMap(changeImagePatch(MyCRD, "nginx:1.7.9")),
			},
			errorExpected: true,
			errorMsg:      "failed to find unique target for patch",
		},
	}

	for _, test := range tests {
		lt, err := NewTransformer(test.patch, rf)
		if err != nil {
			unExpectedError(t, test.name, err)
		}

		err = lt.Transform(test.base)
		if test.errorExpected {
			compareExpectedError(t, test.name, err, test.errorMsg)
			continue
		}
		if err != nil {
			unExpectedError(t, test.name, err)
		}
		compareValues(t, test.name, test.base, test.expected)
	}
}
