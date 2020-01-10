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
		{val: "1.0\nhello", expected: false}, // multiline strings should always be false
	}

	for k := range valueToTagMap {
		testCases = append(testCases, testCase{val: k, expected: true})
	}

	for _, test := range testCases {
		assert.Equal(t, test.expected,
			yaml.IsYaml1_1NonString(&yaml.Node{Kind: yaml.ScalarNode, Value: test.val}), test.val)
	}
}

func TestFormatNonStringStyle(t *testing.T) {
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
	values := []string{"", "~", "null", "Null", "NULL"}
	for i := range values {
		val[values[i]] = "!!null"
	}

	// https://yaml.org/type/bool.html
	values = []string{
		"y", "Y", "yes", "Yes", "YES", "true", "True", "TRUE", "on", "On", "ON", "n", "N", "no",
		"No", "NO", "false", "False", "FALSE", "off", "Off", "OFF"}
	for i := range values {
		val[values[i]] = "!!bool"
	}

	// https://yaml.org/type/float.html
	values = []string{
		".nan", ".NaN", ".NAN", ".inf", ".Inf", ".INF",
		"+.inf", "+.Inf", "+.INF", "-.inf", "-.Inf", "-.INF"}
	for i := range values {
		val[values[i]] = "!!float"
	}

	return val
}()
