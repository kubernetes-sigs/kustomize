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
	th.WriteK("/app/base", `
resources:
- deploy.yaml
configMapGenerator:
- name: my-configmap
  literals:	
  - testValue=1
  - otherValue=10
`)
	th.WriteF("/app/base/deploy.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: storefront
spec:
  replicas: 1
`)
}

func writeTestComponent(th kusttest_test.Harness) {
	th.WriteC("/app/comp", `
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
  - testValue=2
  - compValue=5
`)
	th.WriteF("/app/comp/stub.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: stub
spec:
  replicas: 1
`)
}

func writeOverlayProd(th kusttest_test.Harness) {
	th.WriteK("/app/prod", `
resources:
- ../base
- db

components:
- ../comp
`)
	writeDB(th)
}

func writeDB(th kusttest_test.Harness) {
	deployment("db", "/app/prod/db")(th)
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
			runPath: "/app/prod",
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
  compValue: "5"
  otherValue: "10"
  testValue: "2"
kind: ConfigMap
metadata:
  annotations: {}
  labels: {}
  name: comp-my-configmap-kc6k2kmkh9
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
				writeC("/app/additionalcomp", `
configMapGenerator:
- name: my-configmap
  behavior: merge
  literals:	
  - otherValue=9
`),
				writeK("/app/prod", `
resources:
- ../base
- db

components:
- ../comp
- ../additionalcomp
`),
			},
			runPath: "/app/prod",
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
  compValue: "5"
  otherValue: "9"
  testValue: "2"
kind: ConfigMap
metadata:
  annotations: {}
  labels: {}
  name: comp-my-configmap-55249mf5kb
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
				writeC("/app/additionalcomp", `
components:
- ../comp
configMapGenerator:
- name: my-configmap
  behavior: merge
  literals:
  - otherValue=9
`),
				writeK("/app/prod", `
resources:
- ../base
- db

components:
- ../additionalcomp
`),
			},
			runPath: "/app/prod",
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
  compValue: "5"
  otherValue: "9"
  testValue: "2"
kind: ConfigMap
metadata:
  annotations: {}
  labels: {}
  name: comp-my-configmap-55249mf5kb
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
				writeK("/app/repeated", `
resources:
- ../base
- ../prod
`),
			},
			runPath: "/app/repeated",
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
  otherValue: "10"
  testValue: "1"
kind: ConfigMap
metadata:
  name: my-configmap-2g9c94mhb8
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
  compValue: "5"
  otherValue: "10"
  testValue: "2"
kind: ConfigMap
metadata:
  annotations: {}
  labels: {}
  name: comp-my-configmap-kc6k2kmkh9
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
				writeC("/app/direct-component", `
resources:
- ../base
configMapGenerator:
- name: my-configmap
  behavior: merge
  literals:	
  - compValue=5
  - testValue=2
`),
			},
			runPath: "/app/direct-component",
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
  compValue: "5"
  otherValue: "10"
  testValue: "2"
kind: ConfigMap
metadata:
  annotations: {}
  labels: {}
  name: my-configmap-kc6k2kmkh9
`,
		},
		"missing-optional-component-api-version": {
			input: []FileGen{writeTestBase, writeOverlayProd,
				writeF("/app/comp/"+konfig.DefaultKustomizationFileName(), `
kind: Component
configMapGenerator:
- name: my-configmap
  behavior: merge
  literals:	
  - otherValue=9
`),
			},
			runPath: "/app/prod",
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
  otherValue: "9"
  testValue: "1"
kind: ConfigMap
metadata:
  annotations: {}
  labels: {}
  name: my-configmap-5g7gh5mgt5
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
				deployment("proxy", "/app/comp-a/proxy.yaml"),
				writeC("/app/comp-a", `
resources:
- ../base

nameSuffix: "-a"
`),
				writeC("/app/comp-b", `
resources:
- ../base

nameSuffix: "-b"
`),
				writeK("/app/prod", `
components:
- ../comp-a
- ../comp-b`),
			},
			runPath: "/app/prod",
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
  otherValue: "10"
  testValue: "1"
kind: ConfigMap
metadata:
  name: my-configmap-a-b-2g9c94mhb8
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
  otherValue: "10"
  testValue: "1"
kind: ConfigMap
metadata:
  name: my-configmap-b-2g9c94mhb8
`,
		},

		"multiple-bases-can-add-the-same-component-if-it-doesn-not-define-named-entities": {
			input: []FileGen{
				writeC("/app/comp", `
namespace: prod
`),
				writeK("/app/base-a", `
resources:
- proxy.yaml

components:
- ../comp
`),
				deployment("proxy-a", "/app/base-a/proxy.yaml"),
				writeK("/app/base-b", `
resources:
- proxy.yaml

components:
- ../comp
`),
				deployment("proxy-b", "/app/base-b/proxy.yaml"),
				writeK("/app/prod", `
resources:
- proxy.yaml
- ../base-a
- ../base-b
`),
				deployment("proxy-prod", "/app/prod/proxy.yaml"),
			},
			runPath: "/app/prod",
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
				writeK("/app/compinres", `
resources:
- ../base
- ../comp
`),
			},
			runPath:       "app/compinres",
			expectedError: "expected kind != 'Component' for path '/app/comp'",
		},
		"kustomizations-cannot-be-added-to-components": {
			input: []FileGen{writeTestBase, writeTestComponent,
				writeK("/app/kustincomponents", `
components:
- ../base
- ../comp
`),
			},
			runPath: "/app/kustincomponents",
			expectedError: "accumulating components: accumulateDirectory: \"expected kind 'Component' for path " +
				"'/app/base' but got 'Kustomization'",
		},
		"files-cannot-be-added-to-components-list": {
			input: []FileGen{writeTestBase,
				writeF("/app/filesincomponents/stub.yaml", `
apiVersion: v1
kind: Deployment
metadata:
  name: stub
spec:
  replicas: 1
`),
				writeK("/app/filesincomponents", `
components:
- stub.yaml
- ../comp
`),
			},
			runPath:       "/app/filesincomponents",
			expectedError: "'/app/filesincomponents/stub.yaml' must be a directory to be a root",
		},
		"invalid-component-api-version": {
			input: []FileGen{writeTestBase, writeOverlayProd,
				writeF("/app/comp/"+konfig.DefaultKustomizationFileName(), `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Component
configMapGenerator:
- name: my-configmap
  behavior: merge
  literals:	
  - otherValue=9
`),
			},
			runPath:       "/app/prod",
			expectedError: "apiVersion for Component should be kustomize.config.k8s.io/v1alpha1",
		},
		"components-cannot-add-the-same-resource": {
			input: []FileGen{writeTestBase,
				writeC("/app/comp-a", `
resources:
- proxy.yaml
`),
				deployment("proxy", "/app/comp-a/proxy.yaml"),
				writeC("/app/comp-b", `
resources:
- proxy.yaml
`),
				deployment("proxy", "/app/comp-b/proxy.yaml"),
				writeK("/app/prod", `
resources:
- ../base

components:
- ../comp-a
- ../comp-b`),
			},
			runPath:       "/app/prod",
			expectedError: "may not add resource with an already registered id: ~G_v1_Deployment|~X|proxy",
		},
		"components-cannot-add-the-same-base": {
			input: []FileGen{writeTestBase,
				deployment("proxy", "/app/comp-a/proxy.yaml"),
				writeC("/app/comp-a", `
resources:
- ../base
`),
				writeC("/app/comp-b", `
resources:
- ../base
`),
				writeK("/app/prod", `
components:
- ../comp-a
- ../comp-b`),
			},
			runPath:       "/app/prod",
			expectedError: "may not add resource with an already registered id: ~G_v1_Deployment|~X|storefront",
		},
		"components-cannot-add-bases-containing-the-same-resource": {
			input: []FileGen{writeTestBase,
				writeC("/app/comp-a", `
resources:
- ../base-a
`),
				writeK("/app/base-a", `
resources:
- proxy.yaml
`),
				deployment("proxy", "/app/base-a/proxy.yaml"),
				writeC("/app/comp-b", `
resources:
- ../base-b
`),
				writeK("/app/base-b", `
resources:
- proxy.yaml
`),
				deployment("proxy", "/app/base-b/proxy.yaml"),
				writeK("/app/prod", `
resources:
- ../base

components:
- ../comp-a
- ../comp-b`),
			},
			runPath:       "/app/prod",
			expectedError: "may not add resource with an already registered id: ~G_v1_Deployment|~X|proxy",
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
