package framework_test

import (
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

// This example implements a function that validate resources to ensure
// spec.template.spec.securityContext.runAsNonRoot is set in workload APIs.

func Example_validator() {
	if err := command.AsMain(framework.ResourceListProcessorFunc(validator)); err != nil {
		os.Exit(1)
	}
}

func validator(rl *framework.ResourceList) error {
	var results framework.Results
	for _, obj := range rl.Items {
		if obj.GetApiVersion() == "apps/v1" && (obj.GetKind() == "Deployment" || obj.GetKind() == "StatefulSet" || obj.GetKind() == "DaemonSet" || obj.GetKind() == "ReplicaSet") {
			runAsNonRoot, _, err := obj.GetNestedBool("spec", "template", "spec", "securityContext", "runAsNonRoot")
			if err != nil {
				return framework.ErrorConfigObjectResult(err, obj)
			}
			if !runAsNonRoot {
				results = append(results, framework.ConfigObjectResult("`spec.template.spec.securityContext.runAsNonRoot` must be set to true", obj, framework.Error))
			}
		}
	}
	return results
}
