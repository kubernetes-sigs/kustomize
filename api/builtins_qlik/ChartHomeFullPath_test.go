package builtins_qlik

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/internal/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
)

func TestChartHomeFullPath(t *testing.T) {

	testCases := []struct {
		name                 string
		pluginConfig         func(dir string) string
		pluginInputResources func(dir string) string
		checkAssertions      func(t *testing.T, res resmap.ResMap, dir string, fileContents []byte)
	}{
		{
			name: "Chart home created and file contents copied",
			pluginConfig: func(dir string) string {
				config := fmt.Sprintf(`
apiVersion: qlik.com/v1
kind: ChartHomeFullPath
metadata:
  name: test`)
				return config
			},
			pluginInputResources: func(dir string) string {
				inputResources := fmt.Sprintf(`
apiVersion: apps/v1
kind: HelmChart
metadata:
  name: test
chartName: test
releaseName: test
chartHome: %v`, dir)
				return inputResources
			},
			checkAssertions: func(t *testing.T, resMap resmap.ResMap, dir string, fileContents []byte) {

				res := resMap.GetByIndex(0)

				dir, err := konfig.DefaultAbsPluginHome(filesys.MakeFsOnDisk())
				if err != nil {
					dir = filepath.Join(konfig.HomeDir(), konfig.XdgConfigHomeEnvDefault, konfig.ProgramName, konfig.RelPluginHome)
				}
				chartHome := filepath.Join(dir, "qlik", "v1", "charts")
				chartName, err := res.GetFieldValue("chartName")
				assert.NoError(t, err)

				directoryName := filepath.Join(chartHome, filepath.Join(chartName.(string)))
				//open modified directory
				directory, err := os.Open(directoryName)
				assert.NoError(t, err)
				objects, err := directory.Readdir(-1)
				assert.NoError(t, err)
				defer os.RemoveAll(directoryName)
				//check the temp file was coppied over correctly
				for _, obj := range objects {
					source := filepath.Join(directoryName, obj.Name())
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

			dir, err := ioutil.TempDir("", "charts")
			assert.NoError(t, err)

			os.Mkdir(filepath.Join(dir, "test"), 0777)
			assert.NoError(t, err)

			file, err := ioutil.TempFile(dir, "testFile")
			assert.NoError(t, err)
			defer file.Close()

			fileContents := []byte("test")
			_, err = file.Write(fileContents)
			assert.NoError(t, err)

			resourceFactory := resmap.NewFactory(resource.NewFactory(
				kunstruct.NewKunstructuredFactoryImpl()), transformer.NewFactoryImpl())

			resMap, err := resourceFactory.NewResMapFromBytes([]byte(testCase.pluginInputResources(dir)))
			if err != nil {
				t.Fatalf("Err: %v", err)
			}

			plugin := NewChartHomeFullPathPlugin()
			err = plugin.Config(resmap.NewPluginHelpers(loader.NewFileLoaderAtRoot(filesys.MakeFsInMemory()), valtest_test.MakeFakeValidator(), resourceFactory), []byte(testCase.pluginConfig(dir)))
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
