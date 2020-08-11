package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestLongLineBreaks(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteF("/app/deployment.yaml", `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: test
  labels:
    app.version: "{{ .Values.services.mariadb.app.version }}"
spec:
  template:
    spec:
      containers:
      - name: mariadb
        image: "thisIsAReallyLongRepositoryLinkThatResultsInALineBreakWhenBuildingWithKustomize/mariadb:{{ .Values.services.mariadb.image.version }}"
`)
	th.WriteK("/app", `
resources:
- deployment.yaml
`)
	m := th.Run("/app", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    app.version: '{{ .Values.services.mariadb.app.version }}'
  name: test
spec:
  template:
    spec:
      containers:
      - image: thisIsAReallyLongRepositoryLinkThatResultsInALineBreakWhenBuildingWithKustomize/mariadb:{{ .Values.services.mariadb.image.version }}
        name: mariadb
`)
}
