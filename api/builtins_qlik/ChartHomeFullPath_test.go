package builtins_qlik

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/builtins_qlik/utils/loadertest"
	"sigs.k8s.io/kustomize/api/internal/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
)

func TestChartHomeFullPath(t *testing.T) {

	testCases := []struct {
		name                 string
		pluginConfig         func(dir string) string
		pluginInputResources string
		checkAssertions      func(t *testing.T, res resmap.ResMap, dir string, fileContents []byte)
	}{
		{
			name: "Chart home created and file contents copied",
			pluginConfig: func(dir string) string {
				config := fmt.Sprintf(`
apiVersion: qlik.com/v1
kind: ChartHomeFullPath
metadata:
  name: qliksense
chartHome: %v
`, dir)
				return config
			},
			pluginInputResources: `
apiVersion: apps/v1
kind: HelmChart
metadata:
  name: qliksense
chartName: qliksense
releaseName: qliksense
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap, dir string, fileContents []byte) {
				res := resMap.GetByIndex(0)

				chartHome, err := res.GetFieldValue("chartHome")
				assert.NoError(t, err)

				assert.NotEqual(t, dir, chartHome)

				//open modified directory
				directory, err := os.Open(chartHome.(string))
				assert.NoError(t, err)
				objects, err := directory.Readdir(-1)
				assert.NoError(t, err)

				//check the temp file was coppied over correctly
				for _, obj := range objects {
					source := chartHome.(string) + "/" + obj.Name()
					readFileContents, err := ioutil.ReadFile(source)
					assert.NoError(t, err)
					assert.Equal(t, fileContents, readFileContents)
				}
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// create a temp directory and test file
			dir, err := ioutil.TempDir("", "test")
			assert.NoError(t, err)

			file, err := ioutil.TempFile(dir, "testFile")
			assert.NoError(t, err)
			defer file.Close()

			fileContents := []byte("test")
			_, err = file.Write(fileContents)
			assert.NoError(t, err)

			resourceFactory := resmap.NewFactory(resource.NewFactory(
				kunstruct.NewKunstructuredFactoryImpl()), transformer.NewFactoryImpl())

			resMap, err := resourceFactory.NewResMapFromBytes([]byte(testCase.pluginInputResources))
			if err != nil {
				t.Fatalf("Err: %v", err)
			}

			plugin := NewChartHomeFullPathPlugin()
			err = plugin.Config(resmap.NewPluginHelpers(loadertest.NewFakeLoader("/"), valtest_test.MakeFakeValidator(), resourceFactory), []byte(testCase.pluginConfig(dir)))
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

			testCase.checkAssertions(t, resMap, dir, fileContents)
		})
	}
}
