package builtins_qlik

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/builtins_qlik/utils/loadertest"
	"sigs.k8s.io/kustomize/api/internal/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
)

func TestSuperVars(t *testing.T) {
	pluginInputResources := `
apiVersion: qlik.com/v1
kind: SuperSecret
metadata:
  name: my-secret
  labels:
    myproperty: propertyvalue
stringData:
  myproperty: $(MYPROPERTY)-something
---
apiVersion: qlik.com/v1
kind: SuperConfigMap 
metadata:
  name: my-configmap
  labels:
    myproperty: propertyvalue-2
data:
  myproperty: $(MYPROPERTY2)-something
`
	varReferenceContent := `
varReference:
- path: stringData/myproperty
  kind: SuperSecret 
- path: data/myproperty
  kind: SuperConfigMap 
`

	var testCases = []struct {
		name                   string
		pluginConfig           string
		pluginInputResources   string
		varReferenceContent    string
		transformErrorExpected bool
		checkAssertions        func(*testing.T, resmap.ResMap)
	}{
		{
			name: "success",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperVars 
metadata:
  name: notImportantHere
configurations:
- varreference.yaml
vars:
- name: MYPROPERTY
  objref:
    apiVersion: qlik.com/v1
    kind: SuperSecret
    name: my-secret
  fieldref:
    fieldpath: metadata.labels.myproperty 
- name: MYPROPERTY2
  objref:
    apiVersion: qlik.com/v1
    kind: SuperConfigMap 
    name: my-configmap
  fieldref:
    fieldpath: metadata.labels.myproperty 
`,
			varReferenceContent:    varReferenceContent,
			pluginInputResources:   pluginInputResources,
			transformErrorExpected: false,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				res, err := resMap.GetById(resid.NewResId(resid.Gvk{
					Group:   "qlik.com",
					Version: "v1",
					Kind:    "SuperSecret",
				}, "my-secret"))
				assert.NoError(t, err)
				assert.NotNil(t, res)

				val, err := res.GetFieldValue("stringData.myproperty")
				assert.NoError(t, err)

				assert.Equal(t, "propertyvalue-something", val.(string))

				res, err = resMap.GetById(resid.NewResId(resid.Gvk{
					Group:   "qlik.com",
					Version: "v1",
					Kind:    "SuperConfigMap",
				}, "my-configmap"))
				assert.NoError(t, err)
				assert.NotNil(t, res)

				val, err = res.GetFieldValue("data.myproperty")
				assert.NoError(t, err)

				assert.Equal(t, "propertyvalue-2-something", val.(string))
			},
		},
		{
			name: "some_unresolved_transform_fails",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperVars 
metadata:
  name: notImportantHere
configurations:
- varreference.yaml
vars:
- name: MYPROPERTY
  objref:
    apiVersion: qlik.com/v1
    kind: SuperSecret
    name: my-secret
  fieldref:
    fieldpath: metadata.labels.myproperty 
- name: MYPROPERTY2
  objref:
    apiVersion: qlik.com/v1
    kind: SuperConfigMap 
    name: my-configmap
  fieldref:
    fieldpath: metadata.labels.not-there 
`,
			varReferenceContent:    varReferenceContent,
			pluginInputResources:   pluginInputResources,
			transformErrorExpected: true,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				assert.FailNow(t, "should not be here!")
			},
		},
		{
			name: "some_not_substituted_transform_succeeds",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperVars 
metadata:
  name: notImportantHere
configurations:
- varreference.yaml
vars:
- name: MYPROPERTY
  objref:
    apiVersion: qlik.com/v1
    kind: SuperSecret
    name: my-secret
  fieldref:
    fieldpath: metadata.labels.myproperty
`,
			varReferenceContent:    varReferenceContent,
			pluginInputResources:   pluginInputResources,
			transformErrorExpected: false,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				res, err := resMap.GetById(resid.NewResId(resid.Gvk{
					Group:   "qlik.com",
					Version: "v1",
					Kind:    "SuperSecret",
				}, "my-secret"))
				assert.NoError(t, err)
				assert.NotNil(t, res)

				val, err := res.GetFieldValue("stringData.myproperty")
				assert.NoError(t, err)

				assert.Equal(t, "propertyvalue-something", val.(string))

				res, err = resMap.GetById(resid.NewResId(resid.Gvk{
					Group:   "qlik.com",
					Version: "v1",
					Kind:    "SuperConfigMap",
				}, "my-configmap"))
				assert.NoError(t, err)
				assert.NotNil(t, res)

				val, err = res.GetFieldValue("data.myproperty")
				assert.NoError(t, err)

				assert.Equal(t, "$(MYPROPERTY2)-something", val.(string))
			},
		},
		{
			name: "no_substitution_without_varreference_config",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SuperVars 
metadata:
  name: notImportantHere
vars:
- name: MYPROPERTY
  objref:
    apiVersion: qlik.com/v1
    kind: SuperSecret
    name: my-secret
  fieldref:
    fieldpath: metadata.labels.myproperty 
- name: MYPROPERTY2
  objref:
    apiVersion: qlik.com/v1
    kind: SuperConfigMap 
    name: my-configmap
  fieldref:
    fieldpath: metadata.labels.myproperty 
`,
			varReferenceContent:    "",
			pluginInputResources:   pluginInputResources,
			transformErrorExpected: false,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				res, err := resMap.GetById(resid.NewResId(resid.Gvk{
					Group:   "qlik.com",
					Version: "v1",
					Kind:    "SuperSecret",
				}, "my-secret"))
				assert.NoError(t, err)
				assert.NotNil(t, res)

				val, err := res.GetFieldValue("stringData.myproperty")
				assert.NoError(t, err)

				assert.Equal(t, "$(MYPROPERTY)-something", val.(string))

				res, err = resMap.GetById(resid.NewResId(resid.Gvk{
					Group:   "qlik.com",
					Version: "v1",
					Kind:    "SuperConfigMap",
				}, "my-configmap"))
				assert.NoError(t, err)
				assert.NotNil(t, res)

				val, err = res.GetFieldValue("data.myproperty")
				assert.NoError(t, err)

				assert.Equal(t, "$(MYPROPERTY2)-something", val.(string))
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
			if len(testCase.varReferenceContent) > 0 {
				err = ldr.AddFile("/varreference.yaml", []byte(testCase.varReferenceContent))
				if err != nil {
					t.Fatalf("Err: %v", err)
				}
			}

			plugin := NewSuperVarsPlugin()
			err = plugin.Config(resmap.NewPluginHelpers(ldr, valtest_test.MakeFakeValidator(), resourceFactory), []byte(testCase.pluginConfig))
			if err != nil {
				t.Fatalf("Err: %v", err)
			}

			err = plugin.Transform(resMap)
			if err != nil && !testCase.transformErrorExpected {
				t.Fatalf("Err: %v", err)
			}

			for _, res := range resMap.Resources() {
				fmt.Printf("--res: %v\n", res.String())
			}

			if err == nil {
				testCase.checkAssertions(t, resMap)
			}
		})
	}
}
