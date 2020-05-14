package main

import (
	"testing"
)

func TestGetModuleCurrentVersion(t *testing.T) {
	output := getModuleCurrentVersion("test")
	if output != "v1.0.0" {
		t.Errorf("Unexpected output: %s", output)
	}
}
