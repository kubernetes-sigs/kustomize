package validate

import "testing"

func TestIsValidLabel(t *testing.T) {
	testcases := []struct {
		input, name string
		valid       bool
	}{
		{
			input: "otters:cute",
			valid: true,
			name:  "Valid input format",
		},
		{
			input: "dogs,cats",
			valid: false,
			name:  "Does not contain colon",
		},
		{
			input: ":noKey",
			valid: false,
			name:  "Missing key",
		},
		{
			input: "noValue:",
			valid: false,
			name:  "Missing value",
		},
		{
			input: "exclamation!:point",
			valid: false,
			name:  "Non-alphanumeric input",
		},
		{
			input: "123:45",
			valid: true,
			name:  "Numeric input is allowed",
		},
	}
	for _, tc := range testcases {
		ok, err := IsValidLabel(tc.input)
		if tc.valid && err != nil {
			t.Errorf("unexpected error: for test case %s, expected no error but got: %s", tc.name, err.Error())
		}
		if ok && !tc.valid {
			t.Errorf("for test case %s, expected invalid label format error", tc.name)
		}
		if !ok && tc.valid {
			t.Errorf("unexpected error: for test case %s, expected test to pass", tc.name)
		}
	}
}

func TestIsValidAnnotation(t *testing.T) {
	testcases := []struct {
		input, name string
		valid       bool
	}{
		{
			input: "owls:adorable",
			valid: true,
			name:  "Valid input format",
		},
		{
			input: "cake,cookies",
			valid: false,
			name:  "Does not contain colon",
		},
		{
			input: ":noKey",
			valid: false,
			name:  "Missing key",
		},
		{
			input: "noValue:",
			valid: false,
			name:  "Missing value",
		},
		{
			input: "exclamation!:point",
			valid: false,
			name:  "Input has a bang!",
		},
		{
			input: "987:65",
			valid: true,
			name:  "Numeric input is valid",
		},
	}
	for _, tc := range testcases {
		ok, err := IsValidAnnotation(tc.input)
		if tc.valid && err != nil {
			t.Errorf("unexpected error: for test case %s, expected no error but got: %s", tc.name, err.Error())
		}
		if ok && !tc.valid {
			t.Errorf("for test case %s, expected invalid annotation format error", tc.name)
		}
		if !ok && tc.valid {
			t.Errorf("unexpected error: for test case %s, expected test to pass", tc.name)
		}
	}
}
