package framework_test

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// In this example, we implement a function that injects a logger as a sidecar
// container in workload APIs.

func Example_loggeInjector() {
	if err := command.AsMain(framework.ResourceListProcessorFunc(injectLogger)); err != nil {
		os.Exit(1)
	}
}

// injectLogger injects a logger container into the workload API resources.
// generate implements the goframework.KRMFunction interface.
func injectLogger(rl *framework.ResourceList) error {
	var li LoggerInjection
	if err := rl.FunctionConfig.As(&li); err != nil {
		return err
	}
	for i, obj := range rl.Items {
		if obj.GetApiVersion() == "apps/v1" && (obj.GetKind() == "Deployment" || obj.GetKind() == "StatefulSet" || obj.GetKind() == "DaemonSet" || obj.GetKind() == "ReplicaSet") {
			var container Container
			found, err := obj.Get(&container, "spec", "template", "spec", "containers", fmt.Sprintf("[name=%v]", li.ContainerName))
			if err != nil {
				return err
			}
			if found {
				container.Image = li.ImageName
			} else {
				container = Container{
					Name:  li.ContainerName,
					Image: li.ImageName,
				}
			}
			if err = rl.Items[i].Set(container, "spec", "template", "spec", "containers", fmt.Sprintf("[name=%v]", li.ContainerName)); err != nil {
				return err
			}
		}
	}
	return nil
}

// LoggerInjection is type definition of the functionConfig.
type LoggerInjection struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`

	ContainerName string `json:"containerName" yaml:"containerName"`
	ImageName     string `json:"imageName" yaml:"imageName"`
}

// Container is a copy of corev1.Container.
type Container struct {
	Name  string `json:"name" yaml:"name"`
	Image string `json:"image,omitempty" yaml:"image,omitempty"`
}
