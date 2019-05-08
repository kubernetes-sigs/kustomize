package builtin

import (
	"testing"

	"sigs.k8s.io/kustomize/internal/plugintest"
	"sigs.k8s.io/kustomize/k8sdeps/kv/plugin"
	"sigs.k8s.io/kustomize/pkg/kusttest"
	"sigs.k8s.io/kustomize/pkg/loader"
)

func TestConfigMapGenerator(t *testing.T) {
	tc := plugintest_test.NewPluginTestEnv(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "ConfigMapGenerator")

	th := kusttest_test.NewKustTestHarnessFull(
		t, "/app", loader.RestrictionRootOnly, plugin.ActivePluginConfig())

	th.WriteF("/app/devops.env", `
SERVICE_PORT=32
`)
	th.WriteF("/app/uxteam.env", `
COLOR=red
`)

	rm := th.LoadAndRunGenerator(`
apiVersion: builtin
kind: ConfigMapGenerator
metadata:
  name: myMap
envFiles:
- devops.env
- uxteam.env
literals:
- FRUIT=apple
- VEGETABLE=carrot
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
data:
  COLOR: red
  FRUIT: apple
  SERVICE_PORT: "32"
  VEGETABLE: carrot
kind: ConfigMap
metadata:
  name: myMap
`)
}
