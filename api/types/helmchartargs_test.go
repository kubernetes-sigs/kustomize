// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/types"
)

func TestAsHelmArgs(t *testing.T) {
	p := types.HelmChart{
		Name:                  "chart-name",
		Version:               "1.0.0",
		Repo:                  "https://helm.releases.hashicorp.com",
		ApiVersions:           []string{"foo", "bar"},
		Description:           "desc",
		NameTemplate:          "template",
		SkipTests:             true,
		IncludeCRDs:           true,
		SkipHooks:             true,
		ValuesFile:            "values",
		AdditionalValuesFiles: []string{"values1", "values2"},
		Namespace:             "my-ns",
	}
	require.Equal(t, p.AsHelmArgs("/home/charts"),
		[]string{"template", "/home/charts/chart-name",
			"--namespace", "my-ns",
			"--name-template", "template",
			"--values", "values",
			"-f", "values1", "-f", "values2",
			"--api-versions", "foo", "--api-versions", "bar",
			"--generate-name",
			"--description", "desc",
			"--include-crds",
			"--skip-tests",
			"--no-hooks"})
}
