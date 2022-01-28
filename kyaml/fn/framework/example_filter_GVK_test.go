package framework_test

import (
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

// This example implements a function that reads the desired replicas from the
// functionConfig and updates the replicas field for all deployments.

func Example_filterGVK() {
	if err := command.AsMain(framework.ResourceListProcessorFunc(updateReplicas)); err != nil {
		os.Exit(1)
	}
}

// updateReplicas sets a field in resources selecting by GVK.
func updateReplicas(rl *framework.ResourceList) error {
	if rl.FunctionConfig == nil {
		return framework.ErrMissingFnConfig{}
	}
	replicas, found, err := rl.FunctionConfig.GetNestedInt("replicas")
	if !found || err != nil {
		return err
	}
	for i, obj := range rl.Items {
		if obj.GetApiVersion() == "apps/v1" && obj.GetKind() == "Deployment" {
			rl.Items[i].SetOrDie(replicas, "spec", "replicas")
		}
	}
	return nil
}
