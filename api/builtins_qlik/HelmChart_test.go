package builtins_qlik

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
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

	if err := srv.LinkIndices(); err != nil {
		t.Fatal(err)
	}

	// all flags will get "-d outdir" appended.
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

			err = os.Mkdir(path.Join(testHome, chartName), os.ModePerm)
			assert.NoError(t, err)

			err = ioutil.WriteFile(path.Join(testHome, chartName, "Chart.yaml"), []byte(fmt.Sprintf(`
apiVersion: v1
description: A Helm chart for foo
name: %v
version: 0.1.0
`, chartName)), os.ModePerm)
			assert.NoError(t, err)

			err = ioutil.WriteFile(path.Join(testHome, chartName, "values.yaml"), []byte(`
qliksensetest:
  enabled: true
`), os.ModePerm)
			assert.NoError(t, err)

			err = ioutil.WriteFile(path.Join(testHome, chartName, "requirements.yaml"), []byte(fmt.Sprintf(`
dependencies:
  - name: qliksensetest
    version: 0.1.0
    repository: "%v"
    condition: qliksensetest.enabled
`, srv.URL())), os.ModePerm)
			assert.NoError(t, err)

			lock := &chart.Lock{
				Generated: time.Now(),
				Dependencies: []*chart.Dependency{&chart.Dependency{
					Name:       "qliksensetest",
					Repository: srv.URL(),
					Version:    "0.1.0",
				}},
			}

			digest, err := hashV2Req([]*chart.Dependency{{
				Name:       "qliksensetest",
				Repository: srv.URL(),
				Version:    "0.1.0",
				Condition:  "qliksensetest.enabled",
			}})
			assert.NoError(t, err)
			lock.Digest = digest

			err = writeLock(path.Join(testHome, chartName), lock)
			assert.NoError(t, err)

			err = os.Mkdir(path.Join(testHome, chartName, "templates"), os.ModePerm)
			assert.NoError(t, err)

			err = ioutil.WriteFile(path.Join(testHome, chartName, "templates", "configMap.yaml"), []byte(fmt.Sprintf(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: %v
data:
  something: other
`, chartName)), os.ModePerm)
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

					_ = os.RemoveAll(testHome)
				},
			}
		}(),
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

			testCase.checkAssertions(t, resMap, testCase.expectedResult)
		})
	}
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
