package validators

import (
	"errors"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	v1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"testing"
)

// MapValidatorFunc returns an error if a map contains errors.
type MapValidatorFunc func(map[string]string) error

// MakeAnnotationValidator returns a MapValidatorFunc using apimachinery.
func MakeAnnotationValidator() MapValidatorFunc {
	return func(x map[string]string) error {
		errs := apivalidation.ValidateAnnotations(x, field.NewPath("field"))
		if errs != nil {
			return errors.New(errs.ToAggregate().Error())
		}
		return nil
	}
}

// MakeLabelValidator returns a MapValidatorFunc using apimachinery.
func MakeLabelValidator() MapValidatorFunc {
	return func(x map[string]string) error {
		errs := v1validation.ValidateLabels(x, field.NewPath("field"))
		if errs != nil {
			return errors.New(errs.ToAggregate().Error())
		}
		return nil
	}
}

// FakeValidator can be used in tests.
type FakeValidator struct {
	happy  bool
	called bool
	t      *testing.T
}

// SAD is an error string.
const SAD = "i'm not happy Bob, NOT HAPPY"

// MakeHappyMapValidator makes a FakeValidator that always passes.
func MakeHappyMapValidator(t *testing.T) *FakeValidator {
	return &FakeValidator{happy: true, t: t}
}

// MakeSadMapValidator makes a FakeValidator that always fails.
func MakeSadMapValidator(t *testing.T) *FakeValidator {
	return &FakeValidator{happy: false, t: t}
}

// Validator replaces apimachinery validation in tests.
// Can be set to fail or succeed to test error handling.
// Can confirm if run or not run by surrounding code.
func (v *FakeValidator) Validator(_ map[string]string) error {
	v.called = true
	if v.happy {
		return nil
	}
	return errors.New(SAD)
}

// VerifyCall returns true if Validator was used.
func (v *FakeValidator) VerifyCall() {
	if !v.called {
		v.t.Errorf("should have called Validator")
	}
}

// VerifyNoCall returns true if Validator was not used.
func (v *FakeValidator) VerifyNoCall() {
	if v.called {
		v.t.Errorf("should not have called Validator")
	}
}
