package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

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
