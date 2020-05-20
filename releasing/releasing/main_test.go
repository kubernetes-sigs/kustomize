package main

import (
	"testing"
)

func TestList(t *testing.T) {
	err := listCmdImpl()
	if err != nil {
		t.Error(err)
	}
}

func TestRelease(t *testing.T) {
	args := []string{"api", "minor"}
	err := releaseCmdImpl(args)
	if err != nil {
		t.Error(err)
	}
}
