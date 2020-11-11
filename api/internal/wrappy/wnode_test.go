// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package wrappy

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/kustomize/api/resid"

	"gopkg.in/yaml.v3"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	deploymentLittleJson = `{"apiVersion":"apps/v1","kind":"Deployment",` +
		`"metadata":{"name":"homer","namespace":"simpsons"}}`

	deploymentBiggerJson = `
{
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
    "name": "homer",
    "namespace": "simpsons",
    "labels": {
      "fruit": "apple",
      "veggie": "carrot"
    },
    "annotations": {
      "area": "51",
      "greeting": "Take me to your leader."
    }
  }
}
`
	bigMapYaml = `Kind: Service
complextree:
  - field1:
      - boolfield: true
        floatsubfield: 1.01
        intsubfield: 1010
        stringsubfield: idx1010
      - boolfield: false
        floatsubfield: 1.011
        intsubfield: 1011
        stringsubfield: idx1011
    field2:
      - boolfield: true
        floatsubfield: 1.02
        intsubfield: 1020
        stringsubfield: idx1020
      - boolfield: false
        floatsubfield: 1.021
        intsubfield: 1021
        stringsubfield: idx1021
  - field1:
      - boolfield: true
        floatsubfield: 1.11
        intsubfield: 1110
        stringsubfield: idx1110
      - boolfield: false
        floatsubfield: 1.111
        intsubfield: 1111
        stringsubfield: idx1111
    field2:
      - boolfield: true
        floatsubfield: 1.112
        intsubfield: 1120
        stringsubfield: idx1120
      - boolfield: false
        floatsubfield: 1.1121
        intsubfield: 1121
        stringsubfield: idx1121
metadata:
    labels:
        app: application-name
    name: service-name
spec:
    ports:
        port: 80
that:
  - idx0
  - idx1
  - idx2
  - idx3
these:
  - field1:
      - idx010
      - idx011
    field2:
      - idx020
      - idx021
  - field1:
      - idx110
      - idx111
    field2:
      - idx120
      - idx121
  - field1:
      - idx210
      - idx211
    field2:
      - idx220
      - idx221
this:
    is:
        aBool: true
        aFloat: 1.001
        aNilValue: null
        aNumber: 1000
        anEmptyMap: {}
        anEmptySlice: []
those:
  - field1: idx0foo
    field2: idx0bar
  - field1: idx1foo
    field2: idx1bar
  - field1: idx2foo
    field2: idx2bar
`
)

func makeBigMap() map[string]interface{} {
	return map[string]interface{}{
		"Kind": "Service",
		"metadata": map[string]interface{}{
			"labels": map[string]interface{}{
				"app": "application-name",
			},
			"name": "service-name",
		},
		"spec": map[string]interface{}{
			"ports": map[string]interface{}{
				"port": int64(80),
			},
		},
		"this": map[string]interface{}{
			"is": map[string]interface{}{
				"aNumber":      int64(1000),
				"aFloat":       float64(1.001),
				"aNilValue":    nil,
				"aBool":        true,
				"anEmptyMap":   map[string]interface{}{},
				"anEmptySlice": []interface{}{},
				/*
					TODO: test for unrecognizable (e.g. a function)
						"unrecognizable": testing.InternalExample{
							Name: "fooBar",
						},
				*/
			},
		},
		"that": []interface{}{
			"idx0",
			"idx1",
			"idx2",
			"idx3",
		},
		"those": []interface{}{
			map[string]interface{}{
				"field1": "idx0foo",
				"field2": "idx0bar",
			},
			map[string]interface{}{
				"field1": "idx1foo",
				"field2": "idx1bar",
			},
			map[string]interface{}{
				"field1": "idx2foo",
				"field2": "idx2bar",
			},
		},
		"these": []interface{}{
			map[string]interface{}{
				"field1": []interface{}{"idx010", "idx011"},
				"field2": []interface{}{"idx020", "idx021"},
			},
			map[string]interface{}{
				"field1": []interface{}{"idx110", "idx111"},
				"field2": []interface{}{"idx120", "idx121"},
			},
			map[string]interface{}{
				"field1": []interface{}{"idx210", "idx211"},
				"field2": []interface{}{"idx220", "idx221"},
			},
		},
		"complextree": []interface{}{
			map[string]interface{}{
				"field1": []interface{}{
					map[string]interface{}{
						"stringsubfield": "idx1010",
						"intsubfield":    int64(1010),
						"floatsubfield":  float64(1.010),
						"boolfield":      true,
					},
					map[string]interface{}{
						"stringsubfield": "idx1011",
						"intsubfield":    int64(1011),
						"floatsubfield":  float64(1.011),
						"boolfield":      false,
					},
				},
				"field2": []interface{}{
					map[string]interface{}{
						"stringsubfield": "idx1020",
						"intsubfield":    int64(1020),
						"floatsubfield":  float64(1.020),
						"boolfield":      true,
					},
					map[string]interface{}{
						"stringsubfield": "idx1021",
						"intsubfield":    int64(1021),
						"floatsubfield":  float64(1.021),
						"boolfield":      false,
					},
				},
			},
			map[string]interface{}{
				"field1": []interface{}{
					map[string]interface{}{
						"stringsubfield": "idx1110",
						"intsubfield":    int64(1110),
						"floatsubfield":  float64(1.110),
						"boolfield":      true,
					},
					map[string]interface{}{
						"stringsubfield": "idx1111",
						"intsubfield":    int64(1111),
						"floatsubfield":  float64(1.111),
						"boolfield":      false,
					},
				},
				"field2": []interface{}{
					map[string]interface{}{
						"stringsubfield": "idx1120",
						"intsubfield":    int64(1120),
						"floatsubfield":  float64(1.1120),
						"boolfield":      true,
					},
					map[string]interface{}{
						"stringsubfield": "idx1121",
						"intsubfield":    int64(1121),
						"floatsubfield":  float64(1.1121),
						"boolfield":      false,
					},
				},
			},
		},
	}
}

func TestBasicYamlOperationFromMap(t *testing.T) {
	bytes, err := yaml.Marshal(makeBigMap())
	if err != nil {
		t.Fatalf("unexpected yaml.Marshal err: %v", err)
	}
	if string(bytes) != bigMapYaml {
		t.Fatalf("unexpected string equality")
	}
	rNode, err := kyaml.Parse(string(bytes))
	if err != nil {
		t.Fatalf("unexpected yaml.Marshal err: %v", err)
	}
	rNodeString := rNode.MustString()
	// The result from MustString has more indentation
	// than bigMapYaml.
	rNodeStrings := strings.Split(rNodeString, "\n")
	bigMapStrings := strings.Split(bigMapYaml, "\n")
	if len(rNodeStrings) != len(bigMapStrings) {
		t.Fatalf("line count mismatch")
	}
	for i := range rNodeStrings {
		s1 := strings.TrimSpace(rNodeStrings[i])
		s2 := strings.TrimSpace(bigMapStrings[i])
		if s1 != s2 {
			t.Fatalf("expected '%s'=='%s'", s1, s2)
		}
	}
}

func TestRoundTripJSON(t *testing.T) {
	wn := NewWNode()
	err := wn.UnmarshalJSON([]byte(deploymentLittleJson))
	if err != nil {
		t.Fatalf("unexpected UnmarshalJSON err: %v", err)
	}
	data, err := wn.MarshalJSON()
	if err != nil {
		t.Fatalf("unexpected MarshalJSON err: %v", err)
	}
	actual := string(data)
	if actual != deploymentLittleJson {
		t.Fatalf("expected %s, got %s", deploymentLittleJson, actual)
	}
}

func TestGettingFields(t *testing.T) {
	wn := NewWNode()
	err := wn.UnmarshalJSON([]byte(deploymentBiggerJson))
	if err != nil {
		t.Fatalf("unexpected unmarshaljson err: %v", err)
	}
	gvk := wn.GetGvk()
	expected := "apps"
	actual := gvk.Group
	if expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
	expected = "v1"
	actual = gvk.Version
	if expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
	expected = "Deployment"
	actual = gvk.Kind
	if expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
	actual = wn.GetKind()
	if expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
	expected = "homer"
	actual = wn.GetName()
	if expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
	actualMap := wn.GetLabels()
	v, ok := actualMap["fruit"]
	if !ok || v != "apple" {
		t.Fatalf("unexpected labels '%v'", actualMap)
	}
	actualMap = wn.GetAnnotations()
	v, ok = actualMap["greeting"]
	if !ok || v != "Take me to your leader." {
		t.Fatalf("unexpected annotations '%v'", actualMap)
	}
}

func TestGetFieldValueReturnsMap(t *testing.T) {
	wn := NewWNode()
	if err := wn.UnmarshalJSON([]byte(deploymentBiggerJson)); err != nil {
		t.Fatalf("unexpected unmarshaljson err: %v", err)
	}
	expected := map[string]interface{}{
		"fruit":  "apple",
		"veggie": "carrot",
	}
	actual, err := wn.GetFieldValue("metadata.labels")
	if err != nil {
		t.Fatalf("error getting field value: %v", err)
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatalf("actual map does not deep equal expected map:\n%v", diff)
	}
}

func TestGetFieldValueReturnsSlice(t *testing.T) {
	bytes, err := yaml.Marshal(makeBigMap())
	if err != nil {
		t.Fatalf("unexpected yaml.Marshal err: %v", err)
	}
	rNode, err := kyaml.Parse(string(bytes))
	if err != nil {
		t.Fatalf("unexpected yaml.Marshal err: %v", err)
	}
	wn := FromRNode(rNode)
	expected := []interface{}{"idx0", "idx1", "idx2", "idx3"}
	actual, err := wn.GetFieldValue("that")
	if err != nil {
		t.Fatalf("error getting slice: %v", err)
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatalf("actual slice does not deep equal expected slice:\n%v", diff)
	}
}

func TestGetFieldValueReturnsString(t *testing.T) {
	wn := NewWNode()
	if err := wn.UnmarshalJSON([]byte(deploymentBiggerJson)); err != nil {
		t.Fatalf("unexpected unmarshaljson err: %v", err)
	}
	actual, err := wn.GetFieldValue("metadata.labels.fruit")
	if err != nil {
		t.Fatalf("error getting field value: %v", err)
	}
	v, ok := actual.(string)
	if !ok || v != "apple" {
		t.Fatalf("unexpected value '%v'", actual)
	}
}

func TestGetFieldValueResolvesAlias(t *testing.T) {
	yamlWithAlias := `
foo: &a theValue
bar: *a
`
	rNode, err := kyaml.Parse(yamlWithAlias)
	if err != nil {
		t.Fatalf("unexpected yaml parse error: %v", err)
	}
	wn := FromRNode(rNode)
	actual, err := wn.GetFieldValue("bar")
	if err != nil {
		t.Fatalf("error getting field value: %v", err)
	}
	v, ok := actual.(string)
	if !ok || v != "theValue" {
		t.Fatalf("unexpected value '%v'", actual)
	}
}

func TestGetString(t *testing.T) {
	wn := NewWNode()
	if err := wn.UnmarshalJSON([]byte(deploymentBiggerJson)); err != nil {
		t.Fatalf("unexpected unmarshaljson err: %v", err)
	}
	expected := "carrot"
	actual, err := wn.GetString("metadata.labels.veggie")
	if err != nil {
		t.Fatalf("error getting string: %v", err)
	}
	if expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
}

func TestGetSlice(t *testing.T) {
	bytes, err := yaml.Marshal(makeBigMap())
	if err != nil {
		t.Fatalf("unexpected yaml.Marshal err: %v", err)
	}
	rNode, err := kyaml.Parse(string(bytes))
	if err != nil {
		t.Fatalf("unexpected yaml.Marshal err: %v", err)
	}
	wn := FromRNode(rNode)
	expected := []interface{}{"idx0", "idx1", "idx2", "idx3"}
	actual, err := wn.GetSlice("that")
	if err != nil {
		t.Fatalf("error getting slice: %v", err)
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatalf("actual slice does not deep equal expected slice:\n%v", diff)
	}
}

func TestMap(t *testing.T) {
	wn := NewWNode()
	if err := wn.UnmarshalJSON([]byte(deploymentLittleJson)); err != nil {
		t.Fatalf("unexpected unmarshaljson err: %v", err)
	}

	expected := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"name":      "homer",
			"namespace": "simpsons",
		},
	}

	actual := wn.Map()
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatalf("actual map does not deep equal expected map:\n%v", diff)
	}
}

func TestSetName(t *testing.T) {
	wn := NewWNode()
	if err := wn.UnmarshalJSON([]byte(deploymentBiggerJson)); err != nil {
		t.Fatalf("unexpected unmarshaljson err: %v", err)
	}
	wn.SetName("marge")
	if expected, actual := "marge", wn.GetName(); expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
}

func TestSetNamespace(t *testing.T) {
	wn := NewWNode()
	if err := wn.UnmarshalJSON([]byte(deploymentBiggerJson)); err != nil {
		t.Fatalf("unexpected unmarshaljson err: %v", err)
	}
	wn.SetNamespace("flanders")
	meta, _ := wn.node.GetMeta()
	if expected, actual := "flanders", meta.Namespace; expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
}

func TestSetLabels(t *testing.T) {
	wn := NewWNode()
	if err := wn.UnmarshalJSON([]byte(deploymentBiggerJson)); err != nil {
		t.Fatalf("unexpected unmarshaljson err: %v", err)
	}
	wn.SetLabels(map[string]string{
		"label1": "foo",
		"label2": "bar",
	})
	labels := wn.GetLabels()
	if expected, actual := 2, len(labels); expected != actual {
		t.Fatalf("expected '%d', got '%d'", expected, actual)
	}
	if expected, actual := "foo", labels["label1"]; expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
	if expected, actual := "bar", labels["label2"]; expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
}

func TestGetAnnotations(t *testing.T) {
	wn := NewWNode()
	if err := wn.UnmarshalJSON([]byte(deploymentBiggerJson)); err != nil {
		t.Fatalf("unexpected unmarshaljson err: %v", err)
	}
	wn.SetAnnotations(map[string]string{
		"annotation1": "foo",
		"annotation2": "bar",
	})
	annotations := wn.GetAnnotations()
	if expected, actual := 2, len(annotations); expected != actual {
		t.Fatalf("expected '%d', got '%d'", expected, actual)
	}
	if expected, actual := "foo", annotations["annotation1"]; expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
	if expected, actual := "bar", annotations["annotation2"]; expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
}

func TestSetGvk(t *testing.T) {
	wn := NewWNode()
	if err := wn.UnmarshalJSON([]byte(deploymentBiggerJson)); err != nil {
		t.Fatalf("unexpected unmarshaljson err: %v", err)
	}
	wn.SetGvk(resid.GvkFromString("grp_ver_knd"))
	gvk := wn.GetGvk()
	if expected, actual := "grp", gvk.Group; expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
	if expected, actual := "ver", gvk.Version; expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
	if expected, actual := "knd", gvk.Kind; expected != actual {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
}
