// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kio_test

import (
	"bytes"
	"log"
	"os"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func Example() {
	input := bytes.NewReader([]byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: nginx
spec:
  selector:
    app: nginx
  ports:
    - protocol: TCP
      port: 80
      targetPort: 80
`))

	// setAnnotationFn
	setAnnotationFn := kio.FilterFunc(func(operand []*yaml.RNode) ([]*yaml.RNode, error) {
		for i := range operand {
			resource := operand[i]
			_, err := resource.Pipe(yaml.SetAnnotation("foo", "bar"))
			if err != nil {
				return nil, err
			}
		}
		return operand, nil
	})

	err := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.ByteReader{Reader: input}},
		Filters: []kio.Filter{setAnnotationFn},
		Outputs: []kio.Writer{kio.ByteWriter{Writer: os.Stdout}},
	}.Execute()
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: nginx
	//   labels:
	//     app: nginx
	//   annotations:
	//     foo: 'bar'
	// spec:
	//   replicas: 3
	//   selector:
	//     matchLabels:
	//       app: nginx
	//   template:
	//     metadata:
	//       labels:
	//         app: nginx
	//     spec:
	//       containers:
	//       - name: nginx
	//         image: nginx:1.7.9
	//         ports:
	//         - containerPort: 80
	// ---
	// apiVersion: v1
	// kind: Service
	// metadata:
	//   name: nginx
	//   annotations:
	//     foo: 'bar'
	// spec:
	//   selector:
	//     app: nginx
	//   ports:
	//   - protocol: TCP
	//     port: 80
	//     targetPort: 80
}
