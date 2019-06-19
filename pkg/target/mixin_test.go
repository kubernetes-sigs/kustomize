// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package target_test

import (
	kusttest_test "sigs.k8s.io/kustomize/v3/pkg/kusttest"
	"testing"
)

type mixinTest struct{}

func (ut *mixinTest) writeKustFile0(th *kusttest_test.KustTestHarness) {
	th.WriteK("/mixintest/app/prod/clientb/", `
kind: Kustomization
apiVersion: kustomize.config.k8s.io/v1beta1

resources:
- ../base

patchesStrategicMerge:
- overlayvalues.yaml
`)
}
func (ut *mixinTest) writeKustFile1(th *kusttest_test.KustTestHarness) {
	th.WriteK("/mixintest/app/prod/base/", `
kind: Kustomization
apiVersion: kustomize.config.k8s.io/v1beta1

resources:
- ../../base
- values.yaml

patchesStrategicMerge:
- overlayvalues.yaml
`)
}
func (ut *mixinTest) writeKustFile2(th *kusttest_test.KustTestHarness) {
	th.WriteK("/mixintest/app/prod/clienta/", `
kind: Kustomization
apiVersion: kustomize.config.k8s.io/v1beta1

resources:
- ../base

patchesStrategicMerge:
- overlayvalues.yaml
`)
}
func (ut *mixinTest) writeKustFile3(th *kusttest_test.KustTestHarness) {
	th.WriteK("/mixintest/app/dev/base/", `
kind: Kustomization
apiVersion: kustomize.config.k8s.io/v1beta1

resources:
- ../../base
- values.yaml

patchesStrategicMerge:
- overlayvalues.yaml
`)
}
func (ut *mixinTest) writeKustFile4(th *kusttest_test.KustTestHarness) {
	th.WriteK("/mixintest/app/dev/clienta/", `
kind: Kustomization
apiVersion: kustomize.config.k8s.io/v1beta1

resources:
- ../base

patchesStrategicMerge:
- overlayvalues.yaml
`)
}
func (ut *mixinTest) writeKustFile5(th *kusttest_test.KustTestHarness) {
	th.WriteK("/mixintest/app/base/", `
kind: Kustomization
apiVersion: kustomize.config.k8s.io/v1beta1

resources:
- ../../components/appdeployment
- ../../components/mysql
- ../../components/persistencelayer
- ../../components/rediscache
- ../../components/redissession
- ../../components/varnish
`)
}
func (ut *mixinTest) writeKustFile6(th *kusttest_test.KustTestHarness) {
	th.WriteK("/mixintest/components/mysql/", `
kind: Kustomization
apiVersion: kustomize.config.k8s.io/v1beta1

resources:
- ./service.yaml
- ./values.yaml
`)
}
func (ut *mixinTest) writeKustFile7(th *kusttest_test.KustTestHarness) {
	th.WriteK("/mixintest/components/persistencelayer/", `
kind: Kustomization
apiVersion: kustomize.config.k8s.io/v1beta1

resources:
- ./service.yaml
- ./values.yaml
`)
}
func (ut *mixinTest) writeKustFile8(th *kusttest_test.KustTestHarness) {
	th.WriteK("/mixintest/components/rediscache/", `
kind: Kustomization
apiVersion: kustomize.config.k8s.io/v1beta1

resources:
- ./service.yaml
- ./values.yaml
`)
}
func (ut *mixinTest) writeKustFile9(th *kusttest_test.KustTestHarness) {
	th.WriteK("/mixintest/components/varnish/", `
kind: Kustomization
apiVersion: kustomize.config.k8s.io/v1beta1

resources:
- ./service.yaml
- ./values.yaml
`)
}
func (ut *mixinTest) writeKustFile10(th *kusttest_test.KustTestHarness) {
	th.WriteK("/mixintest/components/redissession/", `
kind: Kustomization
apiVersion: kustomize.config.k8s.io/v1beta1

resources:
- ./service.yaml
- ./values.yaml
`)
}
func (ut *mixinTest) writeKustFile11(th *kusttest_test.KustTestHarness) {
	th.WriteK("/mixintest/components/appdeployment/", `
kind: Kustomization
apiVersion: kustomize.config.k8s.io/v1beta1

resources:
- ./service.yaml
- ./values.yaml
`)
}
func (ut *mixinTest) writeResourceFile0(th *kusttest_test.KustTestHarness) {
	th.WriteF("/mixintest/app/prod/clientb/overlayvalues.yaml", `
apiVersion: v1
kind: Values
metadata:
  name: redissession
  namespace: build
spec:
  port: 18203
  targetPort: 19203
`)
}
func (ut *mixinTest) writeResourceFile1(th *kusttest_test.KustTestHarness) {
	th.WriteF("/mixintest/app/prod/base/overlayvalues.yaml", `
apiVersion: v1
kind: Values
metadata:
  name: rediscache
  namespace: build
spec:
  port: 8704
---
apiVersion: v1
kind: Values
metadata:
  name: redissession
  namespace: build
spec:
  targetPort: 9705
`)
}
func (ut *mixinTest) writeResourceFile2(th *kusttest_test.KustTestHarness) {
	th.WriteF("/mixintest/app/prod/base/values.yaml", `
apiVersion: v1
kind: Values
metadata:
  name: app-prod-base
  namespace: build
spec:
  field1: value1
`)
}
func (ut *mixinTest) writeResourceFile3(th *kusttest_test.KustTestHarness) {
	th.WriteF("/mixintest/app/prod/clienta/overlayvalues.yaml", `
apiVersion: v1
kind: Values
metadata:
  name: persistencelayer
  namespace: build
spec:
  port: 28903
  targetPort: 29903
`)
}
func (ut *mixinTest) writeResourceFile4(th *kusttest_test.KustTestHarness) {
	th.WriteF("/mixintest/app/dev/base/overlayvalues.yaml", `
apiVersion: v1
kind: Values
metadata:
  name: appdeployment
  namespace: build
spec:
  port: 8501
---
apiVersion: v1
kind: Values
metadata:
  name: mysql
  namespace: build
spec:
  targetPort: 9502
`)
}
func (ut *mixinTest) writeResourceFile5(th *kusttest_test.KustTestHarness) {
	th.WriteF("/mixintest/app/dev/base/values.yaml", `
apiVersion: v1
kind: Values
metadata:
  name: app-dev-base
  namespace: build
spec:
  field1: value1
`)
}
func (ut *mixinTest) writeResourceFile6(th *kusttest_test.KustTestHarness) {
	th.WriteF("/mixintest/app/dev/clienta/overlayvalues.yaml", `
apiVersion: v1
kind: Values
metadata:
  name: mysql
  namespace: build
spec:
  targetPort: 12502
`)
}
func (ut *mixinTest) writeResourceFile7(th *kusttest_test.KustTestHarness) {
	th.WriteF("/mixintest/app/base/values.yaml", `
apiVersion: v1
kind: Values
metadata:
  name: app-base
  namespace: build
spec:
  field1: value1
`)
}
func (ut *mixinTest) writeResourceFile8(th *kusttest_test.KustTestHarness) {
	th.WriteF("/mixintest/components/mysql/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: mysql
spec:
  selector:
    app: mysql
  ports:
  - name: web
    port: $(Values.mysql.spec.port)
    targetPort: $(Values.mysql.spec.targetPort)
`)
}
func (ut *mixinTest) writeResourceFile9(th *kusttest_test.KustTestHarness) {
	th.WriteF("/mixintest/components/mysql/values.yaml", `
apiVersion: v1
kind: Values
metadata:
  name: mysql
  namespace: build
spec:
  port: 8002
  targetPort: 9002
`)
}
func (ut *mixinTest) writeResourceFile10(th *kusttest_test.KustTestHarness) {
	th.WriteF("/mixintest/components/persistencelayer/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: persistencelayer
spec:
  selector:
    app: persistencelayer
  ports:
  - name: web
    port: $(Values.persistencelayer.spec.port)
    targetPort: $(Values.persistencelayer.spec.targetPort)
`)
}
func (ut *mixinTest) writeResourceFile11(th *kusttest_test.KustTestHarness) {
	th.WriteF("/mixintest/components/persistencelayer/values.yaml", `
apiVersion: v1
kind: Values
metadata:
  name: persistencelayer
  namespace: build
spec:
  port: 8003
  targetPort: 9003
`)
}
func (ut *mixinTest) writeResourceFile12(th *kusttest_test.KustTestHarness) {
	th.WriteF("/mixintest/components/rediscache/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: rediscache
spec:
  selector:
    app: rediscache
  ports:
  - name: web
    port: $(Values.rediscache.spec.port)
    targetPort: $(Values.rediscache.spec.targetPort)
`)
}
func (ut *mixinTest) writeResourceFile13(th *kusttest_test.KustTestHarness) {
	th.WriteF("/mixintest/components/rediscache/values.yaml", `
apiVersion: v1
kind: Values
metadata:
  name: rediscache
  namespace: build
spec:
  port: 8004
  targetPort: 9004
`)
}
func (ut *mixinTest) writeResourceFile14(th *kusttest_test.KustTestHarness) {
	th.WriteF("/mixintest/components/varnish/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: varnish
spec:
  selector:
    app: varnish
  ports:
  - name: web
    port: $(Values.varnish.spec.port)
    targetPort: $(Values.varnish.spec.targetPort)
`)
}
func (ut *mixinTest) writeResourceFile15(th *kusttest_test.KustTestHarness) {
	th.WriteF("/mixintest/components/varnish/values.yaml", `
apiVersion: v1
kind: Values
metadata:
  name: varnish
  namespace: build
spec:
  port: 8006
  targetPort: 9006
`)
}
func (ut *mixinTest) writeResourceFile16(th *kusttest_test.KustTestHarness) {
	th.WriteF("/mixintest/components/redissession/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: redissession
spec:
  selector:
    app: redissession
  ports:
  - name: web
    port: $(Values.redissession.spec.port)
    targetPort: $(Values.redissession.spec.targetPort)
`)
}
func (ut *mixinTest) writeResourceFile17(th *kusttest_test.KustTestHarness) {
	th.WriteF("/mixintest/components/redissession/values.yaml", `
apiVersion: v1
kind: Values
metadata:
  name: redissession
  namespace: build
spec:
  port: 8005
  targetPort: 9005
`)
}
func (ut *mixinTest) writeResourceFile18(th *kusttest_test.KustTestHarness) {
	th.WriteF("/mixintest/components/appdeployment/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: appdeployment
spec:
  selector:
    app: appdeployment
  ports:
  - name: web
    port: $(Values.appdeployment.spec.port)
    targetPort: $(Values.appdeployment.spec.targetPort)
`)
}
func (ut *mixinTest) writeResourceFile19(th *kusttest_test.KustTestHarness) {
	th.WriteF("/mixintest/components/appdeployment/values.yaml", `
apiVersion: v1
kind: Values
metadata:
  name: appdeployment
  namespace: build
spec:
  port: 8001
  targetPort: 9001
`)
}

func (ut *mixinTest) writeFiles(th *kusttest_test.KustTestHarness) {
	ut.writeKustFile0(th)
	ut.writeKustFile1(th)
	ut.writeKustFile10(th)
	ut.writeKustFile11(th)
	ut.writeKustFile2(th)
	ut.writeKustFile3(th)
	ut.writeKustFile4(th)
	ut.writeKustFile5(th)
	ut.writeKustFile6(th)
	ut.writeKustFile7(th)
	ut.writeKustFile8(th)
	ut.writeKustFile9(th)
	ut.writeResourceFile0(th)
	ut.writeResourceFile1(th)
	ut.writeResourceFile2(th)
	ut.writeResourceFile3(th)
	ut.writeResourceFile4(th)
	ut.writeResourceFile5(th)
	ut.writeResourceFile6(th)
	ut.writeResourceFile7(th)
	ut.writeResourceFile8(th)
	ut.writeResourceFile9(th)
	ut.writeResourceFile10(th)
	ut.writeResourceFile11(th)
	ut.writeResourceFile12(th)
	ut.writeResourceFile13(th)
	ut.writeResourceFile14(th)
	ut.writeResourceFile15(th)
	ut.writeResourceFile16(th)
	ut.writeResourceFile17(th)
	ut.writeResourceFile18(th)
	ut.writeResourceFile19(th)
}

func TestMixinDevClientA(t *testing.T) {
	ut := &mixinTest{}
	th := kusttest_test.NewKustTestHarness(t, "/mixintest/app/dev/clienta")
	ut.writeFiles(th)
	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Service
metadata:
  name: appdeployment
spec:
  ports:
  - name: web
    port: 8501
    targetPort: 9001
  selector:
    app: appdeployment
---
apiVersion: v1
kind: Values
metadata:
  name: appdeployment
  namespace: build
spec:
  port: 8501
  targetPort: 9001
---
apiVersion: v1
kind: Service
metadata:
  name: mysql
spec:
  ports:
  - name: web
    port: 8002
    targetPort: 12502
  selector:
    app: mysql
---
apiVersion: v1
kind: Values
metadata:
  name: mysql
  namespace: build
spec:
  port: 8002
  targetPort: 12502
---
apiVersion: v1
kind: Service
metadata:
  name: persistencelayer
spec:
  ports:
  - name: web
    port: 8003
    targetPort: 9003
  selector:
    app: persistencelayer
---
apiVersion: v1
kind: Values
metadata:
  name: persistencelayer
  namespace: build
spec:
  port: 8003
  targetPort: 9003
---
apiVersion: v1
kind: Service
metadata:
  name: rediscache
spec:
  ports:
  - name: web
    port: 8004
    targetPort: 9004
  selector:
    app: rediscache
---
apiVersion: v1
kind: Values
metadata:
  name: rediscache
  namespace: build
spec:
  port: 8004
  targetPort: 9004
---
apiVersion: v1
kind: Service
metadata:
  name: redissession
spec:
  ports:
  - name: web
    port: 8005
    targetPort: 9005
  selector:
    app: redissession
---
apiVersion: v1
kind: Values
metadata:
  name: redissession
  namespace: build
spec:
  port: 8005
  targetPort: 9005
---
apiVersion: v1
kind: Service
metadata:
  name: varnish
spec:
  ports:
  - name: web
    port: 8006
    targetPort: 9006
  selector:
    app: varnish
---
apiVersion: v1
kind: Values
metadata:
  name: varnish
  namespace: build
spec:
  port: 8006
  targetPort: 9006
---
apiVersion: v1
kind: Values
metadata:
  name: app-dev-base
  namespace: build
spec:
  field1: value1
`)
}

func TestMixinProdClientA(t *testing.T) {
	ut := &mixinTest{}
	th := kusttest_test.NewKustTestHarness(t, "/mixintest/app/prod/clienta")
	ut.writeFiles(th)
	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Service
metadata:
  name: appdeployment
spec:
  ports:
  - name: web
    port: 8001
    targetPort: 9001
  selector:
    app: appdeployment
---
apiVersion: v1
kind: Values
metadata:
  name: appdeployment
  namespace: build
spec:
  port: 8001
  targetPort: 9001
---
apiVersion: v1
kind: Service
metadata:
  name: mysql
spec:
  ports:
  - name: web
    port: 8002
    targetPort: 9002
  selector:
    app: mysql
---
apiVersion: v1
kind: Values
metadata:
  name: mysql
  namespace: build
spec:
  port: 8002
  targetPort: 9002
---
apiVersion: v1
kind: Service
metadata:
  name: persistencelayer
spec:
  ports:
  - name: web
    port: 28903
    targetPort: 29903
  selector:
    app: persistencelayer
---
apiVersion: v1
kind: Values
metadata:
  name: persistencelayer
  namespace: build
spec:
  port: 28903
  targetPort: 29903
---
apiVersion: v1
kind: Service
metadata:
  name: rediscache
spec:
  ports:
  - name: web
    port: 8704
    targetPort: 9004
  selector:
    app: rediscache
---
apiVersion: v1
kind: Values
metadata:
  name: rediscache
  namespace: build
spec:
  port: 8704
  targetPort: 9004
---
apiVersion: v1
kind: Service
metadata:
  name: redissession
spec:
  ports:
  - name: web
    port: 8005
    targetPort: 9705
  selector:
    app: redissession
---
apiVersion: v1
kind: Values
metadata:
  name: redissession
  namespace: build
spec:
  port: 8005
  targetPort: 9705
---
apiVersion: v1
kind: Service
metadata:
  name: varnish
spec:
  ports:
  - name: web
    port: 8006
    targetPort: 9006
  selector:
    app: varnish
---
apiVersion: v1
kind: Values
metadata:
  name: varnish
  namespace: build
spec:
  port: 8006
  targetPort: 9006
---
apiVersion: v1
kind: Values
metadata:
  name: app-prod-base
  namespace: build
spec:
  field1: value1
`)
}

func TestMixinProdClientB(t *testing.T) {
	ut := &mixinTest{}
	th := kusttest_test.NewKustTestHarness(t, "/mixintest/app/prod/clientb")
	ut.writeFiles(th)
	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
kind: Service
metadata:
  name: appdeployment
spec:
  ports:
  - name: web
    port: 8001
    targetPort: 9001
  selector:
    app: appdeployment
---
apiVersion: v1
kind: Values
metadata:
  name: appdeployment
  namespace: build
spec:
  port: 8001
  targetPort: 9001
---
apiVersion: v1
kind: Service
metadata:
  name: mysql
spec:
  ports:
  - name: web
    port: 8002
    targetPort: 9002
  selector:
    app: mysql
---
apiVersion: v1
kind: Values
metadata:
  name: mysql
  namespace: build
spec:
  port: 8002
  targetPort: 9002
---
apiVersion: v1
kind: Service
metadata:
  name: persistencelayer
spec:
  ports:
  - name: web
    port: 8003
    targetPort: 9003
  selector:
    app: persistencelayer
---
apiVersion: v1
kind: Values
metadata:
  name: persistencelayer
  namespace: build
spec:
  port: 8003
  targetPort: 9003
---
apiVersion: v1
kind: Service
metadata:
  name: rediscache
spec:
  ports:
  - name: web
    port: 8704
    targetPort: 9004
  selector:
    app: rediscache
---
apiVersion: v1
kind: Values
metadata:
  name: rediscache
  namespace: build
spec:
  port: 8704
  targetPort: 9004
---
apiVersion: v1
kind: Service
metadata:
  name: redissession
spec:
  ports:
  - name: web
    port: 18203
    targetPort: 19203
  selector:
    app: redissession
---
apiVersion: v1
kind: Values
metadata:
  name: redissession
  namespace: build
spec:
  port: 18203
  targetPort: 19203
---
apiVersion: v1
kind: Service
metadata:
  name: varnish
spec:
  ports:
  - name: web
    port: 8006
    targetPort: 9006
  selector:
    app: varnish
---
apiVersion: v1
kind: Values
metadata:
  name: varnish
  namespace: build
spec:
  port: 8006
  targetPort: 9006
---
apiVersion: v1
kind: Values
metadata:
  name: app-prod-base
  namespace: build
spec:
  field1: value1
`)
}
