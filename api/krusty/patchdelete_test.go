package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestPatchDeleteOfNotExistingAttributesShouldNotAddExtraElements(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteF("resource.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: whatever
spec:
  template:
    spec:
      containers:
      - env:
        - name: EXISTING
          value: EXISTING_VALUE
        - name: FOR_REMOVAL
          value: FOR_REMOVAL_VALUE
        name: whatever
        image: helloworld
`)
	th.WriteF("patch.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: whatever
spec:
  template:
    spec:
      containers:
      - name: whatever
        env:
        - name: FOR_REMOVAL
          $patch: delete
        - name: NOT_EXISTING_FOR_REMOVAL
          $patch: delete
`)
	th.WriteK(".", `
resources:
- resource.yaml
patches:
- path: patch.yaml
  target:
    kind: Deployment
`)

	// It's expected that removal of not existing elements should not introduce extra values,
	// as a patch can be applied to multiple resources, not all of them can have all the elements being deleted.
	expected := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: whatever
spec:
  template:
    spec:
      containers:
      - env:
        - name: EXISTING
          value: EXISTING_VALUE
        image: helloworld
        name: whatever
`
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, expected)
}
