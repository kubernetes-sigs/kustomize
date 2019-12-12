// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package status

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func y2u(t *testing.T, spec string) *unstructured.Unstructured {
	j, err := yaml.YAMLToJSON([]byte(spec))
	assert.NoError(t, err)
	u, _, err := unstructured.UnstructuredJSONScheme.Decode(j, nil, nil)
	assert.NoError(t, err)
	return u.(*unstructured.Unstructured)
}

type testSpec struct {
	spec                 string
	expectedStatus       Status
	expectedConditions   []Condition
	absentConditionTypes []ConditionType
}

func runStatusTest(t *testing.T, tc testSpec) {
	res, err := Compute(y2u(t, tc.spec))
	assert.NoError(t, err)
	assert.Equal(t, tc.expectedStatus, res.Status)

	for _, expectedCondition := range tc.expectedConditions {
		found := false
		for _, condition := range res.Conditions {
			if condition.Type != expectedCondition.Type {
				continue
			}
			found = true
			assert.Equal(t, expectedCondition.Status, condition.Status)
			assert.Equal(t, expectedCondition.Reason, condition.Reason)
		}
		if !found {
			t.Errorf("Expected condition of type %s, but didn't find it", expectedCondition.Type)
		}
	}

	for _, absentConditionType := range tc.absentConditionTypes {
		for _, condition := range res.Conditions {
			if condition.Type == absentConditionType {
				t.Errorf("Expected condition %s to be absent, but found it", absentConditionType)
			}
		}
	}
}

var podNoStatus = `
apiVersion: v1
kind: Pod
metadata:
   generation: 1
   name: test
`

var podReady = `
apiVersion: v1
kind: Pod
metadata:
   generation: 1
   name: test
   namespace: qual
status:
   conditions:
    - type: Ready 
      status: "True"
   phase: Running
`

var podCompletedOK = `
apiVersion: v1
kind: Pod
metadata:
   generation: 1
   name: test
   namespace: qual
status:
   phase: Succeeded
   conditions:
    - type: Ready 
      status: "False"
      reason: PodCompleted

`

var podCompletedFail = `
apiVersion: v1
kind: Pod
metadata:
   generation: 1
   name: test
   namespace: qual
status:
   phase: Failed
   conditions:
    - type: Ready 
      status: "False"
      reason: PodCompleted
`

// Test coverage using GetConditions
func TestPodStatus(t *testing.T) {
	testCases := map[string]testSpec{
		"podNoStatus": {
			spec:           podNoStatus,
			expectedStatus: InProgressStatus,
			expectedConditions: []Condition{{
				Type:   ConditionInProgress,
				Status: corev1.ConditionTrue,
				Reason: "PodNotReady",
			}},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
			},
		},
		"podReady": {
			spec:               podReady,
			expectedStatus:     CurrentStatus,
			expectedConditions: []Condition{},
			absentConditionTypes: []ConditionType{
				ConditionInProgress,
				ConditionFailed,
			},
		},
		"podCompletedSuccessfully": {
			spec:               podCompletedOK,
			expectedStatus:     CurrentStatus,
			expectedConditions: []Condition{},
			absentConditionTypes: []ConditionType{
				ConditionInProgress,
				ConditionFailed,
			},
		},
		"podCompletedFailed": {
			spec:           podCompletedFail,
			expectedStatus: FailedStatus,
			expectedConditions: []Condition{{
				Type:   ConditionFailed,
				Status: corev1.ConditionTrue,
				Reason: "PodFailed",
			}},
			absentConditionTypes: []ConditionType{
				ConditionInProgress,
			},
		},
	}

	for tn, tc := range testCases {
		tc := tc
		t.Run(tn, func(t *testing.T) {
			runStatusTest(t, tc)
		})
	}
}

var pvcNoStatus = `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
   generation: 1
   name: test
`
var pvcBound = `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
   generation: 1
   name: test
   namespace: qual
status:
   phase: Bound
`

func TestPVCStatus(t *testing.T) {
	testCases := map[string]testSpec{
		"pvcNoStatus": {
			spec:           pvcNoStatus,
			expectedStatus: InProgressStatus,
			expectedConditions: []Condition{{
				Type:   ConditionInProgress,
				Status: corev1.ConditionTrue,
				Reason: "NotBound",
			}},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
			},
		},
		"pvcBound": {
			spec:               pvcBound,
			expectedStatus:     CurrentStatus,
			expectedConditions: []Condition{},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
				ConditionInProgress,
			},
		},
	}

	for tn, tc := range testCases {
		tc := tc
		t.Run(tn, func(t *testing.T) {
			runStatusTest(t, tc)
		})
	}
}

var stsNoStatus = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
   generation: 1
   name: test
`
var stsBadStatus = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
   generation: 1
   name: test
   namespace: qual
status:
   observedGeneration: 1
   currentReplicas: 1
`

var stsOK = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
   generation: 1
   name: test
   namespace: qual
spec:
   replicas: 4
status:
   observedGeneration: 1
   currentReplicas: 4
   readyReplicas: 4
   replicas: 4
`

var stsLessReady = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
   generation: 1
   name: test
   namespace: qual
spec:
   replicas: 4
status:
   observedGeneration: 1
   currentReplicas: 4
   readyReplicas: 2
   replicas: 4
`
var stsLessCurrent = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
   generation: 1
   name: test
   namespace: qual
spec:
   replicas: 4
status:
   observedGeneration: 1
   currentReplicas: 2
   readyReplicas: 4
   replicas: 4
`

func TestStsStatus(t *testing.T) {
	testCases := map[string]testSpec{
		"stsNoStatus": {
			spec:           stsNoStatus,
			expectedStatus: InProgressStatus,
			expectedConditions: []Condition{{
				Type:   ConditionInProgress,
				Status: corev1.ConditionTrue,
				Reason: "LessReplicas",
			}},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
			},
		},
		"stsBadStatus": {
			spec:           stsBadStatus,
			expectedStatus: InProgressStatus,
			expectedConditions: []Condition{{
				Type:   ConditionInProgress,
				Status: corev1.ConditionTrue,
				Reason: "LessReplicas",
			}},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
			},
		},
		"stsOK": {
			spec:               stsOK,
			expectedStatus:     CurrentStatus,
			expectedConditions: []Condition{},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
				ConditionInProgress,
			},
		},
		"stsLessReady": {
			spec:           stsLessReady,
			expectedStatus: InProgressStatus,
			expectedConditions: []Condition{{
				Type:   ConditionInProgress,
				Status: corev1.ConditionTrue,
				Reason: "LessReady",
			}},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
			},
		},
		"stsLessCurrent": {
			spec:           stsLessCurrent,
			expectedStatus: InProgressStatus,
			expectedConditions: []Condition{{
				Type:   ConditionInProgress,
				Status: corev1.ConditionTrue,
				Reason: "LessCurrent",
			}},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
			},
		},
	}

	for tn, tc := range testCases {
		tc := tc
		t.Run(tn, func(t *testing.T) {
			runStatusTest(t, tc)
		})
	}
}

var dsNoStatus = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
   name: test
   generation: 1
`
var dsBadStatus = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
   name: test
   namespace: qual
   generation: 1
status:
   observedGeneration: 1
   currentReplicas: 1
`

var dsOK = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
   name: test
   namespace: qual
   generation: 1
status:
   desiredNumberScheduled: 4
   currentNumberScheduled: 4
   updatedNumberScheduled: 4
   numberAvailable: 4
   numberReady: 4
   observedGeneration: 1
`

var dsLessReady = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
   name: test
   namespace: qual
   generation: 1
status:
   observedGeneration: 1
   desiredNumberScheduled: 4
   currentNumberScheduled: 4
   updatedNumberScheduled: 4
   numberAvailable: 4
   numberReady: 2
`
var dsLessAvailable = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
   name: test
   namespace: qual
   generation: 1
status:
   observedGeneration: 1
   desiredNumberScheduled: 4
   currentNumberScheduled: 4
   updatedNumberScheduled: 4
   numberAvailable: 2
   numberReady: 4
`

func TestDaemonsetStatus(t *testing.T) {
	testCases := map[string]testSpec{
		"dsNoStatus": {
			spec:           dsNoStatus,
			expectedStatus: InProgressStatus,
			expectedConditions: []Condition{{
				Type:   ConditionInProgress,
				Status: corev1.ConditionTrue,
				Reason: "NoDesiredNumber",
			}},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
			},
		},
		"dsBadStatus": {
			spec:           dsBadStatus,
			expectedStatus: InProgressStatus,
			expectedConditions: []Condition{{
				Type:   ConditionInProgress,
				Status: corev1.ConditionTrue,
				Reason: "NoDesiredNumber",
			}},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
			},
		},
		"dsOK": {
			spec:               dsOK,
			expectedStatus:     CurrentStatus,
			expectedConditions: []Condition{},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
				ConditionInProgress,
			},
		},
		"dsLessReady": {
			spec:           dsLessReady,
			expectedStatus: InProgressStatus,
			expectedConditions: []Condition{{
				Type:   ConditionInProgress,
				Status: corev1.ConditionTrue,
				Reason: "LessReady",
			}},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
			},
		},
		"dsLessAvailable": {
			spec:           dsLessAvailable,
			expectedStatus: InProgressStatus,
			expectedConditions: []Condition{{
				Type:   ConditionInProgress,
				Status: corev1.ConditionTrue,
				Reason: "LessAvailable",
			}},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
			},
		},
	}

	for tn, tc := range testCases {
		tc := tc
		t.Run(tn, func(t *testing.T) {
			runStatusTest(t, tc)
		})
	}
}

var depNoStatus = `
apiVersion: apps/v1
kind: Deployment
metadata:
   name: test
   generation: 1
`

var depOK = `
apiVersion: apps/v1
kind: Deployment
metadata:
   name: test
   generation: 1
   namespace: qual
status:
   observedGeneration: 1
   updatedReplicas: 1
   readyReplicas: 1
   availableReplicas: 1
   replicas: 1
   conditions:
    - type: Progressing 
      status: "True"
      reason: NewReplicaSetAvailable
    - type: Available 
      status: "True"
`

var depNotProgressing = `
apiVersion: apps/v1
kind: Deployment
metadata:
   name: test
   generation: 1
   namespace: qual
spec:
   progressDeadlineSeconds: 45
status:
   observedGeneration: 1
   updatedReplicas: 1
   readyReplicas: 1
   availableReplicas: 1
   replicas: 1
   observedGeneration: 1
   conditions:
    - type: Progressing 
      status: "False"
      reason: Some reason
    - type: Available 
      status: "True"
`

var depNoProgressDeadlineSeconds = `
apiVersion: apps/v1
kind: Deployment
metadata:
   name: test
   generation: 1
   namespace: qual
status:
   observedGeneration: 1
   updatedReplicas: 1
   readyReplicas: 1
   availableReplicas: 1
   replicas: 1
   observedGeneration: 1
   conditions:
    - type: Available
      status: "True"
`

var depNotAvailable = `
apiVersion: apps/v1
kind: Deployment
metadata:
   name: test
   generation: 1
   namespace: qual
status:
   observedGeneration: 1
   updatedReplicas: 1
   readyReplicas: 1
   availableReplicas: 1
   replicas: 1
   observedGeneration: 1
   conditions:
    - type: Progressing 
      status: "True"
      reason: NewReplicaSetAvailable
    - type: Available 
      status: "False"
`

func TestDeploymentStatus(t *testing.T) {
	testCases := map[string]testSpec{
		"depNoStatus": {
			spec:           depNoStatus,
			expectedStatus: InProgressStatus,
			expectedConditions: []Condition{{
				Type:   ConditionInProgress,
				Status: corev1.ConditionTrue,
				Reason: "LessReplicas",
			}},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
			},
		},
		"depOK": {
			spec:               depOK,
			expectedStatus:     CurrentStatus,
			expectedConditions: []Condition{},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
				ConditionInProgress,
			},
		},
		"depNotProgressing": {
			spec:           depNotProgressing,
			expectedStatus: InProgressStatus,
			expectedConditions: []Condition{{
				Type:   ConditionInProgress,
				Status: corev1.ConditionTrue,
				Reason: "ReplicaSetNotAvailable",
			}},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
			},
		},
		"depNoProgressDeadlineSeconds": {
			spec:               depNoProgressDeadlineSeconds,
			expectedStatus:     CurrentStatus,
			expectedConditions: []Condition{},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
				ConditionInProgress,
			},
		},
		"depNotAvailable": {
			spec:           depNotAvailable,
			expectedStatus: InProgressStatus,
			expectedConditions: []Condition{{
				Type:   ConditionInProgress,
				Status: corev1.ConditionTrue,
				Reason: "DeploymentNotAvailable",
			}},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
			},
		},
	}

	for tn, tc := range testCases {
		tc := tc
		t.Run(tn, func(t *testing.T) {
			runStatusTest(t, tc)
		})
	}
}

var rsNoStatus = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
   name: test
   generation: 1
`

var rsOK1 = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
   name: test
   namespace: qual
   generation: 1
spec:
   replicas: 2
status:
   observedGeneration: 1
   replicas: 2
   readyReplicas: 2
   availableReplicas: 2
   fullyLabeledReplicas: 2
   conditions:
    - type: ReplicaFailure 
      status: "False"
`

var rsOK2 = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
   name: test
   namespace: qual
   generation: 1
spec:
   replicas: 2
status:
   observedGeneration: 1
   fullyLabeledReplicas: 2
   replicas: 2
   readyReplicas: 2
   availableReplicas: 2
`

var rsLessReady = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
   name: test
   namespace: qual
   generation: 1
spec:
   replicas: 4
status:
   observedGeneration: 1
   replicas: 4
   readyReplicas: 2
   availableReplicas: 4
   fullyLabeledReplicas: 4
`

var rsLessAvailable = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
   name: test
   namespace: qual
   generation: 1
spec:
   replicas: 4
status:
   observedGeneration: 1
   replicas: 4
   readyReplicas: 4
   availableReplicas: 2
   fullyLabeledReplicas: 4
`

var rsReplicaFailure = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
   name: test
   namespace: qual
   generation: 1
spec:
   replicas: 4
status:
   observedGeneration: 1
   replicas: 4
   readyReplicas: 4
   fullyLabeledReplicas: 4
   availableReplicas: 4
   conditions:
    - type: ReplicaFailure 
      status: "True"
`

func TestReplicasetStatus(t *testing.T) {
	testCases := map[string]testSpec{
		"rsNoStatus": {
			spec:           rsNoStatus,
			expectedStatus: InProgressStatus,
			expectedConditions: []Condition{{
				Type:   ConditionInProgress,
				Status: corev1.ConditionTrue,
				Reason: "LessLabelled",
			}},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
			},
		},
		"rsOK1": {
			spec:               rsOK1,
			expectedStatus:     CurrentStatus,
			expectedConditions: []Condition{},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
				ConditionInProgress,
			},
		},
		"rsOK2": {
			spec:               rsOK2,
			expectedStatus:     CurrentStatus,
			expectedConditions: []Condition{},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
				ConditionInProgress,
			},
		},
		"rsLessAvailable": {
			spec:           rsLessAvailable,
			expectedStatus: InProgressStatus,
			expectedConditions: []Condition{{
				Type:   ConditionInProgress,
				Status: corev1.ConditionTrue,
				Reason: "LessAvailable",
			}},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
			},
		},
		"rsLessReady": {
			spec:           rsLessReady,
			expectedStatus: InProgressStatus,
			expectedConditions: []Condition{{
				Type:   ConditionInProgress,
				Status: corev1.ConditionTrue,
				Reason: "LessReady",
			}},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
			},
		},
		"rsReplicaFailure": {
			spec:           rsReplicaFailure,
			expectedStatus: InProgressStatus,
			expectedConditions: []Condition{{
				Type:   ConditionInProgress,
				Status: corev1.ConditionTrue,
				Reason: "ReplicaFailure",
			}},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
			},
		},
	}

	for tn, tc := range testCases {
		tc := tc
		t.Run(tn, func(t *testing.T) {
			runStatusTest(t, tc)
		})
	}
}

var pdbNoStatus = `
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
   generation: 1
   name: test
`

var pdbOK1 = `
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
   generation: 1
   name: test
   namespace: qual
status:
   currentHealthy: 2
   desiredHealthy: 2
`

var pdbMoreHealthy = `
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
   generation: 1
   name: test
   namespace: qual
status:
   currentHealthy: 4
   desiredHealthy: 2
`

var pdbLessHealthy = `
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
   generation: 1
   name: test
   namespace: qual
status:
   currentHealthy: 2
   desiredHealthy: 4
`

func TestPDBStatus(t *testing.T) {
	testCases := map[string]testSpec{
		"pdbNoStatus": {
			spec:           pdbNoStatus,
			expectedStatus: InProgressStatus,
			expectedConditions: []Condition{{
				Type:   ConditionInProgress,
				Status: corev1.ConditionTrue,
				Reason: "ZeroDesiredHealthy",
			}},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
			},
		},
		"pdbOK1": {
			spec:               pdbOK1,
			expectedStatus:     CurrentStatus,
			expectedConditions: []Condition{},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
				ConditionInProgress,
			},
		},
		"pdbMoreHealthy": {
			spec:               pdbMoreHealthy,
			expectedStatus:     CurrentStatus,
			expectedConditions: []Condition{},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
				ConditionInProgress,
			},
		},
		"pdbLessHealthy": {
			spec:           pdbLessHealthy,
			expectedStatus: InProgressStatus,
			expectedConditions: []Condition{{
				Type:   ConditionInProgress,
				Status: corev1.ConditionTrue,
				Reason: "BudgetNotMet",
			}},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
			},
		},
	}

	for tn, tc := range testCases {
		tc := tc
		t.Run(tn, func(t *testing.T) {
			runStatusTest(t, tc)
		})
	}
}

var crdNoStatus = `
apiVersion: something/v1
kind: MyCR
metadata:
   generation: 1
   name: test
   namespace: qual
`

var crdMismatchStatusGeneration = `
apiVersion: something/v1
kind: MyCR
metadata:
   name: test
   namespace: qual
   generation: 2
status:
   observedGeneration: 1
`

var crdReady = `
apiVersion: something/v1
kind: MyCR
metadata:
   name: test
   namespace: qual
   generation: 1
status:
   conditions:
    - type: Ready 
      status: "True"
      message: All looks ok
      reason: AllOk
`

var crdNotReady = `
apiVersion: something/v1
kind: MyCR
metadata:
   generation: 1
   name: test
   namespace: qual
status:
   observedGeneration: 1
   conditions:
    - type: Ready 
      status: "False"
`

var crdNoCondition = `
apiVersion: something/v1
kind: MyCR
metadata:
   name: test
   namespace: qual
   generation: 1
status:
   conditions:
    - type: SomeCondition 
      status: "False"
`

func TestCRDGenericStatus(t *testing.T) {
	testCases := map[string]testSpec{
		"crdNoStatus": {
			spec:               crdNoStatus,
			expectedStatus:     CurrentStatus,
			expectedConditions: []Condition{},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
				ConditionInProgress,
			},
		},
		"crdReady": {
			spec:               crdReady,
			expectedStatus:     CurrentStatus,
			expectedConditions: []Condition{},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
				ConditionInProgress,
			},
		},
		"crdNotReady": {
			spec:               crdNotReady,
			expectedStatus:     CurrentStatus,
			expectedConditions: []Condition{},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
				ConditionInProgress,
			},
		},
		"crdNoCondition": {
			spec:               crdNoCondition,
			expectedStatus:     CurrentStatus,
			expectedConditions: []Condition{},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
				ConditionInProgress,
			},
		},
		"crdMismatchStatusGeneration": {
			spec:           crdMismatchStatusGeneration,
			expectedStatus: InProgressStatus,
			expectedConditions: []Condition{{
				Type:   ConditionInProgress,
				Status: corev1.ConditionTrue,
				Reason: "LatestGenerationNotObserved",
			}},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
			},
		},
	}

	for tn, tc := range testCases {
		tc := tc
		t.Run(tn, func(t *testing.T) {
			runStatusTest(t, tc)
		})
	}
}

var jobNoStatus = `
apiVersion: batch/v1
kind: Job
metadata:
   name: test
   namespace: qual
   generation: 1
`

var jobComplete = `
apiVersion: batch/v1
kind: Job
metadata:
   name: test
   namespace: qual
   generation: 1
status:
   succeeded: 1
   active: 0
   conditions:
    - type: Complete 
      status: "True"
`

var jobFailed = `
apiVersion: batch/v1
kind: Job
metadata:
   name: test
   namespace: qual
   generation: 1
spec:
   completions: 4
status:
   succeeded: 3
   failed: 1
   conditions:
    - type: Failed 
      status: "True"
      reason: JobFailed
`

var jobInProgress = `
apiVersion: batch/v1
kind: Job
metadata:
   name: test
   namespace: qual
   generation: 1
spec:
   completions: 10
   parallelism: 2
status:
   startTime: "2019-06-04T01:17:13Z"
   succeeded: 3
   failed: 1
   active: 2
   conditions:
    - type: Failed 
      status: "False"
    - type: Complete 
      status: "False"
`

func TestJobStatus(t *testing.T) {
	testCases := map[string]testSpec{
		"jobNoStatus": {
			spec:           jobNoStatus,
			expectedStatus: InProgressStatus,
			expectedConditions: []Condition{{
				Type:   ConditionInProgress,
				Status: corev1.ConditionTrue,
				Reason: "JobNotStarted",
			}},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
			},
		},
		"jobComplete": {
			spec:               jobComplete,
			expectedStatus:     CurrentStatus,
			expectedConditions: []Condition{},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
				ConditionInProgress,
			},
		},
		"jobFailed": {
			spec:           jobFailed,
			expectedStatus: FailedStatus,
			expectedConditions: []Condition{{
				Type:   ConditionFailed,
				Status: corev1.ConditionTrue,
				Reason: "JobFailed",
			}},
			absentConditionTypes: []ConditionType{
				ConditionInProgress,
			},
		},
		"jobInProgress": {
			spec:               jobInProgress,
			expectedStatus:     CurrentStatus,
			expectedConditions: []Condition{},
			absentConditionTypes: []ConditionType{
				ConditionInProgress,
				ConditionFailed,
			},
		},
	}

	for tn, tc := range testCases {
		tc := tc
		t.Run(tn, func(t *testing.T) {
			runStatusTest(t, tc)
		})
	}
}

var cronjobNoStatus = `
apiVersion: batch/v1
kind: CronJob
metadata:
   name: test
   namespace: qual
   generation: 1
`

var cronjobWithStatus = `
apiVersion: batch/v1
kind: CronJob
metadata:
   name: test
   namespace: qual
   generation: 1
status:
`

func TestCronJobStatus(t *testing.T) {
	testCases := map[string]testSpec{
		"cronjobNoStatus": {
			spec:               cronjobNoStatus,
			expectedStatus:     CurrentStatus,
			expectedConditions: []Condition{},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
				ConditionInProgress,
			},
		},
		"cronjobWithStatus": {
			spec:               cronjobWithStatus,
			expectedStatus:     CurrentStatus,
			expectedConditions: []Condition{},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
				ConditionInProgress,
			},
		},
	}

	for tn, tc := range testCases {
		tc := tc
		t.Run(tn, func(t *testing.T) {
			runStatusTest(t, tc)
		})
	}
}

var serviceDefault = `
apiVersion: v1
kind: Service
metadata:
   name: test
   namespace: qual
   generation: 1
`

var serviceNodePort = `
apiVersion: v1
kind: Service
metadata:
   name: test
   namespace: qual
   generation: 1
spec:
  type: NodePort
`

var serviceLBok = `
apiVersion: v1
kind: Service
metadata:
   name: test
   namespace: qual
   generation: 1
spec:
  type: LoadBalancer
  clusterIP: "1.2.3.4"
`
var serviceLBnok = `
apiVersion: v1
kind: Service
metadata:
   name: test
   namespace: qual
   generation: 1
spec:
  type: LoadBalancer
`

func TestServiceStatus(t *testing.T) {
	testCases := map[string]testSpec{
		"serviceDefault": {
			spec:               serviceDefault,
			expectedStatus:     CurrentStatus,
			expectedConditions: []Condition{},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
				ConditionInProgress,
			},
		},
		"serviceNodePort": {
			spec:               serviceNodePort,
			expectedStatus:     CurrentStatus,
			expectedConditions: []Condition{},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
				ConditionInProgress,
			},
		},
		"serviceLBnok": {
			spec:           serviceLBnok,
			expectedStatus: InProgressStatus,
			expectedConditions: []Condition{{
				Type:   ConditionInProgress,
				Status: corev1.ConditionTrue,
				Reason: "NoIPAssigned",
			}},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
			},
		},
		"serviceLBok": {
			spec:               serviceLBok,
			expectedStatus:     CurrentStatus,
			expectedConditions: []Condition{},
			absentConditionTypes: []ConditionType{
				ConditionFailed,
				ConditionInProgress,
			},
		},
	}

	for tn, tc := range testCases {
		tc := tc
		t.Run(tn, func(t *testing.T) {
			runStatusTest(t, tc)
		})
	}
}
