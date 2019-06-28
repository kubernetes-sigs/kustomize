// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package target_test

import (
	"strings"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/v3/pkg/kusttest"
)

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
func writeDiamondCompositionBase(th *kusttest_test.KustTestHarness) {
	th.WriteK("/app/base", `
resources:
- deployment.yaml
`)

	th.WriteF("/app/base/deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: my-deployment
spec:
  template:
    spec:
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

	th.WriteF("/app/probe/dep-patch.yaml", `
apiVersion: extensions/v1beta1
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
`)
}

func writeDNSOverlay(th *kusttest_test.KustTestHarness) {
	th.WriteK("/app/dns", `
resources:
- ../base

patchesStrategicMerge:
- dep-patch.yaml
`)

	th.WriteF("/app/dns/dep-patch.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: my-deployment
spec:
  template:
    spec:
      dnsPolicy: ClusterFirst
`)
}

func writeRestartOverlay(th *kusttest_test.KustTestHarness) {
	th.WriteK("/app/restart", `
resources:
- ../base

patchesStrategicMerge:
- dep-patch.yaml
`)

	th.WriteF("/app/restart/dep-patch.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: my-deployment
spec:
  template:
    spec:
      restartPolicy: Always
`)
}

func writeComposite(th *kusttest_test.KustTestHarness) {
	th.WriteK("/app/composite", `
resources:
- ../probe
- ../dns
- ../restart
`)
}

func TestCompositeDiamond(t *testing.T) {
	th := kusttest_test.NewKustTestHarness(t, "/app/composite")
	writeDiamondCompositionBase(th)
	writeProbeOverlay(th)
	writeDNSOverlay(th)
	writeRestartOverlay(th)
	writeComposite(th)

	_, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err == nil {
		t.Fatalf("Expected resource accumulation error")
	}
	if !strings.Contains(
		err.Error(), "already registered id: extensions_v1beta1_Deployment|~X|my-deployment") {
		t.Fatalf("Unexpected err: %v", err)
	}
}

// Expected output
//
// apiVersion: extensions/v1beta1
// kind: Deployment
// metadata:
//   name: my-deployment
// spec:
//   template:
//     spec:
//       containers:
//       - image: my-image
//         livenessProbe:
//           httpGet:
//             path: /healthz
//             port: 8080
//         name: my-deployment
//       dnsPolicy: ClusterFirst
//       restartPolicy: Always
