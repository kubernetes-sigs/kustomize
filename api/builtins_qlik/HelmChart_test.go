package builtins_qlik

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/repo/repotest"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/internal/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
	"sigs.k8s.io/yaml"
)

type mockPlugin struct {
}

func TestHelmChart(t *testing.T) {
	srv, err := repotest.NewTempServer("helmTestData/testcharts/*.tgz*")
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Stop()

	if err := srv.LinkIndices(); err != nil {
		t.Fatal(err)
	}

	// all flags will get "-d outdir" appended.
	tests := []struct {
		name            string
		pluginConfig    string
		expectedResult  string
		checkAssertions func(*testing.T, resmap.ResMap, string)
	}{
		{
			name: "Fetch, untar, template",
			pluginConfig: `
apiVersion: apps/v1
kind: HelmChart
metadata:
  name: qliksense
chartName: qliksensetest
releaseName: qliksensetest
chartRepoName: qliksensetest
chartRepo: ` + srv.URL() + `
releaseNamespace: qliksense
`, expectedResult: `
apiVersion: v1
kind: Pod
metadata:
  name: qliksensetest
ports:
- port:
    name: nginx
    protocol: UDP
    targetPort: 8081
spec:
  containers:
  - command:
    - /bin/sleep
    - "9000"
    image: alpine:3.3
    name: qliksense
  restartPolicy: Always

`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap, expectedResult string) {
				result, err := resMap.AsYaml()
				assert.NoError(t, err)

				expected, err := yaml.JSONToYAML([]byte(expectedResult))
				assert.NoError(t, err)
				assert.Equal(t, expected, result)
			},
		},
		{
			name: "Fetch, untar, template with version",
			pluginConfig: `
apiVersion: apps/v1
kind: HelmChart
metadata:
  name: qliksense
chartName: qliksensetest
chartVersion: 0.1.0
releaseName: qliksensetest
chartRepoName: qliksensetest
chartRepo: ` + srv.URL() + `
releaseNamespace: qliksense
`, expectedResult: `
apiVersion: v1
kind: Pod
metadata:
  name: qliksensetest
ports:
- port:
    name: nginx
    protocol: TCP
    targetPort: 8080
spec:
  containers:
  - command:
    - /bin/sleep
    - "9000"
    image: alpine:3.3
    name: waiter
  restartPolicy: Never
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap, expectedResult string) {
				result, err := resMap.AsYaml()
				assert.NoError(t, err)

				expected, err := yaml.JSONToYAML([]byte(expectedResult))
				assert.NoError(t, err)
				assert.Equal(t, expected, result)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			resourceFactory := resmap.NewFactory(resource.NewFactory(
				kunstruct.NewKunstructuredFactoryImpl()), transformer.NewFactoryImpl())

			plugin := NewHelmChartPlugin()
			err = plugin.Config(resmap.NewPluginHelpers((loader.NewFileLoaderAtRoot(filesys.MakeFsInMemory())), valtest_test.MakeFakeValidator(), resourceFactory), []byte(testCase.pluginConfig))
			if err != nil {
				t.Fatalf("Err: %v", err)
			}

			resMap, err := plugin.Generate()
			if err != nil {
				t.Fatalf("Err: %v", err)
			}

			for _, res := range resMap.Resources() {
				fmt.Printf("--res: %v\n", res.String())
			}
			testCase.checkAssertions(t, resMap, testCase.expectedResult)
		})
	}
}
