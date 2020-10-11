// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"fmt"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/kustfile"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v3/internal/commands/testutils"
)

const (
	transformerFileName    = "myWonderfulTransformer.yaml"
	transformerFileContent = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit,
sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
`
)

func TestAddTransformerHappyPath(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	fSys.WriteFile(transformerFileName, []byte(transformerFileContent))
	fSys.WriteFile(transformerFileName+"another", []byte(transformerFileContent))
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddTransformer(fSys)
	args := []string{transformerFileName + "*"}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	content, err := testutils_test.ReadTestKustomization(fSys)
	if err != nil {
		t.Errorf("unexpected read error: %v", err)
	}
	if !strings.Contains(string(content), transformerFileName) {
		t.Errorf("expected transformer name in kustomization")
	}
	if !strings.Contains(string(content), transformerFileName+"another") {
		t.Errorf("expected transformer name in kustomization")
	}
}

func TestAddTransformerAlreadyThere(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	fSys.WriteFile(transformerFileName, []byte(transformerFileName))
	testutils_test.WriteTestKustomization(fSys)

	cmd := newCmdAddTransformer(fSys)
	args := []string{transformerFileName}
	err := cmd.RunE(cmd, args)
	if err != nil {
		t.Fatalf("unexpected cmd error: %v", err)
	}

	// adding an existing transformer shouldn't return an error
	err = cmd.RunE(cmd, args)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}

	// There can be only one. May it be the...
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		t.Fatalf("error retrieving kustomization file: %v", err)
	}
	m, err := mf.Read()
	if err != nil {
		t.Fatalf("error reading kustomization file: %v", err)
	}

	if len(m.Transformers) != 1 || m.Transformers[0] != transformerFileName {
		t.Errorf("expected transformers [%s]; got transformers [%s]", transformerFileName, strings.Join(m.Transformers, ","))
	}
}

func TestAddTransformerNoArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()

	cmd := newCmdAddTransformer(fSys)
	err := cmd.Execute()
	if err == nil {
		t.Errorf("expected error: %v", err)
	}
	if err.Error() != "must specify a transformer file" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}

func TestAddTransformerMissingKustomizationYAML(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	fSys.WriteFile(transformerFileName, []byte(transformerFileContent))
	fSys.WriteFile(transformerFileName+"another", []byte(transformerFileContent))

	cmd := newCmdAddTransformer(fSys)
	args := []string{transformerFileName + "*"}
	err := cmd.RunE(cmd, args)
	if err == nil {
		t.Errorf("expected error: %v", err)
	}
	fmt.Println(err.Error())
	if err.Error() != "Missing kustomization file 'kustomization.yaml'.\n" {
		t.Errorf("incorrect error: %v", err.Error())
	}
}
