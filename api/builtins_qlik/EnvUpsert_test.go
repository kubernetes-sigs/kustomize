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
)

func TestEnvUpsert(t *testing.T) {

	testCases := []struct {
		name                 string
		pluginConfig         string
		pluginInputResources string
		checkAssertions      func(*testing.T, resmap.ResMap)
	}{
		{
			name: "disabled",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: EnvUpsert
metadata:
  name: notImportantHere
enabled: false
target:
  kind: Foo
  name: some-foo
path: fooSpec/fooTemplate/fooContainers/env
env:
- name: FOO
  valueFrom:
    configMapKeyRef:
      name: some-other-config-map
      key: some-other-key
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
        valueFrom:
          configMapKeyRef:
            name: my-config-map
            key: foo
      - name: BAR
        valueFrom:
          configMapKeyRef:
            name: my-config-map
            key: bar
    - name: dont-have-env
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				res := resMap.GetByIndex(0)

				envVars, err := res.GetFieldValue("fooSpec.fooTemplate.fooContainers[0].env")
				assert.NoError(t, err)
				assert.NotNil(t, envVars)
				assert.Equal(t, 2, len(envVars.([]interface{})))

				fooEnvVar := envVars.([]interface{})[0].(map[string]interface{})
				assert.Equal(t, "FOO", fooEnvVar["name"].(string))
				assert.Equal(t, "my-config-map", fooEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["name"].(string))
				assert.Equal(t, "foo", fooEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["key"].(string))

				barEnvVar := envVars.([]interface{})[1].(map[string]interface{})
				assert.Equal(t, "BAR", barEnvVar["name"].(string))
				assert.Equal(t, "my-config-map", barEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["name"].(string))
				assert.Equal(t, "bar", barEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["key"].(string))
			},
		},
		{
			name: "single value update",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: EnvUpsert
metadata:
  name: notImportantHere
enabled: true
target:
  kind: Foo
  name: some-foo
path: fooSpec/fooTemplate/fooContainers/env
env:
- name: FOO
  valueFrom:
    configMapKeyRef:
      name: some-other-config-map
      key: some-other-key
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
        valueFrom:
          configMapKeyRef:
            name: my-config-map
            key: foo
      - name: BAR
        valueFrom:
          configMapKeyRef:
            name: my-config-map
            key: bar
    - name: dont-have-env
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				res := resMap.GetByIndex(0)

				envVars, err := res.GetFieldValue("fooSpec.fooTemplate.fooContainers[0].env")
				assert.NoError(t, err)
				assert.NotNil(t, envVars)
				assert.Equal(t, 2, len(envVars.([]interface{})))

				fooEnvVar := envVars.([]interface{})[0].(map[string]interface{})
				assert.Equal(t, "FOO", fooEnvVar["name"].(string))
				assert.Equal(t, "some-other-config-map", fooEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["name"].(string))
				assert.Equal(t, "some-other-key", fooEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["key"].(string))

				barEnvVar := envVars.([]interface{})[1].(map[string]interface{})
				assert.Equal(t, "BAR", barEnvVar["name"].(string))
				assert.Equal(t, "my-config-map", barEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["name"].(string))
				assert.Equal(t, "bar", barEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["key"].(string))
			},
		},
		{
			name: "multiple values update",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: EnvUpsert
metadata:
  name: notImportantHere
enabled: true
target:
  kind: Foo
  name: some-foo
path: fooSpec/fooTemplate/fooContainers/env
env:
- name: FOO
  valueFrom:
    configMapKeyRef:
      name: some-other-config-map
      key: some-other-foo
- name: BAR
  valueFrom:
    configMapKeyRef:
      name: some-other-config-map
      key: some-other-bar
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
         valueFrom:
           configMapKeyRef:
             name: my-config-map
             key: foo
       - name: BAR
         valueFrom:
           configMapKeyRef:
             name: my-config-map
             key: bar
   - name: dont-have-env
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				res := resMap.GetByIndex(0)

				envVars, err := res.GetFieldValue("fooSpec.fooTemplate.fooContainers[0].env")
				assert.NoError(t, err)
				assert.NotNil(t, envVars)
				assert.Equal(t, 2, len(envVars.([]interface{})))

				fooEnvVar := envVars.([]interface{})[0].(map[string]interface{})
				assert.Equal(t, "FOO", fooEnvVar["name"].(string))
				assert.Equal(t, "some-other-config-map", fooEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["name"].(string))
				assert.Equal(t, "some-other-foo", fooEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["key"].(string))

				barEnvVar := envVars.([]interface{})[1].(map[string]interface{})
				assert.Equal(t, "BAR", barEnvVar["name"].(string))
				assert.Equal(t, "some-other-config-map", barEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["name"].(string))
				assert.Equal(t, "some-other-bar", barEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["key"].(string))
			},
		},
		{
			name: "single value insert",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: EnvUpsert
metadata:
  name: notImportantHere
enabled: true
target:
  kind: Foo
  name: some-foo
path: fooSpec/fooTemplate/fooContainers/env
env:
- name: ABRA
  value: CADABRA
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
          valueFrom:
            configMapKeyRef:
              name: my-config-map
              key: foo
        - name: BAR
          valueFrom:
            configMapKeyRef:
              name: my-config-map
              key: bar
    - name: dont-have-env
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				res := resMap.GetByIndex(0)

				envVars, err := res.GetFieldValue("fooSpec.fooTemplate.fooContainers[0].env")
				assert.NoError(t, err)
				assert.NotNil(t, envVars)
				assert.Equal(t, 3, len(envVars.([]interface{})))

				fooEnvVar := envVars.([]interface{})[0].(map[string]interface{})
				assert.Equal(t, "FOO", fooEnvVar["name"].(string))
				assert.Equal(t, "my-config-map", fooEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["name"].(string))
				assert.Equal(t, "foo", fooEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["key"].(string))

				barEnvVar := envVars.([]interface{})[1].(map[string]interface{})
				assert.Equal(t, "BAR", barEnvVar["name"].(string))
				assert.Equal(t, "my-config-map", barEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["name"].(string))
				assert.Equal(t, "bar", barEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["key"].(string))

				abraEnvVar := envVars.([]interface{})[2].(map[string]interface{})
				assert.Equal(t, "ABRA", abraEnvVar["name"].(string))
				assert.Equal(t, "CADABRA", abraEnvVar["value"].(string))
			},
		},
		{
			name: "multiple values insert",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: EnvUpsert
metadata:
  name: notImportantHere
enabled: true
target:
  kind: Foo
  name: some-foo
path: fooSpec/fooTemplate/fooContainers/env
env:
- name: ABRA
  value: CADABRA
- name: BAZ
  valueFrom:
    configMapKeyRef:
      name: baz-config-map
      key: baz
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
          valueFrom:
            configMapKeyRef:
              name: my-config-map
              key: foo
        - name: BAR
          valueFrom:
            configMapKeyRef:
              name: my-config-map
              key: bar
    - name: dont-have-env
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				res := resMap.GetByIndex(0)

				envVars, err := res.GetFieldValue("fooSpec.fooTemplate.fooContainers[0].env")
				assert.NoError(t, err)
				assert.NotNil(t, envVars)
				assert.Equal(t, 4, len(envVars.([]interface{})))

				fooEnvVar := envVars.([]interface{})[0].(map[string]interface{})
				assert.Equal(t, "FOO", fooEnvVar["name"].(string))
				assert.Equal(t, "my-config-map", fooEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["name"].(string))
				assert.Equal(t, "foo", fooEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["key"].(string))

				barEnvVar := envVars.([]interface{})[1].(map[string]interface{})
				assert.Equal(t, "BAR", barEnvVar["name"].(string))
				assert.Equal(t, "my-config-map", barEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["name"].(string))
				assert.Equal(t, "bar", barEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["key"].(string))

				abraEnvVar := envVars.([]interface{})[2].(map[string]interface{})
				assert.Equal(t, "ABRA", abraEnvVar["name"].(string))
				assert.Equal(t, "CADABRA", abraEnvVar["value"].(string))

				bazEnvVar := envVars.([]interface{})[3].(map[string]interface{})
				assert.Equal(t, "BAZ", bazEnvVar["name"].(string))
				assert.Equal(t, "baz-config-map", bazEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["name"].(string))
				assert.Equal(t, "baz", bazEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["key"].(string))
			},
		},
		{
			name: "single value delete",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: EnvUpsert
metadata:
  name: notImportantHere
enabled: true
target:
  kind: Foo
  name: some-foo
path: fooSpec/fooTemplate/fooContainers/env
env:
- name: FOO
  delete: true
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
          valueFrom:
            configMapKeyRef:
              name: my-config-map
              key: foo
        - name: BAR
          valueFrom:
            configMapKeyRef:
              name: my-config-map
              key: bar
    - name: dont-have-env
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				res := resMap.GetByIndex(0)

				envVars, err := res.GetFieldValue("fooSpec.fooTemplate.fooContainers[0].env")
				assert.NoError(t, err)
				assert.NotNil(t, envVars)
				assert.Equal(t, 1, len(envVars.([]interface{})))

				barEnvVar := envVars.([]interface{})[0].(map[string]interface{})
				assert.Equal(t, "BAR", barEnvVar["name"].(string))
				assert.Equal(t, "my-config-map", barEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["name"].(string))
				assert.Equal(t, "bar", barEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["key"].(string))
			},
		},
		{
			name: "multiple values delete",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: EnvUpsert
metadata:
  name: notImportantHere
enabled: true
target:
  kind: Foo
  name: some-foo
path: fooSpec/fooTemplate/fooContainers/env
env:
- name: FOO
  delete: true
- name: BAR
  delete: true
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
        - name: FIRST
          valueFrom:
            configMapKeyRef:
              name: my-config-map
              key: first
        - name: FOO
          valueFrom:
            configMapKeyRef:
              name: my-config-map
              key: foo
        - name: BAR
          valueFrom:
            configMapKeyRef:
              name: my-config-map
              key: bar
        - name: LAST
          valueFrom:
            configMapKeyRef:
              name: my-config-map
              key: last
    - name: dont-have-env
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				res := resMap.GetByIndex(0)

				envVars, err := res.GetFieldValue("fooSpec.fooTemplate.fooContainers[0].env")
				assert.NoError(t, err)
				assert.NotNil(t, envVars)
				assert.Equal(t, 2, len(envVars.([]interface{})))

				firstEnvVar := envVars.([]interface{})[0].(map[string]interface{})
				assert.Equal(t, "FIRST", firstEnvVar["name"].(string))
				assert.Equal(t, "my-config-map", firstEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["name"].(string))
				assert.Equal(t, "first", firstEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["key"].(string))

				lastEnvVar := envVars.([]interface{})[1].(map[string]interface{})
				assert.Equal(t, "LAST", lastEnvVar["name"].(string))
				assert.Equal(t, "my-config-map", lastEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["name"].(string))
				assert.Equal(t, "last", lastEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["key"].(string))
			},
		},
		{
			name: "update and delete and insert",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: EnvUpsert
metadata:
  name: notImportantHere
enabled: true
target:
  kind: Foo
  name: some-foo
path: fooSpec/fooTemplate/fooContainers/env
env:
- name: FOO
  delete: true
- name: BAR
  valueFrom:
    configMapKeyRef:
      name: another-config-map
      key: bar
- name: ABRA
  value: cadabra
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
          valueFrom:
            configMapKeyRef:
              name: my-config-map
              key: foo
        - name: BAR
          valueFrom:
            configMapKeyRef:
              name: my-config-map
              key: bar
    - name: dont-have-env
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				res := resMap.GetByIndex(0)

				envVars, err := res.GetFieldValue("fooSpec.fooTemplate.fooContainers[0].env")
				assert.NoError(t, err)
				assert.NotNil(t, envVars)
				assert.Equal(t, 2, len(envVars.([]interface{})))

				barEnvVar := envVars.([]interface{})[0].(map[string]interface{})
				assert.Equal(t, "BAR", barEnvVar["name"].(string))
				assert.Equal(t, "another-config-map", barEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["name"].(string))
				assert.Equal(t, "bar", barEnvVar["valueFrom"].(map[string]interface{})["configMapKeyRef"].(map[string]interface{})["key"].(string))

				abraEnvVar := envVars.([]interface{})[1].(map[string]interface{})
				assert.Equal(t, "ABRA", abraEnvVar["name"].(string))
				assert.Equal(t, "cadabra", abraEnvVar["value"].(string))
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

			plugin := NewEnvUpsertPlugin()

			err = plugin.Config(resmap.NewPluginHelpers(loadertest.NewFakeLoader("/"), valtest_test.MakeFakeValidator(), resourceFactory), []byte(testCase.pluginConfig))
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

			testCase.checkAssertions(t, resMap)
		})
	}
}
