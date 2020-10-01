package builtins_qlik

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/builtins_qlik/utils"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
)

func TestSearchReplacePlugin(t *testing.T) {
	type searchReplacePluginTestCaseT struct {
		name                 string
		pluginConfig         string
		pluginInputResources string
		checkAssertions      func(*testing.T, resmap.ResMap)
		loaderRootDir        string
	}

	testCases := []searchReplacePluginTestCaseT{
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
			name: "object reference, no match bo replace",
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
   name: Foo
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
				if "far" != fooEnvVar["value"].(string) {
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
		{
			name: "can replace with a blank from object ref",
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
   name: Foo
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
 name: Foo
 labels:
   myproperty: ""
fooSpec:
 test: test
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
		{
			name: "replace label keys",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SearchReplace
metadata:
  name: notImportantHere
target:
  kind: Deployment
path: spec/template/metadata/labels
search: \b[^"]*-messaging-nats-client\b
replace: foo-messaging-nats-client
`,
			pluginInputResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-1
spec:
  template:
    metadata:
      labels:
        app: some-app
        something-messaging-nats-client: "true"
        release: some-release
    spec:
      containers:
      - name: name-1
        image: image-1:latest
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-2
spec:
  template:
    metadata:
      labels:
        app: some-app
        something-messaging-nats-client: "true"
        release: some-release
    spec:
      containers:
      - name: name-2
        image: image-2:latest
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				for _, res := range resMap.Resources() {
					labels, err := res.GetFieldValue("spec.template.metadata.labels")
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}

					appLabel := labels.(map[string]interface{})["app"].(string)
					if "some-app" != appLabel {
						t.Fatalf("unexpected: %v\n", appLabel)
					}

					natsClientLabe := labels.(map[string]interface{})["foo-messaging-nats-client"].(string)
					if "true" != natsClientLabe {
						t.Fatalf("unexpected: %v\n", natsClientLabe)
					}

					releaseLabel := labels.(map[string]interface{})["release"].(string)
					if "some-release" != releaseLabel {
						t.Fatalf("unexpected: %v\n", releaseLabel)
					}
				}
			},
		},
		{
			name: "replace label key for a custom type and a dollar-variable",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SearchReplace
metadata:
  name: notImportantHere
target:
  kind: Engine
path: spec/metadata/labels
search: \$\(PREFIX\)-messaging-nats-client
replace: foo-messaging-nats-client
`,
			pluginInputResources: `
apiVersion: qixmanager.qlik.com/v1
kind: Engine
metadata:
  name: whatever-engine
spec:
  metadata:
    labels:
      $(PREFIX)-messaging-nats-client: "true"
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				for _, res := range resMap.Resources() {
					labels, err := res.GetFieldValue("spec.metadata.labels")
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}

					natsClientLabe := labels.(map[string]interface{})["foo-messaging-nats-client"].(string)
					if "true" != natsClientLabe {
						t.Fatalf("unexpected: %v\n", natsClientLabe)
					}
				}
			},
		},
		{
			name: "replace root",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SearchReplace
metadata:
  name: notImportantHere
target:
  kind: Deployment
path: /
search: \$\(PREFIX\)
replace: foo
`,
			pluginInputResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-1
spec:
  template:
    metadata:
      labels:
        $(PREFIX)-messaging-nats-client: "true"
        $(PREFIX): bar
    spec:
      containers:
      - name: name-1
        image: $(PREFIX)-image:latest
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				expectingFinalDeploymenYaml := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment-1
spec:
  template:
    metadata:
      labels:
        foo: bar
        foo-messaging-nats-client: "true"
    spec:
      containers:
      - image: foo-image:latest
        name: name-1
`
				if resMapYaml, err := resMap.AsYaml(); err != nil {
					t.Fatalf("unexpected error: %v", err)
				} else if string(resMapYaml) != expectingFinalDeploymenYaml {
					t.Fatalf("unexpected %v, but got: %v", expectingFinalDeploymenYaml, string(resMapYaml))
				}
			},
		},
		{
			name: "base64-encoded target and replacement",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SearchReplace
metadata:
  name: notImportantHere
target:
  kind: Secret
  name: keycloak-secret
path: data/idpConfigs
search: \$\(QLIKSENSE_DOMAIN\)
replaceWithObjRef:
  objref:
    apiVersion: qlik.com/v1
    kind: Secret
    name: gke-configs
  fieldref:
    fieldpath: data.qlikSenseDomain
`,
			pluginInputResources: `
apiVersion: v1
kind: Secret
metadata:
  name: keycloak-secret
type: Opaque
data:
  idpConfigs: JChRTElLU0VOU0VfRE9NQUlOKS5iYXIuY29t
---
apiVersion: v1
kind: Secret
metadata:
  name: gke-configs
type: Opaque
data:
  qlikSenseDomain: Zm9v
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				res, err := resMap.GetById(resid.NewResId(resid.Gvk{
					Group:   "",
					Version: "v1",
					Kind:    "Secret",
				}, "keycloak-secret"))
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				base64EncodedIdpConfigs, err := res.GetFieldValue("data.idpConfigs")
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				decodedIdpConfigsBytes, err := base64.StdEncoding.DecodeString(base64EncodedIdpConfigs.(string))
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if string(decodedIdpConfigsBytes) != "foo.bar.com" {
					t.Fatalf("expected %v to equal %v", string(decodedIdpConfigsBytes), "foo.bar.com")
				}
			},
		},
		{
			name: "base64-encoded target and replacement, replace entire (possibly multiline) string",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SearchReplace
metadata:
  name: notImportantHere
target:
  kind: Secret
  name: target-secret
path: data/tls.crt
search: (?s).*
replaceWithObjRef:
  objref:
    apiVersion: qlik.com/v1
    kind: Secret
    name: source-secret
  fieldref:
    fieldpath: data[tls.foo]
`,
			pluginInputResources: `
apiVersion: v1
kind: Secret
metadata:
  name: target-secret
type: Opaque
data:
  tls.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURWRENDQWp5Z0F3SUJBZ0lKQUowdHFNVDVEV3BzTUEwR0NTcUdTSWIzRFFFQkN3VUFNRDh4R0RBV0JnTlYKQkFNTUQyVnNZWE4wYVdNdVpYaGhiWEJzWlRFak1DRUdBMVVFQ2d3YVpXeGhjM1JwWXkxbGJHRnpkR2xqTFd4dgpZMkZzTFdObGNuUXdIaGNOTVRnd05qRXhNVFV4TVRBeldoY05Namd3TkRFNU1UVXhNVEF6V2pBL01SZ3dGZ1lEClZRUUREQTlsYkdGemRHbGpMbVY0WVcxd2JHVXhJekFoQmdOVkJBb01HbVZzWVhOMGFXTXRaV3hoYzNScFl5MXMKYjJOaGJDMWpaWEowTUlJQklqQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FROEFNSUlCQ2dLQ0FRRUFwUUxJZ1Z4QwpkRjI3UFhzL3M4VFh2VGRVeTVwZFFneG9jWi9MenUxblJXbW1GRUc1Mlo5MmRjS1dJMDljdVg5ZUlZZzE0c21ZCkczSmtjb28vNUt0WUtpNmh5dVBtNlZrdGRyU1dCam1VdGxkbHg3UVRLNkxFVlhFeUU3VDZ6QW1GV3lMZTVJMEIKbDdRQlk2dnVoK3g1dlpkSWd3SzVldzBFZmNJUU1Ra2tiMzVkb00xYm41TEJEVklxUzNmNXUxNTArMTM1RitsWQpOc2lFcWhZaVExZm1PMmkzSzBLOW5TMEl3Nm5vNWp2MkZaNnR5bU9zY2wvaWYzdWQzUzUxOTZNTjJtaFpCRFVaCmlwbThlRVkxNVJOa3VQSXBETzhMYkEwZlFOcUcyYXFGa3JybFcrbEdTaDRYZjZLNmdtZkFRdW15K2xrR3RqVlUKZU5OdGF6NnlMMWNkU3dJREFRQUJvMU13VVRBTEJnTlZIUThFQkFNQ0JMQXdFd1lEVlIwbEJBd3dDZ1lJS3dZQgpCUVVIQXdFd0xRWURWUjBSQkNZd0pJSVBaV3hoYzNScFl5NWxlR0Z0Y0d4bGdoRXFMbVZzWVhOMGFXTXVaWGhoCmJYQnNaVEFOQmdrcWhraUc5dzBCQVFzRkFBT0NBUUVBVDFnMFhqelNDRDBBTkF5cDFOWURLU3ZZVUdHcGpSaFkKdUJRYnRwcDUrUDNzd2xvdTNvMDVwdDNydlZ3QmxwK2tjalFwTEJpN3AyRFNuNTdFWHM5eHFEQTBHRjlMSHdpUwpzcGZVYlhRazJIa1E3SGpHb01FSEJLVWpOVUJoZjJkWVRuK1BtbHhobllpZitoU2dkeDdIZ1JuMTh0K0hqTC9JClVlUUkxMHYvdExiT1diZkdBZmlGYjQyUHEvVjg1aGJlNW9mU1VObHVsNFZNOGVXNk1SNlB2b1dYYWJYcVB2ajMKSXpVOVk2UVFoN2dqYmNQN2RmWUZCd3FFRnUxOHlpMUdFbzcxK0NtUjFFL2VpK2kzZDdZTHlMTWhUOHZnMVNDUgozeTB1TVZZOHkxVHpIdS9OTEl0NVozYm1rNFJjcUo1MDVsYmtrTHdzSVFYUzJpMUxud25weXc9PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
---
apiVersion: v1
kind: Secret
metadata:
  name: source-secret
type: Opaque
data:
  tls.foo: Zm9v
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				res, err := resMap.GetById(resid.NewResId(resid.Gvk{
					Group:   "",
					Version: "v1",
					Kind:    "Secret",
				}, "target-secret"))
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				base64EncodedIdpConfigs, err := res.GetFieldValue("data[tls.crt]")
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				decodedIdpConfigsBytes, err := base64.StdEncoding.DecodeString(base64EncodedIdpConfigs.(string))
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if string(decodedIdpConfigsBytes) != "foo" {
					t.Fatalf("expected %v to equal %v", string(decodedIdpConfigsBytes), "foo.bar.com")
				}
			},
		},
		{
			name: "base64-encoded target only",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SearchReplace
metadata:
  name: notImportantHere
target:
  kind: Secret
  name: keycloak-secret
path: data/idpConfigs
search: \$\(QLIKSENSE_DOMAIN\)
replaceWithObjRef:
  objref:
    apiVersion: qlik.com/v1
    kind: ConfigMap
    mame: gke-configs
  fieldref:
    fieldpath: data.qlikSenseDomain
`,
			pluginInputResources: `
apiVersion: v1
kind: Secret
metadata:
  name: keycloak-secret
type: Opaque
data:
  idpConfigs: JChRTElLU0VOU0VfRE9NQUlOKS5iYXIuY29t
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: gke-configs
data:
  qlikSenseDomain: foo
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				res, err := resMap.GetById(resid.NewResId(resid.Gvk{
					Group:   "",
					Version: "v1",
					Kind:    "Secret",
				}, "keycloak-secret"))
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				base64EncodedIdpConfigs, err := res.GetFieldValue("data.idpConfigs")
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				decodedIdpConfigsBytes, err := base64.StdEncoding.DecodeString(base64EncodedIdpConfigs.(string))
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if string(decodedIdpConfigsBytes) != "foo.bar.com" {
					t.Fatalf("expected %v to equal %v", string(decodedIdpConfigsBytes), "foo.bar.com")
				}
			},
		},
		{
			name: "base64-encoded replacement only",
			pluginConfig: `
apiVersion: qlik.com/v1
kind: SearchReplace
metadata:
  name: notImportantHere
target:
  kind: ConfigMap
  name: keycloak-config
path: data/idpConfigs
search: \$\(QLIKSENSE_DOMAIN\)
replaceWithObjRef:
  objref:
    apiVersion: qlik.com/v1
    kind: Secret
    mame: gke-secrets
  fieldref:
    fieldpath: data.qlikSenseDomain
`,
			pluginInputResources: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: keycloak-config
data:
  idpConfigs: $(QLIKSENSE_DOMAIN).bar.com
---
apiVersion: v1
kind: Secret
metadata:
  name: gke-secrets
type: Opaque
data:
  qlikSenseDomain: Zm9v
`,
			checkAssertions: func(t *testing.T, resMap resmap.ResMap) {
				res, err := resMap.GetById(resid.NewResId(resid.Gvk{
					Group:   "",
					Version: "v1",
					Kind:    "ConfigMap",
				}, "keycloak-config"))
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				idpConfigs, err := res.GetFieldValue("data.idpConfigs")
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if idpConfigs.(string) != "foo.bar.com" {
					t.Fatalf("expected %v to equal %v", idpConfigs.(string), "foo.bar.com")
				}
			},
		},
		func() searchReplacePluginTestCaseT {
			tmpDir, err := ioutil.TempDir("", "")
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			semverTag := "v0.0.1"
			subDir, hash, err := setupGitDirWithSubdir(tmpDir, []string{}, []string{"foo", semverTag, "bar"})
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			return searchReplacePluginTestCaseT{
				name:          "replaceWithGitSemverTag",
				loaderRootDir: subDir,
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
replaceWithGitSemverTag:
  default: v0.0.0
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
					expectedVersion := strings.TrimPrefix(semverTag, "v")
					expectedFooValue := fmt.Sprintf("%s-%s", expectedVersion, hash)
					if expectedFooValue != fooEnvVar["value"].(string) {
						t.Fatalf("unexpected: %v\n", fooEnvVar["value"].(string))
					}

					booEnvVar := envVars.([]interface{})[1].(map[string]interface{})
					if "BOO" != booEnvVar["name"].(string) {
						t.Fatalf("unexpected: %v\n", booEnvVar["name"].(string))
					}
					if "farther than it looks" != booEnvVar["value"].(string) {
						t.Fatalf("unexpected: %v\n", booEnvVar["value"].(string))
					}
					_ = os.RemoveAll(tmpDir)
				},
			}
		}(),
	}
	plugin := SearchReplacePlugin{logger: utils.GetLogger("SearchReplacePlugin")}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resourceFactory := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()), nil)

			resMap, err := resourceFactory.NewResMapFromBytes([]byte(testCase.pluginInputResources))
			if err != nil {
				t.Fatalf("Err: %v", err)
			}

			var ldr ifc.Loader
			if testCase.loaderRootDir == "" {
				ldr = loader.NewFileLoaderAtRoot(filesys.MakeFsInMemory())
			} else {
				ldr, err = loader.NewLoader(loader.RestrictionRootOnly, testCase.loaderRootDir, filesys.MakeFsOnDisk())
				if err != nil {
					t.Fatalf("Err: %v", err)
				}
			}

			h := resmap.NewPluginHelpers(ldr, valtest_test.MakeHappyMapValidator(t), resourceFactory)
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
