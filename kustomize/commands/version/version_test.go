package version

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionRejectsInvalidOutput(t *testing.T) {
	var buf bytes.Buffer
	cmd := NewCmdVersion(&buf)
	cmd.SetArgs([]string{"--output=yml"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error, got nil (output=%q)", buf.String())
	}
	if got, want := err.Error(), "--output must be 'yaml' or 'json'"; !strings.Contains(got, want) {
		t.Fatalf("expected error to contain %q, got %q", want, got)
	}
}
