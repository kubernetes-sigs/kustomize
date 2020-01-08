// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// package kubectlcobra contains cobra commands from kubectl
package kubectlcobra

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"
)

var testNamespace = "test-grouping-namespace"
var groupingObjName = "test-grouping-obj"
var pod1Name = "pod-1"
var pod2Name = "pod-2"
var pod3Name = "pod-3"

var groupingObj = corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Namespace: testNamespace,
		Name:      groupingObjName,
		Labels: map[string]string{
			GroupingLabel: "true",
		},
	},
}

var groupingInfo = &resource.Info{
	Namespace: testNamespace,
	Name:      groupingObjName,
	Object:    &groupingObj,
}

var pod1 = corev1.Pod{
	ObjectMeta: metav1.ObjectMeta{
		Namespace: testNamespace,
		Name:      pod1Name,
	},
}

var pod1Info = &resource.Info{
	Namespace: testNamespace,
	Name:      pod1Name,
	Object:    &pod1,
}

var pod2 = corev1.Pod{
	ObjectMeta: metav1.ObjectMeta{
		Namespace: testNamespace,
		Name:      pod2Name,
	},
}

var pod2Info = &resource.Info{
	Namespace: testNamespace,
	Name:      pod2Name,
	Object:    &pod2,
}

var pod3 = corev1.Pod{
	ObjectMeta: metav1.ObjectMeta{
		Namespace: testNamespace,
		Name:      pod3Name,
	},
}

var pod3Info = &resource.Info{
	Namespace: testNamespace,
	Name:      pod3Name,
	Object:    &pod3,
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
			infos:  []*resource.Info{groupingInfo},
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
			infos:  []*resource.Info{pod1Info, pod2Info, groupingInfo, pod3Info},
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
			infos:  []*resource.Info{groupingInfo},
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
			infos:  []*resource.Info{groupingInfo, pod1Info},
			sorted: true,
		},
		{
			infos:  []*resource.Info{pod1Info, groupingInfo},
			sorted: true,
		},
		{
			infos:  []*resource.Info{pod1Info, pod2Info, groupingInfo, pod3Info},
			sorted: true,
		},
		{
			infos:  []*resource.Info{pod1Info, pod2Info, pod3Info, groupingInfo},
			sorted: true,
		},
		{
			infos:  []*resource.Info{groupingInfo, pod1Info, pod2Info, pod3Info},
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
