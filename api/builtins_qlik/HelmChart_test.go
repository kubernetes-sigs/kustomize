package builtins_qlik

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/provenance"
	"helm.sh/helm/v3/pkg/repo/repotest"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/internal/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
	"sigs.k8s.io/yaml"
)

type HelmChartTestCase struct {
	name            string
	pluginConfig    string
	expectedResult  string
	checkAssertions func(*testing.T, resmap.ResMap, string)
}

func TestHelmChart(t *testing.T) {
	srv, err := repotest.NewTempServer("helmTestData/testcharts/*.tgz*")
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Stop()

	tests := []HelmChartTestCase{
		func() HelmChartTestCase {
			testHome, err := ioutil.TempDir("", "")
			assert.NoError(t, err)

			return HelmChartTestCase{
				name: "Fetch, untar, template",
				pluginConfig: fmt.Sprintf(`
apiVersion: apps/v1
kind: HelmChart
metadata:
  name: qliksense
chartHome: %v
helmHome: %v
chartName: qliksensetest
releaseName: qliksensetest
chartRepoName: qliksensetest
chartRepo: %v
releaseNamespace: qliksense
`, testHome, testHome, srv.URL()),
				expectedResult: `
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

					_ = os.RemoveAll(testHome)
				},
			}
		}(),
		func() HelmChartTestCase {
			testHome, err := ioutil.TempDir("", "")
			assert.NoError(t, err)

			return HelmChartTestCase{
				name: "Fetch, untar, template with version",
				pluginConfig: fmt.Sprintf(`
apiVersion: apps/v1
kind: HelmChart
metadata:
  name: qliksense
chartHome: %v
helmHome: %v
chartName: qliksensetest
chartVersion: 0.1.0
releaseName: qliksensetest
chartRepoName: qliksensetest
chartRepo: %v
releaseNamespace: qliksense
`, testHome, testHome, srv.URL()),
				expectedResult: `
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

					_ = os.RemoveAll(testHome)
				},
			}
		}(),
		func() HelmChartTestCase {
			testHome, err := ioutil.TempDir("", "")
			assert.NoError(t, err)

			chartName := "foo"
			err = writeTestChartWithDeps(testHome, chartName, srv.URL(), "0.1.0")
			assert.NoError(t, err)

			return HelmChartTestCase{
				name: "Local chart with dependencies",
				pluginConfig: fmt.Sprintf(`
apiVersion: qlik.com/v1
chartHome: %v
chartName: %v
helmHome: %v
kind: HelmChart
metadata:
  name: %v
values:
  qliksensetest:
    enabled: true
`, testHome, chartName, testHome, chartName),
				expectedResult: `
apiVersion: v1
data:
  something: other
kind: ConfigMap
metadata:
  name: foo
---
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
				checkAssertions: func(t *testing.T, resMap resmap.ResMap, _ string) {
					//resMapYaml, err := resMap.AsYaml()
					//assert.NoError(t, err)
					//fmt.Printf("%v\n", string(resMapYaml))

					resConfigMap, err := resMap.GetById(resid.NewResId(resid.Gvk{
						Group:   "",
						Version: "v1",
						Kind:    "ConfigMap",
					}, chartName))
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					assert.NoError(t, err)
					assert.NotNil(t, resConfigMap)

					resDependentPod, err := resMap.GetById(resid.NewResId(resid.Gvk{
						Group:   "",
						Version: "v1",
						Kind:    "Pod",
					}, "qliksensetest"))
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					assert.NoError(t, err)
					assert.NotNil(t, resDependentPod)

					//_ = os.RemoveAll(testHome)
				},
			}
		}(),
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			resourceFactory := resmap.NewFactory(resource.NewFactory(
				kunstruct.NewKunstructuredFactoryImpl()), transformer.NewFactoryImpl())

			plugin := NewHelmChartPlugin()
			err = plugin.Config(resmap.NewPluginHelpers(loader.NewFileLoaderAtRoot(filesys.MakeFsInMemory()), valtest_test.MakeFakeValidator(), resourceFactory), []byte(testCase.pluginConfig))
			if err != nil {
				t.Fatalf("Err: %v", err)
			}

			resMap, err := plugin.Generate()
			if err != nil {
				t.Fatalf("Err: %v", err)
			}

			testCase.checkAssertions(t, resMap, testCase.expectedResult)
		})
	}
}

func TestHelmChart_withFetch_withDeps_contendingOnSameCharts(t *testing.T) {
	srv, err := repotest.NewTempServer("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer srv.Stop()

	tempDir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	fooChartName := "foo"
	err = writeTestChartWithDeps(tempDir, fooChartName, srv.URL(), "0.1.0")
	assert.NoError(t, err)
	err = doTarGz(filepath.Join(tempDir, fooChartName), filepath.Join(tempDir, fmt.Sprintf("%v.tgz", fooChartName)))
	assert.NoError(t, err)

	barChartName := "bar"
	err = writeTestChartWithDeps(tempDir, barChartName, srv.URL(), "0.2.0")
	assert.NoError(t, err)
	err = doTarGz(filepath.Join(tempDir, barChartName), filepath.Join(tempDir, fmt.Sprintf("%v.tgz", barChartName)))
	assert.NoError(t, err)

	if _, err := srv.CopyCharts(filepath.Join(tempDir, "*.tgz")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := srv.CopyCharts("helmTestData/testcharts/*.tgz*"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testHome, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(testHome)

	pluginConfigChartFoo := fmt.Sprintf(`
  apiVersion: apps/v1
  kind: HelmChart
  metadata:
    name: halmChart-foo
  chartHome: %v
  helmHome: %v
  chartName: foo
  chartRepo: %v
  releaseNamespace: qliksense
  values:
    qliksensetest:
      enabled: true
`, testHome, testHome, srv.URL())
	expectedGeneratorOutputFoo := `apiVersion: v1
data:
  something: other
kind: ConfigMap
metadata:
  name: foo
---
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
`

	pluginConfigChartBar := fmt.Sprintf(`
  apiVersion: apps/v1
  kind: HelmChart
  metadata:
    name: halmChart-bar
  chartHome: %v
  helmHome: %v
  chartName: bar
  chartRepo: %v
  values:
    qliksensetest:
      enabled: true
`, testHome, testHome, srv.URL())
	expectedGeneratorOutputBar := `apiVersion: v1
data:
  something: other
kind: ConfigMap
metadata:
  name: bar
---
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
`

	logger := log.New(os.Stdout, "", log.LstdFlags|log.LUTC|log.Lmicroseconds|log.Lshortfile)
	resourceFactory := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()), transformer.NewFactoryImpl())
	pluginHelpers := resmap.NewPluginHelpers(loader.NewFileLoaderAtRoot(filesys.MakeFsInMemory()), valtest_test.MakeFakeValidator(), resourceFactory)

	var wg sync.WaitGroup
	numTests := 50
	for i := 0; i < numTests; i++ {
		var config, expected string
		if i%2 == 0 {
			fmt.Println("testing foo...")
			config = pluginConfigChartFoo
			expected = expectedGeneratorOutputFoo
		} else {
			fmt.Println("testing bar...")
			config = pluginConfigChartBar
			expected = expectedGeneratorOutputBar
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			plugin := &HelmChartPlugin{logger: logger}
			err = plugin.Config(pluginHelpers, []byte(config))
			if err != nil {
				t.Fatalf("Err: %v", err)
			}

			resMap, err := plugin.Generate()
			if err != nil {
				t.Fatalf("Err: %v", err)
			}

			resMapYaml, err := resMap.AsYaml()
			assert.NoError(t, err)
			assert.Equal(t, expected, string(resMapYaml))
		}()
	}
	wg.Wait()
}

func TestHelmChart_withFetch_withDeps_contendingOnDiffCharts(t *testing.T) {
	srv, err := repotest.NewTempServer("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer srv.Stop()

	tempDir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testHome, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(testHome)

	numTests := 50
	pluginConfigs := make([]string, 0)
	expectedResults := make([]string, 0)
	for i := 0; i < numTests; i++ {
		fooChartName := fmt.Sprintf("foo-%v", i)
		var depVersion string
		if i%2 == 0 {
			depVersion = "0.1.0"
		} else {
			depVersion = "0.2.0"
		}
		err = writeTestChartWithDeps(tempDir, fooChartName, srv.URL(), depVersion)
		assert.NoError(t, err)
		err = doTarGz(filepath.Join(tempDir, fooChartName), filepath.Join(tempDir, fmt.Sprintf("%v.tgz", fooChartName)))
		assert.NoError(t, err)
		pluginConfigs = append(pluginConfigs, fmt.Sprintf(`
  apiVersion: apps/v1
  kind: HelmChart
  metadata:
    name: halmChart-foo-%v
  chartHome: %v
  helmHome: %v
  chartName: foo-%v
  chartRepo: %v
  releaseNamespace: qliksense
  values:
    qliksensetest:
      enabled: true
`, i, testHome, testHome, i, srv.URL()))
		var expectedResult string
		if i%2 == 0 {
			expectedResult = fmt.Sprintf(`apiVersion: v1
data:
  something: other
kind: ConfigMap
metadata:
  name: foo-%v
---
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
`, i)
		} else {
			expectedResult = fmt.Sprintf(`apiVersion: v1
data:
  something: other
kind: ConfigMap
metadata:
  name: foo-%v
---
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
`, i)
		}
		expectedResults = append(expectedResults, expectedResult)
	}

	if _, err := srv.CopyCharts(filepath.Join(tempDir, "*.tgz")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := srv.CopyCharts("helmTestData/testcharts/*.tgz*"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logger := log.New(os.Stdout, "", log.LstdFlags|log.LUTC|log.Lmicroseconds|log.Lshortfile)
	resourceFactory := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()), transformer.NewFactoryImpl())
	pluginHelpers := resmap.NewPluginHelpers(loader.NewFileLoaderAtRoot(filesys.MakeFsInMemory()), valtest_test.MakeFakeValidator(), resourceFactory)

	log.Println("starting tests...")

	var wg sync.WaitGroup
	for i := 0; i < numTests; i++ {
		pluginConfig := pluginConfigs[i]
		expectedResult := expectedResults[i]

		wg.Add(1)
		go func() {
			defer wg.Done()
			plugin := &HelmChartPlugin{logger: logger}
			err = plugin.Config(pluginHelpers, []byte(pluginConfig))
			if err != nil {
				t.Fatalf("Err: %v", err)
			}

			resMap, err := plugin.Generate()
			if err != nil {
				t.Fatalf("Err: %v", err)
			}

			resMapYaml, err := resMap.AsYaml()
			assert.NoError(t, err)
			assert.Equal(t, expectedResult, string(resMapYaml))
		}()
	}
	wg.Wait()
}

func TestHelmChart_noFetch_withDeps_contendingOnDiffCharts(t *testing.T) {
	srv, err := repotest.NewTempServer("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer srv.Stop()

	testHome, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(testHome)

	numTests := 50
	pluginConfigs := make([]string, 0)
	expectedResults := make([]string, 0)
	for i := 0; i < numTests; i++ {
		fooChartName := fmt.Sprintf("foo-%v", i)
		var depVersion string
		if i%2 == 0 {
			depVersion = "0.1.0"
		} else {
			depVersion = "0.2.0"
		}
		err = writeTestChartWithDeps(testHome, fooChartName, srv.URL(), depVersion)
		assert.NoError(t, err)
		pluginConfigs = append(pluginConfigs, fmt.Sprintf(`
apiVersion: apps/v1
kind: HelmChart
metadata:
  name: halmChart-foo-%v
chartHome: %v
helmHome: %v
chartName: foo-%v
chartRepo: %v
releaseNamespace: qliksense
values:
  qliksensetest:
    enabled: true
`, i, testHome, testHome, i, srv.URL()))
		var expectedResult string
		if i%2 == 0 {
			expectedResult = fmt.Sprintf(`apiVersion: v1
data:
  something: other
kind: ConfigMap
metadata:
  name: foo-%v
---
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
`, i)
		} else {
			expectedResult = fmt.Sprintf(`apiVersion: v1
data:
  something: other
kind: ConfigMap
metadata:
  name: foo-%v
---
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
`, i)
		}
		expectedResults = append(expectedResults, expectedResult)
	}

	if _, err := srv.CopyCharts("helmTestData/testcharts/*.tgz*"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logger := log.New(os.Stdout, "", log.LstdFlags|log.LUTC|log.Lmicroseconds|log.Lshortfile)
	resourceFactory := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()), transformer.NewFactoryImpl())
	pluginHelpers := resmap.NewPluginHelpers(loader.NewFileLoaderAtRoot(filesys.MakeFsInMemory()), valtest_test.MakeFakeValidator(), resourceFactory)

	log.Println("starting tests...")

	var wg sync.WaitGroup
	for i := 0; i < numTests; i++ {
		pluginConfig := pluginConfigs[i]
		expectedResult := expectedResults[i]

		wg.Add(1)
		go func() {
			defer wg.Done()
			plugin := &HelmChartPlugin{logger: logger}
			err = plugin.Config(pluginHelpers, []byte(pluginConfig))
			if err != nil {
				t.Fatalf("Err: %v", err)
			}

			resMap, err := plugin.Generate()
			if err != nil {
				t.Fatalf("Err: %v", err)
			}

			resMapYaml, err := resMap.AsYaml()
			assert.NoError(t, err)
			assert.Equal(t, expectedResult, string(resMapYaml))
		}()
	}
	wg.Wait()
}

func hashV2Req(req []*chart.Dependency) (string, error) {
	dep := make(map[string][]*chart.Dependency)
	dep["dependencies"] = req
	data, err := json.Marshal(dep)
	if err != nil {
		return "", err
	}
	s, err := provenance.Digest(bytes.NewBuffer(data))
	return "sha256:" + s, err
}

func writeLock(chartpath string, lock *chart.Lock) error {
	data, err := yaml.Marshal(lock)
	if err != nil {
		return err
	}
	dest := filepath.Join(chartpath, "requirements.lock")
	return ioutil.WriteFile(dest, data, 0644)
}

func writeTestChartWithDeps(chartHome, chartName, depRepoUrl, depVersion string) error {
	err := os.Mkdir(path.Join(chartHome, chartName), os.ModePerm)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(chartHome, chartName, "Chart.yaml"), []byte(fmt.Sprintf(`
apiVersion: v1
description: A Helm chart for %v
name: %v
version: 0.1.0
`, chartName, chartName)), os.ModePerm)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(chartHome, chartName, "values.yaml"), []byte(`
qliksensetest:
  enabled: false
`), os.ModePerm)
	if err != nil {
		return err
	}

	dep := chart.Dependency{
		Name:       "qliksensetest",
		Repository: depRepoUrl,
		Version:    depVersion,
		Condition:  "qliksensetest.enabled",
	}

	requirements := map[string][]*chart.Dependency{"dependencies": {&dep}}
	requirementsYamlBytes, err := yaml.Marshal(&requirements)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path.Join(chartHome, chartName, "requirements.yaml"), requirementsYamlBytes, os.ModePerm)
	if err != nil {
		return err
	}

	lock := &chart.Lock{
		Generated:    time.Now(),
		Dependencies: []*chart.Dependency{&dep},
	}

	digest, err := hashV2Req([]*chart.Dependency{&dep})
	if err != nil {
		return err
	}
	lock.Digest = digest

	err = writeLock(path.Join(chartHome, chartName), lock)
	if err != nil {
		return err
	}

	err = os.Mkdir(path.Join(chartHome, chartName, "templates"), os.ModePerm)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(chartHome, chartName, "templates", "configMap.yaml"), []byte(fmt.Sprintf(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: %v
data:
  something: other
`, chartName)), os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func doTarGz(src string, dest string) error {
	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	gzipWriter := gzip.NewWriter(destFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// walk through every file in the folder
	parentDir := filepath.Dir(src)
	if err := filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}
		header.Name = strings.TrimPrefix(strings.Replace(file, parentDir, "", 1), string(filepath.Separator))
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if !fi.IsDir() {
			if fileBytes, err := ioutil.ReadFile(file); err != nil {
				return err
			} else if _, err := tarWriter.Write(fileBytes); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
