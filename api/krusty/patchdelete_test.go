package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestPatchDeleteOfNotExistingAttributesShouldNotAddExtraElements(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t)
	defer th.Reset()

	th.WriteF("/app/resource.yaml", `
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
	th.WriteF("/app/patch.yml", `
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
	th.WriteK("/app", `
resources:
- resource.yaml
patches:
- path: patch.yml
  target:
    kind: Deployment
`)

	/*
		// It's expected that removal of not existing elements should not introduce extra values,
		// as a patch can be applied to multipe resources, not all of them can have all deleted elements.
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
	*/

	// Currently kustomize inserts $patch: delete elements into the resulting resources
	erroneousActual := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: whatever
spec:
  template:
    spec:
      containers:
      - env:
        - $patch: delete
          name: NOT_EXISTING_FOR_REMOVAL
        - name: EXISTING
          value: EXISTING_VALUE
        image: helloworld
        name: whatever
`

	m := th.Run("/app", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, erroneousActual)
}
