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
}
