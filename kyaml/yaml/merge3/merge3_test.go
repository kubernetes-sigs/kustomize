// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge3_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	. "sigs.k8s.io/kustomize/kyaml/yaml/merge3"
)

var testCases = [][]testCase{scalarTestCases, listTestCases, mapTestCases, elementTestCases, kustomizationTestCases}

// refTestDefinitions provides the definitions the explicit-schema test cases
// reference through $ref field comments. The builtin Kubernetes schema
// document is no longer embedded, so tests exercising the $ref mechanism
// supply the definitions themselves.
const refTestDefinitions = `{
  "definitions": {
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
        "command": {"type": "array", "items": {"type": "string"}}
      }
    }
  }
}`

func TestMerge(t *testing.T) {
	if err := openapi.AddSchema([]byte(refTestDefinitions)); err != nil {
		t.Fatal(err)
	}
	defer openapi.ResetOpenAPI()
	for i := range testCases {
		for j := range testCases[i] {
			tc := testCases[i][j]
			t.Run(tc.description, func(t *testing.T) {
				actual, err := MergeStrings(tc.local, tc.origin, tc.update, tc.infer)
				if tc.err == nil {
					if !assert.NoError(t, err, tc.description) {
						t.FailNow()
					}
					if !assert.Equal(t,
						strings.TrimSpace(tc.expected), strings.TrimSpace(actual), tc.description) {
						t.FailNow()
					}
				} else {
					if !assert.Errorf(t, err, tc.description) {
						t.FailNow()
					}
					if !assert.Contains(t, tc.err.Error(), err.Error()) {
						t.FailNow()
					}
				}
			})
		}
	}
}

type testCase struct {
	description string
	origin      string
	update      string
	local       string
	expected    string
	err         error
	infer       bool
}
