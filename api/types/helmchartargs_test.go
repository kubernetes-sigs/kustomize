// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/types"
)

func TestAsHelmArgs(t *testing.T) {
	t.Run("use generate-name", func(t *testing.T) {
		p := types.HelmChart{
			Name:                  "chart-name",
			Version:               "1.0.0",
			Repo:                  "https://helm.releases.hashicorp.com",
			ApiVersions:           []string{"foo", "bar"},
			KubeVersion:           "1.27",
			NameTemplate:          "template",
			SkipTests:             true,
			IncludeCRDs:           true,
			SkipHooks:             true,
			ValuesFile:            "values",
			AdditionalValuesFiles: []string{"values1", "values2"},
			Namespace:             "my-ns",
		}
		require.Equal(t, p.AsHelmArgs("/home/charts"),
			[]string{"template", "--generate-name",
				"/home/charts/chart-name",
				"--namespace", "my-ns",
				"--name-template", "template",
				"-f", "values",
				"-f", "values1", "-f", "values2",
				"--api-versions", "foo", "--api-versions", "bar",
				"--kube-version", "1.27",
				"--include-crds",
				"--skip-tests",
				"--no-hooks"})
	})

	t.Run("use release-name", func(t *testing.T) {
		p := types.HelmChart{
			Name:                  "chart-name",
			Version:               "1.0.0",
			Repo:                  "https://helm.releases.hashicorp.com",
			ApiVersions:           []string{"foo", "bar"},
			NameTemplate:          "template",
			ValuesFile:            "values",
			AdditionalValuesFiles: []string{"values1", "values2"},
			Namespace:             "my-ns",
			ReleaseName:           "test",
		}
		require.Equal(t, p.AsHelmArgs("/home/charts"),
			[]string{"template", "test", "/home/charts/chart-name",
				"--namespace", "my-ns",
				"--name-template", "template",
				"-f", "values",
				"-f", "values1", "-f", "values2",
				"--api-versions", "foo", "--api-versions", "bar"})
	})

	t.Run("use helm-debug", func(t *testing.T) {
		p := types.HelmChart{
			Name:                  "chart-name",
			Version:               "1.0.0",
			Repo:                  "https://helm.releases.hashicorp.com",
			ValuesFile:            "values",
			AdditionalValuesFiles: []string{"values1", "values2"},
			Debug:                 true,
		}
		require.Equal(t, p.AsHelmArgs("/home/charts"),
			[]string{"template", "--generate-name", "/home/charts/chart-name",
				"-f", "values",
				"-f", "values1",
				"-f", "values2",
				"--debug"})
	})

	t.Run("use helm-devel", func(t *testing.T) {
		// We first test that the devel flag is only appended when specified
		p := types.HelmChart{
			Name:                  "chart-name",
			Version:               "1.0.0",
			Repo:                  "https://helm.releases.hashicorp.com",
			ValuesFile:            "values",
			AdditionalValuesFiles: []string{"values1", "values2"},
		}
		require.Equal(t, p.AsHelmArgs("/home/charts"),
			[]string{"template", "--generate-name", "/home/charts/chart-name",
				"-f", "values",
				"-f", "values1",
				"-f", "values2"})

		p.Devel = true
		require.Equal(t, p.AsHelmArgs("/home/charts"),
			[]string{"template", "--generate-name", "/home/charts/chart-name",
				"-f", "values",
				"-f", "values1",
				"-f", "values2",
				"--devel"})
	})
}

// Regression test for https://github.com/kubernetes-sigs/kustomize/issues/4593.
// HelmChartArgs.ReleaseNamespace was not copied into HelmChart.Namespace by
// makeHelmChartFromHca, so kustomizations that used the deprecated
// helmChartInflationGenerator with releaseNamespace silently rendered resources
// into the "default" namespace.
func TestSplitHelmParametersPropagatesReleaseNamespace(t *testing.T) {
	args := []types.HelmChartArgs{{
		ChartName:        "nats",
		ChartVersion:     "v0.13.1",
		ReleaseName:      "nats",
		ReleaseNamespace: "custom-ns",
	}}

	charts, _ := types.SplitHelmParameters(args)

	require.Len(t, charts, 1)
	require.Equal(t, "custom-ns", charts[0].Namespace,
		"ReleaseNamespace from the deprecated HelmChartArgs must map onto HelmChart.Namespace")
}
