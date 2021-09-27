// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package valtest_test defines a fakeValidator that can be used in tests
package valtest_test

import (
	"errors"
	"regexp"
	"testing"
)

// fakeValidator can be used in tests.
type fakeValidator struct {
	happy  bool
	called bool
	t      *testing.T
}

// SAD is an error string.
const SAD = "i'm not happy Bob, NOT HAPPY"

// MakeHappyMapValidator makes a fakeValidator that always passes.
func MakeHappyMapValidator(t *testing.T) *fakeValidator {
	return &fakeValidator{happy: true, t: t}
}

// MakeSadMapValidator makes a fakeValidator that always fails.
func MakeSadMapValidator(t *testing.T) *fakeValidator {
	return &fakeValidator{happy: false, t: t}
}

// MakeFakeValidator makes an empty Fake Validator.
func MakeFakeValidator() *fakeValidator {
	return &fakeValidator{}
}

// ErrIfInvalidKey returns nil
func (v *fakeValidator) ErrIfInvalidKey(_ string) error {
	return nil
}

// IsEnvVarName returns nil
func (v *fakeValidator) IsEnvVarName(_ string) error {
	return nil
}

// MakeAnnotationValidator returns a nil function
func (v *fakeValidator) MakeAnnotationValidator() func(map[string]string) error {
	return nil
}

// MakeAnnotationNameValidator returns a nil function
func (v *fakeValidator) MakeAnnotationNameValidator() func([]string) error {
	return nil
}

// MakeLabelValidator returns a nil function
func (v *fakeValidator) MakeLabelValidator() func(map[string]string) error {
	return nil
}

// MakeLabelNameValidator returns a nil function
func (v *fakeValidator) MakeLabelNameValidator() func([]string) error {
	return nil
}

// ValidateNamespace validates namespace by regexp
func (v *fakeValidator) ValidateNamespace(s string) []string {
	pattern := regexp.MustCompile(`^[a-zA-Z].*`)
	if pattern.MatchString(s) {
		return nil
	}
	return []string{"doesn't match"}
}

// Validator replaces apimachinery validation in tests.
// Can be set to fail or succeed to test error handling.
// Can confirm if run or not run by surrounding code.
func (v *fakeValidator) Validator(_ map[string]string) error {
	v.called = true
	if v.happy {
		return nil
	}
	return errors.New(SAD)
}

func (v *fakeValidator) ValidatorArray(_ []string) error {
	v.called = true
	if v.happy {
		return nil
	}
	return errors.New(SAD)
}

// VerifyCall returns true if Validator was used.
func (v *fakeValidator) VerifyCall() {
	if !v.called {
		v.t.Errorf("should have called Validator")
	}
}

// VerifyNoCall returns true if Validator was not used.
func (v *fakeValidator) VerifyNoCall() {
	if v.called {
		v.t.Errorf("should not have called Validator")
	}
}
