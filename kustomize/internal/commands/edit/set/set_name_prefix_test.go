// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/testutils"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
)

const (
	goodPrefixValue = "acme-"
)

func TestSetNamePrefixHappyPath(t *testing.T) {
	fSys := fs.MakeFsInMemory()
	testutils.WriteTestKustomization(fSys)

	cmd := newCmdSetNamePrefix(fSys)
	args := []string{goodPrefixValue}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	content, err := testutils.ReadTestKustomization(fSys)
	if err != nil {
		t.Errorf("unexpected read error: %v", err)
	}
	if !strings.Contains(string(content), goodPrefixValue) {
		t.Errorf("expected prefix value in kustomization file")
	}
}

func TestSetNamePrefixNoArgs(t *testing.T) {
	fSys := fs.MakeFsInMemory()

	cmd := newCmdSetNamePrefix(fSys)
	err := cmd.Execute()
	if err == nil {
		t.Errorf("expected error: %v", err)
	}
	if err.Error() != "must specify exactly one prefix value" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}
