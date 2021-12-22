// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"fmt"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/konfig"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

type FileGen func(kusttest_test.Harness)

func writeC(path string, content string) FileGen {
	return func(th kusttest_test.Harness) {
		th.WriteC(path, content)
	}
}

func writeF(path string, content string) FileGen {
	return func(th kusttest_test.Harness) {
		th.WriteF(path, content)
	}
}

func writeK(path string, content string) FileGen {
	return func(th kusttest_test.Harness) {
		th.WriteK(path, content)
	}
}

func writeTestBase(th kusttest_test.Harness) {
	th.WriteK("base", `
resources:
- deploy.yaml
configMapGenerator:
- name: my-configmap
  literals:	
  - testValue=purple
  - otherValue=green
`)
	th.WriteF("base/deploy.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: storefront
spec:
  replicas: 1
`)
}

func writeTestComponent(th kusttest_test.Harness) {
	th.WriteC("comp", `
namePrefix: comp-
replicas:
- name: storefront
  count: 3
resources:
- stub.yaml
configMapGenerator:
- name: my-configmap
  behavior: merge
  literals:	
  - testValue=blue
  - compValue=red
`)
	th.WriteF("comp/stub.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: stub
spec:
  replicas: 1
`)
}

func writeOverlayProd(th kusttest_test.Harness) {
	th.WriteK("prod", `
resources:
- ../base
- db

components:
- ../comp
`)
	writeDB(th)
}

func writeDB(th kusttest_test.Harness) {
	deployment("db", "prod/db")(th)
}

func deployment(name string, path string) FileGen {
	return writeF(path, fmt.Sprintf(`
apiVersion: v1
kind: Deployment
metadata:
  name: %s
spec:
  type: Logical
`, name))
}

func TestComponent(t *testing.T) {
	testCases := map[string]struct {
		input          []FileGen
		runPath        string
		expectedOutput string
	}{
		// Components are inserted into the resource hierarchy as the parent of those
		// resources that come before it in the resources list of the parent Kustomization.
		"basic-component": {
			input:   []FileGen{writeTestBase, writeTestComponent, writeOverlayProd},
			runPath: "prod",
			expectedOutput: `
apiVersion: v1
kind: Deployment
metadata:
  name: comp-storefront
spec:
  replicas: 3
---
apiVersion: v1
data:
  compValue: red
  otherValue: green
  testValue: blue
kind: ConfigMap
metadata:
  name: comp-my-configmap-97647ckcmg
---
apiVersion: v1
kind: Deployment
metadata:
  name: comp-db
spec:
  type: Logical
---
apiVersion: v1
kind: Deployment
metadata:
  name: comp-stub
spec:
  replicas: 1
`,
		},
		"multiple-components": {
			input: []FileGen{writeTestBase, writeTestComponent, writeDB,
				writeC("additionalcomp", `
configMapGenerator:
- name: my-configmap
  behavior: merge
  literals:	
  - otherValue=orange
`),
				writeK("prod", `
resources:
- ../base
- db

components:
- ../comp
- ../additionalcomp
`),
			},
			runPath: "prod",
			expectedOutput: `
apiVersion: v1
kind: Deployment
metadata:
  name: comp-storefront
spec:
  replicas: 3
---
apiVersion: v1
data:
  compValue: red
  otherValue: orange
  testValue: blue
kind: ConfigMap
metadata:
  name: comp-my-configmap-g486mb229k
---
apiVersion: v1
kind: Deployment
metadata:
  name: comp-db
spec:
  type: Logical
---
apiVersion: v1
kind: Deployment
metadata:
  name: comp-stub
spec:
  replicas: 1
`,
		},
		"nested-components": {
			input: []FileGen{writeTestBase, writeTestComponent, writeDB,
				writeC("additionalcomp", `
components:
- ../comp
configMapGenerator:
- name: my-configmap
  behavior: merge
  literals:
  - otherValue=orange
`),
				writeK("prod", `
resources:
- ../base
- db

components:
- ../additionalcomp
`),
			},
			runPath: "prod",
			expectedOutput: `
apiVersion: v1
kind: Deployment
metadata:
  name: comp-storefront
spec:
  replicas: 3
---
apiVersion: v1
data:
  compValue: red
  otherValue: orange
  testValue: blue
kind: ConfigMap
metadata:
  name: comp-my-configmap-g486mb229k
---
apiVersion: v1
kind: Deployment
metadata:
  name: comp-db
spec:
  type: Logical
---
apiVersion: v1
kind: Deployment
metadata:
  name: comp-stub
spec:
  replicas: 1
`,
		},
		// If a component sets a name prefix on a base, then that base can also be separately included
		// without being affected by the component in another branch of the resource tree
		"basic-component-with-repeated-base": {
			input: []FileGen{writeTestBase, writeTestComponent, writeOverlayProd,
				writeK("repeated", `
resources:
- ../base
- ../prod
`),
			},
			runPath: "repeated",
			expectedOutput: `
apiVersion: v1
kind: Deployment
metadata:
  name: storefront
spec:
  replicas: 1
---
apiVersion: v1
data:
  otherValue: green
  testValue: purple
kind: ConfigMap
metadata:
  name: my-configmap-9cd648hm8f
---
apiVersion: v1
kind: Deployment
metadata:
  name: comp-storefront
spec:
  replicas: 3
---
apiVersion: v1
data:
  compValue: red
  otherValue: green
  testValue: blue
kind: ConfigMap
metadata:
  name: comp-my-configmap-97647ckcmg
---
apiVersion: v1
kind: Deployment
metadata:
  name: comp-db
spec:
  type: Logical
---
apiVersion: v1
kind: Deployment
metadata:
  name: comp-stub
spec:
  replicas: 1
`,
		},
		"applying-component-directly-should-be-same-as-kustomization": {
			input: []FileGen{writeTestBase, writeTestComponent,
				writeC("direct-component", `
resources:
- ../base
configMapGenerator:
- name: my-configmap
  behavior: merge
  literals:	
  - compValue=red
  - testValue=blue
`),
			},
			runPath: "direct-component",
			expectedOutput: `
apiVersion: v1
kind: Deployment
metadata:
  name: storefront
spec:
  replicas: 1
---
apiVersion: v1
data:
  compValue: red
  otherValue: green
  testValue: blue
kind: ConfigMap
metadata:
  name: my-configmap-97647ckcmg
`,
		},
		"missing-optional-component-api-version": {
			input: []FileGen{writeTestBase, writeOverlayProd,
				writeF("comp/"+konfig.DefaultKustomizationFileName(), `
kind: Component
configMapGenerator:
- name: my-configmap
  behavior: merge
  literals:	
  - otherValue=orange
`),
			},
			runPath: "prod",
			expectedOutput: `
apiVersion: v1
kind: Deployment
metadata:
  name: storefront
spec:
  replicas: 1
---
apiVersion: v1
data:
  otherValue: orange
  testValue: purple
kind: ConfigMap
metadata:
  name: my-configmap-6hhdg8gkdg
---
apiVersion: v1
kind: Deployment
metadata:
  name: db
spec:
  type: Logical
`,
		},
		// See how nameSuffix "-b" is also added to the resources included by "comp-a" because they are in the
		// accumulator when "comp-b" is accumulated.  In practice we could use simple Kustomizations for this example.
		"components-can-add-the-same-base-if-the-first-renames-resources": {
			input: []FileGen{writeTestBase,
				deployment("proxy", "comp-a/proxy.yaml"),
				writeC("comp-a", `
resources:
- ../base

nameSuffix: "-a"
`),
				writeC("comp-b", `
resources:
- ../base

nameSuffix: "-b"
`),
				writeK("prod", `
components:
- ../comp-a
- ../comp-b`),
			},
			runPath: "prod",
			expectedOutput: `
apiVersion: v1
kind: Deployment
metadata:
  name: storefront-a-b
spec:
  replicas: 1
---
apiVersion: v1
data:
  otherValue: green
  testValue: purple
kind: ConfigMap
metadata:
  name: my-configmap-a-b-9cd648hm8f
---
apiVersion: v1
kind: Deployment
metadata:
  name: storefront-b
spec:
  replicas: 1
---
apiVersion: v1
data:
  otherValue: green
  testValue: purple
kind: ConfigMap
metadata:
  name: my-configmap-b-9cd648hm8f
`,
		},

		"multiple-bases-can-add-the-same-component-if-it-doesn-not-define-named-entities": {
			input: []FileGen{
				writeC("comp", `
namespace: prod
`),
				writeK("base-a", `
resources:
- proxy.yaml

components:
- ../comp
`),
				deployment("proxy-a", "base-a/proxy.yaml"),
				writeK("base-b", `
resources:
- proxy.yaml

components:
- ../comp
`),
				deployment("proxy-b", "base-b/proxy.yaml"),
				writeK("prod", `
resources:
- proxy.yaml
- ../base-a
- ../base-b
`),
				deployment("proxy-prod", "prod/proxy.yaml"),
			},
			runPath: "prod",
			// Note that the namepsace has not been applied to proxy-prod because it was not in scope when the
			// component was applied
			expectedOutput: `
apiVersion: v1
kind: Deployment
metadata:
  name: proxy-prod
spec:
  type: Logical
---
apiVersion: v1
kind: Deployment
metadata:
  name: proxy-a
  namespace: prod
spec:
  type: Logical
---
apiVersion: v1
kind: Deployment
metadata:
  name: proxy-b
  namespace: prod
spec:
  type: Logical
`,
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			th := kusttest_test.MakeHarness(t)
			for _, f := range tc.input {
				f(th)
			}
			m := th.Run(tc.runPath, th.MakeDefaultOptions())
			th.AssertActualEqualsExpected(m, tc.expectedOutput)
		})
	}
}

func TestComponentErrors(t *testing.T) {
	testCases := map[string]struct {
		input         []FileGen
		runPath       string
		expectedError string
	}{
		"components-cannot-be-added-to-resources": {
			input: []FileGen{writeTestBase, writeTestComponent,
				writeK("compinres", `
resources:
- ../base
- ../comp
`),
			},
			runPath:       "compinres",
			expectedError: "expected kind != 'Component' for path '/comp'",
		},
		"kustomizations-cannot-be-added-to-components": {
			input: []FileGen{writeTestBase, writeTestComponent,
				writeK("kustincomponents", `
components:
- ../base
- ../comp
`),
			},
			runPath: "kustincomponents",
			expectedError: "accumulating components: accumulateDirectory: \"expected kind 'Component' for path " +
				"'/base' but got 'Kustomization'",
		},
		"files-cannot-be-added-to-components-list": {
			input: []FileGen{writeTestBase,
				writeF("filesincomponents/stub.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: stub
spec:
  replicas: 1
`),
				writeK("filesincomponents", `
components:
- stub.yaml
- ../comp
`),
			},
			runPath:       "filesincomponents",
			expectedError: "'/filesincomponents/stub.yaml' must be a directory to be a root",
		},
		"invalid-component-api-version": {
			input: []FileGen{writeTestBase, writeOverlayProd,
				writeF("comp/"+konfig.DefaultKustomizationFileName(), `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Component
configMapGenerator:
- name: my-configmap
  behavior: merge
  literals:	
  - otherValue=orange
`),
			},
			runPath:       "prod",
			expectedError: "apiVersion for Component should be kustomize.config.k8s.io/v1alpha1",
		},
		"components-cannot-add-the-same-resource": {
			input: []FileGen{writeTestBase,
				writeC("comp-a", `
resources:
- proxy.yaml
`),
				deployment("proxy", "comp-a/proxy.yaml"),
				writeC("comp-b", `
resources:
- proxy.yaml
`),
				deployment("proxy", "comp-b/proxy.yaml"),
				writeK("prod", `
resources:
- ../base

components:
- ../comp-a
- ../comp-b`),
			},
			runPath:       "prod",
			expectedError: "may not add resource with an already registered id: Deployment.v1.[noGrp]/proxy.[noNs]",
		},
		"components-cannot-add-the-same-base": {
			input: []FileGen{writeTestBase,
				deployment("proxy", "comp-a/proxy.yaml"),
				writeC("comp-a", `
resources:
- ../base
`),
				writeC("comp-b", `
resources:
- ../base
`),
				writeK("prod", `
components:
- ../comp-a
- ../comp-b`),
			},
			runPath:       "prod",
			expectedError: "may not add resource with an already registered id: Deployment.v1.[noGrp]/storefront.[noNs]",
		},
		"components-cannot-add-bases-containing-the-same-resource": {
			input: []FileGen{writeTestBase,
				writeC("comp-a", `
resources:
- ../base-a
`),
				writeK("base-a", `
resources:
- proxy.yaml
`),
				deployment("proxy", "base-a/proxy.yaml"),
				writeC("comp-b", `
resources:
- ../base-b
`),
				writeK("base-b", `
resources:
- proxy.yaml
`),
				deployment("proxy", "base-b/proxy.yaml"),
				writeK("prod", `
resources:
- ../base

components:
- ../comp-a
- ../comp-b`),
			},
			runPath:       "prod",
			expectedError: "may not add resource with an already registered id: Deployment.v1.[noGrp]/proxy.[noNs]",
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			th := kusttest_test.MakeHarness(t)
			for _, f := range tc.input {
				f(th)
			}
			err := th.RunWithErr(tc.runPath, th.MakeDefaultOptions())
			if err == nil || !strings.Contains(err.Error(), tc.expectedError) {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}
