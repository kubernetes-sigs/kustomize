// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package status_test

import (
	"fmt"
	"log"

	. "sigs.k8s.io/kustomize/kstatus/status"
	"sigs.k8s.io/kustomize/pseudo/k8s/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func ExampleCompute() {
	deploymentManifest := `
apiVersion: apps/v1
kind: Deployment
metadata:
   name: test
   generation: 1
   namespace: qual
status:
   observedGeneration: 1
   updatedReplicas: 1
   readyReplicas: 1
   availableReplicas: 1
   replicas: 1
   conditions:
    - type: Progressing 
      status: "True"
      reason: NewReplicaSetAvailable
    - type: Available 
      status: "True"
`
	deployment := yamlManifestToUnstructured(deploymentManifest)

	res, err := Compute(deployment)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res.Status)
	// Output:
	// Current
}

func ExampleAugment() {
	deploymentManifest := `
apiVersion: apps/v1
kind: Deployment
metadata:
   name: test
   generation: 1
   namespace: qual
status:
   observedGeneration: 1
   updatedReplicas: 1
   readyReplicas: 1
   availableReplicas: 1
   replicas: 1
   conditions:
    - type: Progressing 
      status: "True"
      reason: NewReplicaSetAvailable
    - type: Available 
      status: "True"
`
	deployment := yamlManifestToUnstructured(deploymentManifest)

	err := Augment(deployment)
	if err != nil {
		log.Fatal(err)
	}
	b, err := yaml.Marshal(deployment.Object)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   generation: 1
	//   name: test
	//   namespace: qual
	// status:
	//   availableReplicas: 1
	//   conditions:
	//   - reason: NewReplicaSetAvailable
	//     status: "True"
	//     type: Progressing
	//   - status: "True"
	//     type: Available
	//   observedGeneration: 1
	//   readyReplicas: 1
	//   replicas: 1
	//   updatedReplicas: 1
}

func yamlManifestToUnstructured(manifest string) *unstructured.Unstructured {
	jsonManifest, err := yaml.YAMLToJSON([]byte(manifest))
	if err != nil {
		log.Fatal(err)
	}
	resource, _, err := unstructured.UnstructuredJSONScheme.Decode(jsonManifest, nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	return resource.(*unstructured.Unstructured)
}
