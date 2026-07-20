// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestIsYaml1_1NonString(t *testing.T) {
	type testCase struct {
		val      string
		expected bool
	}

	testCases := []testCase{
		{val: "hello world", expected: false},
		{val: "2.0", expected: true},
		{val: "2", expected: true},
		{val: "true", expected: true},
		{val: "1.0\nhello", expected: false}, // multiline strings should always be false
		{val: "", expected: false},           // empty string should be considered a string
	}

	for k := range valueToTagMap {
		testCases = append(testCases, testCase{val: k, expected: true})
	}

	for _, test := range testCases {
		assert.Equal(t, test.expected,
			yaml.IsYaml1_1NonString(&yaml.Node{Kind: yaml.ScalarNode, Value: test.val}), test.val)
	}
}

// formatTestDefinitions supplies the schema type information this test
// relies on. The builtin Kubernetes schema document is no longer embedded,
// so schema-driven formatting of builtin types requires a supplied schema.
const formatTestDefinitions = `{
  "definitions": {
    "io.k8s.api.apps.v1.Deployment": {
      "x-kubernetes-group-version-kind": [{"group": "apps", "kind": "Deployment", "version": "v1"}],
      "properties": {
        "spec": {"$ref": "#/definitions/io.k8s.api.apps.v1.DeploymentSpec"}
      }
    },
    "io.k8s.api.apps.v1.DeploymentSpec": {
      "properties": {
        "template": {"$ref": "#/definitions/io.k8s.api.core.v1.PodTemplateSpec"}
      }
    },
    "io.k8s.api.core.v1.PodTemplateSpec": {
      "properties": {
        "spec": {"$ref": "#/definitions/io.k8s.api.core.v1.PodSpec"}
      }
    },
    "io.k8s.api.core.v1.PodSpec": {
      "properties": {
        "containers": {
          "type": "array",
          "items": {"$ref": "#/definitions/io.k8s.api.core.v1.Container"},
          "x-kubernetes-patch-merge-key": "name",
          "x-kubernetes-patch-strategy": "merge"
        }
      }
    },
    "io.k8s.api.core.v1.Container": {
      "properties": {
        "name": {"type": "string"},
        "image": {"type": "string"},
        "args": {"type": "array", "items": {"type": "string"}},
        "ports": {
          "type": "array",
          "items": {"$ref": "#/definitions/io.k8s.api.core.v1.ContainerPort"},
          "x-kubernetes-patch-merge-key": "containerPort",
          "x-kubernetes-patch-strategy": "merge"
        }
      }
    },
    "io.k8s.api.core.v1.ContainerPort": {
      "properties": {
        "name": {"type": "string"},
        "containerPort": {"type": "integer"}
      }
    }
  }
}`

func TestFormatNonStringStyle(t *testing.T) {
	if !assert.NoError(t, openapi.AddSchema([]byte(formatTestDefinitions))) {
		t.FailNow()
	}
	defer openapi.ResetOpenAPI()
	n := yaml.MustParse(`apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        args:
        - bar
        - on
        image: nginx:1.7.9
        ports:
        - name: http
          containerPort: "80"
`)
	s := openapi.SchemaForResourceType(
		yaml.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"})

	args, err := n.Pipe(yaml.Lookup(
		"spec", "template", "spec", "containers", "[name=foo]", "args"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.NotNil(t, args) {
		t.FailNow()
	}
	on := args.YNode().Content[1]
	onS := s.Lookup(
		"spec", "template", "spec", "containers", openapi.Elements, "args", openapi.Elements)
	yaml.FormatNonStringStyle(on, *onS.Schema)

	containerPort, err := n.Pipe(yaml.Lookup(
		"spec", "template", "spec", "containers", "[name=foo]", "ports",
		"[name=http]", "containerPort"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.NotNil(t, containerPort) {
		t.FailNow()
	}
	cpS := s.Lookup("spec", "template", "spec", "containers", openapi.Elements,
		"ports", openapi.Elements, "containerPort")
	if !assert.NotNil(t, cpS) {
		t.FailNow()
	}
	yaml.FormatNonStringStyle(containerPort.YNode(), *cpS.Schema)

	actual := n.MustString()
	expected := `apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: foo
        args:
        - bar
        - "on"
        image: nginx:1.7.9
        ports:
        - name: http
          containerPort: 80
`
	assert.Equal(t, expected, actual)
}

// valueToTagMap is a map of values interpreted as non-strings in yaml 1.1 when left
// unquoted.
// To keep compatibility with the yaml parser used by Kubernetes (yaml 1.1) make sure the values
// which are treated as non-string values are kept as non-string values.
// https://github.com/go-yaml/yaml/blob/v2/resolve.go
var valueToTagMap = func() map[string]string {
	val := map[string]string{}

	// https://yaml.org/type/null.html
	values := []string{"~", "null", "Null", "NULL"}
	for i := range values {
		val[values[i]] = yaml.NodeTagNull
	}

	// https://yaml.org/type/bool.html
	values = []string{
		"y", "Y", "yes", "Yes", "YES", "true", "True", "TRUE", "on", "On", "ON", "n", "N", "no",
		"No", "NO", "false", "False", "FALSE", "off", "Off", "OFF"}
	for i := range values {
		val[values[i]] = yaml.NodeTagBool
	}

	// https://yaml.org/type/float.html
	values = []string{
		".nan", ".NaN", ".NAN", ".inf", ".Inf", ".INF",
		"+.inf", "+.Inf", "+.INF", "-.inf", "-.Inf", "-.INF"}
	for i := range values {
		val[values[i]] = yaml.NodeTagFloat
	}

	return val
}()
