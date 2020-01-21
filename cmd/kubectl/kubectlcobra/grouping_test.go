// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// package kubectlcobra contains cobra commands from kubectl
package kubectlcobra

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"
)

var testNamespace = "test-grouping-namespace"
var groupingObjName = "test-grouping-obj"
var pod1Name = "pod-1"
var pod2Name = "pod-2"
var pod3Name = "pod-3"

var testGroupingLabel = "test-app-label"

var groupingObj = unstructured.Unstructured{
	Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name":      groupingObjName,
			"namespace": testNamespace,
			"labels": map[string]interface{}{
				GroupingLabel: testGroupingLabel,
			},
		},
	},
}

var pod1 = unstructured.Unstructured{
	Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]interface{}{
			"name":      pod1Name,
			"namespace": testNamespace,
		},
	},
}

var pod1Info = &resource.Info{
	Namespace: testNamespace,
	Name:      pod1Name,
	Object:    &pod1,
}

var pod2 = unstructured.Unstructured{
	Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]interface{}{
			"name":      pod2Name,
			"namespace": testNamespace,
		},
	},
}

var pod2Info = &resource.Info{
	Namespace: testNamespace,
	Name:      pod2Name,
	Object:    &pod2,
}

var pod3 = unstructured.Unstructured{
	Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]interface{}{
			"name":      pod3Name,
			"namespace": testNamespace,
		},
	},
}

var pod3Info = &resource.Info{
	Namespace: testNamespace,
	Name:      pod3Name,
	Object:    &pod3,
}

var nonUnstructuredGroupingObj = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Namespace: testNamespace,
		Name:      groupingObjName,
		Labels: map[string]string{
			GroupingLabel: "true",
		},
	},
}

var nonUnstructuredGroupingInfo = &resource.Info{
	Namespace: testNamespace,
	Name:      groupingObjName,
	Object:    nonUnstructuredGroupingObj,
}

var nilInfo = &resource.Info{
	Namespace: testNamespace,
	Name:      groupingObjName,
	Object:    nil,
}

var groupingObjLabelWithSpace = unstructured.Unstructured{
	Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name":      groupingObjName,
			"namespace": testNamespace,
			"labels": map[string]interface{}{
				GroupingLabel: "\tgrouping-label ",
			},
		},
	},
}

func TestRetrieveGroupingLabel(t *testing.T) {
	tests := []struct {
		obj           runtime.Object
		groupingLabel string
		isError       bool
	}{
		// Nil grouping object throws error.
		{
			obj:           nil,
			groupingLabel: "",
			isError:       true,
		},
		// Pod is not a grouping object.
		{
			obj:           &pod2,
			groupingLabel: "",
			isError:       true,
		},
		// Retrieves label without preceding/trailing whitespace.
		{
			obj:           &groupingObjLabelWithSpace,
			groupingLabel: "grouping-label",
			isError:       false,
		},
		{
			obj:           &groupingObj,
			groupingLabel: testGroupingLabel,
			isError:       false,
		},
	}

	for _, test := range tests {
		actual, err := retrieveGroupingLabel(test.obj)
		if test.isError && err == nil {
			t.Errorf("Did not receive expected error.\n")
		}
		if !test.isError {
			if err != nil {
				t.Fatalf("Received unexpected error: %s\n", err)
			}
			if test.groupingLabel != actual {
				t.Errorf("Expected grouping label (%s), got (%s)\n", test.groupingLabel, actual)
			}
		}
	}
}

func TestIsGroupingObject(t *testing.T) {
	tests := []struct {
		obj        runtime.Object
		isGrouping bool
	}{
		{
			obj:        nil,
			isGrouping: false,
		},
		{
			obj:        &groupingObj,
			isGrouping: true,
		},
		{
			obj:        &pod2,
			isGrouping: false,
		},
	}

	for _, test := range tests {
		grouping := isGroupingObject(test.obj)
		if test.isGrouping && !grouping {
			t.Errorf("Grouping object not identified: %#v", test.obj)
		}
		if !test.isGrouping && grouping {
			t.Errorf("Non-grouping object identifed as grouping obj: %#v", test.obj)
		}
	}
}

func TestFindGroupingObject(t *testing.T) {
	tests := []struct {
		infos  []*resource.Info
		exists bool
		name   string
	}{
		{
			infos:  []*resource.Info{},
			exists: false,
			name:   "",
		},
		{
			infos:  []*resource.Info{nil},
			exists: false,
			name:   "",
		},
		{
			infos:  []*resource.Info{copyGroupingInfo()},
			exists: true,
			name:   groupingObjName,
		},
		{
			infos:  []*resource.Info{pod1Info},
			exists: false,
			name:   "",
		},
		{
			infos:  []*resource.Info{pod1Info, pod2Info, pod3Info},
			exists: false,
			name:   "",
		},
		{
			infos:  []*resource.Info{pod1Info, pod2Info, copyGroupingInfo(), pod3Info},
			exists: true,
			name:   groupingObjName,
		},
	}

	for _, test := range tests {
		groupingObj, found := findGroupingObject(test.infos)
		if test.exists && !found {
			t.Errorf("Should have found grouping object")
		}
		if !test.exists && found {
			t.Errorf("Grouping object found, but it does not exist: %#v", groupingObj)
		}
		if test.exists && found && test.name != groupingObj.Name {
			t.Errorf("Grouping object name does not match: %s/%s", test.name, groupingObj.Name)
		}
	}
}

func TestSortGroupingObject(t *testing.T) {
	tests := []struct {
		infos  []*resource.Info
		sorted bool
	}{
		{
			infos:  []*resource.Info{},
			sorted: false,
		},
		{
			infos:  []*resource.Info{copyGroupingInfo()},
			sorted: true,
		},
		{
			infos:  []*resource.Info{pod1Info},
			sorted: false,
		},
		{
			infos:  []*resource.Info{pod1Info, pod2Info},
			sorted: false,
		},
		{
			infos:  []*resource.Info{copyGroupingInfo(), pod1Info},
			sorted: true,
		},
		{
			infos:  []*resource.Info{pod1Info, copyGroupingInfo()},
			sorted: true,
		},
		{
			infos:  []*resource.Info{pod1Info, pod2Info, copyGroupingInfo(), pod3Info},
			sorted: true,
		},
		{
			infos:  []*resource.Info{pod1Info, pod2Info, pod3Info, copyGroupingInfo()},
			sorted: true,
		},
		{
			infos:  []*resource.Info{copyGroupingInfo(), pod1Info, pod2Info, pod3Info},
			sorted: true,
		},
	}

	for _, test := range tests {
		wasSorted := sortGroupingObject(test.infos)
		if wasSorted && !test.sorted {
			t.Errorf("Grouping object was sorted, but it shouldn't have been")
		}
		if !wasSorted && test.sorted {
			t.Errorf("Grouping object was NOT sorted, but it should have been")
		}
		if wasSorted {
			first := test.infos[0]
			if !isGroupingObject(first.Object) {
				t.Errorf("Grouping object was not sorted into first position")
			}
		}
	}
}

func TestAddRetrieveInventoryToFromGroupingObject(t *testing.T) {
	tests := []struct {
		infos    []*resource.Info
		expected []*Inventory
		isError  bool
	}{
		// No grouping object is an error.
		{
			infos:   []*resource.Info{},
			isError: true,
		},
		// No grouping object is an error.
		{
			infos:   []*resource.Info{pod1Info, pod2Info},
			isError: true,
		},
		// Grouping object without other objects is OK.
		{
			infos:   []*resource.Info{copyGroupingInfo(), nilInfo},
			isError: true,
		},
		{
			infos:   []*resource.Info{nonUnstructuredGroupingInfo},
			isError: true,
		},
		{
			infos:    []*resource.Info{copyGroupingInfo()},
			expected: []*Inventory{},
			isError:  false,
		},
		// More than one grouping object is an error.
		{
			infos:    []*resource.Info{copyGroupingInfo(), copyGroupingInfo()},
			expected: []*Inventory{},
			isError:  true,
		},
		// More than one grouping object is an error.
		{
			infos:    []*resource.Info{copyGroupingInfo(), pod1Info, copyGroupingInfo()},
			expected: []*Inventory{},
			isError:  true,
		},
		// Basic test case: one grouping object, one pod.
		{
			infos: []*resource.Info{copyGroupingInfo(), pod1Info},
			expected: []*Inventory{
				&Inventory{
					Namespace: testNamespace,
					Name:      pod1Name,
					GroupKind: schema.GroupKind{
						Group: "",
						Kind:  "Pod",
					},
				},
			},
			isError: false,
		},
		{
			infos: []*resource.Info{pod1Info, copyGroupingInfo()},
			expected: []*Inventory{
				&Inventory{
					Namespace: testNamespace,
					Name:      pod1Name,
					GroupKind: schema.GroupKind{
						Group: "",
						Kind:  "Pod",
					},
				},
			},
			isError: false,
		},
		{
			infos: []*resource.Info{pod1Info, pod2Info, copyGroupingInfo(), pod3Info},
			expected: []*Inventory{
				&Inventory{
					Namespace: testNamespace,
					Name:      pod1Name,
					GroupKind: schema.GroupKind{
						Group: "",
						Kind:  "Pod",
					},
				},
				&Inventory{
					Namespace: testNamespace,
					Name:      pod2Name,
					GroupKind: schema.GroupKind{
						Group: "",
						Kind:  "Pod",
					},
				},
				&Inventory{
					Namespace: testNamespace,
					Name:      pod3Name,
					GroupKind: schema.GroupKind{
						Group: "",
						Kind:  "Pod",
					},
				},
			},
			isError: false,
		},
		{
			infos: []*resource.Info{pod1Info, pod2Info, pod3Info, copyGroupingInfo()},
			expected: []*Inventory{
				&Inventory{
					Namespace: testNamespace,
					Name:      pod1Name,
					GroupKind: schema.GroupKind{
						Group: "",
						Kind:  "Pod",
					},
				},
				&Inventory{
					Namespace: testNamespace,
					Name:      pod2Name,
					GroupKind: schema.GroupKind{
						Group: "",
						Kind:  "Pod",
					},
				},
				&Inventory{
					Namespace: testNamespace,
					Name:      pod3Name,
					GroupKind: schema.GroupKind{
						Group: "",
						Kind:  "Pod",
					},
				},
			},
			isError: false,
		},
		{
			infos: []*resource.Info{copyGroupingInfo(), pod1Info, pod2Info, pod3Info},
			expected: []*Inventory{
				&Inventory{
					Namespace: testNamespace,
					Name:      pod1Name,
					GroupKind: schema.GroupKind{
						Group: "",
						Kind:  "Pod",
					},
				},
				&Inventory{
					Namespace: testNamespace,
					Name:      pod2Name,
					GroupKind: schema.GroupKind{
						Group: "",
						Kind:  "Pod",
					},
				},
				&Inventory{
					Namespace: testNamespace,
					Name:      pod3Name,
					GroupKind: schema.GroupKind{
						Group: "",
						Kind:  "Pod",
					},
				},
			},
			isError: false,
		},
	}

	for _, test := range tests {
		err := addInventoryToGroupingObj(test.infos)
		if test.isError && err == nil {
			t.Errorf("Should have produced an error, but returned none.")
		}
		if !test.isError {
			if err != nil {
				t.Fatalf("Received error when expecting none (%s)\n", err)
			}
			retrieved, err := retrieveInventoryFromGroupingObj(test.infos)
			if err != nil {
				t.Fatalf("Error retrieving inventory: %s\n", err)
			}
			if len(test.expected) != len(retrieved) {
				t.Errorf("Expected inventory for %d resources, actual %d",
					len(test.expected), len(retrieved))
			}
			for _, expected := range test.expected {
				found := false
				for _, actual := range retrieved {
					if expected.Equals(actual) {
						found = true
						continue
					}
				}
				if !found {
					t.Errorf("Expected inventory (%s) not found", expected)
				}
			}
			// If the grouping object has an inventory, check the
			// grouping object has an inventory hash.
			groupingInfo, exists := findGroupingObject(test.infos)
			if exists && len(test.expected) > 0 {
				invHash := retrieveInventoryHash(groupingInfo)
				if len(invHash) == 0 {
					t.Errorf("Grouping object missing inventory hash")
				}
			}
		}
	}
}

func TestAddSuffixToName(t *testing.T) {
	tests := []struct {
		info     *resource.Info
		suffix   string
		expected string
		isError  bool
	}{
		// Nil info should return error.
		{
			info:     nil,
			suffix:   "",
			expected: "",
			isError:  true,
		},
		// Empty suffix should return error.
		{
			info:     copyGroupingInfo(),
			suffix:   "",
			expected: "",
			isError:  true,
		},
		// Empty suffix should return error.
		{
			info:     copyGroupingInfo(),
			suffix:   " \t",
			expected: "",
			isError:  true,
		},
		{
			info:     copyGroupingInfo(),
			suffix:   "hashsuffix",
			expected: groupingObjName + "-hashsuffix",
			isError:  false,
		},
	}

	for _, test := range tests {
		//t.Errorf("%#v [%s]", test.info, test.suffix)
		err := addSuffixToName(test.info, test.suffix)
		if test.isError {
			if err == nil {
				t.Errorf("Should have produced an error, but returned none.")
			}
		}
		if !test.isError {
			if err != nil {
				t.Fatalf("Received error when expecting none (%s)\n", err)
			}
			actualName, err := getObjectName(test.info.Object)
			if err != nil {
				t.Fatalf("Error getting object name: %s", err)
			}
			if actualName != test.info.Name {
				t.Errorf("Object name (%s) does not match info name (%s)\n", actualName, test.info.Name)
			}
			if test.expected != actualName {
				t.Errorf("Expected name (%s), got (%s)\n", test.expected, actualName)
			}
		}
	}
}

func getObjectName(obj runtime.Object) (string, error) {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return "", fmt.Errorf("Grouping object is not Unstructured format")
	}
	return u.GetName(), nil
}

func copyGroupingInfo() *resource.Info {
	groupingObjCopy := groupingObj.DeepCopy()
	var groupingInfo = &resource.Info{
		Namespace: testNamespace,
		Name:      groupingObjName,
		Object:    groupingObjCopy,
	}
	return groupingInfo
}
