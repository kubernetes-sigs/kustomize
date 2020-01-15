// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// package kubectlcobra contains cobra commands from kubectl
package kubectlcobra

import (
	"testing"

	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/kubectl/pkg/cmd/apply"
)

func TestPrependGroupingObject(t *testing.T) {
	tests := []struct {
		infos []*resource.Info
	}{
		{
			infos: []*resource.Info{copyGroupingInfo()},
		},
		{
			infos: []*resource.Info{pod1Info, pod3Info, copyGroupingInfo()},
		},
		{
			infos: []*resource.Info{pod1Info, pod2Info, copyGroupingInfo(), pod3Info},
		},
	}

	for _, test := range tests {
		applyOptions := createApplyOptions(test.infos)
		f := PrependGroupingObject(applyOptions)
		err := f()
		if err != nil {
			t.Errorf("Error running pre-processor callback: %s", err)
		}
		infos, _ := applyOptions.GetObjects()
		if len(test.infos) != len(infos) {
			t.Fatalf("Wrong number of objects after prepending grouping object")
		}
		groupingInfo := infos[0]
		if !isGroupingObject(groupingInfo.Object) {
			t.Fatalf("First object is not the grouping object")
		}
		inventory, _ := retrieveInventoryFromGroupingObj(infos)
		if len(inventory) != (len(infos) - 1) {
			t.Errorf("Wrong number of inventory items stored in grouping object")
		}
	}

}

// createApplyOptions is a helper function to assemble the ApplyOptions
// with the passed objects (infos).
func createApplyOptions(infos []*resource.Info) *apply.ApplyOptions {
	applyOptions := &apply.ApplyOptions{}
	applyOptions.SetObjects(infos)
	return applyOptions
}
