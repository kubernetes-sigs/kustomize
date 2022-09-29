// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/internal/utils"
	"sigs.k8s.io/kustomize/api/krusty"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const generateDeploymentWithOriginDotSh = `#!/bin/sh

cat <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  labels:
    app: nginx
  annotations:
    tshirt-size: small # this injects the resource reservations
    config.kubernetes.io/origin: |
      path: somefile.yaml
spec:
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
        image: nginx
EOF
`

func TestAnnoTransformerBuiltinLocal(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  ports:
  - port: 7002
`)
	th.WriteK(".", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- service.yaml
buildMetadata: [transformerAnnotations]
namePrefix: foo-
`)
	options := th.MakeDefaultOptions()
	m := th.Run(".", options)
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Service
metadata:
  annotations:
    alpha.config.kubernetes.io/transformations: |
      - configuredIn: kustomization.yaml
        configuredBy:
          apiVersion: builtin
          kind: PrefixTransformer
  name: foo-myService
spec:
  ports:
  - port: 7002
`)
}

func TestAnnoOriginAndTransformerBuiltinLocal(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  ports:
  - port: 7002
`)
	th.WriteK(".", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- service.yaml
buildMetadata: [originAnnotations, transformerAnnotations]
namePrefix: foo-
`)
	options := th.MakeDefaultOptions()
	m := th.Run(".", options)
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Service
metadata:
  annotations:
    alpha.config.kubernetes.io/transformations: |
      - configuredIn: kustomization.yaml
        configuredBy:
          apiVersion: builtin
          kind: PrefixTransformer
    config.kubernetes.io/origin: |
      path: service.yaml
  name: foo-myService
spec:
  ports:
  - port: 7002
`)
}

func TestAnnoTransformerLocalFilesWithOverlay(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
namePrefix: b-
resources:
- namespace.yaml
- role.yaml
- service.yaml
- deployment.yaml
`)
	th.WriteF("base/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: myService
`)
	th.WriteF("base/namespace.yaml", `
apiVersion: v1
kind: Namespace
metadata:
  name: myNs
`)
	th.WriteF("base/role.yaml", `
apiVersion: v1
kind: Role
metadata:
  name: myRole
`)
	th.WriteF("base/deployment.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: myDep
`)
	th.WriteK("prod", `
namePrefix: p-
resources:
- ../base
- service.yaml
- namespace.yaml
buildMetadata: [transformerAnnotations]
`)
	th.WriteF("prod/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: myService2
`)
	th.WriteF("prod/namespace.yaml", `
apiVersion: v1
kind: Namespace
metadata:
  name: myNs2
`)
	m := th.Run("prod", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Namespace
metadata:
  annotations:
    alpha.config.kubernetes.io/transformations: |
      - configuredIn: ../base/kustomization.yaml
        configuredBy:
          apiVersion: builtin
          kind: PrefixTransformer
      - configuredIn: kustomization.yaml
        configuredBy:
          apiVersion: builtin
          kind: PrefixTransformer
  name: myNs
---
apiVersion: v1
kind: Role
metadata:
  annotations:
    alpha.config.kubernetes.io/transformations: |
      - configuredIn: ../base/kustomization.yaml
        configuredBy:
          apiVersion: builtin
          kind: PrefixTransformer
      - configuredIn: kustomization.yaml
        configuredBy:
          apiVersion: builtin
          kind: PrefixTransformer
  name: p-b-myRole
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    alpha.config.kubernetes.io/transformations: |
      - configuredIn: ../base/kustomization.yaml
        configuredBy:
          apiVersion: builtin
          kind: PrefixTransformer
      - configuredIn: kustomization.yaml
        configuredBy:
          apiVersion: builtin
          kind: PrefixTransformer
  name: p-b-myService
---
apiVersion: v1
kind: Deployment
metadata:
  annotations:
    alpha.config.kubernetes.io/transformations: |
      - configuredIn: ../base/kustomization.yaml
        configuredBy:
          apiVersion: builtin
          kind: PrefixTransformer
      - configuredIn: kustomization.yaml
        configuredBy:
          apiVersion: builtin
          kind: PrefixTransformer
  name: p-b-myDep
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    alpha.config.kubernetes.io/transformations: |
      - configuredIn: kustomization.yaml
        configuredBy:
          apiVersion: builtin
          kind: PrefixTransformer
  name: p-myService2
---
apiVersion: v1
kind: Namespace
metadata:
  annotations:
    alpha.config.kubernetes.io/transformations: |
      - configuredIn: kustomization.yaml
        configuredBy:
          apiVersion: builtin
          kind: PrefixTransformer
  name: myNs2
`)
}

func TestAnnoOriginRemoteBuiltinTransformer(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()
	b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	tmpDir, err := filesys.NewTmpConfirmedDir()
	assert.NoError(t, err)
	assert.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "kustomization.yaml"), []byte(`
resources:
- github.com/kubernetes-sigs/kustomize/examples/multibases/production/?ref=v1.0.6
buildMetadata: [transformerAnnotations]
`)))
	m, err := b.Run(
		fSys,
		tmpDir.String())
	if utils.IsErrTimeout(err) {
		// Don't fail on timeouts.
		t.SkipNow()
	}
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	yml, err := m.AsYaml()
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: v1
kind: Pod
metadata:
  annotations:
    alpha.config.kubernetes.io/transformations: |
      - repo: https://github.com/kubernetes-sigs/kustomize.git
        ref: v1.0.6
        configuredIn: examples/multibases/production/kustomization.yaml
        configuredBy:
          apiVersion: builtin
          kind: PrefixTransformer
  labels:
    app: myapp
  name: prod-myapp-pod
spec:
  containers:
  - image: nginx:1.7.9
    name: nginx
`, string(yml))
	assert.NoError(t, fSys.RemoveAll(tmpDir.String()))
}

func TestAnnoTransformerBuiltinInline(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteF("resource.yaml", `
apiVersion: apps/v1
kind: ConfigMap
metadata:
  name: whatever
data: {}
`)
	th.WriteK(".", `
resources:
- resource.yaml
transformers:
- |-
  apiVersion: builtin
  kind: NamespaceTransformer
  metadata:
    name: not-important-to-example
    namespace: test
  fieldSpecs:
  - path: metadata/namespace
    create: true
buildMetadata: [transformerAnnotations]
`)

	expected := `
apiVersion: apps/v1
data: {}
kind: ConfigMap
metadata:
  annotations:
    alpha.config.kubernetes.io/transformations: |
      - configuredIn: kustomization.yaml
        configuredBy:
          apiVersion: builtin
          kind: NamespaceTransformer
          name: not-important-to-example
          namespace: test
  name: whatever
  namespace: test
`
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, expected)
}

func TestAnnoOriginCustomInlineTransformer(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()

	th := kusttest_test.MakeHarnessWithFs(t, fSys)
	o := th.MakeOptionsPluginsEnabled()
	o.PluginConfig.FnpLoadingOptions.EnableExec = true

	tmpDir, err := filesys.NewTmpConfirmedDir()
	assert.NoError(t, err)
	th.WriteK(tmpDir.String(), `
transformers:
- |-
  kind: executable
  metadata:
    name: demo
    annotations:
      config.kubernetes.io/function: |
        exec:
          path: ./generateDeployment.sh
  spec:
buildMetadata: [transformerAnnotations]
`)

	// generateDeploymentWithOriginDotSh creates a resource that already has an origin annotation,
	// which will cause kustomize to record the plugin origin data as a transformation
	th.WriteF(filepath.Join(tmpDir.String(), "generateDeployment.sh"), generateDeploymentWithOriginDotSh)

	assert.NoError(t, os.Chmod(filepath.Join(tmpDir.String(), "generateDeployment.sh"), 0777))
	th.WriteF(filepath.Join(tmpDir.String(), "gener.yaml"), `
kind: executable
metadata:
  name: demo
  annotations:
    config.kubernetes.io/function: |
      exec:
        path: ./generateDeployment.sh
spec:
`)

	m := th.Run(tmpDir.String(), o)
	assert.NoError(t, err)
	yml, err := m.AsYaml()
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    alpha.config.kubernetes.io/transformations: |
      - configuredIn: kustomization.yaml
        configuredBy:
          kind: executable
          name: demo
    tshirt-size: small
  labels:
    app: nginx
  name: nginx
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx
        name: nginx
`, string(yml))
	assert.NoError(t, fSys.RemoveAll(tmpDir.String()))
}

func TestAnnoOriginCustomExecTransformerWithOverlay(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()

	th := kusttest_test.MakeHarnessWithFs(t, fSys)
	o := th.MakeOptionsPluginsEnabled()
	o.PluginConfig.FnpLoadingOptions.EnableExec = true

	tmpDir, err := filesys.NewTmpConfirmedDir()
	assert.NoError(t, err)
	base := filepath.Join(tmpDir.String(), "base")
	prod := filepath.Join(tmpDir.String(), "prod")
	assert.NoError(t, fSys.Mkdir(base))
	assert.NoError(t, fSys.Mkdir(prod))
	th.WriteK(base, `
transformers:
- gener.yaml
`)
	th.WriteK(prod, `
resources:
- ../base
buildMetadata: [transformerAnnotations]
`)
	th.WriteF(filepath.Join(base, "gener.yaml"), `
kind: executable
metadata:
  name: demo
  annotations:
    config.kubernetes.io/function: |
      exec:
        path: ./generateDeployment.sh
spec:
`)
	// generateDeploymentWithOriginDotSh creates a resource that already has an origin annotation,
	// which will cause kustomize to record the plugin origin data as a transformation
	th.WriteF(filepath.Join(base, "generateDeployment.sh"), generateDeploymentWithOriginDotSh)
	assert.NoError(t, os.Chmod(filepath.Join(base, "generateDeployment.sh"), 0777))

	m := th.Run(prod, o)
	assert.NoError(t, err)
	yml, err := m.AsYaml()
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    alpha.config.kubernetes.io/transformations: |
      - configuredIn: ../base/gener.yaml
        configuredBy:
          kind: executable
          name: demo
    tshirt-size: small
  labels:
    app: nginx
  name: nginx
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx
        name: nginx
`, string(yml))
	assert.NoError(t, fSys.RemoveAll(tmpDir.String()))
}
