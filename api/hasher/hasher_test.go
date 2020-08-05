// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package hasher

import (
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestSortArrayAndComputeHash(t *testing.T) {
	array1 := []string{"a", "b", "c", "d"}
	array2 := []string{"c", "b", "d", "a"}
	h1, err := SortArrayAndComputeHash(array1)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if h1 == "" {
		t.Errorf("failed to hash %v", array1)
	}
	h2, err := SortArrayAndComputeHash(array2)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if h2 == "" {
		t.Errorf("failed to hash %v", array2)
	}
	if h1 != h2 {
		t.Errorf("hash is not consistent with reordered list: %s %s", h1, h2)
	}
}

func TestHash(t *testing.T) {
	// hash the empty string to be sure that sha256 is being used
	expect := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	sum := Hash("")
	if expect != sum {
		t.Errorf("expected hash %q but got %q", expect, sum)
	}
}

func TestConfigMapHash(t *testing.T) {
	cases := []struct {
		desc   string
		cmYaml string
		hash   string
		err    string
	}{
		// empty map
		{"empty data", `
apiVersion: v1
kind: ConfigMap`, "6ct58987ht", ""},
		// one key
		{"one key", `
apiVersion: v1
kind: ConfigMap
data: 
  one: ""`, "9g67k2htb6", ""},
		// three keys (tests sorting order)
		{"three keys", `
apiVersion: v1
kind: ConfigMap
data:
  two: 2
  one: ""
  three: 3`, "7757f9kkct", ""},
		// empty binary data map
		{"empty binary data", `
apiVersion: v1
kind: ConfigMap`, "6ct58987ht", ""},
		// one key with binary data
		{"one key with binary data", `
apiVersion: v1
kind: ConfigMap
binaryData:
  one: ""`, "6mtk2m274t", ""},
		// three keys with binary data (tests sorting order)
		{"three keys with binary data", `
apiVersion: v1
kind: ConfigMap
binaryData:
  two: 2
  one: ""
  three: 3`, "9th7kc28dg", ""},
		// two keys, one with string and another with binary data
		{"two keys with one each", `
apiVersion: v1
kind: ConfigMap
data:
  one: ""
binaryData:
  two: ""`, "698h7c7t9m", ""},
	}

	for _, c := range cases {
		node, err := yaml.Parse(c.cmYaml)
		if err != nil {
			t.Fatal(err)
		}
		h, err := HashRNode(node)
		if SkipRest(t, c.desc, err, c.err) {
			continue
		}
		if c.hash != h {
			t.Errorf("case %q, expect hash %q but got %q", c.desc, c.hash, h)
		}
	}
}

func TestSecretHash(t *testing.T) {
	cases := []struct {
		desc       string
		secretYaml string
		hash       string
		err        string
	}{
		// empty map
		{"empty data", `
apiVersion: v1
kind: Secret
type: my-type`, "5gmgkf8578", ""},
		// one key
		{"one key", `
apiVersion: v1
kind: Secret
type: my-type
data:
  one: ""`, "74bd68bm66", ""},
		// three keys (tests sorting order)
		{"three keys", `
apiVersion: v1
kind: Secret
type: my-type
data:
  two: 2
  one: ""
  three: 3`, "4gf75c7476", ""},
		// with stringdata
		{"stringdata", `
apiVersion: v1
kind: Secret
type: my-type
data:
  one: ""
stringData:
  two: 2`, "c4h4264gdb", ""},
		// empty stringdata
		{"empty stringdata", `
apiVersion: v1
kind: Secret
type: my-type
data:
  one: ""`, "74bd68bm66", ""},
	}

	for _, c := range cases {
		node, err := yaml.Parse(c.secretYaml)
		if err != nil {
			t.Fatal(err)
		}
		h, err := HashRNode(node)
		if SkipRest(t, c.desc, err, c.err) {
			continue
		}
		if c.hash != h {
			t.Errorf("case %q, expect hash %q but got %q", c.desc, c.hash, h)
		}
	}
}

func TestUnstructuredHash(t *testing.T) {
	cases := []struct {
		desc         string
		unstructured string
		hash         string
		err          string
	}{
		{"minimal", `
apiVersion: test/v1
kind: TestResource
metadata:
  name: my-resource`, "244782mkb7", ""},
		{"with spec", `
apiVersion: test/v1
kind: TestResource
metadata:
  name: my-resource
spec:
  foo: 1
  bar: abc`, "59m2mdccg4", ""},
	}

	for _, c := range cases {
		node, err := yaml.Parse(c.unstructured)
		if err != nil {
			t.Fatal(err)
		}
		h, err := HashRNode(node)
		if SkipRest(t, c.desc, err, c.err) {
			continue
		}
		if c.hash != h {
			t.Errorf("case %q, expect hash %q but got %q", c.desc, c.hash, h)
		}
	}
}

func TestEncodeConfigMap(t *testing.T) {
	cases := []struct {
		desc   string
		cmYaml string
		expect string
		err    string
	}{
		// empty map
		{"empty data", `
apiVersion: v1
kind: ConfigMap`, `{"data":"","kind":"ConfigMap","name":""}`, ""},
		// one key
		{"one key", `
apiVersion: v1
kind: ConfigMap
data: 
  one: ""`, `{"data":{"one":""},"kind":"ConfigMap","name":""}`, ""},
		// three keys (tests sorting order)
		{"three keys", `
apiVersion: v1
kind: ConfigMap
data:
  two: 2
  one: ""
  three: 3`, `{"data":{"one":"","three":3,"two":2},"kind":"ConfigMap","name":""}`, ""},
		// empty binary map
		{"empty data", `
apiVersion: v1
kind: ConfigMap`, `{"data":"","kind":"ConfigMap","name":""}`, ""},
		// one key with binary data
		{"one key", `
apiVersion: v1
kind: ConfigMap
binaryData:
  one: ""`, `{"binaryData":{"one":""},"data":"","kind":"ConfigMap","name":""}`, ""},
		// three keys with binary data (tests sorting order)
		{"three keys", `
apiVersion: v1
kind: ConfigMap
binaryData:
  two: 2
  one: ""
  three: 3`, `{"binaryData":{"one":"","three":3,"two":2},"data":"","kind":"ConfigMap","name":""}`, ""},
		// two keys, one string and one binary values
		{"two keys with one each", `
apiVersion: v1
kind: ConfigMap
data:
  one: ""
binaryData:
  two: ""`, `{"binaryData":{"two":""},"data":{"one":""},"kind":"ConfigMap","name":""}`, ""},
	}
	for _, c := range cases {
		node, err := yaml.Parse(c.cmYaml)
		if err != nil {
			t.Fatal(err)
		}
		s, err := encodeConfigMap(node)
		if SkipRest(t, c.desc, err, c.err) {
			continue
		}
		if s != c.expect {
			t.Errorf("case %q, expect %q but got %q from encode %#v", c.desc, c.expect, s, c.cmYaml)
		}
	}
}

func TestEncodeSecret(t *testing.T) {
	cases := []struct {
		desc       string
		secretYaml string
		expect     string
		err        string
	}{
		// empty map
		{"empty data", `
apiVersion: v1
kind: Secret
type: my-type`, `{"data":"","kind":"Secret","name":"","type":"my-type"}`, ""},
		// one key
		{"one key", `
apiVersion: v1
kind: Secret
type: my-type
data:
  one: ""`, `{"data":{"one":""},"kind":"Secret","name":"","type":"my-type"}`, ""},
		// three keys (tests sorting order) - note json.Marshal base64 encodes the values because they come in as []byte
		{"three keys", `
apiVersion: v1
kind: Secret
type: my-type
data:
  two: 2
  one: ""
  three: 3`, `{"data":{"one":"","three":3,"two":2},"kind":"Secret","name":"","type":"my-type"}`, ""},
		// with stringdata
		{"stringdata", `
apiVersion: v1
kind: Secret
type: my-type
data:
  one: ""
stringData:
  two: 2`, `{"data":{"one":""},"kind":"Secret","name":"","stringData":{"two":2},"type":"my-type"}`, ""},
		// empty stringdata
		{"empty stringdata", `
apiVersion: v1
kind: Secret
type: my-type
data:
  one: ""`, `{"data":{"one":""},"kind":"Secret","name":"","type":"my-type"}`, ""},
	}
	for _, c := range cases {
		node, err := yaml.Parse(c.secretYaml)
		if err != nil {
			t.Fatal(err)
		}
		s, err := encodeSecret(node)
		if SkipRest(t, c.desc, err, c.err) {
			continue
		}
		if s != c.expect {
			t.Errorf("case %q, expect %q but got %q from encode %#v", c.desc, c.expect, s, c.secretYaml)
		}
	}
}

// SkipRest returns true if there was a non-nil error or if we expected an error that didn't happen,
// and logs the appropriate error on the test object.
// The return value indicates whether we should skip the rest of the test case due to the error result.
func SkipRest(t *testing.T, desc string, err error, contains string) bool {
	if err != nil {
		if len(contains) == 0 {
			t.Errorf("case %q, expect nil error but got %q", desc, err.Error())
		} else if !strings.Contains(err.Error(), contains) {
			t.Errorf("case %q, expect error to contain %q but got %q", desc, contains, err.Error())
		}
		return true
	} else if len(contains) > 0 {
		t.Errorf("case %q, expect error to contain %q but got nil error", desc, contains)
		return true
	}
	return false
}
