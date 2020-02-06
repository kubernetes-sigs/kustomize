package builtins_qlik

import (
	"fmt"
	"testing"

	"sigs.k8s.io/kustomize/api/builtins_qlik/utils"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/internal/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
)

func TestSearchReplacePlugin(t *testing.T) {

	testCases := []struct {
		name                 string
		pluginConfig         string
		pluginInputResources string
		checkAssertions      func(*testing.T, resmap.ResMap)
	}{
		{
			name: "relaxed is dangerous",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SearchReplace
metadata:
  name: notImportantHere
target:
  kind: Foo
  name: some-foo
path: fooSpec/fooTemplate/fooContainers/env/value
search: far
replace: not far
`,
			pluginInputResources: `
apiVersion: qlik.com/v1
kind: Foo
metadata:
  name: some-foo
fooSpec:
  fooTemplate:
    fooContainers:
    - name: have-env
      env:
      - name: FOO
        value: far
      - name: BOO
        value: farther than it looks
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				res := resMap.GetByIndex(0)

				envVars, err := res.GetFieldValue("fooSpec.fooTemplate.fooContainers[0].env")
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				fooEnvVar := envVars.([]interface{})[0].(map[string]interface{})
				if "FOO" != fooEnvVar["name"].(string) {
					t.Fatalf("unexpected: %v\n", fooEnvVar["name"].(string))
				}
				if "not far" != fooEnvVar["value"].(string) {
					t.Fatalf("unexpected: %v\n", fooEnvVar["value"].(string))
				}

				booEnvVar := envVars.([]interface{})[1].(map[string]interface{})
				if "BOO" != booEnvVar["name"].(string) {
					t.Fatalf("unexpected: %v\n", booEnvVar["name"].(string))
				}
				if "not farther than it looks" != booEnvVar["value"].(string) {
					t.Fatalf("unexpected: %v\n", booEnvVar["value"].(string))
				}
			},
		},
		{
			name: "strict is safer",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SearchReplace
metadata:
  name: notImportantHere
target:
  kind: Foo
  name: some-foo
path: fooSpec/fooTemplate/fooContainers/env/value
search: ^far$
replace: not far
`,
			pluginInputResources: `
apiVersion: qlik.com/v1
kind: Foo
metadata:
  name: some-foo
fooSpec:
  fooTemplate:
    fooContainers:
    - name: have-env
      env:
      - name: FOO
        value: far
      - name: BOO
        value: farther than it looks
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				res := resMap.GetByIndex(0)

				envVars, err := res.GetFieldValue("fooSpec.fooTemplate.fooContainers[0].env")
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				fooEnvVar := envVars.([]interface{})[0].(map[string]interface{})
				if "FOO" != fooEnvVar["name"].(string) {
					t.Fatalf("unexpected: %v\n", fooEnvVar["name"].(string))
				}
				if "not far" != fooEnvVar["value"].(string) {
					t.Fatalf("unexpected: %v\n", fooEnvVar["value"].(string))
				}

				booEnvVar := envVars.([]interface{})[1].(map[string]interface{})
				if "BOO" != booEnvVar["name"].(string) {
					t.Fatalf("unexpected: %v\n", booEnvVar["name"].(string))
				}
				if "farther than it looks" != booEnvVar["value"].(string) {
					t.Fatalf("unexpected: %v\n", booEnvVar["value"].(string))
				}
			},
		},
		{
			name: "object reference, GVK-only match",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SearchReplace
metadata:
  name: notImportantHere
target:
  kind: Foo
  name: some-foo
path: fooSpec/fooTemplate/fooContainers/env/value
search: ^far$
replaceWithObjRef:
  objref:
    apiVersion: qlik.com/v1
    kind: Bar
  fieldref:
    fieldpath: metadata.labels.myproperty
`,
			pluginInputResources: `
apiVersion: qlik.com/v1
kind: Foo
metadata:
  name: some-foo
fooSpec:
  fooTemplate:
    fooContainers:
    - name: have-env
      env:
      - name: FOO
        value: far
      - name: BOO
        value: farther than it looks
--- 
apiVersion: qlik.com/v1
kind: Bar
metadata:
  name: some-bar
  labels:
    myproperty: not far
fooSpec:
  test: test
--- 
apiVersion: qlik.com/v1
kind: Foo
metadata:
  name: some-Foo
  labels:
    myproperty: not good
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				res := resMap.GetByIndex(0)

				envVars, err := res.GetFieldValue("fooSpec.fooTemplate.fooContainers[0].env")
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				fooEnvVar := envVars.([]interface{})[0].(map[string]interface{})
				if "FOO" != fooEnvVar["name"].(string) {
					t.Fatalf("unexpected: %v\n", fooEnvVar["name"].(string))
				}
				if "not far" != fooEnvVar["value"].(string) {
					t.Fatalf("unexpected: %v\n", fooEnvVar["value"].(string))
				}

				booEnvVar := envVars.([]interface{})[1].(map[string]interface{})
				if "BOO" != booEnvVar["name"].(string) {
					t.Fatalf("unexpected: %v\n", booEnvVar["name"].(string))
				}
				if "farther than it looks" != booEnvVar["value"].(string) {
					t.Fatalf("unexpected: %v\n", booEnvVar["value"].(string))
				}
			},
		},
		{
			name: "object reference, first GVK-only match",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SearchReplace
metadata:
  name: notImportantHere
target:
  kind: Foo
  name: some-foo
path: fooSpec/fooTemplate/fooContainers/env/value
search: ^far$
replaceWithObjRef:
  objref:
    apiVersion: qlik.com/
    kind: Bar
  fieldref:
    fieldpath: metadata.labels.myproperty
`,
			pluginInputResources: `
apiVersion: qlik.com/v1
kind: Foo
metadata:
  name: some-foo
fooSpec:
  fooTemplate:
    fooContainers:
    - name: have-env
      env:
      - name: FOO
        value: far
      - name: BOO
        value: farther than it looks
--- 
apiVersion: qlik.com/v1
kind: Bar
metadata:
  name: some-bar-1
  labels:
    myproperty: not far
fooSpec:
  test: test
---
apiVersion: qlik.com/v1
kind: Bar
metadata:
  name: some-bar-2
  labels:
    myproperty: too far
fooSpec:
  test: test
--- 
apiVersion: qlik.com/v1
kind: Foo
metadata:
  name: some-Foo
  labels:
    myproperty: not good
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				res := resMap.GetByIndex(0)

				envVars, err := res.GetFieldValue("fooSpec.fooTemplate.fooContainers[0].env")
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				fooEnvVar := envVars.([]interface{})[0].(map[string]interface{})
				if "FOO" != fooEnvVar["name"].(string) {
					t.Fatalf("unexpected: %v\n", fooEnvVar["name"].(string))
				}
				if "not far" != fooEnvVar["value"].(string) {
					t.Fatalf("unexpected: %v\n", fooEnvVar["value"].(string))
				}

				booEnvVar := envVars.([]interface{})[1].(map[string]interface{})
				if "BOO" != booEnvVar["name"].(string) {
					t.Fatalf("unexpected: %v\n", booEnvVar["name"].(string))
				}
				if "farther than it looks" != booEnvVar["value"].(string) {
					t.Fatalf("unexpected: %v\n", booEnvVar["value"].(string))
				}
			},
		},
		{
			name: "object reference, GVK and name match",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SearchReplace
metadata:
  name: notImportantHere
target:
  kind: Foo
  name: some-foo
path: fooSpec/fooTemplate/fooContainers/env/value
search: ^far$
replaceWithObjRef:
  objref:
    apiVersion: qlik.com/
    kind: Bar
    name: some-bar
  fieldref:
    fieldpath: metadata.labels.myproperty
`,
			pluginInputResources: `
apiVersion: qlik.com/v1
kind: Foo
metadata:
  name: some-foo
fooSpec:
  fooTemplate:
    fooContainers:
    - name: have-env
      env:
      - name: FOO
        value: far
      - name: BOO
        value: farther than it looks
--- 
apiVersion: qlik.com/v1
kind: Bar
metadata:
  name: some-chocolate-bar
  labels:
    myproperty: not far enough
fooSpec:
  test: test
--- 
apiVersion: qlik.com/v1
kind: Bar
metadata:
  name: some-bar
  labels:
    myproperty: not far
fooSpec:
  test: test
--- 
apiVersion: qlik.com/v1
kind: Foo
metadata:
  name: some-Foo
  labels:
    myproperty: not good
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				res := resMap.GetByIndex(0)

				envVars, err := res.GetFieldValue("fooSpec.fooTemplate.fooContainers[0].env")
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				fooEnvVar := envVars.([]interface{})[0].(map[string]interface{})
				if "FOO" != fooEnvVar["name"].(string) {
					t.Fatalf("unexpected: %v\n", fooEnvVar["name"].(string))
				}
				if "not far" != fooEnvVar["value"].(string) {
					t.Fatalf("unexpected: %v\n", fooEnvVar["value"].(string))
				}

				booEnvVar := envVars.([]interface{})[1].(map[string]interface{})
				if "BOO" != booEnvVar["name"].(string) {
					t.Fatalf("unexpected: %v\n", booEnvVar["name"].(string))
				}
				if "farther than it looks" != booEnvVar["value"].(string) {
					t.Fatalf("unexpected: %v\n", booEnvVar["value"].(string))
				}
			},
		},
		{
			name: "can replace with a blank",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SearchReplace
metadata:
  name: notImportantHere
target:
  kind: Foo
  name: some-foo
path: fooSpec/fooTemplate/fooContainers/env/value
search: ^far$
replace: ""
`,
			pluginInputResources: `
apiVersion: qlik.com/v1
kind: Foo
metadata:
  name: some-foo
fooSpec:
  fooTemplate:
    fooContainers:
    - name: have-env
      env:
      - name: FOO
        value: far
      - name: BOO
        value: farther than it looks
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				res := resMap.GetByIndex(0)

				envVars, err := res.GetFieldValue("fooSpec.fooTemplate.fooContainers[0].env")
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				fooEnvVar := envVars.([]interface{})[0].(map[string]interface{})
				if "FOO" != fooEnvVar["name"].(string) {
					t.Fatalf("unexpected: %v\n", fooEnvVar["name"].(string))
				}
				if "" != fooEnvVar["value"].(string) {
					t.Fatalf("unexpected: %v\n", fooEnvVar["value"].(string))
				}

				booEnvVar := envVars.([]interface{})[1].(map[string]interface{})
				if "BOO" != booEnvVar["name"].(string) {
					t.Fatalf("unexpected: %v\n", booEnvVar["name"].(string))
				}
				if "farther than it looks" != booEnvVar["value"].(string) {
					t.Fatalf("unexpected: %v\n", booEnvVar["value"].(string))
				}
			},
		},
	}
	plugin := SearchReplacePlugin{logger: utils.GetLogger("SearchReplacePlugin")}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resourceFactory := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()), transformer.NewFactoryImpl())

			resMap, err := resourceFactory.NewResMapFromBytes([]byte(testCase.pluginInputResources))
			if err != nil {
				t.Fatalf("Err: %v", err)
			}

			h := resmap.NewPluginHelpers(loader.NewFileLoaderAtRoot(filesys.MakeFsInMemory()), valtest_test.MakeHappyMapValidator(t), resourceFactory)
			if err := plugin.Config(h, []byte(testCase.pluginConfig)); err != nil {
				t.Fatalf("Err: %v", err)
			}

			if err := plugin.Transform(resMap); err != nil {
				t.Fatalf("Err: %v", err)
			}

			for _, res := range resMap.Resources() {
				fmt.Printf("--res: %v\n", res.String())
			}

			testCase.checkAssertions(t, resMap)
		})
	}
}
