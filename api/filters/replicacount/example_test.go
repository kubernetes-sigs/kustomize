package replicacount

import (
	"bytes"
	"log"
	"os"

	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

func ExampleFilter() {
	err := kio.Pipeline{
		Inputs: []kio.Reader{&kio.ByteReader{Reader: bytes.NewBufferString(`
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
spec:
  template:
    replicas: 5
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance
spec:
  template:
    replicas: 5
`)}},
		Filters: []kio.Filter{Filter{
			Replica: types.Replica{
				Count: 42,
				Name:  "instance",
			},
			FieldSpec: types.FieldSpec{
				Path: "spec/template/replicas",
			},
		}},
		Outputs: []kio.Writer{kio.ByteWriter{Writer: os.Stdout}},
	}.Execute()
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// apiVersion: example.com/v1
	// kind: Foo
	// metadata:
	//   name: instance
	// spec:
	//   template:
	//     replicas: 42
	// ---
	// apiVersion: example.com/v1
	// kind: Bar
	// metadata:
	//   name: instance
	// spec:
	//   template:
	//     replicas: 42
}
