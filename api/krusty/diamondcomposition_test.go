// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	"sigs.k8s.io/kustomize/api/types"
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

const patchDNSPolicy = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
spec:
  template:
    spec:
      dnsPolicy: ClusterFirst
`
const patchJsonDNSPolicy = `[{"op": "add", "path": "/spec/template/spec/dnsPolicy", "value": "ClusterFirst"}]`

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

func writeDeploymentBase(th kusttest_test.Harness) {
	th.WriteK("base", `
resources:
- deployment.yaml
`)

	th.WriteF("base/deployment.yaml", `
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

func writeProbeOverlay(th kusttest_test.Harness) {
	th.WriteK("probe", `
resources:
- ../base
patchesStrategicMerge:
- dep-patch.yaml
`)
	th.WriteF("probe/dep-patch.yaml", patchAddProbe)
}

func writeDNSOverlay(th kusttest_test.Harness) {
	th.WriteK("dns", `
resources:
- ../base
patchesStrategicMerge:
- dep-patch.yaml
`)
	th.WriteF("dns/dep-patch.yaml", patchDNSPolicy)
}

func writeRestartOverlay(th kusttest_test.Harness) {
	th.WriteK("restart", `
resources:
- ../base
patchesStrategicMerge:
- dep-patch.yaml
`)
	th.WriteF("restart/dep-patch.yaml", patchRestartPolicy)
}

// Here's a composite kustomization, that combines multiple overlays
// (replicas, dns and metadata) which patch the same base resource.
//
// The base resource is a deployment and the overlays patch aspects
// of it, without using any of the `namePrefix`, `nameSuffix` or `namespace`
// kustomization keywords.
//
//	     composite
//	   /     |     \
//	probe   dns  restart
//	   \     |     /
//	       base
func TestIssue1251_CompositeDiamond_Failure(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeDeploymentBase(th)
	writeProbeOverlay(th)
	writeDNSOverlay(th)
	writeRestartOverlay(th)

	th.WriteK("composite", `
resources:
- ../probe
- ../dns
- ../restart
`)

	err := th.RunWithErr("composite", th.MakeDefaultOptions())
	if err == nil {
		t.Fatalf("Expected resource accumulation error")
	}
	if !strings.Contains(
		err.Error(), "already registered id: Deployment.v1.apps/my-deployment.[noNs]") {
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
	th := kusttest_test.MakeHarness(t)
	writeDeploymentBase(th)

	// probe overlays base.
	writeProbeOverlay(th)

	// dns overlays probe.
	writeDNSOverlay(th)
	th.WriteK("dns", `
resources:
- ../probe
patchesStrategicMerge:
- dep-patch.yaml
`)

	// restart overlays dns.
	writeRestartOverlay(th)
	th.WriteK("restart", `
resources:
- ../dns
patchesStrategicMerge:
- dep-patch.yaml
`)

	m := th.Run("restart", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, expectedPatchedDeployment)
}

func TestIssue1251_Patches_Local(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeDeploymentBase(th)

	th.WriteK("composite", `
resources:
- ../base
patchesStrategicMerge:
- patchAddProbe.yaml
- patchDnsPolicy.yaml
- patchRestartPolicy.yaml
`)
	th.WriteF("composite/patchRestartPolicy.yaml", patchRestartPolicy)
	th.WriteF("composite/patchDnsPolicy.yaml", patchDNSPolicy)
	th.WriteF("composite/patchAddProbe.yaml", patchAddProbe)

	m := th.Run("composite", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, expectedPatchedDeployment)
}

func definePatchDirStructure(th kusttest_test.Harness) {
	writeDeploymentBase(th)

	th.WriteF("patches/patchRestartPolicy.yaml", patchRestartPolicy)
	th.WriteF("patches/patchDnsPolicy.yaml", patchDNSPolicy)
	th.WriteF("patches/patchAddProbe.yaml", patchAddProbe)
}

// Fails due to file load restrictor.
func TestIssue1251_Patches_ProdVsDev_Failure(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	definePatchDirStructure(th)

	th.WriteK("prod", `
resources:
- ../base
patchesStrategicMerge:
- ../patches/patchAddProbe.yaml
- ../patches/patchDnsPolicy.yaml
`)

	err := th.RunWithErr("prod", th.MakeDefaultOptions())
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(
		err.Error(),
		"security; file '/patches/patchAddProbe.yaml' is not in or below '/prod'") {
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
//
//	TestIssue1251_Patches_ProdVsDev_Failure
//
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
	th := kusttest_test.MakeHarness(t)
	definePatchDirStructure(th)

	th.WriteK("prod", `
resources:
- ../base
patchesStrategicMerge:
- ../patches/patchAddProbe.yaml
- ../patches/patchDnsPolicy.yaml
`)
	opts := th.MakeDefaultOptions()
	opts.LoadRestrictions = types.LoadRestrictionsNone

	m := th.Run("prod", opts)
	th.AssertActualEqualsExpected(m, prodDevMergeResult1)

	th = kusttest_test.MakeHarness(t)
	definePatchDirStructure(th)

	th.WriteK("dev", `
resources:
- ../base
patchesStrategicMerge:
- ../patches/patchDnsPolicy.yaml
- ../patches/patchRestartPolicy.yaml
`)

	m = th.Run("dev", opts)
	th.AssertActualEqualsExpected(m, prodDevMergeResult2)
}

func TestIssue1251_Plugins_ProdVsDev(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchJson6902Transformer")
	defer th.Reset()

	defineTransformerDirStructure(th)
	th.WriteK("prod", `
resources:
- ../base
transformers:
- ../patches/addProbe
- ../patches/addDnsPolicy
`)

	m := th.Run("prod", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, prodDevMergeResult1)

	defineTransformerDirStructure(th)
	th.WriteK("dev", `
resources:
- ../base
transformers:
- ../patches/addRestartPolicy
- ../patches/addDnsPolicy
`)

	m = th.Run("dev", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, prodDevMergeResult2)
}

func TestIssue1251_Plugins_Local(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchJson6902Transformer")
	defer th.Reset()

	writeDeploymentBase(th.Harness)

	writeJsonTransformerPluginConfig(
		th, "composite", "addDnsPolicy", patchJsonDNSPolicy)
	writeJsonTransformerPluginConfig(
		th, "composite", "addRestartPolicy", patchJsonRestartPolicy)
	writeJsonTransformerPluginConfig(
		th, "composite", "addProbe", patchJsonAddProbe)

	th.WriteK("composite", `
resources:
- ../base
transformers:
- addDnsPolicyConfig.yaml
- addRestartPolicyConfig.yaml
- addProbeConfig.yaml
`)
	m := th.Run("composite", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, expectedPatchedDeployment)
}

func writeJsonTransformerPluginConfig(
	th *kusttest_test.HarnessEnhanced, path, name, patch string) {
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
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchJson6902Transformer")
	defer th.Reset()
	writeDeploymentBase(th.Harness)

	th.WriteK("patches", `
resources:
- addDnsPolicyConfig.yaml
- addRestartPolicyConfig.yaml
- addProbeConfig.yaml
`)
	writeJsonTransformerPluginConfig(
		th, "patches", "addDnsPolicy", patchJsonDNSPolicy)
	writeJsonTransformerPluginConfig(
		th, "patches", "addRestartPolicy", patchJsonRestartPolicy)
	writeJsonTransformerPluginConfig(
		th, "patches", "addProbe", patchJsonAddProbe)

	th.WriteK("composite", `
resources:
- ../base
transformers:
- ../patches
`)
	m := th.Run("composite", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, expectedPatchedDeployment)
}

func defineTransformerDirStructure(th *kusttest_test.HarnessEnhanced) {
	writeDeploymentBase(th.Harness)

	th.WriteK("patches/addDnsPolicy", `
resources:
- addDnsPolicyConfig.yaml
`)
	writeJsonTransformerPluginConfig(
		th, "patches/addDnsPolicy", "addDnsPolicy", patchJsonDNSPolicy)

	th.WriteK("patches/addRestartPolicy", `
resources:
- addRestartPolicyConfig.yaml
`)
	writeJsonTransformerPluginConfig(
		th, "patches/addRestartPolicy", "addRestartPolicy", patchJsonRestartPolicy)

	th.WriteK("patches/addProbe", `
resources:
- addProbeConfig.yaml
`)
	writeJsonTransformerPluginConfig(
		th, "patches/addProbe", "addProbe", patchJsonAddProbe)
}
