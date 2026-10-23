package version

import (
	"strings"
	"testing"
)

func TestValidateInvalidOutput(t *testing.T) {
	o := NewOptions(nil)
	o.Output = "ym1"

	err := o.Validate(nil)
	if err == nil {
		t.Fatal("expected an error for invalid output, got nil")
	}
	if !strings.Contains(err.Error(), "--output must be 'yaml' or 'json'") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
func TestValidateValidOutput(t *testing.T) {
	for _, output := range []string{"", "yaml", "json"} {
		o := NewOptions(nil)
		o.Output = output

		if err := o.Validate(nil); err != nil {
			t.Fatalf("output %q should be valid, got error: %v", output, err)
		}
	}
}