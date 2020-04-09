// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"strings"
	"testing"

	. "sigs.k8s.io/kustomize/api/krusty"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	"sigs.k8s.io/kustomize/api/types"
)

const httpsService = `
apiVersion: v1
kind: Service
metadata:
  name: my-https-svc
spec:
  ports:
  - port: 443
    protocol: TCP
    name: https
  selector:
    app: my-app
`

func writeStatefulSetBase(th kusttest_test.Harness) {
	th.WriteK("/app/base", `
resources:
- statefulset.yaml
`)
	th.WriteF("/app/base/statefulset.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: my-sts
spec:
  serviceName: my-svc
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: app
        image: my-image
  volumeClaimTemplates:
  - spec:
      storageClassName: default
`)
}

func writeHTTPSOverlay(th kusttest_test.Harness) {
	th.WriteK("/app/https", `
resources:
- ../base
- https-svc.yaml
patchesStrategicMerge:
- sts-patch.yaml
`)
	th.WriteF("/app/https/https-svc.yaml", httpsService)
	th.WriteF("/app/https/sts-patch.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: my-sts
spec:
  serviceName: my-https-svc
`)
}

func writeHTTPSTransformerRaw(th kusttest_test.Harness) {
	th.WriteF("/app/https/service/https-svc.yaml", httpsService)
	th.WriteF("/app/https/transformer/transformer.yaml", `
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: svcNameTran
target: 
  group: apps
  version: v1
  kind: StatefulSet
  name: my-sts
patch: |-
  apiVersion: apps/v1
  kind: StatefulSet
  metadata:
    name: my-sts
  spec:
    serviceName: my-https-svc
`)
}

func writeHTTPSTransformerBase(th kusttest_test.Harness) {
	th.WriteK("/app/https/service", `
resources:
- https-svc.yaml
`)
	th.WriteK("/app/https/transformer", `
resources:
- transformer.yaml
`)
	writeHTTPSTransformerRaw(th)
}

func writeConfigFromEnvOverlay(th kusttest_test.Harness) {
	th.WriteK("/app/config", `
resources:
- ../base
configMapGenerator:
- name: my-config
  literals:
  - MY_ENV=foo
generatorOptions:
  disableNameSuffixHash: true
patchesStrategicMerge:
- sts-patch.yaml
`)
	th.WriteF("/app/config/sts-patch.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: my-sts
spec:
  template:
    spec:
      containers:
      - name: app
        envFrom:
        - configMapRef:
            name: my-config
`)
}

func writeConfigFromEnvTransformerRaw(th kusttest_test.Harness) {
	th.WriteF("/app/config/map/generator.yaml", `
apiVersion: builtin
kind: ConfigMapGenerator
metadata:
  name: my-config
options:
  disableNameSuffixHash: true
literals:
- MY_ENV=foo
`)
	th.WriteF("/app/config/transformer/transformer.yaml", `
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: envFromConfigTrans
target: 
  group: apps
  version: v1
  kind: StatefulSet
  name: my-sts
patch: |-
  apiVersion: apps/v1
  kind: StatefulSet
  metadata:
    name: my-sts
  spec:
    template:
      spec:
        containers:
        - name: app
          envFrom:
          - configMapRef:
              name: my-config
`)
}
func writeConfigFromEnvTransformerBase(th kusttest_test.Harness) {
	th.WriteK("/app/config/map", `
resources:
- generator.yaml
`)
	th.WriteK("/app/config/transformer", `
resources:
- transformer.yaml
`)
	writeConfigFromEnvTransformerRaw(th)
}

func writeTolerationsOverlay(th kusttest_test.Harness) {
	th.WriteK("/app/tolerations", `
resources:
- ../base
patchesStrategicMerge:
- sts-patch.yaml
`)
	th.WriteF("/app/tolerations/sts-patch.yaml", `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: my-sts
spec:
  template:
    spec:
      tolerations:
      - effect: NoExecute
        key: node.kubernetes.io/not-ready
        tolerationSeconds: 30
`)
}

func writeTolerationsTransformerRaw(th kusttest_test.Harness) {
	th.WriteF("/app/tolerations/transformer.yaml", `
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: tolTrans
target: 
  group: apps
  version: v1
  kind: StatefulSet
  name: my-sts
patch: |-
  apiVersion: apps/v1
  kind: StatefulSet
  metadata:
    name: my-sts
  spec:
    template:
      spec:
        tolerations:
        - effect: NoExecute
          key: node.kubernetes.io/not-ready
          tolerationSeconds: 30
`)
}

func writeTolerationsTransformerBase(th kusttest_test.Harness) {
	th.WriteK("/app/tolerations", `
resources:
- transformer.yaml
`)
	writeTolerationsTransformerRaw(th)
}

func writeStorageOverlay(th kusttest_test.Harness) {
	th.WriteK("/app/storage", `
resources:
- ../base
patchesJson6902:
- target:
    group: apps
    version: v1
    kind: StatefulSet
    name: my-sts
  path: sts-patch.json
`)
	th.WriteF("/app/storage/sts-patch.json", `
[{"op": "replace", "path": "/spec/volumeClaimTemplates/0/spec/storageClassName", "value": "my-sc"}]
`)
}

func writeStorageTransformerRaw(th kusttest_test.Harness) {
	th.WriteF("/app/storage/transformer.yaml", `
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: storageTrans
target: 
  group: apps
  version: v1
  kind: StatefulSet
  name: my-sts
patch: |-
  [{"op": "replace", "path": "/spec/volumeClaimTemplates/0/spec/storageClassName", "value": "my-sc"}]
`)
}

func writeStorageTransformerBase(th kusttest_test.Harness) {
	th.WriteK("/app/storage", `
resources:
- transformer.yaml
`)
	writeStorageTransformerRaw(th)
}

func writePatchingOverlays(th kusttest_test.Harness) {
	writeStorageOverlay(th)
	writeConfigFromEnvOverlay(th)
	writeTolerationsOverlay(th)
	writeHTTPSOverlay(th)
}

func writePatchingTransformersRaw(th kusttest_test.Harness) {
	writeStorageTransformerRaw(th)
	writeConfigFromEnvTransformerRaw(th)
	writeTolerationsTransformerRaw(th)
	writeHTTPSTransformerRaw(th)
}

// Similar to writePatchingTransformersRaw, except here the
// transformers and related artifacts are addressable as _bases_.
// They are listed in a kustomization file, and consumers of
// the plugin refer to the kustomization instead of to the local
// file in the "transformers:" field.
//
// Using bases makes the set of files relocatable with
// respect to the overlays, and avoids the need to relax load
// restrictions on file paths reaching outside the `dev` and
// `prod` kustomization roots.  I.e. with bases tests can use
// NewKustTestHarness instead of NewKustTestHarnessNoLoadRestrictor.
//
// Using transformer plugins from _bases_ means the plugin config
// must be self-contained, i.e. the config may not have fields that
// refer to local files, since those files won't be present when
// the plugin is instantiated and used.
func writePatchingTransformerBases(th kusttest_test.Harness) {
	writeStorageTransformerBase(th)
	writeConfigFromEnvTransformerBase(th)
	writeTolerationsTransformerBase(th)
	writeHTTPSTransformerBase(th)
}

// Here's a complex kustomization scenario that combines multiple overlays
// on a common base:
//
//                 dev                       prod
//                  |                         |
//                  |                         |
//        + ------- +          + ------------ + ------------- +
//        |         |          |              |               |
//        |         |          |              |               |
//        v         |          v              v               v
//     storage      + -----> config       tolerations       https
//        |                    |              |               |
//        |                    |              |               |
//        |                    + --- +  + --- +               |
//        |                          |  |                     |
//        |                          v  v                     |
//        + -----------------------> base <------------------ +
//
// The base resource is a statefulset. Each intermediate overlay manages or
// generates new resources and patches different aspects of the same base
// resource, without using any of the `namePrefix`, `nameSuffix` or `namespace`
// kustomization keywords.
//
// Intermediate overlays:
//   - storage: Changes the storage class of the stateful set with a JSON patch.
//   - config: Generates a config map and adds a field as an environment
//             variable.
//   - tolerations: Adds a new tolerations field in the spec.
//   - https: Adds a new service resource and changes the service name in the
//            stateful set.
//
// Top overlays:
//   - dev: Combines the storage and config intermediate overlays.
//   - prod: Combines the config, tolerations and https intermediate overlays.

func TestComplexComposition_Dev_Failure(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeStatefulSetBase(th)
	writePatchingOverlays(th)
	th.WriteK("/app/dev", `
resources:
- ../storage
- ../config
`)
	err := th.RunWithErr("/app/dev", th.MakeDefaultOptions())
	if err == nil {
		t.Fatalf("Expected resource accumulation error")
	}
	if !strings.Contains(
		err.Error(), "already registered id: apps_v1_StatefulSet|~X|my-sts") {
		t.Fatalf("Unexpected err: %v", err)
	}
}

const devDesiredResult = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: my-sts
spec:
  selector:
    matchLabels:
      app: my-app
  serviceName: my-svc
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - envFrom:
        - configMapRef:
            name: my-config
        image: my-image
        name: app
  volumeClaimTemplates:
  - spec:
      storageClassName: my-sc
---
apiVersion: v1
data:
  MY_ENV: foo
kind: ConfigMap
metadata:
  name: my-config
`

func TestComplexComposition_Dev_SuccessWithRawTransformers(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeStatefulSetBase(th)
	writePatchingTransformersRaw(th)
	th.WriteK("/app/dev", `
resources:
- ../base
generators:
- ../config/map/generator.yaml
transformers:
- ../config/transformer/transformer.yaml
- ../storage/transformer.yaml
`)
	m := th.Run("/app/dev", func() Options {
		o := th.MakeDefaultOptions()
		o.LoadRestrictions = types.LoadRestrictionsNone
		return o
	}())
	th.AssertActualEqualsExpected(m, devDesiredResult)
}

func TestComplexComposition_Dev_SuccessWithBaseTransformers(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeStatefulSetBase(th)
	writePatchingTransformerBases(th)
	th.WriteK("/app/dev", `
resources:
- ../base
generators:
- ../config/map
transformers:
- ../config/transformer
- ../storage
`)
	m := th.Run("/app/dev", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, devDesiredResult)
}

func TestComplexComposition_Prod_Failure(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeStatefulSetBase(th)
	writePatchingOverlays(th)
	th.WriteK("/app/prod", `
resources:
- ../config
- ../tolerations
- ../https
`)
	err := th.RunWithErr("/app/prod", th.MakeDefaultOptions())
	if err == nil {
		t.Fatalf("Expected resource accumulation error")
	}
	if !strings.Contains(
		err.Error(), "already registered id: apps_v1_StatefulSet|~X|my-sts") {
		t.Fatalf("Unexpected err: %v", err)
	}
}

const prodDesiredResult = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: my-sts
spec:
  selector:
    matchLabels:
      app: my-app
  serviceName: my-https-svc
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - envFrom:
        - configMapRef:
            name: my-config
        image: my-image
        name: app
      tolerations:
      - effect: NoExecute
        key: node.kubernetes.io/not-ready
        tolerationSeconds: 30
  volumeClaimTemplates:
  - spec:
      storageClassName: default
---
apiVersion: v1
kind: Service
metadata:
  name: my-https-svc
spec:
  ports:
  - name: https
    port: 443
    protocol: TCP
  selector:
    app: my-app
---
apiVersion: v1
data:
  MY_ENV: foo
kind: ConfigMap
metadata:
  name: my-config
`

func TestComplexComposition_Prod_SuccessWithRawTransformers(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeStatefulSetBase(th)
	writePatchingTransformersRaw(th)
	th.WriteK("/app/prod", `
resources:
- ../base
- ../https/service/https-svc.yaml
generators:
- ../config/map/generator.yaml
transformers:
- ../config/transformer/transformer.yaml
- ../https/transformer/transformer.yaml
- ../tolerations/transformer.yaml
`)
	m := th.Run("/app/prod", func() Options {
		o := th.MakeDefaultOptions()
		o.LoadRestrictions = types.LoadRestrictionsNone
		return o
	}())
	th.AssertActualEqualsExpected(m, prodDesiredResult)
}

func TestComplexComposition_Prod_SuccessWithBaseTransformers(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	writeStatefulSetBase(th)
	writePatchingTransformerBases(th)
	th.WriteK("/app/prod", `
resources:
- ../base
- ../https/service
generators:
- ../config/map
transformers:
- ../config/transformer
- ../https/transformer
- ../tolerations
`)
	m := th.Run("/app/prod", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, prodDesiredResult)
}
