package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestMultibyteCharInConfigMap(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
  - resources.yaml
`)
	th.WriteF("resources.yaml", `
apiVersion: v1
kind: ConfigMap
metadata:
  name: game-config
data:
  key: あ 
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  key: あ
kind: ConfigMap
metadata:
  name: game-config
`)
}
