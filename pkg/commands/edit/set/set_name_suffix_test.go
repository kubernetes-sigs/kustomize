// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/v3/pkg/fs"
)

const (
	goodSuffixValue = "-acme"
)

func TestSetNameSuffixHappyPath(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteTestKustomization()

	cmd := newCmdSetNameSuffix(fakeFS)
	args := []string{goodSuffixValue}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	content, err := fakeFS.ReadTestKustomization()
	if err != nil {
		t.Errorf("unexpected read error: %v", err)
	}
	if !strings.Contains(string(content), goodSuffixValue) {
		t.Errorf("expected suffix value in kustomization file")
	}
}

func TestSetNameSuffixNoArgs(t *testing.T) {
	fakeFS := fs.MakeFakeFS()

	cmd := newCmdSetNameSuffix(fakeFS)
	err := cmd.Execute()
	if err == nil {
		t.Errorf("expected error: %v", err)
	}
	if err.Error() != "must specify exactly one suffix value" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}
