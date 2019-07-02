// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package target_test

import (
	kusttest_test "sigs.k8s.io/kustomize/v3/pkg/kusttest"
	"strings"
	"testing"
)

type recreate1298Test struct{}

func (ut *recreate1298Test) writeKustFileDev(th *kusttest_test.KustTestHarness) {
	th.WriteK("/recreate1298/dev/", `
resources:
- elasticsearch-dev-service.yaml
vars:
- name: elasticsearch-dev-service-name
  objref:
    kind: Service
    name: elasticsearch
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
- name: elasticsearch-dev-protocol
  objref:
    kind: Service
    name: elasticsearch
    apiVersion: v1
  fieldref:
    fieldpath: spec.ports[0].protocol
`)
}
func (ut *recreate1298Test) writeKustFileTest(th *kusttest_test.KustTestHarness) {
	th.WriteK("/recreate1298/test/", `
resources:
- elasticsearch-test-service.yaml
vars:
- name: elasticsearch-test-service-name
  objref:
    kind: Service
    name: elasticsearch
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
- name: elasticsearch-test-protocol
  objref:
    kind: Service
    name: elasticsearch
    apiVersion: v1
  fieldref:
    fieldpath: spec.ports[0].protocol
`)
}
func (ut *recreate1298Test) writeKustFileOneFolder(th *kusttest_test.KustTestHarness) {
	th.WriteK("/recreate1298/onefolder/", `
resources:
- elasticsearch-test-service.yaml
- elasticsearch-dev-service.yaml
vars:
- name: elasticsearch-test-service-name
  objref:
    kind: Service
    name: elasticsearch
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
- name: elasticsearch-test-protocol
  objref:
    kind: Service
    name: elasticsearch
    apiVersion: v1
  fieldref:
    fieldpath: spec.ports[0].protocol
- name: elasticsearch-dev-service-name
  objref:
    kind: Service
    name: elasticsearch
    apiVersion: v1
  fieldref:
    fieldpath: metadata.name
- name: elasticsearch-dev-protocol
  objref:
    kind: Service
    name: elasticsearch
    apiVersion: v1
  fieldref:
    fieldpath: spec.ports[0].protocol
`)
}
func (ut *recreate1298Test) writeKustFileTwoFolders(th *kusttest_test.KustTestHarness) {
	th.WriteK("/recreate1298/twofolders/", `
resources:
- ../dev
- ../test
`)
}
func (ut *recreate1298Test) writeResourcesDev(th *kusttest_test.KustTestHarness, filename string) {
	th.WriteF("/recreate1298/"+filename, `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: elasticsearch
  namespace: dev
spec:
  template:
    spec:
      containers:
        - name: elasticsearch
          env:
            - name: DISCOVERY_SERVICE
              value: "$(elasticsearch-dev-service-name).monitoring.svc.cluster.local"
            - name: DISCOVERY_PROTOCOL
              value: "$(elasticsearch-dev-protocol)"
---
apiVersion: v1
kind: Service
metadata:
  name: elasticsearch
  namespace: dev
spec:
  ports:
    - name: transport
      port: 9300
      protocol: TCP
  clusterIP: None
`)
}
func (ut *recreate1298Test) writeResourcesTest(th *kusttest_test.KustTestHarness, filename string) {
	th.WriteF("/recreate1298/"+filename, `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: elasticsearch
  namespace: test
spec:
  template:
    spec:
      containers:
        - name: elasticsearch
          env:
            - name: DISCOVERY_SERVICE
              value: "$(elasticsearch-test-service-name).monitoring.svc.cluster.local"
            - name: DISCOVERY_PROTOCOL
              value: "$(elasticsearch-test-protocol)"
---
apiVersion: v1
kind: Service
metadata:
  name: elasticsearch
  namespace: test
spec:
  ports:
    - name: transport
      port: 9300
      protocol: UDP
  clusterIP: None
`)
}

const expectedOutput string = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: elasticsearch
  namespace: dev
spec:
  template:
    spec:
      containers:
      - env:
        - name: DISCOVERY_SERVICE
          value: elasticsearch.monitoring.svc.cluster.local
        - name: DISCOVERY_PROTOCOL
          value: TCP
        name: elasticsearch
---
apiVersion: v1
kind: Service
metadata:
  name: elasticsearch
  namespace: dev
spec:
  clusterIP: None
  ports:
  - name: transport
    port: 9300
    protocol: TCP
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: elasticsearch
  namespace: test
spec:
  template:
    spec:
      containers:
      - env:
        - name: DISCOVERY_SERVICE
          value: elasticsearch.monitoring.svc.cluster.local
        - name: DISCOVERY_PROTOCOL
          value: UDP
        name: elasticsearch
---
apiVersion: v1
kind: Service
metadata:
  name: elasticsearch
  namespace: test
spec:
  clusterIP: None
  ports:
  - name: transport
    port: 9300
    protocol: UDP
`

func TestVariablesInTwoFolder(t *testing.T) {
	ut := &recreate1298Test{}
	th := kusttest_test.NewKustTestHarness(t, "/recreate1298/twofolders")
	ut.writeKustFileTwoFolders(th)
	ut.writeKustFileTest(th)
	ut.writeKustFileDev(th)
	ut.writeResourcesDev(th, "dev/elasticsearch-dev-service.yaml")
	ut.writeResourcesTest(th, "test/elasticsearch-test-service.yaml")
	m, err := th.MakeKustTarget().MakeCustomizedResMap()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	th.AssertActualEqualsExpected(m, expectedOutput)
}

func TestVariablesInOneFolder(t *testing.T) {
	ut := &recreate1298Test{}
	th := kusttest_test.NewKustTestHarness(t, "/recreate1298/onefolder")
	ut.writeKustFileOneFolder(th)
	ut.writeResourcesDev(th, "onefolder/elasticsearch-dev-service.yaml")
	ut.writeResourcesTest(th, "onefolder/elasticsearch-test-service.yaml")
	_, err := th.MakeKustTarget().MakeCustomizedResMap()
	// This requires the namespace to be added to the variable declaration.
	// The expected output would then be identical regardless if the
	// the files are in one or two folders.
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "unable to disambiguate") {
		t.Fatalf("unexpected error %v", err)
	}

}
