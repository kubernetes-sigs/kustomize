package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestGkeGenerator(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteK(".", `
generators:
- |-
  apiVersion: builtin
  kind: IAMPolicyGenerator
  metadata:
    name: my-gke-generator
  cloud: gke
  kubernetesService:
    name: k8s-sa-name
  serviceAccount:
    name: gsa-name
    projectId: project-id
`)
	expected := `
apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    iam.gke.io/gcp-service-account: gsa-name@project-id.iam.gserviceaccount.com
  name: k8s-sa-name
`
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, expected)
}

func TestGkeGeneratorWithNamespace(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteK(".", `
generators:
- |-
  apiVersion: builtin
  kind: IAMPolicyGenerator
  metadata:
    name: my-gke-generator
  cloud: gke
  kubernetesService: 
    namespace: k8s-namespace
    name: k8s-sa-name
  serviceAccount:
    name: gsa-name
    projectId: project-id
`)
	expected := `
apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    iam.gke.io/gcp-service-account: gsa-name@project-id.iam.gserviceaccount.com
  name: k8s-sa-name
  namespace: k8s-namespace
`
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, expected)
}

func TestGkeGeneratorWithTwo(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteK(".", `
generators:
- gkegenerator1.yaml
- gkegenerator2.yaml
`)

	th.WriteF("gkegenerator1.yaml", `
apiVersion: builtin
kind: IAMPolicyGenerator
metadata:
  name: my-gke-generator1
cloud: gke
kubernetesService: 
  namespace: k8s-namespace-1
  name: k8s-sa-name-1
serviceAccount:
  name: gsa-name-1
  projectId: project-id-1
`)
	th.WriteF("gkegenerator2.yaml", `
apiVersion: builtin
kind: IAMPolicyGenerator
metadata:
  name: my-gke-generator2
cloud: gke
kubernetesService:
  name: k8s-sa-name-2
serviceAccount:
  name: gsa-name-2
  projectId: project-id-2
`)
	expected := `
apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    iam.gke.io/gcp-service-account: gsa-name-1@project-id-1.iam.gserviceaccount.com
  name: k8s-sa-name-1
  namespace: k8s-namespace-1
---
apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    iam.gke.io/gcp-service-account: gsa-name-2@project-id-2.iam.gserviceaccount.com
  name: k8s-sa-name-2
`
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, expected)
}
