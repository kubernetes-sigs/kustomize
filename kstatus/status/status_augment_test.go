// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package status

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	corev1 "sigs.k8s.io/kustomize/pseudo/k8s/api/core/v1"
)

var pod = `
apiVersion: v1
kind: Pod
metadata:
   generation: 1
   name: test
   namespace: qual
status:
   phase: Running
`

var custom = `
apiVersion: v1beta1
kind: SomeCustomKind
metadata:
   generation: 1
   name: test
   namespace: default
`

var timestamp = time.Now().Add(-1 * time.Minute).UTC().Format(time.RFC3339)

func addConditions(t *testing.T, u *unstructured.Unstructured, conditions []map[string]interface{}) {
	conds := make([]interface{}, 0)
	for _, c := range conditions {
		conds = append(conds, c)
	}
	err := unstructured.SetNestedSlice(u.Object, conds, "status", "conditions")
	if err != nil {
		t.Fatal(err)
	}
}

func TestAugmentConditions(t *testing.T) {
	testCases := map[string]struct {
		manifest           string
		withConditions     []map[string]interface{}
		expectedConditions []Condition
	}{
		"no existing conditions": {
			manifest:       pod,
			withConditions: []map[string]interface{}{},
			expectedConditions: []Condition{
				{
					Type:   ConditionInProgress,
					Status: corev1.ConditionTrue,
					Reason: "PodNotReady",
				},
			},
		},
		"has other existing conditions": {
			manifest: pod,
			withConditions: []map[string]interface{}{
				{
					"lastTransitionTime": timestamp,
					"lastUpdateTime":     timestamp,
					"type":               "Ready",
					"status":             "False",
					"reason":             "Pod has not started",
				},
			},
			expectedConditions: []Condition{
				{
					Type:   ConditionInProgress,
					Status: corev1.ConditionTrue,
					Reason: "PodNotReady",
				},
				{
					Type:   "Ready",
					Status: corev1.ConditionFalse,
					Reason: "Pod has not started",
				},
			},
		},
		"already has condition of standard type InProgress": {
			manifest: pod,
			withConditions: []map[string]interface{}{
				{
					"lastTransitionTime": timestamp,
					"lastUpdateTime":     timestamp,
					"type":               ConditionInProgress.String(),
					"status":             "True",
					"reason":             "PodIsAbsolutelyNotReady",
				},
			},
			expectedConditions: []Condition{
				{
					Type:   ConditionInProgress,
					Status: corev1.ConditionTrue,
					Reason: "PodIsAbsolutelyNotReady",
				},
			},
		},
		"already has condition of standard type Failed": {
			manifest: pod,
			withConditions: []map[string]interface{}{
				{
					"lastTransitionTime": timestamp,
					"lastUpdateTime":     timestamp,
					"type":               ConditionFailed.String(),
					"status":             "True",
					"reason":             "PodHasFailed",
				},
			},
			expectedConditions: []Condition{
				{
					Type:   ConditionFailed,
					Status: corev1.ConditionTrue,
					Reason: "PodHasFailed",
				},
			},
		},
		"custom resource with no conditions": {
			manifest:           custom,
			withConditions:     []map[string]interface{}{},
			expectedConditions: []Condition{},
		},
	}

	for tn, tc := range testCases {
		tc := tc
		t.Run(tn, func(t *testing.T) {
			u := y2u(t, tc.manifest)
			addConditions(t, u, tc.withConditions)

			err := Augment(u)
			if err != nil {
				t.Error(err)
			}

			obj, err := GetObjectWithConditions(u.Object)
			if err != nil {
				t.Error(err)
			}

			assert.Equal(t, len(tc.expectedConditions), len(obj.Status.Conditions))

			for _, expectedCondition := range tc.expectedConditions {
				found := false
				for _, condition := range obj.Status.Conditions {
					if expectedCondition.Type.String() != condition.Type {
						continue
					}
					found = true
					assert.Equal(t, expectedCondition.Type.String(), condition.Type)
					assert.Equal(t, expectedCondition.Reason, condition.Reason)
				}
				assert.True(t, found)
			}
		})
	}
}
