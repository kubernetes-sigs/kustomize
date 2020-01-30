package builtins_qlik

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/internal/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
	"sigs.k8s.io/yaml"
)

func TestHelmValues(t *testing.T) {

	testCases := []struct {
		name                 string
		pluginConfig         string
		pluginInputResources string
		expectedResult       string
		checkAssertions      func(*testing.T, resmap.ResMap, string)
	}{
		{
			name: "HelmValues success",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: HelmValues
metadata:
  name: qliksense
chartName: qliksense
releaseName: qliksense
values:
  config:
    accessControl:
      testing: 1234
  qix-sessions:
    testing: true
`,
			pluginInputResources: `
apiVersion: apps/v1
kind: HelmChart
metadata:
  name: qliksense
chartName: qliksense
releaseName: qliksense
values:
  config:
    accessControl:
      testing: 4321
`,
			expectedResult: `
apiVersion: apps/v1
chartName: qliksense
kind: HelmChart
metadata:
  name: qliksense
releaseName: qliksense
values:
  config:
    accessControl:
      testing: 4321
  qix-sessions:
    testing: true
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
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resourceFactory := resmap.NewFactory(resource.NewFactory(
				kunstruct.NewKunstructuredFactoryImpl()), transformer.NewFactoryImpl())

			resMap, err := resourceFactory.NewResMapFromBytes([]byte(testCase.pluginInputResources))
			if err != nil {
				t.Fatalf("Err: %v", err)
			}

			plugin := NewHelmValuesPlugin()
			err = plugin.Config(resmap.NewPluginHelpers(loader.NewFileLoaderAtRoot(filesys.MakeFsInMemory()), valtest_test.MakeFakeValidator(), resourceFactory), []byte(testCase.pluginConfig))
			if err != nil {
				t.Fatalf("Err: %v", err)
			}

			err = plugin.Transform(resMap)
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
