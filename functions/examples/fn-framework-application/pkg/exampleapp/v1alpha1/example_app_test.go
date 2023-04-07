package v1alpha1_test

import (
	"testing"

	"sigs.k8s.io/kustomize/functions/examples/fn-framework-application/pkg/dispatcher"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/frameworktestutil"
)

func TestExampleApp_GoldenFiles(t *testing.T) {
	c := frameworktestutil.CommandResultsChecker{
		Command: dispatcher.NewCommand,
		// TestDataDirectory:        "testdata/success",
		// UpdateExpectedFromActual: true,
	}
	c.Assert(t)
}
