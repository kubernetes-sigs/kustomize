// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package imagetag

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	filtertest "sigs.k8s.io/kustomize/api/testutils/filtertest"
	"sigs.k8s.io/kustomize/api/types"
)

func TestImageTagUpdater_Filter(t *testing.T) {
	testCases := map[string]struct {
		input          string
		expectedOutput string
		filter         Filter
		fsSlice        types.FsSlice
	}{
		"ignore CustomResourceDefinition": {
			input: `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: whatever
spec:
  containers:
  - image: whatever
`,
			expectedOutput: `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: whatever
spec:
  containers:
  - image: whatever
`,
			filter: Filter{
				ImageTag: types.Image{
					Name:    "whatever",
					NewName: "theImageShouldNotChangeInACrd",
				},
			},
			fsSlice: []types.FieldSpec{
				{
					Path: "spec/containers/image",
				},
			},
		},

		"legacy multiple images in containers": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
spec:
  containers:
  - image: nginx:1.2.1
  - image: nginx:2.1.2
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
spec:
  containers:
  - image: apache@12345
  - image: apache@12345
`,
			filter: Filter{
				ImageTag: types.Image{
					Name:    "nginx",
					NewName: "apache",
					Digest:  "12345",
				},
			},
			fsSlice: []types.FieldSpec{
				{
					Path: "spec/containers/image",
				},
			},
		},
		"legacy both containers and initContainers": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
spec:
  containers:
  - image: nginx:1.2.1
  - image: tomcat:1.2.3
  initContainers:
  - image: nginx:1.2.1
  - image: apache:1.2.3
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
spec:
  containers:
  - image: apache:3.2.1
  - image: tomcat:1.2.3
  initContainers:
  - image: apache:3.2.1
  - image: apache:1.2.3
`,
			filter: Filter{
				ImageTag: types.Image{
					Name:    "nginx",
					NewName: "apache",
					NewTag:  "3.2.1",
				},
			},
			fsSlice: []types.FieldSpec{
				{
					Path: "spec/containers/image",
				},
				{
					Path: "spec/initContainers/image",
				},
			},
		},
		"legacy updates at multiple depths": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
spec:
  containers:
  - image: nginx:1.2.1
  - image: tomcat:1.2.3
  template:
    spec:
      initContainers:
      - image: nginx:1.2.1
      - image: apache:1.2.3
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
spec:
  containers:
  - image: apache:3.2.1
  - image: tomcat:1.2.3
  template:
    spec:
      initContainers:
      - image: apache:3.2.1
      - image: apache:1.2.3
`,
			filter: Filter{
				ImageTag: types.Image{
					Name:    "nginx",
					NewName: "apache",
					NewTag:  "3.2.1",
				},
			},
			fsSlice: []types.FieldSpec{
				{
					Path: "spec/containers/image",
				},
				{
					Path: "spec/template/spec/initContainers/image",
				},
			},
		},
		"update with digest": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
spec:
  image: nginx:1.2.1
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
spec:
  image: apache@12345
`,
			filter: Filter{
				ImageTag: types.Image{
					Name:    "nginx",
					NewName: "apache",
					Digest:  "12345",
				},
			},
			fsSlice: []types.FieldSpec{
				{
					Path: "spec/image",
				},
			},
		},

		"multiple matches in sequence": {
			input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
spec:
  containers:
  - image: nginx:1.2.1
  - image: not_nginx@54321
  - image: nginx:1.2.1
`,
			expectedOutput: `
apiVersion: example.com/v1
kind: Foo
metadata:
  name: instance
spec:
  containers:
  - image: apache:3.2.1
  - image: not_nginx@54321
  - image: apache:3.2.1
`,
			filter: Filter{
				ImageTag: types.Image{
					Name:    "nginx",
					NewName: "apache",
					NewTag:  "3.2.1",
				},
			},
			fsSlice: []types.FieldSpec{
				{
					Path: "spec/containers/image",
				},
			},
		},

		"new Tag": {
			input: `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:latest
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: nginx
        name: nginx-notag
      - image: nginx@sha256:111111111111111111
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine`,
			expectedOutput: `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:v2
        name: nginx-tagged
      - image: nginx:v2
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: nginx:v2
        name: nginx-notag
      - image: nginx:v2
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`,
			filter: Filter{
				ImageTag: types.Image{
					Name:   "nginx",
					NewTag: "v2",
				},
			},
			fsSlice: []types.FieldSpec{
				{
					Path: "spec/template/spec/containers[]/image",
				},
				{
					Path: "spec/template/spec/initContainers[]/image",
				},
			},
		},
		"newImage": {
			input: `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:latest
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: nginx
        name: nginx-notag
      - image: nginx@sha256:111111111111111111
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`,
			expectedOutput: `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: busybox:1.7.9
        name: nginx-tagged
      - image: busybox:latest
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: busybox
        name: nginx-notag
      - image: busybox@sha256:111111111111111111
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`,
			filter: Filter{
				ImageTag: types.Image{
					Name:    "nginx",
					NewName: "busybox",
				},
			},
			fsSlice: []types.FieldSpec{
				{
					Path: "spec/template/spec/containers[]/image",
				},
				{
					Path: "spec/template/spec/initContainers[]/image",
				},
			},
		},
		"newImageAndTag": {
			input: `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:latest
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: nginx
        name: nginx-notag
      - image: nginx@sha256:111111111111111111
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`,
			expectedOutput: `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: busybox:v3
        name: nginx-tagged
      - image: busybox:v3
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: busybox:v3
        name: nginx-notag
      - image: busybox:v3
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`,
			filter: Filter{
				ImageTag: types.Image{
					Name:    "nginx",
					NewName: "busybox",
					NewTag:  "v3",
				},
			},
			fsSlice: []types.FieldSpec{
				{
					Path: "spec/template/spec/containers[]/image",
				},
				{
					Path: "spec/template/spec/initContainers[]/image",
				},
			},
		},
		"newDigest": {
			input: `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:latest
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: nginx
        name: nginx-notag
      - image: nginx@sha256:111111111111111111
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`,
			expectedOutput: `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx@sha256:222222222222222222
        name: nginx-tagged
      - image: nginx@sha256:222222222222222222
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: nginx@sha256:222222222222222222
        name: nginx-notag
      - image: nginx@sha256:222222222222222222
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`,
			filter: Filter{
				ImageTag: types.Image{
					Name:   "nginx",
					Digest: "sha256:222222222222222222",
				},
			},
			fsSlice: []types.FieldSpec{
				{
					Path: "spec/template/spec/containers/image",
				},
				{
					Path: "spec/template/spec/initContainers/image",
				},
			},
		},
		"newImageAndDigest": {
			input: `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:latest
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: nginx
        name: nginx-notag
      - image: nginx@sha256:111111111111111111
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`,
			expectedOutput: `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: busybox@sha256:222222222222222222
        name: nginx-tagged
      - image: busybox@sha256:222222222222222222
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: busybox@sha256:222222222222222222
        name: nginx-notag
      - image: busybox@sha256:222222222222222222
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`,
			filter: Filter{
				ImageTag: types.Image{
					Name:    "nginx",
					NewName: "busybox",
					Digest:  "sha256:222222222222222222",
				},
			},
			fsSlice: []types.FieldSpec{
				{
					Path: "spec/template/spec/containers[]/image",
				},
				{
					Path: "spec/template/spec/initContainers[]/image",
				},
			},
		},
		"emptyContainers": {
			input: `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  containers:
`,
			expectedOutput: `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  containers: []
`,
			filter: Filter{
				ImageTag: types.Image{
					Name:   "nginx",
					NewTag: "v2",
				},
			},
			fsSlice: []types.FieldSpec{
				{
					Path: "spec/containers[]/image",
					//					CreateIfNotPresent: true,
				},
			},
		},
		"tagWithBraces": {
			input: `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: some.registry.io/my-image:{GENERATED_TAG}
        name: my-image
`,
			expectedOutput: `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: some.registry.io/my-image:my-fixed-tag
        name: my-image
`,
			filter: Filter{
				ImageTag: types.Image{
					Name:   "some.registry.io/my-image",
					NewTag: "my-fixed-tag",
				},
			},
			fsSlice: []types.FieldSpec{
				{
					Path: "spec/template/spec/containers[]/image",
				},
				{
					Path: "spec/template/spec/initContainers[]/image",
				},
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			filter := tc.filter
			filter.FsSlice = tc.fsSlice
			if !assert.Equal(t,
				strings.TrimSpace(tc.expectedOutput),
				strings.TrimSpace(filtertest.RunFilter(t, tc.input, filter))) {
				t.FailNow()
			}
		})
	}
}
