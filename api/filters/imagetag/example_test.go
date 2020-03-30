// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package imagetag

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
  containers:
  - name: FooBar
    image: nginx
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance
spec:
  containers:
  - name: BarFoo
    image: nginx:1.2.1
`)}},
		Filters: []kio.Filter{Filter{
			ImageTag: types.Image{
				Name:    "nginx",
				NewName: "apache",
				Digest:  "12345",
			},
			FsSlice: []types.FieldSpec{
				{
					Path: "spec/containers[]/image",
				},
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
	//   containers:
	//   - name: FooBar
	//     image: apache@12345
	// ---
	// apiVersion: example.com/v1
	// kind: Bar
	// metadata:
	//   name: instance
	// spec:
	//   containers:
	//   - name: BarFoo
	//     image: apache@12345
}

func ExampleLegacyFilter() {
	err := kio.Pipeline{
		Inputs: []kio.Reader{&kio.ByteReader{Reader: bytes.NewBufferString(`
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
spec:
  containers:
  - name: FooBar
    image: nginx
---
apiVersion: example.com/v1
kind: Bar
metadata:
  name: instance
spec:
  containers:
  - name: BarFoo
    image: nginx:1.2.1
`)}},
		Filters: []kio.Filter{LegacyFilter{
			ImageTag: types.Image{
				Name:    "nginx",
				NewName: "apache",
				Digest:  "12345",
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
	//   containers:
	//   - name: FooBar
	//     image: apache@12345
	// ---
	// apiVersion: example.com/v1
	// kind: Bar
	// metadata:
	//   name: instance
	// spec:
	//   containers:
	//   - name: BarFoo
	//     image: apache@12345
}
