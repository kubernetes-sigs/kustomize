package builtins_qlik

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/builtins_qlik/utils/loadertest"
	"sigs.k8s.io/kustomize/api/internal/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
)

func TestSuperConfigMap_simpleTransformer(t *testing.T) {
	pluginInputResources := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config-map
data:
  foo: bar
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: my-container
        image: some-image
        env:
        - name: FOO
          valueFrom:
            configMapKeyRef:
              name: my-config-map
              key: foo
`
	testCases := []struct {
		name                 string
		pluginConfig         string
		pluginInputResources string
		checkAssertions      func(*testing.T, resmap.ResMap)
	}{
		{
			name: "withoutHash_withoutAppendData",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperConfigMap
metadata:
  name: my-config-map
disableNameSuffixHash: true
`,
			pluginInputResources: pluginInputResources,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				foundConfigMapResource := false
				for _, res := range resMap.Resources() {
					if res.GetKind() == "ConfigMap" {
						foundConfigMapResource = true
						assert.Equal(t, "my-config-map", res.GetName())
						assert.False(t, res.NeedHashSuffix())

						data, err := res.GetFieldValue("data")
						assert.NoError(t, err)
						assert.True(t, len(data.(map[string]interface{})) == 1)

						value, err := res.GetFieldValue("data.foo")
						assert.NoError(t, err)
						assert.Equal(t, "bar", value)

						break
					}
				}
				assert.True(t, foundConfigMapResource)
			},
		},
		{
			name: "withoutHash_withAppendData",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperConfigMap
metadata:
  name: my-config-map
data:
  baz: boo
  abra: cadabra
disableNameSuffixHash: true
`,
			pluginInputResources: pluginInputResources,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				foundConfigMapResource := false
				for _, res := range resMap.Resources() {
					if res.GetKind() == "ConfigMap" {
						foundConfigMapResource = true
						assert.Equal(t, "my-config-map", res.GetName())
						assert.False(t, res.NeedHashSuffix())

						data, err := res.GetFieldValue("data")
						assert.NoError(t, err)
						assert.True(t, len(data.(map[string]interface{})) == 3)

						value, err := res.GetFieldValue("data.foo")
						assert.NoError(t, err)
						assert.Equal(t, "bar", value)

						value, err = res.GetFieldValue("data.baz")
						assert.NoError(t, err)
						assert.Equal(t, "boo", value)

						value, err = res.GetFieldValue("data.abra")
						assert.NoError(t, err)
						assert.Equal(t, "cadabra", value)

						break
					}
				}
				assert.True(t, foundConfigMapResource)
			},
		},
		{
			name: "withHash_withoutAppendData",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperConfigMap
metadata:
  name: my-config-map
`,
			pluginInputResources: pluginInputResources,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				foundConfigMapResource := false
				for _, res := range resMap.Resources() {
					if res.GetKind() == "ConfigMap" {
						foundConfigMapResource = true

						assert.Equal(t, "my-config-map", res.GetName())
						assert.True(t, res.NeedHashSuffix())

						data, err := res.GetFieldValue("data")
						assert.NoError(t, err)
						assert.True(t, len(data.(map[string]interface{})) == 1)

						value, err := res.GetFieldValue("data.foo")
						assert.NoError(t, err)
						assert.Equal(t, "bar", value)

						break
					}
				}
				assert.True(t, foundConfigMapResource)
			},
		},
		{
			name: "withHash_withAppendData",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperConfigMap
metadata:
  name: my-config-map
data:
  baz: boo
  abra: cadabra
`,
			pluginInputResources: pluginInputResources,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				foundConfigMapResource := false
				for _, res := range resMap.Resources() {
					if res.GetKind() == "ConfigMap" {
						foundConfigMapResource = true

						assert.Equal(t, "my-config-map", res.GetName())
						assert.True(t, res.NeedHashSuffix())

						data, err := res.GetFieldValue("data")
						assert.NoError(t, err)
						assert.True(t, len(data.(map[string]interface{})) == 3)

						value, err := res.GetFieldValue("data.foo")
						assert.NoError(t, err)
						assert.Equal(t, "bar", value)

						value, err = res.GetFieldValue("data.baz")
						assert.NoError(t, err)
						assert.Equal(t, "boo", value)

						value, err = res.GetFieldValue("data.abra")
						assert.NoError(t, err)
						assert.Equal(t, "cadabra", value)

						break
					}
				}
				assert.True(t, foundConfigMapResource)
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

			plugin := NewSuperConfigMapTransformerPlugin()

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

func TestSuperConfigMap_assumeTargetWillExistTransformer(t *testing.T) {
	pluginInputResources := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment-1
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: my-container
        image: some-image
        env:
        - name: FOO
          valueFrom:
            configMapKeyRef:
              name: my-config-map
              key: foo
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment-2
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: my-container-2
        image: some-image
        env:
        - name: FOO
          valueFrom:
            configMapKeyRef:
              name: my-config-map
              key: foo
`
	assertReferencesUpdatedWithHashes := func(t *testing.T, resMap resmap.ResMap) {
		for _, res := range resMap.Resources() {
			if res.GetKind() == "ConfigMap" {
				assert.FailNow(t, "configMap should not be present in the stream")
				break
			}
		}

		foundDeployments := map[string]bool{"my-deployment-1": false, "my-deployment-2": false}
		for _, deploymentName := range []string{"my-deployment-1", "my-deployment-2"} {
			for _, res := range resMap.Resources() {
				if res.GetKind() == "Deployment" && res.GetName() == deploymentName {
					foundDeployments[deploymentName] = true

					value, err := res.GetFieldValue("spec.template.spec.containers[0].env[0].valueFrom.configMapKeyRef.name")
					assert.NoError(t, err)

					match, err := regexp.MatchString("^my-config-map-[0-9a-z]+$", value.(string))
					assert.NoError(t, err)
					assert.True(t, match)

					break
				}
			}
		}
		for deploymentName := range foundDeployments {
			assert.True(t, foundDeployments[deploymentName])
		}
	}

	assertReferencesNotUpdated := func(t *testing.T, resMap resmap.ResMap) {
		for _, res := range resMap.Resources() {
			if res.GetKind() == "ConfigMap" {
				assert.FailNow(t, "configMap should not be present in the stream")
				break
			}
		}

		foundDeployments := map[string]bool{"my-deployment-1": false, "my-deployment-2": false}
		for _, deploymentName := range []string{"my-deployment-1", "my-deployment-2"} {
			for _, res := range resMap.Resources() {
				if res.GetKind() == "Deployment" && res.GetName() == deploymentName {
					foundDeployments[deploymentName] = true

					value, err := res.GetFieldValue("spec.template.spec.containers[0].env[0].valueFrom.configMapKeyRef.name")
					assert.NoError(t, err)

					assert.NoError(t, err)
					assert.Equal(t, "my-config-map", value)

					break
				}
			}
		}
		for deploymentName := range foundDeployments {
			assert.True(t, foundDeployments[deploymentName])
		}
	}

	testCases := []struct {
		name                 string
		pluginConfig         string
		pluginInputResources string
		checkAssertions      func(*testing.T, resmap.ResMap)
	}{
		{
			name: "assumeTargetWillExist_isTrue_byDefault",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperConfigMap
metadata:
 name: my-config-map
data:
 foo: bar
 baz: boo
`,
			pluginInputResources: pluginInputResources,
			checkAssertions:      assertReferencesUpdatedWithHashes,
		},
		{
			name: "assumeTargetWillExist_canBeTurnedOff",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperConfigMap
metadata:
  name: my-config-map
data:
  foo: bar
  baz: boo
assumeTargetWillExist: false
`,
			pluginInputResources: pluginInputResources,
			checkAssertions:      assertReferencesNotUpdated,
		},
		{
			name: "withHash_withAppendData",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperConfigMap
metadata:
  name: my-config-map
data:
  foo: bar
  baz: boo
assumeTargetWillExist: true
`,
			pluginInputResources: pluginInputResources,
			checkAssertions:      assertReferencesUpdatedWithHashes,
		},
		{
			name: "doesNothing_withoutHash",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperConfigMap
metadata:
  name: my-config-map
data:
  foo: bar
  baz: boo
assumeTargetWillExist: true
disableNameSuffixHash: true
`,
			pluginInputResources: pluginInputResources,
			checkAssertions:      assertReferencesNotUpdated,
		},
		{
			name: "appendNameSuffixHash_forEmptyData",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperConfigMap
metadata:
 name: my-config-map
assumeTargetWillExist: true
`,
			pluginInputResources: pluginInputResources,
			checkAssertions:      assertReferencesUpdatedWithHashes,
		},
		{
			name: "appendNameSuffixHash_withPrefix",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperConfigMap
metadata:
 name: my-config-map
assumeTargetWillExist: true
prefix: some-service-
`,
			pluginInputResources: pluginInputResources,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				for _, res := range resMap.Resources() {
					if res.GetKind() == "ConfigMap" {
						assert.FailNow(t, "configMap should not be present in the stream")
						break
					}
				}

				foundDeployments := map[string]bool{"my-deployment-1": false, "my-deployment-2": false}
				for _, deploymentName := range []string{"my-deployment-1", "my-deployment-2"} {
					for _, res := range resMap.Resources() {
						if res.GetKind() == "Deployment" && res.GetName() == deploymentName {
							foundDeployments[deploymentName] = true

							value, err := res.GetFieldValue("spec.template.spec.containers[0].env[0].valueFrom.configMapKeyRef.name")
							assert.NoError(t, err)
							refName := value.(string)

							match, err := regexp.MatchString("^some-service-my-config-map-[0-9a-z]+$", refName)
							assert.NoError(t, err)
							assert.True(t, match)

							resourceFactory := resmap.NewFactory(resource.NewFactory(
								kunstruct.NewKunstructuredFactoryImpl()), nil)

							plugin := NewSuperConfigMapGeneratorPlugin()
							err = plugin.Config(resmap.NewPluginHelpers(loadertest.NewFakeLoader("/"), valtest_test.MakeFakeValidator(), resourceFactory), []byte(`
apiVersion: qlik.com/v1
kind: SuperConfigMap
metadata:
 name: my-config-map
assumeTargetWillExist: true
prefix: some-service-
`))
							if err != nil {
								t.Fatalf("Err: %v", err)
							}
							generateResMap, err := plugin.Generate()
							assert.NoError(t, err)

							tempRes := generateResMap.GetByIndex(0)
							assert.NotNil(t, tempRes)
							assert.True(t, tempRes.NeedHashSuffix())

							tempRes.SetName(fmt.Sprintf("some-service-%s", tempRes.GetName()))

							hash, err := kunstruct.NewKunstructuredFactoryImpl().Hasher().Hash(tempRes)
							assert.NoError(t, err)
							assert.Equal(t, fmt.Sprintf("%s-%s", tempRes.GetName(), hash), refName)

							break
						}
					}
				}
				for deploymentName := range foundDeployments {
					assert.True(t, foundDeployments[deploymentName])
				}
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resourceFactory := resmap.NewFactory(resource.NewFactory(
				kunstruct.NewKunstructuredFactoryImpl()), nil)

			resMap, err := resourceFactory.NewResMapFromBytes([]byte(testCase.pluginInputResources))
			if err != nil {
				t.Fatalf("Err: %v", err)
			}

			plugin := NewSuperConfigMapTransformerPlugin()
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

func TestSuperConfigMap_generator(t *testing.T) {
	testCases := []struct {
		name                 string
		pluginConfig         string
		pluginInputResources string
		checkAssertions      func(*testing.T, resmap.ResMap)
	}{
		{
			name: "withoutHash_withoutData",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperConfigMap
metadata:
 name: my-config-map
behavior: create
disableNameSuffixHash: true
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				foundConfigMapResource := false
				for _, res := range resMap.Resources() {
					if res.GetKind() == "ConfigMap" {
						foundConfigMapResource = true
						assert.Equal(t, "my-config-map", res.GetName())
						assert.False(t, res.NeedHashSuffix())

						data, err := res.GetFieldValue("data")
						assert.Error(t, err)
						assert.Nil(t, data)

						break
					}
				}
				assert.True(t, foundConfigMapResource)
			},
		},
		{
			name: "withoutHash_withData",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperConfigMap
metadata:
  name: my-config-map
data:
  foo: bar
  baz: whatever
behavior: create
disableNameSuffixHash: true
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				foundConfigMapResource := false
				for _, res := range resMap.Resources() {
					if res.GetKind() == "ConfigMap" {
						foundConfigMapResource = true
						assert.Equal(t, "my-config-map", res.GetName())
						assert.False(t, res.NeedHashSuffix())

						data, err := res.GetFieldValue("data")
						assert.NoError(t, err)
						assert.True(t, len(data.(map[string]interface{})) == 2)

						value, err := res.GetFieldValue("data.foo")
						assert.NoError(t, err)
						assert.Equal(t, "bar", value)

						value, err = res.GetFieldValue("data.baz")
						assert.NoError(t, err)
						assert.Equal(t, "whatever", value)

						break
					}
				}
				assert.True(t, foundConfigMapResource)
			},
		},
		{
			name: "withHash_withoutData",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperConfigMap
metadata:
  name: my-config-map
behavior: create
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				foundConfigMapResource := false
				for _, res := range resMap.Resources() {
					if res.GetKind() == "ConfigMap" {
						foundConfigMapResource = true
						assert.Equal(t, "my-config-map", res.GetName())
						assert.True(t, res.NeedHashSuffix())

						data, err := res.GetFieldValue("data")
						assert.Error(t, err)
						assert.Nil(t, data)

						break
					}
				}
				assert.True(t, foundConfigMapResource)
			},
		},
		{
			name: "withHash_withStringData",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperConfigMap
metadata:
  name: my-config-map
data:
  foo: bar
  baz: whatever
behavior: create
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				foundConfigMapResource := false
				for _, res := range resMap.Resources() {
					if res.GetKind() == "ConfigMap" {
						foundConfigMapResource = true
						assert.Equal(t, "my-config-map", res.GetName())
						assert.True(t, res.NeedHashSuffix())

						data, err := res.GetFieldValue("data")
						assert.NoError(t, err)
						assert.True(t, len(data.(map[string]interface{})) == 2)

						value, err := res.GetFieldValue("data.foo")
						assert.NoError(t, err)
						assert.Equal(t, "bar", value)

						value, err = res.GetFieldValue("data.baz")
						assert.NoError(t, err)
						assert.Equal(t, "whatever", value)

						break
					}
				}
				assert.True(t, foundConfigMapResource)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resourceFactory := resmap.NewFactory(resource.NewFactory(
				kunstruct.NewKunstructuredFactoryImpl()), nil)

			plugin := NewSuperConfigMapGeneratorPlugin()
			err := plugin.Config(resmap.NewPluginHelpers(loadertest.NewFakeLoader("/"), valtest_test.MakeFakeValidator(), resourceFactory), []byte(testCase.pluginConfig))
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

			testCase.checkAssertions(t, resMap)
		})
	}
}
