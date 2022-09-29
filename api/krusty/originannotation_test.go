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

func TestAnnoOriginLocalFiles(t *testing.T) {
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
buildMetadata: [originAnnotations]
`)
	options := th.MakeDefaultOptions()
	m := th.Run(".", options)
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Service
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: service.yaml
  name: myService
spec:
  ports:
  - port: 7002
`)
}

func TestAnnoOriginLocalFilesWithOverlay(t *testing.T) {
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
buildMetadata: [originAnnotations]
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
    config.kubernetes.io/origin: |
      path: ../base/namespace.yaml
  name: myNs
---
apiVersion: v1
kind: Role
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: ../base/role.yaml
  name: p-b-myRole
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: ../base/service.yaml
  name: p-b-myService
---
apiVersion: v1
kind: Deployment
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: ../base/deployment.yaml
  name: p-b-myDep
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: service.yaml
  name: p-myService2
---
apiVersion: v1
kind: Namespace
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: namespace.yaml
  name: myNs2
`)
}

// This is a copy of TestGeneratorBasics in configmaps_test.go,
// except that we've enabled the addAnnoOrigin option.
func TestGeneratorWithAnnoOrigin(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
namePrefix: blah-
configMapGenerator:
- name: bob
  literals:
  - fruit=apple
  - vegetable=broccoli
  envs:
  - foo.env
  env: bar.env
  files:
  - passphrase=phrase.dat
  - forces.txt
- name: json
  literals:
  - 'v2=[{"path": "var/druid/segment-cache"}]'
  - >- 
    druid_segmentCache_locations=[{"path": 
    "var/druid/segment-cache", 
    "maxSize": 32000000000, 
    "freeSpacePercent": 1.0}]
secretGenerator:
- name: bob
  literals:
  - fruit=apple
  - vegetable=broccoli
  envs:
  - foo.env
  files:
  - passphrase=phrase.dat
  - forces.txt
  env: bar.env
buildMetadata: [originAnnotations]
`)
	th.WriteF("foo.env", `
MOUNTAIN=everest
OCEAN=pacific
`)
	th.WriteF("bar.env", `
BIRD=falcon
`)
	th.WriteF("phrase.dat", `
Life is short.
But the years are long.
Not while the evil days come not.
`)
	th.WriteF("forces.txt", `
gravitational
electromagnetic
strong nuclear
weak nuclear
`)
	opts := th.MakeDefaultOptions()
	m := th.Run(".", opts)
	th.AssertActualEqualsExpected(
		m, `
apiVersion: v1
data:
  BIRD: falcon
  MOUNTAIN: everest
  OCEAN: pacific
  forces.txt: |2

    gravitational
    electromagnetic
    strong nuclear
    weak nuclear
  fruit: apple
  passphrase: |2

    Life is short.
    But the years are long.
    Not while the evil days come not.
  vegetable: broccoli
kind: ConfigMap
metadata:
  annotations:
    config.kubernetes.io/origin: |
      configuredIn: kustomization.yaml
      configuredBy:
        apiVersion: builtin
        kind: ConfigMapGenerator
  name: blah-bob-g9df72cd5b
---
apiVersion: v1
data:
  druid_segmentCache_locations: '[{"path":  "var/druid/segment-cache",  "maxSize":
    32000000000,  "freeSpacePercent": 1.0}]'
  v2: '[{"path": "var/druid/segment-cache"}]'
kind: ConfigMap
metadata:
  annotations:
    config.kubernetes.io/origin: |
      configuredIn: kustomization.yaml
      configuredBy:
        apiVersion: builtin
        kind: ConfigMapGenerator
  name: blah-json-5298bc8g99
---
apiVersion: v1
data:
  BIRD: ZmFsY29u
  MOUNTAIN: ZXZlcmVzdA==
  OCEAN: cGFjaWZpYw==
  forces.txt: |
    CmdyYXZpdGF0aW9uYWwKZWxlY3Ryb21hZ25ldGljCnN0cm9uZyBudWNsZWFyCndlYWsgbn
    VjbGVhcgo=
  fruit: YXBwbGU=
  passphrase: |
    CkxpZmUgaXMgc2hvcnQuCkJ1dCB0aGUgeWVhcnMgYXJlIGxvbmcuCk5vdCB3aGlsZSB0aG
    UgZXZpbCBkYXlzIGNvbWUgbm90Lgo=
  vegetable: YnJvY2NvbGk=
kind: Secret
metadata:
  annotations:
    config.kubernetes.io/origin: |
      configuredIn: kustomization.yaml
      configuredBy:
        apiVersion: builtin
        kind: SecretGenerator
  name: blah-bob-58g62h555c
type: Opaque
`)
}

func TestAnnoOriginLocalBuiltinGenerator(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- service.yaml
configMapGenerator:
- name: bob
  literals:
  - fruit=Indian Gooseberry
  - year=2020
  - crisis=true
buildMetadata: [originAnnotations]

`)
	th.WriteF("service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: demo
spec:
  clusterIP: None
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(
		m, `
apiVersion: v1
kind: Service
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: service.yaml
  name: demo
spec:
  clusterIP: None
---
apiVersion: v1
data:
  crisis: "true"
  fruit: Indian Gooseberry
  year: "2020"
kind: ConfigMap
metadata:
  annotations:
    config.kubernetes.io/origin: |
      configuredIn: kustomization.yaml
      configuredBy:
        apiVersion: builtin
        kind: ConfigMapGenerator
  name: bob-79t79mt227
`)
}

func TestAnnoOriginConfigMapGeneratorMerge(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
configMapGenerator:
- name: bob
  literals:
  - fruit=Indian Gooseberry
  - year=2020
  - crisis=true
`)
	th.WriteK("overlay", `
resources:
- ../base
configMapGenerator:
- name: bob
  behavior: merge
  literals:
  - month=12
buildMetadata: [originAnnotations]
`)
	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `apiVersion: v1
data:
  crisis: "true"
  fruit: Indian Gooseberry
  month: "12"
  year: "2020"
kind: ConfigMap
metadata:
  annotations:
    config.kubernetes.io/origin: |
      configuredIn: ../base/kustomization.yaml
      configuredBy:
        apiVersion: builtin
        kind: ConfigMapGenerator
  name: bob-bk46gm59c6
`)
}

func TestAnnoOriginConfigMapGeneratorReplace(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
configMapGenerator:
- name: bob
  literals:
  - fruit=Indian Gooseberry
  - year=2020
  - crisis=true
`)
	th.WriteK("overlay", `
resources:
- ../base
configMapGenerator:
- name: bob
  behavior: replace
  literals:
  - month=12
buildMetadata: [originAnnotations]
`)
	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `apiVersion: v1
data:
  month: "12"
kind: ConfigMap
metadata:
  annotations:
    config.kubernetes.io/origin: |
      configuredIn: kustomization.yaml
      configuredBy:
        apiVersion: builtin
        kind: ConfigMapGenerator
  name: bob-f8t5fhtbhc
`)
}

func TestAnnoOriginCustomExecGenerator(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()

	th := kusttest_test.MakeHarnessWithFs(t, fSys)
	o := th.MakeOptionsPluginsEnabled()
	o.PluginConfig.FnpLoadingOptions.EnableExec = true

	tmpDir, err := filesys.NewTmpConfirmedDir()
	assert.NoError(t, err)
	th.WriteK(tmpDir.String(), `
resources:
- short_secret.yaml
generators:
- gener.yaml
buildMetadata: [originAnnotations]
`)

	// Create some additional resource just to make sure everything is added
	th.WriteF(filepath.Join(tmpDir.String(), "short_secret.yaml"),
		`
apiVersion: v1
kind: Secret
metadata:
  labels:
    airshipit.org/ephemeral-user-data: "true"
  name: node1-bmc-secret
type: Opaque
stringData:
  userData: |
    bootcmd:
    - mkdir /mnt/vda
`)
	th.WriteF(filepath.Join(tmpDir.String(), "generateDeployment.sh"), generateDeploymentDotSh)

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
	assert.Equal(t, `apiVersion: v1
kind: Secret
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: short_secret.yaml
  labels:
    airshipit.org/ephemeral-user-data: "true"
  name: node1-bmc-secret
stringData:
  userData: |
    bootcmd:
    - mkdir /mnt/vda
type: Opaque
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    config.kubernetes.io/origin: |
      configuredIn: gener.yaml
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

func TestAnnoOriginCustomInlineExecGenerator(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()

	th := kusttest_test.MakeHarnessWithFs(t, fSys)
	o := th.MakeOptionsPluginsEnabled()
	o.PluginConfig.FnpLoadingOptions.EnableExec = true

	tmpDir, err := filesys.NewTmpConfirmedDir()
	assert.NoError(t, err)
	th.WriteK(tmpDir.String(), `
resources:
- short_secret.yaml
generators:
- |-
  kind: executable
  metadata:
    name: demo
    annotations:
      config.kubernetes.io/function: |
        exec:
          path: ./generateDeployment.sh
  spec:
buildMetadata: [originAnnotations]
`)

	// Create some additional resource just to make sure everything is added
	th.WriteF(filepath.Join(tmpDir.String(), "short_secret.yaml"),
		`
apiVersion: v1
kind: Secret
metadata:
  labels:
    airshipit.org/ephemeral-user-data: "true"
  name: node1-bmc-secret
type: Opaque
stringData:
  userData: |
    bootcmd:
    - mkdir /mnt/vda
`)
	th.WriteF(filepath.Join(tmpDir.String(), "generateDeployment.sh"), generateDeploymentDotSh)
	assert.NoError(t, os.Chmod(filepath.Join(tmpDir.String(), "generateDeployment.sh"), 0777))
	m := th.Run(tmpDir.String(), o)
	assert.NoError(t, err)
	yml, err := m.AsYaml()
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: v1
kind: Secret
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: short_secret.yaml
  labels:
    airshipit.org/ephemeral-user-data: "true"
  name: node1-bmc-secret
stringData:
  userData: |
    bootcmd:
    - mkdir /mnt/vda
type: Opaque
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    config.kubernetes.io/origin: |
      configuredIn: kustomization.yaml
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

func TestAnnoOriginCustomExecGeneratorWithOverlay(t *testing.T) {
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
resources:
- short_secret.yaml
generators:
- gener.yaml
`)
	th.WriteK(prod, `
resources:
- ../base
buildMetadata: [originAnnotations]
`)
	th.WriteF(filepath.Join(base, "short_secret.yaml"),
		`
apiVersion: v1
kind: Secret
metadata:
  labels:
    airshipit.org/ephemeral-user-data: "true"
  name: node1-bmc-secret
type: Opaque
stringData:
  userData: |
    bootcmd:
    - mkdir /mnt/vda
`)
	th.WriteF(filepath.Join(base, "generateDeployment.sh"), generateDeploymentDotSh)

	assert.NoError(t, os.Chmod(filepath.Join(base, "generateDeployment.sh"), 0777))
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

	m := th.Run(prod, o)
	assert.NoError(t, err)
	yml, err := m.AsYaml()
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: v1
kind: Secret
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: ../base/short_secret.yaml
  labels:
    airshipit.org/ephemeral-user-data: "true"
  name: node1-bmc-secret
stringData:
  userData: |
    bootcmd:
    - mkdir /mnt/vda
type: Opaque
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    config.kubernetes.io/origin: |
      configuredIn: ../base/gener.yaml
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

func TestAnnoOriginCustomInlineExecGeneratorWithOverlay(t *testing.T) {
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
resources:
- short_secret.yaml
generators:
- |-
  kind: executable
  metadata:
    name: demo
    annotations:
      config.kubernetes.io/function: |
        exec:
          path: ./generateDeployment.sh
  spec:
`)
	th.WriteK(prod, `
resources:
- ../base
buildMetadata: [originAnnotations]
`)
	th.WriteF(filepath.Join(base, "short_secret.yaml"),
		`
apiVersion: v1
kind: Secret
metadata:
  labels:
    airshipit.org/ephemeral-user-data: "true"
  name: node1-bmc-secret
type: Opaque
stringData:
  userData: |
    bootcmd:
    - mkdir /mnt/vda
`)
	th.WriteF(filepath.Join(base, "generateDeployment.sh"), generateDeploymentDotSh)
	assert.NoError(t, os.Chmod(filepath.Join(base, "generateDeployment.sh"), 0777))
	m := th.Run(prod, o)
	assert.NoError(t, err)
	yml, err := m.AsYaml()
	assert.NoError(t, err)
	assert.Equal(t, `apiVersion: v1
kind: Secret
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: ../base/short_secret.yaml
  labels:
    airshipit.org/ephemeral-user-data: "true"
  name: node1-bmc-secret
stringData:
  userData: |
    bootcmd:
    - mkdir /mnt/vda
type: Opaque
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    config.kubernetes.io/origin: |
      configuredIn: ../base/kustomization.yaml
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

func TestAnnoOriginRemoteBuiltinGenerator(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()
	b := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	tmpDir, err := filesys.NewTmpConfirmedDir()
	assert.NoError(t, err)
	assert.NoError(t, fSys.WriteFile(filepath.Join(tmpDir.String(), "kustomization.yaml"), []byte(`
resources:
- github.com/kubernetes-sigs/kustomize/examples/ldap/base/?ref=v1.0.6
buildMetadata: [originAnnotations]
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
	assert.Contains(t, string(yml), `kind: ConfigMap
metadata:
  annotations:
    config.kubernetes.io/origin: |
      repo: https://github.com/kubernetes-sigs/kustomize.git
      ref: v1.0.6
      configuredIn: examples/ldap/base/kustomization.yaml
      configuredBy:
        apiVersion: builtin
        kind: ConfigMapGenerator
  name: ldap-configmap-4d7m6k5b42`)
	assert.NoError(t, fSys.RemoveAll(tmpDir.String()))
}

func TestAnnoOriginInlineBuiltinGenerator(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- service.yaml
generators:
- |-
  apiVersion: builtin
  kind: ConfigMapGenerator
  metadata:
    name: notImportantHere
  name: bob
  literals:
  - fruit=Indian Gooseberry
  - year=2020
  - crisis=true
buildMetadata: [originAnnotations]
`)

	th.WriteF("service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: apple
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Service
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: service.yaml
  name: apple
---
apiVersion: v1
data:
  crisis: "true"
  fruit: Indian Gooseberry
  year: "2020"
kind: ConfigMap
metadata:
  annotations:
    config.kubernetes.io/origin: |
      configuredIn: kustomization.yaml
      configuredBy:
        apiVersion: builtin
        kind: ConfigMapGenerator
        name: notImportantHere
  name: bob-79t79mt227
`)
}

func TestAnnoOriginGeneratorFromFile(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- service.yaml
generators:
- configmap.yaml
buildMetadata: [originAnnotations]
`)
	th.WriteF("configmap.yaml", `
apiVersion: builtin
kind: ConfigMapGenerator
metadata:
  name: notImportantHere
name: bob
literals:
- fruit=Indian Gooseberry
- year=2020
- crisis=true
`)
	th.WriteF("service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: apple
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Service
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: service.yaml
  name: apple
---
apiVersion: v1
data:
  crisis: "true"
  fruit: Indian Gooseberry
  year: "2020"
kind: ConfigMap
metadata:
  annotations:
    config.kubernetes.io/origin: |
      configuredIn: configmap.yaml
      configuredBy:
        apiVersion: builtin
        kind: ConfigMapGenerator
        name: notImportantHere
  name: bob-79t79mt227
`)
}

func TestAnnoOriginBuiltinGeneratorFromFileWithOverlay(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("base", `
resources:
- short_secret.yaml
generators:
- configmap.yaml
`)
	th.WriteF("base/configmap.yaml", `apiVersion: builtin
kind: ConfigMapGenerator
metadata:
  name: notImportantHere
name: bob
literals:
- fruit=Indian Gooseberry
- year=2020
- crisis=true
`)
	th.WriteK("prod", `
resources:
- ../base
buildMetadata: [originAnnotations]
`)
	th.WriteF("base/short_secret.yaml",
		`
apiVersion: v1
kind: Secret
metadata:
  labels:
    airshipit.org/ephemeral-user-data: "true"
  name: node1-bmc-secret
type: Opaque
stringData:
  userData: |
    bootcmd:
    - mkdir /mnt/vda
`)
	m := th.Run("prod", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `apiVersion: v1
kind: Secret
metadata:
  annotations:
    config.kubernetes.io/origin: |
      path: ../base/short_secret.yaml
  labels:
    airshipit.org/ephemeral-user-data: "true"
  name: node1-bmc-secret
stringData:
  userData: |
    bootcmd:
    - mkdir /mnt/vda
type: Opaque
---
apiVersion: v1
data:
  crisis: "true"
  fruit: Indian Gooseberry
  year: "2020"
kind: ConfigMap
metadata:
  annotations:
    config.kubernetes.io/origin: |
      configuredIn: ../base/configmap.yaml
      configuredBy:
        apiVersion: builtin
        kind: ConfigMapGenerator
        name: notImportantHere
  name: bob-79t79mt227
`)
}

func TestAnnoOriginGeneratorInTransformersField(t *testing.T) {
	fSys := filesys.MakeFsOnDisk()

	th := kusttest_test.MakeHarnessWithFs(t, fSys)
	o := th.MakeOptionsPluginsEnabled()
	o.PluginConfig.FnpLoadingOptions.EnableExec = true

	tmpDir, err := filesys.NewTmpConfirmedDir()
	assert.NoError(t, err)
	th.WriteK(tmpDir.String(), `
transformers:
- gener.yaml
buildMetadata: [originAnnotations]
`)

	th.WriteF(filepath.Join(tmpDir.String(), "generateDeployment.sh"), generateDeploymentDotSh)

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
    config.kubernetes.io/origin: |
      configuredIn: gener.yaml
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

func TestAnnoOriginGeneratorInTransformersFieldWithOverlay(t *testing.T) {
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

	th.WriteF(filepath.Join(base, "generateDeployment.sh"), generateDeploymentDotSh)

	assert.NoError(t, os.Chmod(filepath.Join(base, "generateDeployment.sh"), 0777))
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
	th.WriteK(prod, `
resources:
- ../base
nameSuffix: -foo
buildMetadata: [originAnnotations, transformerAnnotations]
`)

	m := th.Run(prod, o)
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
          apiVersion: builtin
          kind: SuffixTransformer
    config.kubernetes.io/origin: |
      configuredIn: ../base/gener.yaml
      configuredBy:
        kind: executable
        name: demo
    tshirt-size: small
  labels:
    app: nginx
  name: nginx-foo
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
