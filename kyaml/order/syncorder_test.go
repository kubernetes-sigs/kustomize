// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package order

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestSyncOrder(t *testing.T) {
	testCases := []struct {
		name     string
		from     string
		to       string
		expected string
	}{
		{
			name: "sort data fields configmap with comments",
			from: `apiVersion: v1
kind: ConfigMap
metadata:
  name: setters-config
data:
  # This should be the name of your Config Controller instance
  cluster-name: cluster-name
  # This should be the project where you deployed Config Controller
  project-id: project-id # pro
  project-number: "1234567890123"
  # You can leave these defaults
  namespace: config-control
  deployment-repo: deployment-repo
  source-repo: source-repo
`,
			to: `apiVersion: v1
kind: ConfigMap
metadata: # kpt-merge: /setters-config
  name: setters-config
data:
  # You can leave these defaults
  namespace: config-control
  # This should be the name of your Config Controller instance
  cluster-name: cluster-name
  deployment-repo: deployment-repo
  # This should be the project where you deployed Config Controller
  project-id: project-id # project
  project-number: "1234567890123"
  source-repo: source-repo
`,
			expected: `apiVersion: v1
kind: ConfigMap
metadata: # kpt-merge: /setters-config
  name: setters-config
data:
  # This should be the name of your Config Controller instance
  cluster-name: cluster-name
  # This should be the project where you deployed Config Controller
  project-id: project-id # project
  project-number: "1234567890123"
  # You can leave these defaults
  namespace: config-control
  deployment-repo: deployment-repo
  source-repo: source-repo
`,
		},
		{
			name: "sort data fields configmap but retain order of extra fields",
			from: `apiVersion: v1
kind: ConfigMap
data:
  baz: bar
  cluster-name: cluster-name
  foo: config-control
`,
			to: `kind: ConfigMap
apiVersion: v1
metadata:
  name: foo
data:
  color: orange
  foo: config-control
  abc: def
  cluster-name: cluster-name
`,
			expected: `apiVersion: v1
kind: ConfigMap
data:
  cluster-name: cluster-name
  foo: config-control
  color: orange
  abc: def
metadata:
  name: foo
`,
		},
		{
			name: "sort containers list node with comments",
			from: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: before
spec:
  containers:
  - name: nginx
    # nginx image
    image: "nginx:1.16.1"
    ports:
    - protocol: TCP # tcp protocol
      containerPort: 80
`,
			to: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: after
spec:
  containers:
  - ports:
    - containerPort: 80
      protocol: TCP # tcp protocol
    # nginx image
    image: "nginx:1.16.2"
    # nginx container
    name: nginx
# end of resource
`,
			expected: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: after
spec:
  containers:
  # nginx container
  - name: nginx
    # nginx image
    image: "nginx:1.16.2"
    ports:
    - protocol: TCP # tcp protocol
      containerPort: 80
# end of resource
`,
		},
		{
			name: "Do not alter sequence order",
			from: `apiVersion: v1
kind: KRMFile
metadata:
  name: before
pipeline:
  mutators:
  - image: apply-setters:v0.1
    configPath: setters.yaml
  - image: set-namespace:v0.1
    configPath: ns.yaml
`,
			to: `apiVersion: v1
kind: KRMFile
metadata:
  name: after
pipeline:
  mutators:
  - configPath: sr.yaml
    image: search-replace:v0.1
  - image: apply-setters:v0.1
    configPath: setters.yaml
  - image: set-namespace:v0.1
    configPath: ns.yaml
`,
			expected: `apiVersion: v1
kind: KRMFile
metadata:
  name: after
pipeline:
  mutators:
  - image: search-replace:v0.1
    configPath: sr.yaml
  - image: apply-setters:v0.1
    configPath: setters.yaml
  - image: set-namespace:v0.1
    configPath: ns.yaml
`,
		},
		{
			name: "Complex ASM reorder example",
			from: `apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: (unknown)
  creationTimestamp: null
  name: controlplanerevisions.mesh.cloud.google.com
spec:
  group: mesh.cloud.google.com
  names:
    kind: ControlPlaneRevision
    listKind: ControlPlaneRevisionList
    plural: controlplanerevisions
    singular: controlplanerevision
  scope: Namespaced
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      description: ControlPlaneRevision is the Schema for the ControlPlaneRevision API
      properties:
        apiVersion:
          description: 'APIVersion'
          type: string
        kind:
          description: 'Kind'
          type: string
        metadata:
          type: object
        spec:
          description: ControlPlaneRevisionSpec defines the desired state of ControlPlaneRevision
          properties:
            channel:
              description: ReleaseChannel determines the aggressiveness of upgrades.
              enum:
              - regular
              - rapid
              - stable
              type: string
            type:
              description: ControlPlaneRevisionType determines how the revision should be managed.
              enum:
              - managed_service
              type: string
          type: object
        status:
          description: ControlPlaneRevisionStatus defines the observed state of ControlPlaneRevision.
          properties:
            conditions:
              items:
                description: ControlPlaneRevisionCondition is a repeated struct definining the current conditions of a ControlPlaneRevision.
                properties:
                  lastTransitionTime:
                    description: Last time the condition transitioned from one status to another
                    format: date-time
                    type: string
                  message:
                    description: Human-readable message indicating details about last transition
                    type: string
                  reason:
                    description: Unique, one-word, CamelCase reason for the condition's last transition
                    type: string
                  status:
                    description: Status is the status of the condition. Can be True, False, or Unknown.
                    type: string
                  type:
                    description: Type is the type of the condition.
                    type: string
                type: object
              type: array
          type: object
      type: object
  version: v1alpha1
  versions:
  - name: v1alpha1
    served: true
    storage: true
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`,
			to: `apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: controlplanerevisions.mesh.cloud.google.com
  annotations:
    controller-gen.kubebuilder.io/version: (unknown)
  creationTimestamp: null
spec:
  group: mesh.cloud.google.com
  names:
    kind: ControlPlaneRevision
    listKind: ControlPlaneRevisionList
    plural: controlplanerevisions
    singular: controlplanerevision
  scope: Namespaced
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      type: object
      description: ControlPlaneRevision is the Schema for the ControlPlaneRevision API
      properties:
        apiVersion:
          type: string
          description: 'APIVersion'
        kind:
          type: string
          description: 'Kind'
        metadata:
          type: object
        spec:
          type: object
          description: ControlPlaneRevisionSpec defines the desired state of ControlPlaneRevision
          properties:
            type:
              type: string
              description: ControlPlaneRevisionType determines how the revision should be managed.
              enum:
                - managed_service
            channel:
              type: string
              description: ReleaseChannel determines the aggressiveness of upgrades.
              enum:
                - regular
                - rapid
                - stable
        status:
          type: object
          description: ControlPlaneRevisionStatus defines the observed state of ControlPlaneRevision.
          properties:
            conditions:
              type: array
              items:
                type: object
                description: ControlPlaneRevisionCondition is a repeated struct definining the current conditions of a ControlPlaneRevision.
                properties:
                  type:
                    type: string
                    description: Type is the type of the condition.
                  status:
                    type: string
                    description: Status is the status of the condition. Can be True, False, or Unknown.
                  lastTransitionTime:
                    type: string
                    description: Last time the condition transitioned from one status to another
                    format: date-time
                  message:
                    type: string
                    description: Human-readable message indicating details about last transition
                  reason:
                    type: string
                    description: Unique, one-word, CamelCase reason for the condition's last transition
  version: v1alpha1
  versions:
    - name: v1alpha1
      served: true
      storage: true
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`,
			expected: `test.from`,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			from, err := yaml.Parse(tc.from)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			to, err := yaml.Parse(tc.to)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			err = SyncOrder(from, to)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			out := &bytes.Buffer{}
			kio.ByteWriter{
				Writer:                out,
				KeepReaderAnnotations: false,
			}.Write([]*yaml.RNode{to})

			// this means "to" is just a reordered version of "from" and after syncing order,
			// resultant "to" must be equal to "from"
			if tc.expected == "test.from" {
				tc.expected = tc.from
			}

			if !assert.Equal(t, tc.expected, out.String()) {
				t.FailNow()
			}
		})
	}
}
