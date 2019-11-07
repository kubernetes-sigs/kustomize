/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package validation

import (
	"fmt"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestValidateLabels(t *testing.T) {
	successCases := []map[string]string{
		{"simple": "bar"},
		{"now-with-dashes": "bar"},
		{"1-starts-with-num": "bar"},
		{"1234": "bar"},
		{"simple/simple": "bar"},
		{"now-with-dashes/simple": "bar"},
		{"now-with-dashes/now-with-dashes": "bar"},
		{"now.with.dots/simple": "bar"},
		{"now-with.dashes-and.dots/simple": "bar"},
		{"1-num.2-num/3-num": "bar"},
		{"1234/5678": "bar"},
		{"1.2.3.4/5678": "bar"},
		{"UpperCaseAreOK123": "bar"},
		{"goodvalue": "123_-.BaR"},
	}
	for i := range successCases {
		errs := ValidateLabels(successCases[i], field.NewPath("field"))
		if len(errs) != 0 {
			t.Errorf("case[%d] expected success, got %#v", i, errs)
		}
	}

	namePartErrMsg := "name part must consist of"
	nameErrMsg := "a qualified name must consist of"
	labelErrMsg := "a valid label must be an empty string or consist of"
	maxLengthErrMsg := "must be no more than"

	labelNameErrorCases := []struct {
		labels map[string]string
		expect string
	}{
		{map[string]string{"nospecialchars^=@": "bar"}, namePartErrMsg},
		{map[string]string{"cantendwithadash-": "bar"}, namePartErrMsg},
		{map[string]string{"only/one/slash": "bar"}, nameErrMsg},
		{map[string]string{strings.Repeat("a", 254): "bar"}, maxLengthErrMsg},
	}
	for i := range labelNameErrorCases {
		errs := ValidateLabels(labelNameErrorCases[i].labels, field.NewPath("field"))
		if len(errs) != 1 {
			t.Errorf("case[%d]: expected failure", i)
		} else {
			if !strings.Contains(errs[0].Detail, labelNameErrorCases[i].expect) {
				t.Errorf("case[%d]: error details do not include %q: %q", i, labelNameErrorCases[i].expect, errs[0].Detail)
			}
		}
	}

	labelValueErrorCases := []struct {
		labels map[string]string
		expect string
	}{
		{map[string]string{"toolongvalue": strings.Repeat("a", 64)}, maxLengthErrMsg},
		{map[string]string{"backslashesinvalue": "some\\bad\\value"}, labelErrMsg},
		{map[string]string{"nocommasallowed": "bad,value"}, labelErrMsg},
		{map[string]string{"strangecharsinvalue": "?#$notsogood"}, labelErrMsg},
	}
	for i := range labelValueErrorCases {
		errs := ValidateLabels(labelValueErrorCases[i].labels, field.NewPath("field"))
		if len(errs) != 1 {
			t.Errorf("case[%d]: expected failure", i)
		} else {
			if !strings.Contains(errs[0].Detail, labelValueErrorCases[i].expect) {
				t.Errorf("case[%d]: error details do not include %q: %q", i, labelValueErrorCases[i].expect, errs[0].Detail)
			}
		}
	}
}

func TestValidDryRun(t *testing.T) {
	tests := [][]string{
		{},
		{"All"},
		{"All", "All"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test), func(t *testing.T) {
			if errs := ValidateDryRun(field.NewPath("dryRun"), test); len(errs) != 0 {
				t.Errorf("%v should be a valid dry-run value: %v", test, errs)
			}
		})
	}
}

func TestInvalidDryRun(t *testing.T) {
	tests := [][]string{
		{"False"},
		{"All", "False"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test), func(t *testing.T) {
			if len(ValidateDryRun(field.NewPath("dryRun"), test)) == 0 {
				t.Errorf("%v shouldn't be a valid dry-run value", test)
			}
		})
	}

}

func boolPtr(b bool) *bool {
	return &b
}

func TestValidPatchOptions(t *testing.T) {
	tests := []struct {
		opts      metav1.PatchOptions
		patchType types.PatchType
	}{
		{
			opts: metav1.PatchOptions{
				Force:        boolPtr(true),
				FieldManager: "kubectl",
			},
			patchType: types.ApplyPatchType,
		},
		{
			opts: metav1.PatchOptions{
				FieldManager: "kubectl",
			},
			patchType: types.ApplyPatchType,
		},
		{
			opts:      metav1.PatchOptions{},
			patchType: types.MergePatchType,
		},
		{
			opts: metav1.PatchOptions{
				FieldManager: "patcher",
			},
			patchType: types.MergePatchType,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test.opts), func(t *testing.T) {
			errs := ValidatePatchOptions(&test.opts, test.patchType)
			if len(errs) != 0 {
				t.Fatalf("Expected no failures, got: %v", errs)
			}
		})
	}
}

func TestInvalidPatchOptions(t *testing.T) {
	tests := []struct {
		opts      metav1.PatchOptions
		patchType types.PatchType
	}{
		// missing manager
		{
			opts:      metav1.PatchOptions{},
			patchType: types.ApplyPatchType,
		},
		// force on non-apply
		{
			opts: metav1.PatchOptions{
				Force: boolPtr(true),
			},
			patchType: types.MergePatchType,
		},
		// manager and force on non-apply
		{
			opts: metav1.PatchOptions{
				FieldManager: "kubectl",
				Force:        boolPtr(false),
			},
			patchType: types.MergePatchType,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test.opts), func(t *testing.T) {
			errs := ValidatePatchOptions(&test.opts, test.patchType)
			if len(errs) == 0 {
				t.Fatal("Expected failures, got none.")
			}
		})
	}
}

func TestValidateFieldManagerValid(t *testing.T) {
	tests := []string{
		"filedManager",
		"你好", // Hello
		"🍔",
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			errs := ValidateFieldManager(test, field.NewPath("fieldManager"))
			if len(errs) != 0 {
				t.Errorf("Validation failed: %v", errs)
			}
		})
	}
}

func TestValidateFieldManagerInvalid(t *testing.T) {
	tests := []string{
		"field\nmanager", // Contains invalid character \n
		"fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", // Has 129 chars
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			errs := ValidateFieldManager(test, field.NewPath("fieldManager"))
			if len(errs) == 0 {
				t.Errorf("Validation should have failed")
			}
		})
	}
}

func TestValidateMangedFieldsInvalid(t *testing.T) {
	tests := []metav1.ManagedFieldsEntry{
		{
			Operation: metav1.ManagedFieldsOperationUpdate,
			// FieldsType is missing
		},
		{
			Operation:  metav1.ManagedFieldsOperationUpdate,
			FieldsType: "RandomVersion",
		},
		{
			Operation:  "RandomOperation",
			FieldsType: "FieldsV1",
		},
		{
			// Operation is missing
			FieldsType: "FieldsV1",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v", test), func(t *testing.T) {
			errs := ValidateManagedFields([]metav1.ManagedFieldsEntry{test}, field.NewPath("managedFields"))
			if len(errs) == 0 {
				t.Errorf("Validation should have failed")
			}
		})
	}
}

func TestValidateMangedFieldsValid(t *testing.T) {
	tests := []metav1.ManagedFieldsEntry{
		{
			Operation:  metav1.ManagedFieldsOperationUpdate,
			FieldsType: "FieldsV1",
		},
		{
			Operation:  metav1.ManagedFieldsOperationApply,
			FieldsType: "FieldsV1",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v", test), func(t *testing.T) {
			err := ValidateManagedFields([]metav1.ManagedFieldsEntry{test}, field.NewPath("managedFields"))
			if err != nil {
				t.Errorf("Validation failed: %v", err)
			}
		})
	}
}
