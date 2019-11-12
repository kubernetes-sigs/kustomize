// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package status

import (
	"time"

	"github.com/pkg/errors"
	corev1 "sigs.k8s.io/kustomize/pseudo/k8s/api/core/v1"
	"sigs.k8s.io/kustomize/pseudo/k8s/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	// The set of standard conditions defined in this package. These follow the "abnormality-true"
	// convention where conditions should have a true value for abnormal/error situations and the absence
	// of a condition should be interpreted as a false value, i.e. everything is normal.
	ConditionFailed     ConditionType = "Failed"
	ConditionInProgress ConditionType = "InProgress"

	// The set of status conditions which can be assigned to resources.
	InProgressStatus  Status = "InProgress"
	FailedStatus      Status = "Failed"
	CurrentStatus     Status = "Current"
	TerminatingStatus Status = "Terminating"
)

// ConditionType defines the set of condition types allowed inside a Condition struct.
type ConditionType string

// String returns the ConditionType as a string.
func (c ConditionType) String() string {
	return string(c)
}

// Status defines the set of statuses a resource can have.
type Status string

// String returns the status as a string.
func (s Status) String() string {
	return string(s)
}

// Result contains the results of a call to compute the status of
// a resource.
type Result struct {
	//Status
	Status Status
	// Message
	Message string
	// Conditions list of extracted conditions from Resource
	Conditions []Condition
}

// Condition defines the general format for conditions on Kubernetes resources.
// In practice, each kubernetes resource defines their own format for conditions, but
// most (maybe all) follows this structure.
type Condition struct {
	// Type condition type
	Type ConditionType
	// Status String that describes the condition status
	Status corev1.ConditionStatus
	// Reason one work CamelCase reason
	Reason string
	// Message Human readable reason string
	Message string
}

// Compute finds the status of a given unstructured resource. It does not
// fetch the state of the resource from a cluster, so the provided unstructured
// must have the complete state, including status.
//
// The returned result contains the status of the resource, which will be
// one of
//  * InProgress
//  * Current
//  * Failed
//  * Terminating
// It also contains a message that provides more information on why
// the resource has the given status. Finally, the result also contains
// a list of standard resources that would belong on the given resource.
func Compute(u *unstructured.Unstructured) (*Result, error) {
	res, err := checkGenericProperties(u)
	if err != nil || res != nil {
		return res, err
	}

	fn := GetLegacyConditionsFn(u)
	if fn != nil {
		return fn(u)
	}

	// The resource is not one of the built-in types with specific
	// rules and we were unable to make a decision based on the
	// generic rules. In this case we assume that the absence of any known
	// conditions means the resource is current.
	return &Result{
		Status:     CurrentStatus,
		Message:    "Resource is current",
		Conditions: []Condition{},
	}, err
}

// Augment takes a resource and augments the resource with the
// standard status conditions.
func Augment(u *unstructured.Unstructured) error {
	res, err := Compute(u)
	if err != nil {
		return err
	}

	conditions, found, err := unstructured.NestedSlice(u.Object, "status", "conditions")
	if err != nil {
		return err
	}

	if !found {
		conditions = make([]interface{}, 0)
	}

	currentTime := time.Now().UTC().Format(time.RFC3339)

	for _, resCondition := range res.Conditions {
		present := false
		for _, c := range conditions {
			condition, ok := c.(map[string]interface{})
			if !ok {
				return errors.New("condition does not have the expected structure")
			}
			conditionType, ok := condition["type"].(string)
			if !ok {
				return errors.New("condition type does not have the expected type")
			}
			if conditionType == string(resCondition.Type) {
				conditionStatus, ok := condition["status"].(string)
				if !ok {
					return errors.New("condition status does not have the expected type")
				}
				if conditionStatus != string(resCondition.Status) {
					condition["lastTransitionTime"] = currentTime
				}
				condition["status"] = string(resCondition.Status)
				condition["lastUpdateTime"] = currentTime
				condition["reason"] = resCondition.Reason
				condition["message"] = resCondition.Message
				present = true
			}
		}
		if !present {
			conditions = append(conditions, map[string]interface{}{
				"lastTransitionTime": currentTime,
				"lastUpdateTime":     currentTime,
				"message":            resCondition.Message,
				"reason":             resCondition.Reason,
				"status":             string(resCondition.Status),
				"type":               string(resCondition.Type),
			})
		}
	}
	return unstructured.SetNestedSlice(u.Object, conditions, "status", "conditions")
}
