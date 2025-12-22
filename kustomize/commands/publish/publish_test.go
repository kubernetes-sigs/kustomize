// Copyright 2025 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package publish_test

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/konfig"
	loctest "sigs.k8s.io/kustomize/api/testutils/localizertest"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/publish"
)

type FakePassiveClock struct {
	time time.Time
}

func NewFakePassiveClock(t time.Time) *FakePassiveClock {
	return &FakePassiveClock{
		time: t,
	}
}

func (f *FakePassiveClock) Now() time.Time {
	return f.time
}

func (f *FakePassiveClock) Since(ts time.Time) time.Duration {
	return f.time.Sub(ts)
}

// const deployment = `apiVersion: apps/v1
// kind: Deployment
// metadata:
//   name: nginx-deployment
//   labels:
//     app: nginx
// spec:
//   replicas: 3
//   selector:
//     matchLabels:
//       app: nginx
//   template:
//     metadata:
//       labels:
//         app: nginx
//     spec:
//       containers:
//       - name: nginx
//         image: nginx:1.14.2
//         ports:
//         - containerPort: 80
// `

// const kustomization = `apiVersion: kustomize.config.k8s.io/v1beta1
// kind: Kustomization
// resources:
//   - deployment.yaml
// `

// func TestScopeFlag(t *testing.T) {
// 	kustomizations := map[string]string{
// 		filepath.Join("target", "kustomization.yaml"): fmt.Sprintf(`resources:
// - %s
// `, filepath.Join("..", "base")),
// 		filepath.Join("base", "kustomization.yaml"): kustomization,
// 		filepath.Join("base", "deployment.yaml"): deployment,
// 	}
// 	expected, actual, testDir := loctest.PrepareFs(t, []string{
// 		"target",
// 		"base",
// 	}, kustomizations)

// 	cmd := localize.NewCmdLocalize(actual)
// 	require.NoError(t, cmd.Flags().Set("scope", testDir.String()))
// 	err := cmd.RunE(cmd, []string{
// 		testDir.Join("target"),
// 		testDir.Join("dst"),
// 	})
// 	require.NoError(t, err)

// 	loctest.SetupDir(t, expected, testDir.Join("dst"), kustomizations)
// 	loctest.CheckFs(t, testDir.String(), expected, actual)
// }

// func TestNoVerifyFlag(t *testing.T) {
// 	kustomization := map[string]string{
// 		"kustomization.yaml": `resources:
// - deployment.yaml
// `,
// 		"deployment.yaml": deployment,
// 	}
// 	expected, actual, target := loctest.PrepareFs(t, nil, kustomization)

// 	buffy := new(bytes.Buffer)
// 	log.SetOutput(buffy)
// 	defer func() {
// 		log.SetOutput(os.Stderr)
// 	}()
// 	cmd := localize.NewCmdLocalize(actual)
// 	require.NoError(t, cmd.Flags().Set("no-verify", "true"))
// 	err := cmd.RunE(cmd, []string{
// 		target.String(),
// 		target.Join("dst"),
// 	})
// 	require.NoError(t, err)

// 	loctest.SetupDir(t, expected, target.Join("dst"), kustomization)
// 	loctest.CheckFs(t, target.String(), expected, actual)

// 	successMsg := fmt.Sprintf(`SUCCESS: localized "%s" to directory %s
// `, target.String(), target.Join("dst"))
// 	verifyMsg := "VERIFICATION"
// 	require.NotContains(t, buffy.String(), verifyMsg)
// 	require.Contains(t, buffy.String(), successMsg)
// }

// func TestFailingBuildCmd(t *testing.T) {
// 	kustomization := map[string]string{
// 		"kustomization.yaml": helmKustomization,
// 	}
// 	_, actual, target := loctest.PrepareFs(t, nil, kustomization)

// 	buffy := new(bytes.Buffer)
// 	log.SetOutput(buffy)
// 	defer func() {
// 		log.SetOutput(os.Stderr)
// 	}()
// 	cmd := localize.NewCmdLocalize(actual)
// 	err := cmd.RunE(cmd, []string{
// 		target.String(),
// 		target.Join("dst"),
// 	})
// 	require.Error(t, err)

// 	verifyMsg := "If your target directory requires flags to build"
// 	require.Contains(t, buffy.String(), verifyMsg)
// }

// func TestOptionalArgs(t *testing.T) {
// 	for name, args := range map[string][]string{
// 		"no_target": {},
// 		"no_dst":    {"."},
// 	} {
// 		t.Run(name, func(t *testing.T) {
// 			kust := map[string]string{
// 				"kustomization.yaml": `resources:
// - deployment.yaml
// `,
// 				"deployment.yaml": deployment,
// 			}
// 			expected, actual, testDir := loctest.PrepareFs(t, []string{
// 				"target",
// 			}, nil)
// 			target := testDir.Join("target")
// 			loctest.SetupDir(t, actual, target, kust)
// 			loctest.SetWorkingDir(t, target)

// 			buffy := new(bytes.Buffer)
// 			log.SetOutput(buffy)
// 			defer func() {
// 				log.SetOutput(os.Stderr)
// 			}()
// 			cmd := localize.NewCmdLocalize(actual)
// 			err := cmd.RunE(cmd, args)
// 			require.NoError(t, err)

// 			loctest.SetupDir(t, expected, target, kust)
// 			dst := filepath.Join(target, "localized-target")
// 			loctest.SetupDir(t, expected, dst, kust)
// 			loctest.CheckFs(t, testDir.String(), expected, actual)

// 			verifyMsg := "VERIFICATION SUCCESS"
// 			require.Contains(t, buffy.String(), verifyMsg)
// 			successMsg := fmt.Sprintf(`SUCCESS: localized "." to directory %s
// `, dst)
// 			require.Contains(t, buffy.String(), successMsg)
// 		})
// 	}
// }

func TestPublishRequireKustomizationFile(t *testing.T) {
	kustomization := map[string]string{
		"src/README.md": `# NO VALID FILE
`,
	}
	clock := NewFakePassiveClock(time.Date(int(2025), time.July, int(28), int(20), int(56), int(0), int(0), time.UTC))

	_, actual, target := loctest.PrepareFs(t, []string{"src"}, kustomization)
	loctest.SetWorkingDir(t, target.Join("src"))

	buffy := new(bytes.Buffer)
	log.SetOutput(buffy)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	cmd := publish.NewCmdPublish(actual, clock)
	err := cmd.RunE(cmd, []string{
		target.Join("dst"),
	})
	require.Error(t, err, `Missing kustomization file '%s'.\n`, konfig.DefaultKustomizationFileName())
	require.NoDirExistsf(t, target.Join("dst"), "OCI Registry directory created")
	require.Empty(t, buffy.String())
}

func TestPublishAcceptRequireKustomizationFile(t *testing.T) {
	for _, filename := range konfig.RecognizedKustomizationFileNames() {
		t.Run(filename, func(t *testing.T) {
			kustomization := map[string]string{
				filepath.Join("src", filename): `namePrefix: test-
`,
			}
			clock := NewFakePassiveClock(time.Date(int(2025), time.July, int(28), int(20), int(56), int(0), int(0), time.UTC))

			_, actual, target := loctest.PrepareFs(t, []string{"src"}, kustomization)
			loctest.SetWorkingDir(t, target.Join("src"))

			buffy := new(bytes.Buffer)
			log.SetOutput(buffy)
			defer func() {
				log.SetOutput(os.Stderr)
			}()
			cmd := publish.NewCmdPublish(actual, clock)
			err := cmd.RunE(cmd, []string{
				target.Join("dst"),
			})
			require.NoError(t, err)
			require.DirExistsf(t, target.Join("dst"), "OCI Registry directory not created")
			successMsg := fmt.Sprintf(`SUCCESS: published %s:%s@`, target.Join("dst"), "latest")
			require.Contains(t, buffy.String(), successMsg)
		})
	}
}

func TestOutput(t *testing.T) {
	kustomization := map[string]string{
		"src/kustomization.yaml": `namePrefix: test-
`,
	}

	clock := NewFakePassiveClock(time.Date(int(2025), time.July, int(28), int(20), int(56), int(0), int(0), time.UTC))

	expected, actual, target := loctest.PrepareFs(t, []string{"src"}, kustomization)
	loctest.SetWorkingDir(t, target.Join("src"))

	buffy := new(bytes.Buffer)
	log.SetOutput(buffy)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	cmd := publish.NewCmdPublish(actual, clock)
	err := cmd.RunE(cmd, []string{
		target.Join("dst"),
	})
	require.NoError(t, err)

	loctest.SetupDir(t, expected, target.Join("dst"), kustomization)
	loctest.CheckFs(t, target.String(), expected, actual)

	verifyMsg := "VERIFICATION SUCCESS"
	require.Contains(t, buffy.String(), verifyMsg)
	successMsg := fmt.Sprintf(`SUCCESS: copied "%s" to directory %s\n`, target.String(), target.Join("dst"))
	require.Contains(t, buffy.String(), successMsg)
}

// func TestAlpha(t *testing.T) {
// 	_, actual, _ := loctest.PrepareFs(t, nil, map[string]string{
// 		"kustomization.yaml": `namePrefix: test-`,
// 	})

// 	cmd := localize.NewCmdLocalize(actual)
// 	require.Contains(t, cmd.Short, "[Alpha]")
// 	require.Contains(t, cmd.Long, "[Alpha]")
// }

// func TestTooManyArgs(t *testing.T) {
// 	_, actual, target := loctest.PrepareFs(t, nil, map[string]string{
// 		"kustomization.yaml": `namePrefix: test-`,
// 	})

// 	cmd := localize.NewCmdLocalize(actual)
// 	err := cmd.Args(cmd, []string{
// 		target.String(),
// 		target.Join("dst"),
// 		target.String(),
// 	})
// 	require.EqualError(t, err, "accepts at most 2 arg(s), received 3")
// }
