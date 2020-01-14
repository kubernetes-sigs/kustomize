package builtins_qlik

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/builtins_qlik/utils/loadertest"
	"sigs.k8s.io/kustomize/api/internal/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
	"sigs.k8s.io/yaml"
)

func TestValuesFile(t *testing.T) {
	valuesFileContent := `
  config:
    accessControl:
      testing: 1234
    qix-sessions:
      testing: true
    test123:
      working: 123
`

	var testCases = []struct {
		name                 string
		pluginConfig         string
		pluginInputResources string
		valuesFileContent    string
		expectedResult       string
		checkAssertions      func(*testing.T, resmap.ResMap, string)
	}{
		{
			name: "ValuesFile success",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: ValuesFile
metadata:
  name: qliksense
enabled: true
valuesFile: values.tml.yaml
dataSource: 
  ejson:
    filePath: test.json
`,
			valuesFileContent: valuesFileContent,
			pluginInputResources: `
apiVersion: apps/v1
kind: HelmValues
metadata:
  name: collections
values:
  config:
    qix-sessions:
      testing: false
`,
			expectedResult: `
apiVersion: apps/v1
kind: HelmValues
metadata:
  name: collections
values:
  config:
    accessControl:
      testing: 1234
    qix-sessions:
      testing: true
    test123:
      working: 123
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

			ldr := loadertest.NewFakeLoader("/")
			if len(testCase.valuesFileContent) > 0 {
				err = ldr.AddFile("/values.tml.yaml", []byte(testCase.valuesFileContent))
				if err != nil {
					t.Fatalf("Err: %v", err)
				}
			}

			plugin := NewValuesFilePlugin()
			err = plugin.Config(resmap.NewPluginHelpers(ldr, valtest_test.MakeFakeValidator(), resourceFactory), []byte(testCase.pluginConfig))
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

			if err == nil {
				testCase.checkAssertions(t, resMap, testCase.expectedResult)
			}
		})
	}
}
