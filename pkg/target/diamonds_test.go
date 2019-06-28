// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package target_test

import (
	kusttest_test "sigs.k8s.io/kustomize/v3/pkg/kusttest"
	"testing"
)

// Here's a structure of two kustomizations,
// `dev` and `prod`, individually deployable,
// that depend on a diamond that combines
// multiple tenants (kirk, spock and bones),
// each sharing a common base.
//
// The objects used are contrived to avoid
// clouding the example with authentic
// but verbose Deployment boilerplate.
//
// Patches are applied at various levels,
// requiring more specificity as needed.
//
//         dev      prod
//             \   /
//            tenants
//          /    |    \
//      kirk   spock  bones
//          \    |    /
//             base
//
func writeDiamondBase(th *kusttest_test.KustTestHarness) {
	th.WriteK("/app/base", `
resources:
- deploy.yaml
`)
	th.WriteF("/app/base/deploy.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: storefront
spec:
  numReplicas: 1
`)
}

func writeKirk(th *kusttest_test.KustTestHarness) {
	th.WriteK("/app/kirk", `
namePrefix: kirk-
resources:
- ../base
- configmap.yaml
patchesStrategicMerge:
- dep-patch.yaml
`)
	th.WriteF("/app/kirk/dep-patch.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: storefront
spec:
  type: Confident
`)
	th.WriteF("/app/kirk/configmap.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: settings
data:
  phaser: caress
`)
}

func writeSpock(th *kusttest_test.KustTestHarness) {
	th.WriteK("/app/spock", `
namePrefix: spock-
resources:
- ../base
patchesStrategicMerge:
- dep-patch.yaml
`)
	th.WriteF("/app/spock/dep-patch.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: storefront
spec:
  type: Logical
`)
}

func writeBones(th *kusttest_test.KustTestHarness) {
	th.WriteK("/app/bones", `
namePrefix: bones-
resources:
- ../base
patchesStrategicMerge:
- dep-patch.yaml
`)
	th.WriteF("/app/bones/dep-patch.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: storefront
spec:
  type: Concerned
`)
}

func writeTenants(th *kusttest_test.KustTestHarness) {
	th.WriteK("/app/tenants", `
namePrefix: t-
resources:
- ../kirk
- ../spock
- ../bones
- configMap.yaml
patchesStrategicMerge:
- bones-patch.yaml
`)
	th.WriteF("/app/tenants/bones-patch.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: bones-storefront
spec:
  mood: Cantankerous
`)
	th.WriteF("/app/tenants/configMap.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: federation
data:
  zone: neutral
  guardian: forever
`)
}

func TestBasicDiamond(t *testing.T) {
	th := kusttest_test.NewKustTestHarness(t, "/app/prod")
	writeDiamondBase(th)
	writeKirk(th)
	writeSpock(th)
	writeBones(th)
	writeTenants(th)
	th.WriteK("/app/prod", `
namePrefix: prod-
resources:
- ../tenants
patchesStrategicMerge:
- patches.yaml
`)
	// The patch only has to be specific enough
	// to match the item.
	th.WriteF("/app/prod/patches.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: t-kirk-storefront
spec:
  numReplicas: 10000
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: federation
data:
  guardian: ofTheGalaxy
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: t-federation
data:
  zone: twilight
`)

	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Deployment
metadata:
  name: prod-t-kirk-storefront
spec:
  numReplicas: 10000
  type: Confident
---
apiVersion: v1
data:
  phaser: caress
kind: ConfigMap
metadata:
  name: prod-t-kirk-settings
---
apiVersion: v1
kind: Deployment
metadata:
  name: prod-t-spock-storefront
spec:
  numReplicas: 1
  type: Logical
---
apiVersion: v1
kind: Deployment
metadata:
  name: prod-t-bones-storefront
spec:
  mood: Cantankerous
  numReplicas: 1
  type: Concerned
---
apiVersion: v1
data:
  guardian: ofTheGalaxy
  zone: twilight
kind: ConfigMap
metadata:
  name: prod-t-federation
`)
}

// This example demonstrate a simple sharing
// of a configmap and variables between
// component1 and component2 before being
// aggregated into myapp
//
//             myapp
//          /    |    \
//    component1 | component2
//          \    |    /
//             common
//

type diamonImportTest struct{}

func (ut *diamonImportTest) writeKustCommon(th *kusttest_test.KustTestHarness) {
	th.WriteK("/diamondimport/common/", `
resources:
- configmap.yaml

vars:
- name: ConfigMap.global.data.user
  objref:
    apiVersion: v1
    kind: ConfigMap
    name: global
  fieldref:
    fieldpath: data.user
`)
}
func (ut *diamonImportTest) writeKustComponent2(th *kusttest_test.KustTestHarness) {
	th.WriteK("/diamondimport/component2/", `
resources:
- ../common
- deployment.yaml
`)
}
func (ut *diamonImportTest) writeKustComponent1(th *kusttest_test.KustTestHarness) {
	th.WriteK("/diamondimport/component1/", `
resources:
- ../common
- deployment.yaml
`)
}
func (ut *diamonImportTest) writeKustMyApp(th *kusttest_test.KustTestHarness) {
	th.WriteK("/diamondimport/myapp/", `
resources:
- ../common
- ../component1
- ../component2
`)
}
func (ut *diamonImportTest) writeCommonConfigMap(th *kusttest_test.KustTestHarness) {
	th.WriteF("/diamondimport/common/configmap.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: global
data:
  settings: |
     database: mydb
     port: 3000
  user: myuser
`)
}
func (ut *diamonImportTest) writeComponent2(th *kusttest_test.KustTestHarness) {
	th.WriteF("/diamondimport/component2/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: component2
  labels:
     app: component2
spec:
  replicas: 1
  selector:
    matchLabels:
      app: component2
  template:
    metadata:
      labels:
        app: component2
    spec:
      containers:
      - name: component2
        image: k8s.gcr.io/busybox
        env:
        - name: APP_USER
          value: $(ConfigMap.global.data.user)
        command: [ "/bin/sh", "-c", "cat /etc/config/component2 && sleep 60" ]
        volumeMounts:
        - name: config-volume
          mountPath: /etc/config
      volumes:
      - name: config-volume
        configMap:
          name: global
          items:
          - key: settings
            path: component2
`)
}
func (ut *diamonImportTest) writeComponent1(th *kusttest_test.KustTestHarness) {
	th.WriteF("/diamondimport/component1/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: component1
  labels:
     app: component1
spec:
  replicas: 1
  selector:
    matchLabels:
      app: component1
  template:
    metadata:
      labels:
        app: component1
    spec:
      containers:
      - name: component1
        image: k8s.gcr.io/busybox
        env:
        - name: APP_USER
          value: $(ConfigMap.global.data.user)
        command: [ "/bin/sh", "-c", "cat /etc/config/component1 && sleep 60" ]
        volumeMounts:
        - name: config-volume
          mountPath: /etc/config
      volumes:
      - name: config-volume
        configMap:
          name: global
          items:
          - key: settings
            path: component1
`)
}
func TestSimpleDiamondImport(t *testing.T) {
	ut := &diamonImportTest{}
	th := kusttest_test.NewKustTestHarness(t, "/diamondimport/myapp")
	ut.writeKustCommon(th)
	ut.writeKustComponent1(th)
	ut.writeKustComponent2(th)
	ut.writeKustMyApp(th)
	ut.writeCommonConfigMap(th)
	ut.writeComponent1(th)
	ut.writeComponent2(th)
	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		// Error before this Resource.Append fix.
		// may not add resource with an already registered id: ~G_v1_ConfigMap|~X|global
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  settings: |
    database: mydb
    port: 3000
  user: myuser
kind: ConfigMap
metadata:
  name: global
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: component1
  name: component1
spec:
  replicas: 1
  selector:
    matchLabels:
      app: component1
  template:
    metadata:
      labels:
        app: component1
    spec:
      containers:
      - command:
        - /bin/sh
        - -c
        - cat /etc/config/component1 && sleep 60
        env:
        - name: APP_USER
          value: myuser
        image: k8s.gcr.io/busybox
        name: component1
        volumeMounts:
        - mountPath: /etc/config
          name: config-volume
      volumes:
      - configMap:
          items:
          - key: settings
            path: component1
          name: global
        name: config-volume
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: component2
  name: component2
spec:
  replicas: 1
  selector:
    matchLabels:
      app: component2
  template:
    metadata:
      labels:
        app: component2
    spec:
      containers:
      - command:
        - /bin/sh
        - -c
        - cat /etc/config/component2 && sleep 60
        env:
        - name: APP_USER
          value: myuser
        image: k8s.gcr.io/busybox
        name: component2
        volumeMounts:
        - mountPath: /etc/config
          name: config-volume
      volumes:
      - configMap:
          items:
          - key: settings
            path: component2
          name: global
        name: config-volume
`)
}
