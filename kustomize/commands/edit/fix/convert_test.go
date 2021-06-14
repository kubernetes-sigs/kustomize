// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fix

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v4/commands/internal/testutils"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestFixVarsSimple(t *testing.T) {
	kustomization := []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- pod.yaml

vars:
- name: SOME_SECRET_NAME
  objref:
    kind: Secret
    name: my-secret
    apiVersion: v1
`)
	pod := []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - image: myimage
    name: hello
    env:
    - name: SECRET_TOKEN
      value: $(SOME_SECRET_NAME)
`)

	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(fSys, kustomization)
	fSys.WriteFile("pod.yaml", pod)
	cmd := NewCmdFix(fSys, os.Stdout)
	assert.NoError(t, cmd.Flags().Set("vars", "true"))
	assert.NoError(t, cmd.RunE(cmd, nil))
	content, err := testutils_test.ReadTestKustomization(fSys)
	assert.NoError(t, err)

	assert.Equal(t, `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- pod.yaml

replacements:
- source:
    kind: Secret
    name: my-secret
    version: v1
  targets:
  - fieldPaths:
    - spec.containers.0.env.0.value
    select:
      kind: Pod
      name: my-pod
      version: v1
`, string(content))

	content, err = fSys.ReadFile("pod.yaml")
	assert.NoError(t, err)
	assert.Equal(t, `
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - image: myimage
    name: hello
    env:
    - name: SECRET_TOKEN
      value: SOME_SECRET_NAME_PLACEHOLDER
`, string(content))
}

func TestFixVarsDelimiterPrefix(t *testing.T) {
	kustomization := []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- pod.yaml

vars:
- name: SOME_SECRET_NAME
  objref:
    kind: Secret
    name: my-secret
    apiVersion: v1
`)
	pod := []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - image: myimage
    name: hello
    env:
    - name: SECRET_TOKEN
      value: $(SOME_SECRET_NAME)/path
`)

	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(fSys, kustomization)
	fSys.WriteFile("pod.yaml", pod)
	cmd := NewCmdFix(fSys, os.Stdout)
	assert.NoError(t, cmd.Flags().Set("vars", "true"))
	assert.NoError(t, cmd.RunE(cmd, nil))
	content, err := testutils_test.ReadTestKustomization(fSys)
	assert.NoError(t, err)

	assert.Equal(t, `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- pod.yaml

replacements:
- source:
    kind: Secret
    name: my-secret
    version: v1
  targets:
  - fieldPaths:
    - spec.containers.0.env.0.value
    options:
      delimiter: /
    select:
      kind: Pod
      name: my-pod
      version: v1
`, string(content))

	content, err = fSys.ReadFile("pod.yaml")
	assert.NoError(t, err)
	assert.Equal(t, `
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - image: myimage
    name: hello
    env:
    - name: SECRET_TOKEN
      value: SOME_SECRET_NAME_PLACEHOLDER/path
`, string(content))
}

func TestFixVarsDelimiterSuffix(t *testing.T) {
	kustomization := []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- pod.yaml

vars:
- name: SOME_SECRET_NAME
  objref:
    kind: Secret
    name: my-secret
    apiVersion: v1
`)
	pod := []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - image: myimage
    name: hello
    env:
    - name: SECRET_TOKEN
      value: path/$(SOME_SECRET_NAME)
`)

	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(fSys, kustomization)
	fSys.WriteFile("pod.yaml", pod)
	cmd := NewCmdFix(fSys, os.Stdout)
	assert.NoError(t, cmd.Flags().Set("vars", "true"))
	assert.NoError(t, cmd.RunE(cmd, nil))
	content, err := testutils_test.ReadTestKustomization(fSys)
	assert.NoError(t, err)

	assert.Equal(t, `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- pod.yaml

replacements:
- source:
    kind: Secret
    name: my-secret
    version: v1
  targets:
  - fieldPaths:
    - spec.containers.0.env.0.value
    options:
      delimiter: /
      index: 1
    select:
      kind: Pod
      name: my-pod
      version: v1
`, string(content))

	content, err = fSys.ReadFile("pod.yaml")
	assert.NoError(t, err)
	assert.Equal(t, `
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - image: myimage
    name: hello
    env:
    - name: SECRET_TOKEN
      value: path/SOME_SECRET_NAME_PLACEHOLDER
`, string(content))
}

func TestFixVarsDelimiter(t *testing.T) {
	kustomization := []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- pod.yaml

vars:
- name: SOME_SECRET_NAME
  objref:
    kind: Secret
    name: my-secret
    apiVersion: v1
`)
	pod := []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - image: myimage
    name: hello
    env:
    - name: SECRET_TOKEN
      value: my/path/$(SOME_SECRET_NAME)/secret/path
`)

	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(fSys, kustomization)
	fSys.WriteFile("pod.yaml", pod)
	cmd := NewCmdFix(fSys, os.Stdout)
	assert.NoError(t, cmd.Flags().Set("vars", "true"))
	assert.NoError(t, cmd.RunE(cmd, nil))
	content, err := testutils_test.ReadTestKustomization(fSys)
	assert.NoError(t, err)

	assert.Equal(t, `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- pod.yaml

replacements:
- source:
    kind: Secret
    name: my-secret
    version: v1
  targets:
  - fieldPaths:
    - spec.containers.0.env.0.value
    options:
      delimiter: /
      index: 2
    select:
      kind: Pod
      name: my-pod
      version: v1
`, string(content))

	content, err = fSys.ReadFile("pod.yaml")
	assert.NoError(t, err)
	assert.Equal(t, `
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - image: myimage
    name: hello
    env:
    - name: SECRET_TOKEN
      value: my/path/SOME_SECRET_NAME_PLACEHOLDER/secret/path
`, string(content))
}

func TestFixVarsNotDelimited(t *testing.T) {
	kustomization := []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- pod.yaml

vars:
- name: SOME_SECRET_NAME
  objref:
    kind: Secret
    name: my-secret
    apiVersion: v1
`)
	pod := []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - image: myimage
    name: hello
    env:
    - name: SECRET_TOKEN
      value: var$(SOME_SECRET_NAME)/path
`)

	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(fSys, kustomization)
	fSys.WriteFile("pod.yaml", pod)
	cmd := NewCmdFix(fSys, os.Stdout)
	assert.NoError(t, cmd.Flags().Set("vars", "true"))
	err := cmd.RunE(cmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error with pod.yaml: cannot convert all vars to replacements; $(SOME_SECRET_NAME) is not delimited")
}

func TestFixVarsWithPatchBasic(t *testing.T) {
	kustomization := []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- pod.yaml

patches:
- path: patch.yaml
  target:
    kind: Pod
    name: my-pod

vars:
- name: SOME_SECRET_NAME
  objref:
    kind: Secret
    name: my-secret
    apiVersion: v1
`)
	pod := []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
  labels:
    foo: $(SOME_SECRET_NAME)
spec:
  containers:
  - image: myimage
    name: hello
`)
	patch := []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - image: myimage
    name: hello
    env:
    - name: SECRET_TOKEN
      value: $(SOME_SECRET_NAME)
`)

	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(fSys, kustomization)
	fSys.WriteFile("pod.yaml", pod)
	fSys.WriteFile("patch.yaml", patch)
	cmd := NewCmdFix(fSys, os.Stdout)
	assert.NoError(t, cmd.Flags().Set("vars", "true"))
	assert.NoError(t, cmd.RunE(cmd, nil))
	content, err := testutils_test.ReadTestKustomization(fSys)
	assert.NoError(t, err)

	assert.Equal(t, `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- pod.yaml

patches:
- path: patch.yaml
  target:
    kind: Pod
    name: my-pod

replacements:
- source:
    kind: Secret
    name: my-secret
    version: v1
  targets:
  - fieldPaths:
    - metadata.labels.foo
    select:
      kind: Pod
      name: my-pod
      version: v1
  - fieldPaths:
    - spec.containers.0.env.0.value
    select:
      kind: Pod
      name: my-pod
`, string(content))

	content, err = fSys.ReadFile("pod.yaml")
	assert.NoError(t, err)
	assert.Equal(t, `
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
  labels:
    foo: SOME_SECRET_NAME_PLACEHOLDER
spec:
  containers:
  - image: myimage
    name: hello
`, string(content))

	content, err = fSys.ReadFile("patch.yaml")
	assert.NoError(t, err)
	assert.Equal(t, `
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - image: myimage
    name: hello
    env:
    - name: SECRET_TOKEN
      value: SOME_SECRET_NAME_PLACEHOLDER
`, string(content))
}

func TestFixVarsWithPatchDifferentName(t *testing.T) {
	kustomization := []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- pod.yaml

patches:
- path: patch.yaml
  target:
    kind: Pod
    name: my-pod

vars:
- name: SOME_SECRET_NAME
  objref:
    kind: Secret
    name: my-secret
    apiVersion: v1
`)
	pod := []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
  labels:
    foo: $(SOME_SECRET_NAME)
spec:
  containers:
  - image: myimage
    name: hello
`)
	patch := []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: doesnt-matter
spec:
  containers:
  - image: myimage
    name: hello
    env:
    - name: SECRET_TOKEN
      value: $(SOME_SECRET_NAME)
`)

	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(fSys, kustomization)
	fSys.WriteFile("pod.yaml", pod)
	fSys.WriteFile("patch.yaml", patch)
	cmd := NewCmdFix(fSys, os.Stdout)
	assert.NoError(t, cmd.Flags().Set("vars", "true"))
	assert.NoError(t, cmd.RunE(cmd, nil))
	content, err := testutils_test.ReadTestKustomization(fSys)
	assert.NoError(t, err)

	// the second replacement target is from the patch, and used the patch's target selector
	// as the replacement's target selector. Note that the replacement target uses the pod
	// name 'my-pod', and not the name provided in the patch.
	assert.Equal(t, `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- pod.yaml

patches:
- path: patch.yaml
  target:
    kind: Pod
    name: my-pod

replacements:
- source:
    kind: Secret
    name: my-secret
    version: v1
  targets:
  - fieldPaths:
    - metadata.labels.foo
    select:
      kind: Pod
      name: my-pod
      version: v1
  - fieldPaths:
    - spec.containers.0.env.0.value
    select:
      kind: Pod
      name: my-pod
`, string(content))

	content, err = fSys.ReadFile("pod.yaml")
	assert.NoError(t, err)
	assert.Equal(t, `
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
  labels:
    foo: SOME_SECRET_NAME_PLACEHOLDER
spec:
  containers:
  - image: myimage
    name: hello
`, string(content))

	content, err = fSys.ReadFile("patch.yaml")
	assert.NoError(t, err)
	assert.Equal(t, `
apiVersion: v1
kind: Pod
metadata:
  name: doesnt-matter
spec:
  containers:
  - image: myimage
    name: hello
    env:
    - name: SECRET_TOKEN
      value: SOME_SECRET_NAME_PLACEHOLDER
`, string(content))
}

func TestFixVarsWithPatchNameChange(t *testing.T) {
	kustomization := []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- pod.yaml

patches:
- path: patch.yaml
  target:
    kind: Pod
    name: my-pod
  options:
    allowNameChange: true

vars:
- name: SOME_SECRET_NAME
  objref:
    kind: Secret
    name: my-secret
    apiVersion: v1
`)
	pod := []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - image: myimage
    name: hello
`)
	patch := []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: doesnt-matter
spec:
  containers:
  - image: myimage
    name: hello
    env:
    - name: SECRET_TOKEN
      value: $(SOME_SECRET_NAME)
`)

	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(fSys, kustomization)
	fSys.WriteFile("pod.yaml", pod)
	fSys.WriteFile("patch.yaml", patch)
	cmd := NewCmdFix(fSys, os.Stdout)
	assert.NoError(t, cmd.Flags().Set("vars", "true"))
	assert.NoError(t, cmd.RunE(cmd, nil))
	content, err := testutils_test.ReadTestKustomization(fSys)
	assert.NoError(t, err)

	// The replacement target from the patch uses the patch's name because
	// allowNameChange is set to true.
	assert.Equal(t, `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- pod.yaml

patches:
- options:
    allowNameChange: true
  path: patch.yaml
  target:
    kind: Pod
    name: my-pod

replacements:
- source:
    kind: Secret
    name: my-secret
    version: v1
  targets:
  - fieldPaths:
    - spec.containers.0.env.0.value
    select:
      kind: Pod
      name: doesnt-matter
`, string(content))

	content, err = fSys.ReadFile("pod.yaml")
	assert.NoError(t, err)
	assert.Equal(t, `
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - image: myimage
    name: hello
`, string(content))

	content, err = fSys.ReadFile("patch.yaml")
	assert.NoError(t, err)
	assert.Equal(t, `
apiVersion: v1
kind: Pod
metadata:
  name: doesnt-matter
spec:
  containers:
  - image: myimage
    name: hello
    env:
    - name: SECRET_TOKEN
      value: SOME_SECRET_NAME_PLACEHOLDER
`, string(content))
}

func TestFixVarsWithPatchMultipleResources(t *testing.T) {
	kustomization := []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- pod.yaml

patches:
- path: patch.yaml
  target:
    kind: Pod

vars:
- name: SOME_SECRET_NAME
  objref:
    kind: Secret
    name: my-secret
    apiVersion: v1
`)
	pod := []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: my-pod-1
spec:
  containers:
  - image: myimage
    name: hello
---
apiVersion: v1
kind: Pod
metadata:
  name: my-pod-2
spec:
  containers:
  - image: myimage
    name: hello
`)
	patch := []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: doesnt-matter
spec:
  containers:
  - image: myimage
    name: hello
    env:
    - name: SECRET_TOKEN
      value: $(SOME_SECRET_NAME)
`)

	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(fSys, kustomization)
	fSys.WriteFile("pod.yaml", pod)
	fSys.WriteFile("patch.yaml", patch)
	cmd := NewCmdFix(fSys, os.Stdout)
	assert.NoError(t, cmd.Flags().Set("vars", "true"))
	assert.NoError(t, cmd.RunE(cmd, nil))
	content, err := testutils_test.ReadTestKustomization(fSys)
	assert.NoError(t, err)

	// The replacement target targets all resources of type `Pod` because the
	// patch targets all resources of type `Pod`.
	assert.Equal(t, `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- pod.yaml

patches:
- path: patch.yaml
  target:
    kind: Pod

replacements:
- source:
    kind: Secret
    name: my-secret
    version: v1
  targets:
  - fieldPaths:
    - spec.containers.0.env.0.value
    select:
      kind: Pod
`, string(content))

	content, err = fSys.ReadFile("pod.yaml")
	assert.NoError(t, err)
	assert.Equal(t, `
apiVersion: v1
kind: Pod
metadata:
  name: my-pod-1
spec:
  containers:
  - image: myimage
    name: hello
---
apiVersion: v1
kind: Pod
metadata:
  name: my-pod-2
spec:
  containers:
  - image: myimage
    name: hello
`, string(content))

	content, err = fSys.ReadFile("patch.yaml")
	assert.NoError(t, err)
	assert.Equal(t, `
apiVersion: v1
kind: Pod
metadata:
  name: doesnt-matter
spec:
  containers:
  - image: myimage
    name: hello
    env:
    - name: SECRET_TOKEN
      value: SOME_SECRET_NAME_PLACEHOLDER
`, string(content))
}

func TestFixVarsWithPatchMultipleResourcesAndKindChange(t *testing.T) {
	kustomization := []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- pod.yaml

patches:
- path: patch.yaml
  target:
    kind: Pod
  options:
    allowKindChange: true

vars:
- name: SOME_SECRET_NAME
  objref:
    kind: Secret
    name: my-secret
    apiVersion: v1
`)
	pod := []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: my-pod-1
spec:
  containers:
  - image: myimage
    name: hello
---
apiVersion: v1
kind: Pod
metadata:
  name: my-pod-2
spec:
  containers:
  - image: myimage
    name: hello
`)
	patch := []byte(`
apiVersion: v1
kind: Custom
metadata:
  name: doesnt-matter
spec:
  containers:
  - image: myimage
    name: hello
    env:
    - name: SECRET_TOKEN
      value: $(SOME_SECRET_NAME)
`)

	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(fSys, kustomization)
	fSys.WriteFile("pod.yaml", pod)
	fSys.WriteFile("patch.yaml", patch)
	cmd := NewCmdFix(fSys, os.Stdout)
	assert.NoError(t, cmd.Flags().Set("vars", "true"))
	assert.NoError(t, cmd.RunE(cmd, nil))
	content, err := testutils_test.ReadTestKustomization(fSys)
	assert.NoError(t, err)

	// The replacement target from the patch uses the patch's Kind because
	// allowKindChange is set to true.
	assert.Equal(t, `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- pod.yaml

patches:
- options:
    allowKindChange: true
  path: patch.yaml
  target:
    kind: Pod

replacements:
- source:
    kind: Secret
    name: my-secret
    version: v1
  targets:
  - fieldPaths:
    - spec.containers.0.env.0.value
    select:
      kind: Custom
`, string(content))

	content, err = fSys.ReadFile("pod.yaml")
	assert.NoError(t, err)
	assert.Equal(t, `
apiVersion: v1
kind: Pod
metadata:
  name: my-pod-1
spec:
  containers:
  - image: myimage
    name: hello
---
apiVersion: v1
kind: Pod
metadata:
  name: my-pod-2
spec:
  containers:
  - image: myimage
    name: hello
`, string(content))

	content, err = fSys.ReadFile("patch.yaml")
	assert.NoError(t, err)
	assert.Equal(t, `
apiVersion: v1
kind: Custom
metadata:
  name: doesnt-matter
spec:
  containers:
  - image: myimage
    name: hello
    env:
    - name: SECRET_TOKEN
      value: SOME_SECRET_NAME_PLACEHOLDER
`, string(content))
}

func TestFixVarsWithOverlay(t *testing.T) {
	kustomizationOverlay := []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- base

vars:
- name: SOME_SECRET_NAME
  objref:
    kind: Secret
    name: my-secret
    apiVersion: v1
`)
	kustomizationBase := []byte(`
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- pod.yaml
`)
	pod := []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
    - image: myimage
      name: hello
      env:
        - name: SECRET_TOKEN
          value: $(SOME_SECRET_NAME)
`)

	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(fSys, kustomizationOverlay)
	fSys.WriteFile("base/pod.yaml", pod)
	fSys.WriteFile("base/kustomization.yaml", kustomizationBase)
	cmd := NewCmdFix(fSys, os.Stdout)
	assert.NoError(t, cmd.Flags().Set("vars", "true"))
	assert.NoError(t, cmd.RunE(cmd, nil))
	content, err := testutils_test.ReadTestKustomization(fSys)
	assert.NoError(t, err)
	assert.Equal(t, `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- base

replacements:
- source:
    kind: Secret
    name: my-secret
    version: v1
  targets:
  - fieldPaths:
    - spec.containers.0.env.0.value
    select:
      kind: Pod
      name: my-pod
      version: v1
`, string(content))

	content, err = fSys.ReadFile("base/kustomization.yaml")
	assert.NoError(t, err)
	assert.Equal(t, `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- pod.yaml
`, string(content))

	content, err = fSys.ReadFile("base/pod.yaml")
	assert.NoError(t, err)
	assert.Equal(t, `
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
    - image: myimage
      name: hello
      env:
        - name: SECRET_TOKEN
          value: SOME_SECRET_NAME_PLACEHOLDER
`, string(content))
}
