package krusty_test

import (
	"strings"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestIssue3489(t *testing.T) {
	const assets = `{
	"tenantId": "XXXXX-XXXXXX-XXXXX-XXXXXX-XXXXXX",
	"subscriptionId": "XXXXX-XXXXXX-XXXXX-XXXXXX-XXXXXX",
	"resourceGroup": "DNS-EUW-XXX-RG",
	"useManagedIdentityExtension": true,
	"userAssignedIdentityID": "XXXXX-XXXXXX-XXXXX-XXXXXX-XXXXXX"
}
`
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
namespace: kube-system
resources:
- external-dns
- external-dns-private
`)
	th.WriteK("external-dns", `
resources:
- ../base
commonLabels:
  app: external-dns
  instance: public
images:
- name: k8s.gcr.io/external-dns/external-dns
  newName: xxx.azurecr.io/external-dns
  newTag: v0.7.4_sylr.1
- name: quay.io/sylr/external-dns
  newName: xxx.azurecr.io/external-dns
  newTag: v0.7.4_sylr.1
secretGenerator:
- name: azure-config-file
  behavior: replace
  files:
  - assets/azure.json
patches:
- target:
    group: apps
    version: v1
    kind: Deployment
    name: external-dns
  patch: |-
    - op: replace
      path: /spec/template/spec/containers/0/args
      value:
      - --txt-owner-id="aks"
      - --txt-prefix=external-dns-
      - --source=service
      - --provider=azure
      - --registry=txt
      - --domain-filter=dev.company.com
`)

	th.WriteF("external-dns/assets/azure.json", assets)
	th.WriteK("external-dns-private", `
resources:
- ../base
nameSuffix: -private
commonLabels:
  app: external-dns
  instance: private
images:
- name: k8s.gcr.io/external-dns/external-dns
  newName: xxx.azurecr.io/external-dns
  newTag: v0.7.4_sylr.1
- name: quay.io/sylr/external-dns
  newName: xxx.azurecr.io/external-dns
  newTag: v0.7.4_sylr.1
secretGenerator:
- name: azure-config-file
  behavior: replace
  files:
  - assets/azure.json
patches:
- target:
    group: apps
    version: v1
    kind: Deployment
    name: external-dns
  patch: |-
    - op: replace
      path: /spec/template/spec/containers/0/args
      value:
      - --txt-owner-id="aks"
      - --txt-prefix=external-dns-private-
      - --source=service
      - --provider=azure-private-dns
      - --registry=txt
      - --domain-filter=static.company.az
`)
	th.WriteF("external-dns-private/assets/azure.json", assets)
	th.WriteK("base", `
resources:
- clusterrole.yaml
- clusterrolebinding.yaml
- deployment.yaml
- serviceaccount.yaml
commonLabels:
  app: external-dns
  instance: public
images:
- name: k8s.gcr.io/external-dns/external-dns
  newName: quay.io/sylr/external-dns
  newTag: v0.7.4-73-g00a9a0c7
secretGenerator:
- name: azure-config-file
  files:
  - assets/azure.json
`)
	th.WriteF("base/assets/azure.json", assets)
	th.WriteF("base/clusterrolebinding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: external-dns-viewer
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: external-dns
subjects:
- kind: ServiceAccount
  name: external-dns
`)
	th.WriteF("base/clusterrole.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: external-dns
rules:
- apiGroups: ['']
  resources: ['endpoints', 'pods', 'services', 'nodes']
  verbs: ['get', 'watch', 'list']
- apiGroups: ['extensions', 'networking.k8s.io']
  resources: ['ingresses']
  verbs: ['get', 'watch', 'list']
`)
	th.WriteF("base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: external-dns
spec:
  strategy:
    type: Recreate
  selector:
    matchLabels: {}
  template:
    metadata: {}
    spec:
      serviceAccountName: external-dns
      containers:
      - name: external-dns
        image: k8s.gcr.io/external-dns/external-dns
        args:
        - --domain-filter=""
        - --txt-owner-id=""
        - --txt-prefix=external-dns-
        - --source=service
        - --provider=azure
        - --registry=txt
        resources: {}
        volumeMounts:
        - name: azure-config-file
          mountPath: /etc/kubernetes
          readOnly: true
      volumes:
      - name: azure-config-file
        secret:
          secretName: azure-config-file
`)
	th.WriteF("base/serviceaccount.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: external-dns
`)
	// TODO(3489): This shouldn't be an error.
	err := th.RunWithErr(".", th.MakeDefaultOptions())
	if !strings.Contains(err.Error(), "found multiple possible referrals") {
		t.Fatalf("unexpected error: %q", err)
	}
}

func TestEmptyFieldSpecValue(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK("/app", `
generators:
- generators.yaml
configurations:
- kustomizeconfig.yaml
`)
	th.WriteF("/app/generators.yaml", `
apiVersion: builtin
kind: ConfigMapGenerator
metadata:
  name: secret-example
labels:
  app.kubernetes.io/name: secret-example
literals:
- this_is_a_secret_name=
`)
	th.WriteF("/app/kustomizeconfig.yaml", `
nameReference:
- kind: Secret
  version: v1
  fieldSpecs:
  - path: data/this_is_a_secret_name
    kind: ConfigMap
`)
	m := th.Run("/app", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  this_is_a_secret_name: ""
kind: ConfigMap
metadata:
  name: secret-example-7hf4fh868h
`)
}
