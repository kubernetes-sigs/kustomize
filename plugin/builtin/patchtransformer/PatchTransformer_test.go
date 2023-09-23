// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

const (
	someDeploymentResources = `
apiVersion: apps/v1
metadata:
  name: myDeploy
  labels:
    old-label: old-value
kind: Deployment
spec:
  replica: 2
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - name: nginx
        image: nginx
---
apiVersion: apps/v1
metadata:
  name: yourDeploy
  labels:
    new-label: new-value
kind: Deployment
spec:
  replica: 1
  template:
    metadata:
      labels:
        new-label: new-value
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
---
apiVersion: apps/v1
metadata:
  name: myDeploy
  label:
    old-label: old-value
kind: MyKind
spec:
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - name: nginx
        image: nginx
`
)

func TestPatchTransformerMissingFile(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckError(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
path: patch.yaml
`, someDeploymentResources, func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(),
			"'/patch.yaml' doesn't exist") {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}

func TestPatchTransformerBadPatch(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckError(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
patch: "thisIsNotAPatch"
`, someDeploymentResources, func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(),
			"unable to parse SM or JSON patch from ") {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}

func TestPatchTransformerMissingSelector(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckError(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
patch: '[{"op": "add", "path": "/spec/template/spec/dnsPolicy", "value": "ClusterFirst"}]'
`, someDeploymentResources, func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(),
			"must specify a target for JSON patch") {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}

func TestPatchTransformerBlankPatch(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckError(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
patch: "  "
target:
  name: .*Deploy
`, someDeploymentResources, func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(),
			"must specify one of patch and path in") {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}

func TestPatchTransformerBothEmptyPathAndPatch(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckError(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
`, someDeploymentResources, func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(),
			"must specify one of patch and path in") {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}

func TestPatchTransformerBothNonEmptyPathAndPatch(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckError(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
Path: patch.yaml
Patch: "something"
`, someDeploymentResources, func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(),
			"patch and path can't be set at the same time") {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}

const (
	multipleSMPatchesFile = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  template:
    spec:
      containers:
      - image: public.ecr.aws/nginx/nginx:mainline
        name: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: yourDeploy
spec:
  template:
    metadata:
      labels:
        new-label: new-value-with-multipleSMPatchesFile
`

	multipleSMPatchesSuccesfulResult = `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    old-label: old-value
  name: myDeploy
spec:
  replica: 2
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - image: public.ecr.aws/nginx/nginx:mainline
        name: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    new-label: new-value
  name: yourDeploy
spec:
  replica: 1
  template:
    metadata:
      labels:
        new-label: new-value-with-multipleSMPatchesFile
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx
---
apiVersion: apps/v1
kind: MyKind
metadata:
  label:
    old-label: old-value
  name: myDeploy
spec:
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - image: nginx
        name: nginx
`
)

func TestMultipleSMPatchesWithFilePath(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()

	th.WriteF(`multiplepatches.yaml`, multipleSMPatchesFile)

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
Path: multiplepatches.yaml
`, someDeploymentResources, multipleSMPatchesSuccesfulResult)
}

func TestMultipleSMPatchesWithPatch(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
patch: |-
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: myDeploy
  spec:
    template:
      spec:
        containers:
        - image: public.ecr.aws/nginx/nginx:mainline
          name: nginx
  ---
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: yourDeploy
  spec:
    template:
      metadata:
        labels:
          new-label: new-value-with-multipleSMPatchesFile
`, someDeploymentResources, multipleSMPatchesSuccesfulResult)
}

func TestMultipleSMPatchesAndTarget(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()

	th.WriteF(`multiplepatches.yaml`, multipleSMPatchesFile)

	th.RunTransformerAndCheckError(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
Path: multiplepatches.yaml
target:
  name: .*Deploy
  kind: Deployment
`, someDeploymentResources, func(t *testing.T, err error) {
		t.Helper()
		require.ErrorContains(t, err, "Multiple Strategic-Merge Patches in one `patches` entry is not allowed to set `patches.target` field")
	})
}

const (
	oneDeployment = `
apiVersion: apps/v1
metadata:
  name: oneDeploy
kind: Deployment
spec:
  replica: 1
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
      - name: sidecar
        image: busybox:1.36.1
`
	multiplePatchTransformerConfig = `
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
patch: |-
  - op: replace
    path: /spec/template/spec/containers/0/image
    value: nginx:latest
  ---
  - op: replace
    path: /spec/template/spec/containers/1/image
    value: busybox:latest
target:
  name: .*Deploy
  kind: Deployment
`
)

func TestPatchTransformerNotAllowedMultipleJsonPatches(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()

	// TODO(5049): Multiple JSON patch Yaml documents are not allowed and need to return error.
	th.RunTransformerAndCheckResult(multiplePatchTransformerConfig, oneDeployment, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: oneDeploy
spec:
  replica: 1
  template:
    spec:
      containers:
      - image: nginx:latest
        name: nginx
      - image: busybox:1.36.1
        name: sidecar
`)
	// th.RunTransformerAndCheckError(multiplePatchTransformerConfig, oneDeployment, func(t *testing.T, err error) {
	// 	t.Helper()
	// 	if err == nil {
	// 		t.Fatalf("expected error")
	// 	}
	// 	if !strings.Contains(err.Error(),
	// 		"Multiple Json6902 Patch in 'patches' is not allowed.") {
	// 		t.Fatalf("unexpected err: %v", err)
	// 	}
	// })
}

func TestPatchTransformerFromFiles(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()

	th.WriteF("patch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeploy
spec:
  replica: 3
`)

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
path: patch.yaml
target:
  name: .*Deploy
`,
		someDeploymentResources,
		`
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    old-label: old-value
  name: myDeploy
spec:
  replica: 3
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - image: nginx
        name: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    new-label: new-value
  name: yourDeploy
spec:
  replica: 3
  template:
    metadata:
      labels:
        new-label: new-value
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx
---
apiVersion: apps/v1
kind: MyKind
metadata:
  label:
    old-label: old-value
  name: myDeploy
spec:
  replica: 3
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - image: nginx
        name: nginx
`)
}

func TestPatchTransformerSmpSidecars(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()

	th.WriteF("patch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: not-important
spec:
  template:
    spec:
      containers:
      - name: istio-proxy
        image: docker.io/istio/proxyv2
        args:
        - proxy
        - sidecar
`)
	rm := th.LoadAndRunTransformer(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
path: patch.yaml
target:
  name: myDeploy
`, someDeploymentResources)
	th.AssertActualEqualsExpectedNoIdAnnotations(rm, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    old-label: old-value
  name: myDeploy
spec:
  replica: 2
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - args:
        - proxy
        - sidecar
        image: docker.io/istio/proxyv2
        name: istio-proxy
      - image: nginx
        name: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    new-label: new-value
  name: yourDeploy
spec:
  replica: 1
  template:
    metadata:
      labels:
        new-label: new-value
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx
---
apiVersion: apps/v1
kind: MyKind
metadata:
  label:
    old-label: old-value
  name: myDeploy
spec:
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - args:
        - proxy
        - sidecar
        image: docker.io/istio/proxyv2
        name: istio-proxy
      - image: nginx
        name: nginx
`)
}

func TestPatchTransformerWithInlineJson(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
patch: '[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value": "nginx:latest"}]'
target:
  name: .*Deploy
  kind: Deployment
`, someDeploymentResources,
		`
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    old-label: old-value
  name: myDeploy
spec:
  replica: 2
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - image: nginx:latest
        name: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    new-label: new-value
  name: yourDeploy
spec:
  replica: 1
  template:
    metadata:
      labels:
        new-label: new-value
    spec:
      containers:
      - image: nginx:latest
        name: nginx
---
apiVersion: apps/v1
kind: MyKind
metadata:
  label:
    old-label: old-value
  name: myDeploy
spec:
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - image: nginx
        name: nginx
`)
}

func TestPatchTransformerWithInlineYaml(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
target:
  name: .*Deploy
  kind: Deployment
patch: |-
  apiVersion: apps/v1
  metadata:
    name: myDeploy
  kind: Deployment
  spec:
    replica: 77
`, someDeploymentResources, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    old-label: old-value
  name: myDeploy
spec:
  replica: 77
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - image: nginx
        name: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    new-label: new-value
  name: yourDeploy
spec:
  replica: 77
  template:
    metadata:
      labels:
        new-label: new-value
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx
---
apiVersion: apps/v1
kind: MyKind
metadata:
  label:
    old-label: old-value
  name: myDeploy
spec:
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - image: nginx
        name: nginx
`)
}

func TestPatchTransformerWithInlineYamlRegexTarget(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
target:
  name: .*Deploy
  kind: Deployment|MyKind
  group: \w{4}
  version: v\d
patch: |-
  apiVersion: apps/v1
  metadata:
    name: myDeploy
  kind: Deployment
  spec:
    replica: 77
`, someDeploymentResources, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    old-label: old-value
  name: myDeploy
spec:
  replica: 77
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - image: nginx
        name: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    new-label: new-value
  name: yourDeploy
spec:
  replica: 77
  template:
    metadata:
      labels:
        new-label: new-value
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx
---
apiVersion: apps/v1
kind: MyKind
metadata:
  label:
    old-label: old-value
  name: myDeploy
spec:
  replica: 77
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - image: nginx
        name: nginx
`)
}

func TestPatchTransformerWithPatchDelete(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
target:
  name: myDeploy
  kind: Deployment
patch: |-
  apiVersion: apps/v1
  metadata:
    name: myDeploy
  kind: Deployment
  $patch: delete
`, someDeploymentResources, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    new-label: new-value
  name: yourDeploy
spec:
  replica: 1
  template:
    metadata:
      labels:
        new-label: new-value
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx
---
apiVersion: apps/v1
kind: MyKind
metadata:
  label:
    old-label: old-value
  name: myDeploy
spec:
  template:
    metadata:
      labels:
        old-label: old-value
    spec:
      containers:
      - image: nginx
        name: nginx
`)
}

const anIngressResource = `apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: my-ingress
spec:
  rules:
  - host: foo.bar.com
    http:
      paths:
      - path: /
        backend:
          serviceName: homepage
          servicePort: 8888
      - path: /api
        backend:
          serviceName: my-api
          servicePort: 7701
      - path: /test
        backend:
          serviceName: hello
          servicePort: 7702
`

func TestPatchTransformerJson(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()
	th.WriteF("patch.json", `[
  {"op": "replace",
   "path": "/spec/rules/0/host",
   "value": "foo.bar.io"},

  {"op": "replace",
   "path": "/spec/rules/0/http/paths/0/backend/servicePort",
   "value": 80},

  {"op": "add",
   "path": "/spec/rules/0/http/paths/1",
   "value": { "path": "/healthz", "backend": {"servicePort":7700} }}
]
`)

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
path: patch.json
target:
  kind: Ingress
`, anIngressResource, `
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: my-ingress
spec:
  rules:
  - host: foo.bar.io
    http:
      paths:
      - backend:
          serviceName: homepage
          servicePort: 80
        path: /
      - backend:
          servicePort: 7700
        path: /healthz
      - backend:
          serviceName: my-api
          servicePort: 7701
        path: /api
      - backend:
          serviceName: hello
          servicePort: 7702
        path: /test
`)
}

func TestPatchTransformerJsonAsYaml(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()
	th.WriteF("patch.yaml", `
- op: add
  path: /spec/rules/0/http/paths/-
  value:
    path: '/canada'
    backend:
      serviceName: hoser
      servicePort: 7703
`)

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: notImportantHere
path: patch.yaml
target:
  kind: Ingress
`, anIngressResource, `
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: my-ingress
spec:
  rules:
  - host: foo.bar.com
    http:
      paths:
      - backend:
          serviceName: homepage
          servicePort: 8888
        path: /
      - backend:
          serviceName: my-api
          servicePort: 7701
        path: /api
      - backend:
          serviceName: hello
          servicePort: 7702
        path: /test
      - backend:
          serviceName: hoser
          servicePort: 7703
        path: /canada
`)
}

// test for https://github.com/kubernetes-sigs/kustomize/issues/2767
// currently documents broken state.  resulting ports: should have both
// take-over-the-world and disappearing-act on 8080
func TestPatchTransformerSimilarArrays(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: test-transformer
patch: |-
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: test-transformer
    labels:
      test-transformer: did-my-job
target:
  kind: Deployment
  name: test-deployment
`, `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          name: disappearing-act
          protocol: TCP
        - containerPort: 8080
          name: take-over-the-world
          protocol: TCP
`, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    test-transformer: did-my-job
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          name: take-over-the-world
          protocol: TCP
`)
}

func TestPatchTransformerAnchor(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PatchTransformer")
	defer th.Reset()

	th.RunTransformerAndCheckResult(`
apiVersion: builtin
kind: PatchTransformer
metadata:
  name: test-transformer
patch: |-
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: test-deployment
  spec:
    selector:
      matchLabels:
        app: &name test-label
    template:
      metadata:
        labels:
          app: *name
target:
  kind: Deployment
  name: test-deployment
`, `apiVersion: apps/v1
kind: Deployment
metadata:
  name: &name test-deployment
spec:
  selector:
    matchLabels:
      app: *name
  template:
    metadata:
      labels:
        app: *name
    spec:
      containers:
      - image: test-image
        name: *name
`, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  selector:
    matchLabels:
      app: test-label
  template:
    metadata:
      labels:
        app: test-label
    spec:
      containers:
      - image: test-image
        name: test-deployment
`)
}
