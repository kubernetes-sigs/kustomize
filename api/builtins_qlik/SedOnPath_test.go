package builtins_qlik

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/builtins_qlik/utils/loadertest"
	"sigs.k8s.io/kustomize/api/internal/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
)

func TestBasicSed(t *testing.T) {
	p := SedOnPathPlugin{
		Regex: []string{"s/hello/goodbye/g"},
	}
	result, err := p.executeSed("hello there!")
	assert.NoError(t, err)
	//must compare with \n cause sed automatically adds
	assert.Equal(t, "goodbye there!\n", result)
}

func TestSedOnPath(t *testing.T) {
	basicPluginInputResources := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: the-deployment
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: the-container-1
        image: the-image-1:1
        command: 
        - /foo
        - --port=8080
        - --bar=baz
        - --TempContentServiceUrl=http://temporary-contents:6080
      - name: the-container-2
        image: the-image-2:1
        command: 
        - /abra
        - --port=8080
        - --cadabra=bam
        - --TempContentServiceUrl=http://temporary-contents:6080
`

	testCases := []struct {
		name                    string
		pluginConfig            string
		pluginInputResources    string
		expectingTransformError bool
		checkAssertions         func(*testing.T, resmap.ResMap)
	}{
		{
			name: "value_at_path_is_map",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SedOnPath
metadata:
  name: notImportantHere
path: spec/template
regex:
- s/.*//g
`,
			pluginInputResources:    basicPluginInputResources,
			expectingTransformError: true,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				assert.FailNow(t, "should not be here!")
			},
		},
		{
			name: "value_at_path_is_string",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SedOnPath
metadata:
  name: notImportantHere
path: spec/template/spec/containers/name
regex:
- s/the-container/the-awesome-container/g
`,
			pluginInputResources:    basicPluginInputResources,
			expectingTransformError: false,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				res := resMap.GetByIndex(0)
				assert.NotNil(t, res)

				containers, err := res.GetFieldValue("spec.template.spec.containers")
				assert.NoError(t, err)

				for _, container := range containers.([]interface{}) {
					assert.True(t, strings.HasPrefix(container.(map[string]interface{})["name"].(string), "the-awesome-container-"))
				}
			},
		},
		{
			name: "value_at_path_is_array_of_strings",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SedOnPath
metadata:
  name: notImportantHere
path: spec/template/spec/containers/command
regex:
- s/--port=.*$/--port=1234/g
- s/--TempContentServiceUrl=.*$/--TempContentServiceUrl=http:\/\/\$\(PREFIX\)-contents:6080/g
`,
			pluginInputResources:    basicPluginInputResources,
			expectingTransformError: false,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				res := resMap.GetByIndex(0)
				assert.NotNil(t, res)

				containers, err := res.GetFieldValue("spec.template.spec.containers")
				assert.NoError(t, err)

				portArgSubCounter := 0
				for _, container := range containers.([]interface{}) {
					args := container.(map[string]interface{})["command"].([]interface{})
					for _, arg := range args {
						zArg := arg.(string)
						if strings.TrimSpace(zArg) == "--port=1234" {
							portArgSubCounter++
						}
					}
				}
				assert.Equal(t, 2, portArgSubCounter)

				tempContentServiceUrlSubCounter := 0
				for _, container := range containers.([]interface{}) {
					args := container.(map[string]interface{})["command"].([]interface{})
					for _, arg := range args {
						zArg := arg.(string)
						if strings.TrimSpace(zArg) == "--TempContentServiceUrl=http://$(PREFIX)-contents:6080" {
							tempContentServiceUrlSubCounter++
						}
					}
				}
				assert.Equal(t, 2, tempContentServiceUrlSubCounter)
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

			plugin := NewSedOnPathPlugin()
			err = plugin.Config(resmap.NewPluginHelpers(loadertest.NewFakeLoader("/"), valtest_test.MakeFakeValidator(), resourceFactory), []byte(testCase.pluginConfig))
			if err != nil {
				t.Fatalf("Err: %v", err)
			}

			err = plugin.Transform(resMap)
			if err != nil && testCase.expectingTransformError {
				return
			}

			if err != nil {
				t.Fatalf("Err: %v", err)
			}

			for _, res := range resMap.Resources() {
				fmt.Printf("--res: %v\n", res.String())
			}

			testCase.checkAssertions(t, resMap)
		})
	}
}
