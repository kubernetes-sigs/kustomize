package builtins_qlik

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/internal/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
)

func TestSuperSecret_simpleTransformer(t *testing.T) {
	pluginInputResources := `
apiVersion: v1
kind: Secret
metadata:
  name: mySecret
type: Opaque
data:
  PASSWORD: d2hhdGV2ZXI=
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: myPod
        image: some-image
        volumeMounts:
        - name: foo
          mountPath: "/etc/foo"
          readOnly: true
      volumes:
      - name: foo
        secret:
          secretName: mySecret
`
	withHashWithAppendDataAssertions := func(t *testing.T, resMap resmap.ResMap) {
		foundSecretResource := false
		for _, res := range resMap.Resources() {
			if res.GetKind() == "Secret" {
				foundSecretResource = true

				assert.Equal(t, "mySecret", res.GetName())
				assert.True(t, res.NeedHashSuffix())

				data, err := res.GetFieldValue("data")
				assert.NoError(t, err)
				assert.True(t, len(data.(map[string]interface{})) == 4)

				value, err := res.GetFieldValue("data.PASSWORD")
				assert.NoError(t, err)
				assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("whatever")), value)

				value, err = res.GetFieldValue("data.foo")
				assert.NoError(t, err)
				assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("bar")), value)

				value, err = res.GetFieldValue("data.baz")
				assert.NoError(t, err)
				assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("whatever")), value)

				value, err = res.GetFieldValue("data.anotherPassword")
				assert.NoError(t, err)
				assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("boom")), value)

				break
			}
		}
		assert.True(t, foundSecretResource)
	}

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
kind: SuperSecret
metadata:
  name: mySecret
options:
  disableNameSuffixHash: true
`,
			pluginInputResources: pluginInputResources,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				foundSecretResource := false
				for _, res := range resMap.Resources() {
					if res.GetKind() == "Secret" {
						foundSecretResource = true
						assert.Equal(t, "mySecret", res.GetName())
						assert.False(t, res.NeedHashSuffix())

						data, err := res.GetFieldValue("data")
						assert.NoError(t, err)
						assert.True(t, len(data.(map[string]interface{})) == 1)

						value, err := res.GetFieldValue("data.PASSWORD")
						assert.NoError(t, err)
						assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("whatever")), value)

						break
					}
				}
				assert.True(t, foundSecretResource)
			},
		},
		{
			name: "withoutHash_withAppendData",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperSecret
metadata:
  name: mySecret
stringData:
  foo: bar
  baz: whatever
data:
  anotherPassword: Ym9vbQ==
options:
  disableNameSuffixHash: true
`,
			pluginInputResources: pluginInputResources,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				foundSecretResource := false
				for _, res := range resMap.Resources() {
					if res.GetKind() == "Secret" {
						foundSecretResource = true
						assert.Equal(t, "mySecret", res.GetName())
						assert.False(t, res.NeedHashSuffix())

						data, err := res.GetFieldValue("data")
						assert.NoError(t, err)
						assert.True(t, len(data.(map[string]interface{})) == 4)

						value, err := res.GetFieldValue("data.PASSWORD")
						assert.NoError(t, err)
						assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("whatever")), value)

						value, err = res.GetFieldValue("data.foo")
						assert.NoError(t, err)
						assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("bar")), value)

						value, err = res.GetFieldValue("data.baz")
						assert.NoError(t, err)
						assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("whatever")), value)

						value, err = res.GetFieldValue("data.anotherPassword")
						assert.NoError(t, err)
						assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("boom")), value)

						break
					}
				}
				assert.True(t, foundSecretResource)
			},
		},
		{
			name: "withHash_withoutAppendData",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperSecret
metadata:
  name: mySecret
`,
			pluginInputResources: pluginInputResources,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				foundSecretResource := false
				for _, res := range resMap.Resources() {
					if res.GetKind() == "Secret" {
						foundSecretResource = true

						assert.Equal(t, "mySecret", res.GetName())
						assert.True(t, res.NeedHashSuffix())

						data, err := res.GetFieldValue("data")
						assert.NoError(t, err)
						assert.True(t, len(data.(map[string]interface{})) == 1)

						value, err := res.GetFieldValue("data.PASSWORD")
						assert.NoError(t, err)
						assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("whatever")), value)

						break
					}
				}
				assert.True(t, foundSecretResource)
			},
		},
		{
			name: "withHash_withAppendData",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperSecret
metadata:
  name: mySecret
stringData:
  foo: bar
  baz: whatever
data:
  anotherPassword: Ym9vbQ==
`,
			pluginInputResources: pluginInputResources,
			checkAssertions:      withHashWithAppendDataAssertions,
		},
		{
			name: "withHash_withAppendData_withOverwrite",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperSecret
metadata:
  name: mySecret
stringData:
  foo: bar
  baz: whatever
data:
  anotherPassword: Ym9vbQ==
`,
			pluginInputResources: `
apiVersion: v1
kind: Secret
metadata:
  name: mySecret
type: Opaque
data:
  foo: replace-me!!!
  PASSWORD: d2hhdGV2ZXI=
`,
			checkAssertions: withHashWithAppendDataAssertions,
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

			plugin := NewSuperSecretTransformerPlugin()

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

			testCase.checkAssertions(t, resMap)
		})
	}
}

func TestSuperSecret_assumeTargetWillExistTransformer(t *testing.T) {
	pluginInputResources := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment1
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: myPod1
        image: some-image
        volumeMounts:
        - name: foo
          mountPath: "/etc/foo"
          readOnly: true
      volumes:
      - name: foo
        secret:
          secretName: mySecret
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment2
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: myPod2
        image: some-image
        volumeMounts:
        - name: foo
          mountPath: "/etc/foo"
          readOnly: true
      volumes:
      - name: foo
        secret:
          secretName: mySecret
`
	assertReferencesUpdatedWithHashes := func(t *testing.T, resMap resmap.ResMap) {
		for _, res := range resMap.Resources() {
			if res.GetKind() == "Secret" {
				assert.FailNow(t, "secret should not be present in the stream")
				break
			}
		}

		foundDeployments := map[string]bool{"myDeployment1": false, "myDeployment2": false}
		for _, deploymentName := range []string{"myDeployment1", "myDeployment2"} {
			for _, res := range resMap.Resources() {
				if res.GetKind() == "Deployment" && res.GetName() == deploymentName {
					foundDeployments[deploymentName] = true

					value, err := res.GetFieldValue("spec.template.spec.volumes[0].secret.secretName")
					assert.NoError(t, err)

					match, err := regexp.MatchString("^mySecret-[0-9a-z]+$", value.(string))
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
			if res.GetKind() == "Secret" {
				assert.FailNow(t, "secret should not be present in the stream")
				break
			}
		}

		foundDeployments := map[string]bool{"myDeployment1": false, "myDeployment2": false}
		for _, deploymentName := range []string{"myDeployment1", "myDeployment2"} {
			for _, res := range resMap.Resources() {
				if res.GetKind() == "Deployment" && res.GetName() == deploymentName {
					foundDeployments[deploymentName] = true

					value, err := res.GetFieldValue("spec.template.spec.volumes[0].secret.secretName")
					assert.NoError(t, err)
					assert.Equal(t, "mySecret", value)

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
kind: SuperSecret
metadata:
 name: mySecret
stringData:
 foo: bar
 baz: whatever
`,
			pluginInputResources: pluginInputResources,
			checkAssertions:      assertReferencesUpdatedWithHashes,
		},
		{
			name: "assumeTargetWillExist_canBeTurnedOff",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperSecret
metadata:
 name: mySecret
stringData:
 foo: bar
 baz: whatever
assumeTargetWillExist: false
`,
			pluginInputResources: pluginInputResources,
			checkAssertions:      assertReferencesNotUpdated,
		},
		{
			name: "withHash_withAppendData",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperSecret
metadata:
 name: mySecret
stringData:
 foo: bar
 baz: whatever
assumeTargetWillExist: true
`,
			pluginInputResources: pluginInputResources,
			checkAssertions:      assertReferencesUpdatedWithHashes,
		},
		{
			name: "doesNothing_withoutHash",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperSecret
metadata:
 name: mySecret
stringData:
 foo: bar
 baz: whatever
assumeTargetWillExist: true
options:
  disableNameSuffixHash: true
`,
			pluginInputResources: pluginInputResources,
			checkAssertions:      assertReferencesNotUpdated,
		},
		{
			name: "appendNameSuffixHash_forEmptyData",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperSecret
metadata:
  name: mySecret
assumeTargetWillExist: true
`,
			pluginInputResources: pluginInputResources,
			checkAssertions:      assertReferencesUpdatedWithHashes,
		},
		{
			name: "appendNameSuffixHash_withPrefix",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperSecret
metadata:
  name: mySecret
assumeTargetWillExist: true
prefix: some-service-
`,
			pluginInputResources: pluginInputResources,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				for _, res := range resMap.Resources() {
					if res.GetKind() == "Secret" {
						assert.FailNow(t, "secret should not be present in the stream")
						break
					}
				}

				foundDeployments := map[string]bool{"myDeployment1": false, "myDeployment2": false}
				for _, deploymentName := range []string{"myDeployment1", "myDeployment2"} {
					for _, res := range resMap.Resources() {
						if res.GetKind() == "Deployment" && res.GetName() == deploymentName {
							foundDeployments[deploymentName] = true

							value, err := res.GetFieldValue("spec.template.spec.volumes[0].secret.secretName")
							assert.NoError(t, err)
							refName := value.(string)

							match, err := regexp.MatchString("^some-service-mySecret-[0-9a-z]+$", refName)
							assert.NoError(t, err)
							assert.True(t, match)

							resourceFactory := resmap.NewFactory(resource.NewFactory(
								kunstruct.NewKunstructuredFactoryImpl()), nil)

							plugin := NewSuperSecretGeneratorPlugin()

							err = plugin.Config(resmap.NewPluginHelpers(loader.NewFileLoaderAtRoot(filesys.MakeFsInMemory()), valtest_test.MakeFakeValidator(), resourceFactory), []byte(`
apiVersion: qlik.com/v1
kind: SuperSecret
metadata:
 name: mySecret
assumeTargetWillExist: true
prefix: some-service-
`))

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
				kunstruct.NewKunstructuredFactoryImpl()), transformer.NewFactoryImpl())

			resMap, err := resourceFactory.NewResMapFromBytes([]byte(testCase.pluginInputResources))
			if err != nil {
				t.Fatalf("Err: %v", err)
			}

			plugin := NewSuperSecretTransformerPlugin()

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

			testCase.checkAssertions(t, resMap)
		})
	}
}

func TestSuperSecret_generator(t *testing.T) {
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
kind: SuperSecret
metadata:
  name: mySecret
behavior: create
options:
  disableNameSuffixHash: true
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				foundSecretResource := false
				for _, res := range resMap.Resources() {
					if res.GetKind() == "Secret" {
						foundSecretResource = true
						assert.Equal(t, "mySecret", res.GetName())
						assert.False(t, res.NeedHashSuffix())

						data, err := res.GetFieldValue("data")
						assert.Error(t, err)
						assert.Nil(t, data)

						break
					}
				}
				assert.True(t, foundSecretResource)
			},
		},
		{
			name: "withoutHash_withData",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperSecret
metadata:
  name: mySecret
stringData:
  foo: bar
  baz: whatever
data:
  anotherPassword: Ym9vbQ==
  nonBase64Key: should-generate-blank
behavior: create
options:
  disableNameSuffixHash: true
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				foundSecretResource := false
				for _, res := range resMap.Resources() {
					if res.GetKind() == "Secret" {
						foundSecretResource = true
						assert.Equal(t, "mySecret", res.GetName())
						assert.False(t, res.NeedHashSuffix())

						data, err := res.GetFieldValue("data")
						assert.NoError(t, err)
						assert.True(t, len(data.(map[string]interface{})) == 4)

						value, err := res.GetFieldValue("data.foo")
						assert.NoError(t, err)
						assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("bar")), value)

						value, err = res.GetFieldValue("data.baz")
						assert.NoError(t, err)
						assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("whatever")), value)

						value, err = res.GetFieldValue("data.anotherPassword")
						assert.NoError(t, err)
						assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("boom")), value)

						value, err = res.GetFieldValue("data.nonBase64Key")
						assert.NoError(t, err)
						assert.Equal(t, "", value)

						break
					}
				}
				assert.True(t, foundSecretResource)
			},
		},
		{
			name: "withHash_withoutData",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperSecret
metadata:
  name: mySecret
behavior: create
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				foundSecretResource := false
				for _, res := range resMap.Resources() {
					if res.GetKind() == "Secret" {
						foundSecretResource = true
						assert.Equal(t, "mySecret", res.GetName())
						assert.True(t, res.NeedHashSuffix())

						data, err := res.GetFieldValue("data")
						assert.Error(t, err)
						assert.Nil(t, data)

						break
					}
				}
				assert.True(t, foundSecretResource)
			},
		},
		{
			name: "withHash_withData",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperSecret
metadata:
  name: mySecret
stringData:
  foo: bar
  baz: whatever
data:
  anotherPassword: Ym9vbQ==
behavior: create
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				foundSecretResource := false
				for _, res := range resMap.Resources() {
					if res.GetKind() == "Secret" {
						foundSecretResource = true
						assert.Equal(t, "mySecret", res.GetName())
						assert.True(t, res.NeedHashSuffix())

						data, err := res.GetFieldValue("data")
						assert.NoError(t, err)
						assert.True(t, len(data.(map[string]interface{})) == 3)

						value, err := res.GetFieldValue("data.foo")
						assert.NoError(t, err)
						assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("bar")), value)

						value, err = res.GetFieldValue("data.baz")
						assert.NoError(t, err)
						assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("whatever")), value)

						value, err = res.GetFieldValue("data.anotherPassword")
						assert.NoError(t, err)
						assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("boom")), value)

						break
					}
				}
				assert.True(t, foundSecretResource)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resourceFactory := resmap.NewFactory(resource.NewFactory(
				kunstruct.NewKunstructuredFactoryImpl()), nil)

			plugin := NewSuperSecretGeneratorPlugin()
			err := plugin.Config(resmap.NewPluginHelpers(loader.NewFileLoaderAtRoot(filesys.MakeFsInMemory()), valtest_test.MakeFakeValidator(), resourceFactory), []byte(testCase.pluginConfig))
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
