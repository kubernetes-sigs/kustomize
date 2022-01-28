package framework_test

import (
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

// In this example, we read a field from the input object and print it to the log.

func Example_aReadField() {
	if err := command.AsMain(framework.ResourceListProcessorFunc(readField)); err != nil {
		os.Exit(1)
	}
}

func readField(rl *framework.ResourceList) error {
	for _, obj := range rl.Items {
		if obj.GetApiVersion() == "apps/v1" && obj.GetKind() == "Deployment" {
			replicas, found, err := obj.GetNestedInt("spec", "replicas")
			if !found || err != nil {
				return err
			}
			framework.Logf("replicas is %v\n", replicas)
		}
	}
	return nil
}
