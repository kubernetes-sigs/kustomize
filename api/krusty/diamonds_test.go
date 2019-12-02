// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
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
func writeDiamondBase(th kusttest_test.Harness) {
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

func writeKirk(th kusttest_test.Harness) {
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

func writeSpock(th kusttest_test.Harness) {
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

func writeBones(th kusttest_test.Harness) {
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

func writeTenants(th kusttest_test.Harness) {
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
	th := kusttest_test.MakeHarness(t)
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

	m := th.Run("/app/prod", th.MakeDefaultOptions())
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
