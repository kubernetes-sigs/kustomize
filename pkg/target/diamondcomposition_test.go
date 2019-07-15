// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package target_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/v3/pkg/kusttest"
	plugins_test "sigs.k8s.io/kustomize/v3/pkg/plugins/test"
)

const patchAddProbe = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
spec:
  template:
    spec:
      containers:
      - name: my-deployment
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
`

const container = `{ "image": "my-image", "livenessProbe": { "httpGet" : {"path": "/healthz", "port": 8080 } }, "name": "my-deployment"}`

const patchJsonAddProbe = `[{"op": "replace", "path": "/spec/template/spec/containers/0", "value": ` +
	container + `}]`

const patchDnsPolicy = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
spec:
  template:
    spec:
      dnsPolicy: ClusterFirst
`
const patchJsonDnsPolicy = `[{"op": "add", "path": "/spec/template/spec/dnsPolicy", "value": "ClusterFirst"}]`

const patchRestartPolicy = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
spec:
  template:
    spec:
      restartPolicy: Always
`
const patchJsonRestartPolicy = `[{"op": "add", "path": "/spec/template/spec/restartPolicy", "value": "Always"}]`

func writeDeploymentBase(th *kusttest_test.KustTestHarness) {
	th.WriteK("/app/base", `
resources:
- deployment.yaml
`)

	th.WriteF("/app/base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
spec:
  template:
    spec:
      dnsPolicy: "None"
      containers:
      - name: my-deployment
        image: my-image
`)
}

func writeProbeOverlay(th *kusttest_test.KustTestHarness) {
	th.WriteK("/app/probe", `
resources:
- ../base
patchesStrategicMerge:
- dep-patch.yaml
`)
	th.WriteF("/app/probe/dep-patch.yaml", patchAddProbe)
}

func writeDNSOverlay(th *kusttest_test.KustTestHarness) {
	th.WriteK("/app/dns", `
resources:
- ../base
patchesStrategicMerge:
- dep-patch.yaml
`)
	th.WriteF("/app/dns/dep-patch.yaml", patchDnsPolicy)
}

func writeRestartOverlay(th *kusttest_test.KustTestHarness) {
	th.WriteK("/app/restart", `
resources:
- ../base
patchesStrategicMerge:
- dep-patch.yaml
`)
	th.WriteF("/app/restart/dep-patch.yaml", patchRestartPolicy)
}

// Here's a composite kustomization, that combines multiple overlays
// (replicas, dns and metadata) which patch the same base resource.
//
// The base resource is a deployment and the overlays patch aspects
// of it, without using any of the `namePrefix`, `nameSuffix` or `namespace`
// kustomization keywords.
//
//            composite
//          /     |     \
//       probe   dns  restart
//          \     |     /
//              base
//
func TestIssue1251_CompositeDiamond_Failure(t *testing.T) {
	th := kusttest_test.NewKustTestHarness(t, "/app/composite")
	writeDeploymentBase(th)
	writeProbeOverlay(th)
	writeDNSOverlay(th)
	writeRestartOverlay(th)

	th.WriteK("/app/composite", `
resources:
- ../probe
- ../dns
- ../restart
`)

	_, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err == nil {
		t.Fatalf("Expected resource accumulation error")
	}
	if !strings.Contains(
		err.Error(), "already registered id: apps_v1_Deployment|~X|my-deployment") {
		t.Fatalf("Unexpected err: %v", err)
	}
}

const expectedPatchedDeployment = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
spec:
  template:
    spec:
      containers:
      - image: my-image
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
        name: my-deployment
      dnsPolicy: ClusterFirst
      restartPolicy: Always
`

// This test reuses some methods from TestIssue1251_CompositeDiamond,
// but overwrites the kustomization files in the overlays.
func TestIssue1251_Patches_Overlayed(t *testing.T) {
	th := kusttest_test.NewKustTestHarness(t, "/app/restart")
	writeDeploymentBase(th)

	// probe overlays base.
	writeProbeOverlay(th)

	// dns overlays probe.
	writeDNSOverlay(th)
	th.WriteK("/app/dns", `
resources:
- ../probe
patchesStrategicMerge:
- dep-patch.yaml
`)

	// restart overlays dns.
	writeRestartOverlay(th)
	th.WriteK("/app/restart", `
resources:
- ../dns
patchesStrategicMerge:
- dep-patch.yaml
`)

	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, expectedPatchedDeployment)
}

func TestIssue1251_Patches_Local(t *testing.T) {
	th := kusttest_test.NewKustTestHarness(t, "/app/composite")
	writeDeploymentBase(th)

	th.WriteK("/app/composite", `
resources:
- ../base
patchesStrategicMerge:
- patchAddProbe.yaml
- patchDnsPolicy.yaml
- patchRestartPolicy.yaml
`)
	th.WriteF("/app/composite/patchRestartPolicy.yaml", patchRestartPolicy)
	th.WriteF("/app/composite/patchDnsPolicy.yaml", patchDnsPolicy)
	th.WriteF("/app/composite/patchAddProbe.yaml", patchAddProbe)

	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, expectedPatchedDeployment)
}

func definePatchDirStructure(th *kusttest_test.KustTestHarness) {
	writeDeploymentBase(th)

	th.WriteF("/app/patches/patchRestartPolicy.yaml", patchRestartPolicy)
	th.WriteF("/app/patches/patchDnsPolicy.yaml", patchDnsPolicy)
	th.WriteF("/app/patches/patchAddProbe.yaml", patchAddProbe)
}

// Fails due to file load restrictor.
func TestIssue1251_Patches_ProdVsDev_Failure(t *testing.T) {
	th := kusttest_test.NewKustTestPluginHarness(t, "/app/prod")
	definePatchDirStructure(th)

	th.WriteK("/app/prod", `
resources:
- ../base
patchesStrategicMerge:
- ../patches/patchAddProbe.yaml
- ../patches/patchDnsPolicy.yaml
`)

	_, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(
		err.Error(),
		"security; file '/app/patches/patchAddProbe.yaml' is not in or below '/app/prod'") {
		t.Fatalf("unexpected error: %v", err)
	}
}

const prodDevMergeResult1 = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
spec:
  template:
    spec:
      containers:
      - image: my-image
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
        name: my-deployment
      dnsPolicy: ClusterFirst
`

const prodDevMergeResult2 = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
spec:
  template:
    spec:
      containers:
      - image: my-image
        name: my-deployment
      dnsPolicy: ClusterFirst
      restartPolicy: Always
`

// This test does what
//    TestIssue1251_Patches_ProdVsDev_Failure
// failed to do, because this test does the equivalent
// os specifying `--load_restrictor none` on the build.
//
// This allows the use patch files located outside the
// kustomization root, and not in a kustomization
// themselves.
//
// Doing so means the kustomization using them is no
// longer relocatable, not addressible via a git URL,
// and not git clonable. It's no longer self-contained.
//
// Likewise suppressing load restrictions happens for
// the entire build (i.e. everything can reach outside
// the kustomization root), opening the user to whatever
// threat the load restrictor was meant to address.
func TestIssue1251_Patches_ProdVsDev(t *testing.T) {
	th := kusttest_test.NewKustTestNoLoadRestrictorHarness(t, "/app/prod")
	definePatchDirStructure(th)

	th.WriteK("/app/prod", `
resources:
- ../base
patchesStrategicMerge:
- ../patches/patchAddProbe.yaml
- ../patches/patchDnsPolicy.yaml
`)
	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	th.AssertActualEqualsExpected(m, prodDevMergeResult1)

	th = kusttest_test.NewKustTestNoLoadRestrictorHarness(t, "/app/dev")
	definePatchDirStructure(th)

	th.WriteK("/app/dev", `
resources:
- ../base
patchesStrategicMerge:
- ../patches/patchDnsPolicy.yaml
- ../patches/patchRestartPolicy.yaml
`)

	m, err = th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, prodDevMergeResult2)
}

func TestIssue1251_Plugins_ProdVsDev(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchJson6902Transformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app/prod")
	defineTransformerDirStructure(th)
	th.WriteK("/app/prod", `
resources:
- ../base
transformers:
- ../patches/addProbe
- ../patches/addDnsPolicy
`)

	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, prodDevMergeResult1)

	th = kusttest_test.NewKustTestPluginHarness(t, "/app/dev")
	defineTransformerDirStructure(th)
	th.WriteK("/app/dev", `
resources:
- ../base
transformers:
- ../patches/addRestartPolicy
- ../patches/addDnsPolicy
`)

	m, err = th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, prodDevMergeResult2)
}

func TestIssue1251_Plugins_Local(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchJson6902Transformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app/composite")
	writeDeploymentBase(th)

	writeJsonTransformerPluginConfig(
		th, "/app/composite", "addDnsPolicy", patchJsonDnsPolicy)
	writeJsonTransformerPluginConfig(
		th, "/app/composite", "addRestartPolicy", patchJsonRestartPolicy)
	writeJsonTransformerPluginConfig(
		th, "/app/composite", "addProbe", patchJsonAddProbe)

	th.WriteK("/app/composite", `
resources:
- ../base
transformers:
- addDnsPolicyConfig.yaml
- addRestartPolicyConfig.yaml
- addProbeConfig.yaml
`)
	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, expectedPatchedDeployment)
}

func writeJsonTransformerPluginConfig(
	th *kusttest_test.KustTestHarness, path, name, patch string) {
	th.WriteF(filepath.Join(path, name+"Config.yaml"),
		fmt.Sprintf(`
apiVersion: builtin
kind: PatchJson6902Transformer
metadata:
  name: %s
target:
  group: apps
  version: v1
  kind: Deployment
  name: my-deployment
jsonOp: '%s'
`, name, patch))
}

// Remote in the sense that they are bundled in a different kustomization.
func TestIssue1251_Plugins_Bundled(t *testing.T) {
	tc := plugins_test.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "PatchJson6902Transformer")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app/composite")
	writeDeploymentBase(th)

	th.WriteK("/app/patches", `
resources:
- addDnsPolicyConfig.yaml
- addRestartPolicyConfig.yaml
- addProbeConfig.yaml
`)
	writeJsonTransformerPluginConfig(
		th, "/app/patches", "addDnsPolicy", patchJsonDnsPolicy)
	writeJsonTransformerPluginConfig(
		th, "/app/patches", "addRestartPolicy", patchJsonRestartPolicy)
	writeJsonTransformerPluginConfig(
		th, "/app/patches", "addProbe", patchJsonAddProbe)

	th.WriteK("/app/composite", `
resources:
- ../base
transformers:
- ../patches
`)
	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, expectedPatchedDeployment)
}

func defineTransformerDirStructure(th *kusttest_test.KustTestHarness) {
	writeDeploymentBase(th)

	th.WriteK("/app/patches/addDnsPolicy", `
resources:
- addDnsPolicyConfig.yaml
`)
	writeJsonTransformerPluginConfig(
		th, "/app/patches/addDnsPolicy", "addDnsPolicy", patchJsonDnsPolicy)

	th.WriteK("/app/patches/addRestartPolicy", `
resources:
- addRestartPolicyConfig.yaml
`)
	writeJsonTransformerPluginConfig(
		th, "/app/patches/addRestartPolicy", "addRestartPolicy", patchJsonRestartPolicy)

	th.WriteK("/app/patches/addProbe", `
resources:
- addProbeConfig.yaml
`)
	writeJsonTransformerPluginConfig(
		th, "/app/patches/addProbe", "addProbe", patchJsonAddProbe)
}
