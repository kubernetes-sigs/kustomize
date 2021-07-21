// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package frameworktestutil

import (
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestProcessorResultsChecker_UpdateExpectedFromActual(t *testing.T) {
	dir := filepath.FromSlash("testdata/update_expectations/processor")
	checker := ProcessorResultsChecker{
		TestDataDirectory:        dir,
		UpdateExpectedFromActual: true,
		Processor:                testProcessor,
	}
	// This should result in the test being skipped. If no tests are found, it will instead fail.
	checker.Assert(t)
	require.Contains(t, checker.TestCasesRun(), filepath.Join(dir, "important_subdir"))

	checker.UpdateExpectedFromActual = false
	// This time should inherently pass
	checker.Assert(t)
	require.Contains(t, checker.TestCasesRun(), filepath.Join(dir, "important_subdir"))
}

func TestCommandResultsChecker_UpdateExpectedFromActual(t *testing.T) {
	dir := filepath.FromSlash("testdata/update_expectations/command")
	checker := CommandResultsChecker{
		TestDataDirectory:        dir,
		UpdateExpectedFromActual: true,
		Command:                  testCommand,
	}
	// This should result in the test being skipped. If no tests are found, it will instead fail.
	checker.Assert(t)
	require.Contains(t, checker.TestCasesRun(), filepath.Join(dir, "important_subdir"))

	checker.UpdateExpectedFromActual = false
	// This time should inherently pass
	checker.Assert(t)
	require.Contains(t, checker.TestCasesRun(), filepath.Join(dir, "important_subdir"))
}

func testCommand() *cobra.Command {
	return command.Build(testProcessor(), command.StandaloneEnabled, false)
}

func testProcessor() framework.ResourceListProcessor {
	return framework.SimpleProcessor{
		Filter: kio.FilterFunc(func(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
			for _, node := range nodes {
				err := node.SetAnnotations(map[string]string{"updated": "true"})
				if err != nil {
					return nil, err
				}
			}
			return nodes, nil
		}),
	}
}
