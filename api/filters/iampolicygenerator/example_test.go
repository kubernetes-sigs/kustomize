// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package iampolicygenerator

import (
	"log"
	"os"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func ExampleFilter() {
	f := Filter{}
	var err = yaml.Unmarshal([]byte(`
cloud: gke
kubernetesService: 
  namespace: k8s-namespace
  name: k8s-sa-name
serviceAccount:
  name: gsa-name
  projectId: project-id
`), &f)
	if err != nil {
		log.Fatal(err)
	}

	err = kio.Pipeline{
		Inputs:  []kio.Reader{},
		Filters: []kio.Filter{f},
		Outputs: []kio.Writer{kio.ByteWriter{Writer: os.Stdout}},
	}.Execute()
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// apiVersion: v1
	// kind: ServiceAccount
	// metadata:
	//   annotations:
	//     iam.gke.io/gcp-service-account: gsa-name@project-id.iam.gserviceaccount.com
	//   name: k8s-sa-name
	//   namespace: k8s-namespace
}
