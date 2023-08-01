// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localize_test

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	loctest "sigs.k8s.io/kustomize/api/testutils/localizertest"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/localize"
)

const (
	// file built via:
	// flux push artifact oci://ghcr.io/frenchben/kustomize-manifest:latest --path examples/oci-test --source="git@github.com:kubernetes-sig/kustomize.git"
	testOciArtifact = "oci://ghcr.io/frenchben/kustomize-manifest"
	deployment      = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.14.2
        ports:
        - containerPort: 80
`
)

func TestScopeFlag(t *testing.T) {
	kustomizations := map[string]string{
		filepath.Join("target", "kustomization.yaml"): fmt.Sprintf(`resources:
- %s
`, filepath.Join("..", "base")),
		filepath.Join("base", "kustomization.yaml"): `resources:
- deployment.yaml
`,
		filepath.Join("base", "deployment.yaml"): deployment,
	}
	expected, actual, testDir := loctest.PrepareFs(t, []string{
		"target",
		"base",
	}, kustomizations)

	cmd := localize.NewCmdLocalize(actual)
	require.NoError(t, cmd.Flags().Set("scope", testDir.String()))
	err := cmd.RunE(cmd, []string{
		testDir.Join("target"),
		testDir.Join("dst"),
	})
	require.NoError(t, err)

	loctest.SetupDir(t, expected, testDir.Join("dst"), kustomizations)
	loctest.CheckFs(t, testDir.String(), expected, actual)
}

func TestOptionalArgs(t *testing.T) {
	for name, args := range map[string][]string{
		"no_target": {},
		"no_dst":    {"."},
	} {
		t.Run(name, func(t *testing.T) {
			kust := map[string]string{
				"kustomization.yaml": `resources:
- deployment.yaml
`,
				"deployment.yaml": deployment,
			}
			expected, actual, testDir := loctest.PrepareFs(t, []string{
				"target",
			}, nil)
			target := testDir.Join("target")
			loctest.SetupDir(t, actual, target, kust)
			loctest.SetWorkingDir(t, target)

			buffy := new(bytes.Buffer)
			log.SetOutput(buffy)
			defer func() {
				log.SetOutput(os.Stderr)
			}()
			cmd := localize.NewCmdLocalize(actual)
			err := cmd.RunE(cmd, args)
			require.NoError(t, err)

			loctest.SetupDir(t, expected, target, kust)
			dst := filepath.Join(target, "localized-target")
			loctest.SetupDir(t, expected, dst, kust)
			loctest.CheckFs(t, testDir.String(), expected, actual)

			successMsg := fmt.Sprintf(`SUCCESS: localized "." to directory %s
`, dst)
			require.Contains(t, buffy.String(), successMsg)
		})
	}
}

func TestOutput(t *testing.T) {
	kustomization := map[string]string{
		"kustomization.yaml": `namePrefix: test-
`,
	}
	expected, actual, target := loctest.PrepareFs(t, nil, kustomization)

	buffy := new(bytes.Buffer)
	log.SetOutput(buffy)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	cmd := localize.NewCmdLocalize(actual)
	err := cmd.RunE(cmd, []string{
		target.String(),
		target.Join("dst"),
	})
	require.NoError(t, err)

	loctest.SetupDir(t, expected, target.Join("dst"), kustomization)
	// loctest.CheckFs(t, target.String(), expected, actual)

	successMsg := fmt.Sprintf("SUCCESS: localized \"%s\" to directory %s", target.String(), target.Join("dst"))
	require.Contains(t, buffy.String(), successMsg)
}

func TestAlpha(t *testing.T) {
	_, actual, _ := loctest.PrepareFs(t, nil, map[string]string{
		"kustomization.yaml": `namePrefix: test-`,
	})

	cmd := localize.NewCmdLocalize(actual)
	require.Contains(t, cmd.Short, "[Alpha]")
	require.Contains(t, cmd.Long, "[Alpha]")
}

func TestTooManyArgs(t *testing.T) {
	_, actual, target := loctest.PrepareFs(t, nil, map[string]string{
		"kustomization.yaml": `namePrefix: test-`,
	})

	cmd := localize.NewCmdLocalize(actual)
	err := cmd.Args(cmd, []string{
		target.String(),
		target.Join("dst"),
		target.String(),
	})
	require.EqualError(t, err, "accepts at most 2 arg(s), received 3")
}

func TestGenericArtifactPull(t *testing.T) {
	// Create mapping of OCI image content
	kustomizations := map[string]string{
		"kustomization.yaml": `resources:
- deployment.yaml
`,
		"deployment.yaml": deployment,
	}
	expected, actual, target := loctest.PrepareFs(t, nil, nil)
	cmd := localize.NewCmdLocalize(actual)

	// Create output buffer
	buffy := new(bytes.Buffer)
	log.SetOutput(buffy)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	// Run the localize command with our OCI artifact
	err := cmd.RunE(cmd, []string{
		testOciArtifact,
		target.Join("dst"),
	})
	require.NoError(t, err)

	// Verify that our OCI artifact extract matches the expected "kustomizations" file content
	loctest.SetupDir(t, expected, target.Join("dst"), kustomizations)
	loctest.CheckFs(t, target.String(), expected, actual)

	successMsg := fmt.Sprintf("SUCCESS: localized \"%s\" to directory %s", testOciArtifact, target.Join("dst"))
	require.Contains(t, buffy.String(), successMsg)
}
