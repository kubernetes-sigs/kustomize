// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "sigs.k8s.io/yaml/goyaml.v3"
)

func TestPathMatcher_Filter(t *testing.T) {
	node := MustParse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
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
        args: [-c, conf.yaml]
        ports:
        - containerPort: 80
      - name: sidecar
        image: sidecar:1.0.0
        ports:
        - containerPort: 8081
        - containerPort: 9090
`)

	updates := []struct {
		path  []string
		value string
	}{
		{[]string{
			"spec", "template", "spec", "containers", "[name=.*]"},
			"- name: nginx\n  image: nginx:1.7.9\n  args: [-c, conf.yaml]\n  ports:\n  - containerPort: 80\n" +
				"- name: sidecar\n  image: sidecar:1.0.0\n  ports:\n  - containerPort: 8081\n  - containerPort: 9090\n"},
		{[]string{
			"spec", "template", "spec", "containers", "[name=.*]", "image"},
			"- nginx:1.7.9\n- sidecar:1.0.0\n"},
		{[]string{
			"spec", "template", "spec", "containers", "[name=n.*]", "image"},
			"- nginx:1.7.9\n"},
		{[]string{
			"spec", "template", "spec", "containers", "[name=s.*]", "image"},
			"- sidecar:1.0.0\n"},
		{[]string{
			"spec", "template", "spec", "containers", "[name=.*x]", "image"},
			"- nginx:1.7.9\n"},
		{[]string{
			"spec", "template", "spec", "containers", "[name=.*]", "ports"},
			"- - containerPort: 80\n- - containerPort: 8081\n  - containerPort: 9090\n"},
		{[]string{
			"spec", "template", "spec", "containers", "[name=.*]", "ports", "[containerPort=8.*]"},
			"- containerPort: 80\n- containerPort: 8081\n"},
		{[]string{
			"spec", "template", "spec", "containers", "[name=.*]", "ports", "[containerPort=.*1]"},
			"- containerPort: 8081\n"},
		{[]string{
			"spec", "template", "spec", "containers", "[name=.*]", "ports", "[containerPort=9.*]"},
			"- containerPort: 9090\n"},
		{[]string{
			"spec", "template", "spec", "containers", "[name=s.*]", "ports", "[containerPort=8.*]"},
			"- containerPort: 8081\n"},
		{[]string{
			"spec", "template", "spec", "containers", "[name=s.*]", "ports", "[containerPort=.*2]"},
			""},
		{[]string{
			"spec", "template", "spec", "containers", "*", "image"},
			"- nginx:1.7.9\n- sidecar:1.0.0\n"},
		{[]string{
			"spec", "template", "spec", "containers", "*", "ports", "*"},
			"- containerPort: 80\n- containerPort: 8081\n- containerPort: 9090\n"},
		{[]string{
			"spec", "template", "spec", "containers", "[name=.*]", "args", "1"},
			"- conf.yaml\n"},
	}
	for i, u := range updates {
		result, err := node.Pipe(&PathMatcher{Path: u.path})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, u.value, result.MustString(), fmt.Sprintf("%d", i))
	}
}

func TestPathMatcher_Filter_Create(t *testing.T) {
	testCases := map[string]struct {
		path                    []string
		matches                 []string
		modifiedNodeMustContain string
		create                  yaml.Kind
		expectErr               string
	}{
		"create non-primitive sequence item that does not exist": {
			path: []string{"spec", "template", "spec", "containers", "[name=please-create-me]"},
			matches: []string{
				"name: please-create-me\n",
			},
			modifiedNodeMustContain: "- name: please-create-me",
			create:                  yaml.MappingNode,
		},
		"create non-primitive item in empty sequence by index": {
			path:                    []string{"spec", "template", "spec", "containers", "[name=nginx]", "envFrom", "0"},
			matches:                 []string{"{}\n"},
			modifiedNodeMustContain: "envFrom:\n        - {}\n",
			create:                  yaml.MappingNode,
		},
		"create primitive item in empty sequence by index": {
			path:                    []string{"spec", "template", "spec", "containers", "[name=sidecar]", "args", "0"},
			matches:                 []string{"\n"},
			modifiedNodeMustContain: "args:\n        -\n",
			create:                  yaml.ScalarNode,
		},
		"append primitive item to sequence by index": {
			path:                    []string{"spec", "template", "spec", "containers", "[name=nginx]", "args", "2"},
			matches:                 []string{"\n"},
			create:                  yaml.ScalarNode,
			modifiedNodeMustContain: "args: [-c, conf.yaml, '']",
		},
		"append non-primitive item to sequence by index": {
			path:                    []string{"spec", "template", "spec", "containers", "[name=nginx]", "ports", "1"},
			matches:                 []string{"{}\n"},
			modifiedNodeMustContain: "ports: [{containerPort: 80}, {}]",
			create:                  yaml.MappingNode,
		},
		"appending non-primitive element in middle of sequence": {
			path:                    []string{"spec", "template", "spec", "containers", "2", "imagePullPolicy"},
			matches:                 []string{"\n"},
			create:                  yaml.ScalarNode,
			modifiedNodeMustContain: "\n      - imagePullPolicy:\n",
		},
		"fail to create non-primitive item by non-zero index in created sequence": {
			path:      []string{"spec", "template", "spec", "containers", "[name=nginx]", "envFrom", "1"},
			matches:   []string{},
			create:    yaml.MappingNode,
			expectErr: "index 1 specified but only 0 elements found",
		},
		"fail to create primitive item by non-zero index in created sequence": {
			path:      []string{"spec", "template", "spec", "containers", "[name=sidecar]", "args", "1"},
			matches:   []string{},
			create:    yaml.ScalarNode,
			expectErr: "index 1 specified but only 0 elements found",
		},
		"fail to create non-primitive item by distant index in existing sequence": {
			path:      []string{"spec", "template", "spec", "containers", "3"},
			matches:   []string{},
			create:    yaml.MappingNode,
			expectErr: "index 3 specified but only 2 elements found",
		},
		"fail to create primitive item by distant index in existing sequence": {
			path:      []string{"spec", "template", "spec", "containers", "[name=nginx]", "args", "3"},
			matches:   []string{},
			create:    yaml.ScalarNode,
			expectErr: "index 3 specified but only 2 elements found",
		},
		"create primitive sequence item that does not exist": {
			path: []string{"metadata", "finalizers", "[=create-me]"},
			matches: []string{
				"create-me\n",
			},
			modifiedNodeMustContain: "finalizers:\n  - create-me\n",
			create:                  yaml.ScalarNode,
		},
		"create series of maps that do not exist": {
			path: []string{"spec", "selector", "matchLabels", "does-not-exist"},
			matches: []string{
				"{}\n",
			},
			modifiedNodeMustContain: "selector:\n    matchLabels:\n      app: nginx\n      does-not-exist: {}\n",
			create:                  yaml.MappingNode,
		},
		"create scalar below series of maps and sequences that do not exist": {
			path: []string{"spec", "template", "spec", "containers", "[name=please-create-me]", "env", "[key=please-create-me]", "value"},
			matches: []string{
				"\n",
			},
			modifiedNodeMustContain: "- name: please-create-me\n        env:\n        - key: please-create-me\n          value:\n",
			create:                  yaml.ScalarNode,
		},
		"find sequence items that already exist": {
			path: []string{"spec", "template", "spec", "containers", "[name=.*]"},
			matches: []string{
				"name: nginx\nimage: nginx:1.7.9\nargs: [-c, conf.yaml]\nports: [{containerPort: 80}]\nenv:\n- key: CONTAINER_NAME\n  value: nginx\n",
				"name: sidecar\nimage: sidecar:1.0.0\nports:\n- containerPort: 8081\n- containerPort: 9090\n",
			},
			create: yaml.MappingNode,
		},
		"find and create sequence below wildcard that exists on some sequence items": {
			path: []string{"spec", "template", "spec", "containers", "[name=.*]", "env"},
			matches: []string{
				"- key: CONTAINER_NAME\n  value: nginx\n",
				"[]\n",
			},
			create: yaml.SequenceNode,
		},
		"find field below wildcard that exists on all sequence items": {
			path: []string{"spec", "template", "spec", "containers", "[name=.*]", "ports"},
			matches: []string{
				"[{containerPort: 80}]\n",
				"- containerPort: 8081\n- containerPort: 9090\n",
			},
			create: yaml.SequenceNode,
		},
		"find field below query that targets a specific item": {
			path: []string{"spec", "template", "spec", "containers", "[name=nginx]", "env"},
			matches: []string{
				"- key: CONTAINER_NAME\n  value: nginx\n",
			},
			create: yaml.SequenceNode,
		},
		"create field below query that targets any value of a field that does not exist": {
			path: []string{"spec", "template", "spec", "containers", "[foo=.*]", "env"},
			matches: []string{
				"[]\n",
			},
			// This is kinda weird. The query doesn't match anything, and we can't tell that it is a
			// wildcard rather than a literal, so we use the value to create the field.
			modifiedNodeMustContain: "- foo: .*\n        env: []\n",
			create:                  yaml.SequenceNode,
		},
	}
	nodeStr := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
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
        args: [-c, conf.yaml]
        ports: [{containerPort: 80}]
        env:
        - key: CONTAINER_NAME
          value: nginx
      - name: sidecar
        image: sidecar:1.0.0
        ports:
        - containerPort: 8081
        - containerPort: 9090
`
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			node := MustParse(nodeStr)
			result, err := node.Pipe(&PathMatcher{Path: tc.path, Create: tc.create})
			if tc.expectErr != "" {
				require.EqualError(t, err, tc.expectErr)
				return
			}
			require.NoError(t, err)
			matches, err := result.Elements()
			require.NoError(t, err)
			require.Equalf(t, len(tc.matches), len(matches), "Full sequence wrapper of result:\n%s", result.MustString())

			modifiedNode := node.MustString()
			for i, expected := range tc.matches {
				assert.Equal(t, tc.create, matches[i].YNode().Kind)
				assert.Equal(t, expected, matches[i].MustString())
				assert.Contains(t, modifiedNode, tc.modifiedNodeMustContain)
			}
		})
	}
}
