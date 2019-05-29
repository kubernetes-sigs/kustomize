// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package validator provides functions to validate labels, annotations,
// namespaces and configmap/secret keys using apimachinery functions.
package validator

import (
	"errors"
	"fmt"
	"strings"

	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	v1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// KustValidator validates Labels and annotations by apimachinery
type KustValidator struct{}

// NewKustValidator returns a KustValidator object
func NewKustValidator() *KustValidator {
	return &KustValidator{}
}

func (v *KustValidator) ErrIfInvalidKey(k string) error {
	if errs := validation.IsConfigMapKey(k); len(errs) != 0 {
		return fmt.Errorf(
			"%q is not a valid key name: %s",
			k, strings.Join(errs, ";"))
	}
	return nil
}

func (v *KustValidator) IsEnvVarName(k string) error {
	if errs := validation.IsEnvVarName(k); len(errs) != 0 {
		return fmt.Errorf(
			"%q is not a valid key name: %s",
			k, strings.Join(errs, ";"))
	}
	return nil
}

// MakeAnnotationValidator returns a MapValidatorFunc using apimachinery.
func (v *KustValidator) MakeAnnotationValidator() func(map[string]string) error {
	return func(x map[string]string) error {
		errs := apivalidation.ValidateAnnotations(x, field.NewPath("field"))
		if len(errs) > 0 {
			return errors.New(errs.ToAggregate().Error())
		}
		return nil
	}
}

// MakeLabelValidator returns a MapValidatorFunc using apimachinery.
func (v *KustValidator) MakeLabelValidator() func(map[string]string) error {
	return func(x map[string]string) error {
		errs := v1validation.ValidateLabels(x, field.NewPath("field"))
		if len(errs) > 0 {
			return errors.New(errs.ToAggregate().Error())
		}
		return nil
	}
}

// ValidateNamespace validates a string is a valid namespace using apimachinery.
func (v *KustValidator) ValidateNamespace(s string) []string {
	return validation.IsDNS1123Label(s)
}
