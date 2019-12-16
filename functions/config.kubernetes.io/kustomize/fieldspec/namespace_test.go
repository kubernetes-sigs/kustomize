// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fieldspec_test

import (
	"testing"

	. "sigs.k8s.io/kustomize/functions/config.kubernetes.io/kustomize/fieldspec"
	"sigs.k8s.io/kustomize/functions/config.kubernetes.io/kustomize/testutil"
)

func TestNamespaceFilter(t *testing.T) {
	testutil.RunTestCases(t, namespaceTestCases)
}

var namespaceTestCases = []testutil.FieldSpecTestCase{
	// Test Case
	{
		Name: "add-namespace",
		Input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  Name: Instance
---
apiVersion: example.com/v1
kind: Bar
metadata:
  Name: Instance
`,
		Expected: `
apiVersion: example.com/v1
kind: Foo
metadata:
  Name: Instance
  namespace: foo
---
apiVersion: example.com/v1
kind: Bar
metadata:
  Name: Instance
  namespace: foo
`,
		Instance: KustomizeNamespaceFilter{KustomizeNamespace: "foo"},
	},

	// Test Case
	{
		Name: "update-namespace",
		Input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  Name: Instance
  # update this namespace
  namespace: bar
---
apiVersion: example.com/v1
kind: Bar
metadata:
  Name: Instance
  namespace: bar
`,
		Expected: `
apiVersion: example.com/v1
kind: Foo
metadata:
  Name: Instance
  # update this namespace
  namespace: foo
---
apiVersion: example.com/v1
kind: Bar
metadata:
  Name: Instance
  namespace: foo
`,
		Instance: KustomizeNamespaceFilter{KustomizeNamespace: "foo"},
	},

	// Test Case
	{
		Name: "owner",
		Input: `
apiVersion: example.com/v1
kind: Foo
metadata:
  Name: Instance
  # update this namespace
  namespace: bar
---
apiVersion: example.com/v1
kind: Bar
metadata:
  Name: Instance
  namespace: bar
`,
		Expected: `
apiVersion: example.com/v1
kind: Foo
metadata:
  Name: Instance
  # update this namespace
  namespace: foo # {"ownedBy":"owner"}
---
apiVersion: example.com/v1
kind: Bar
metadata:
  Name: Instance
  namespace: foo # {"ownedBy":"owner"}

`,
		Instance: (&KustomizeNamespaceFilter{KustomizeNamespace: "foo"}).SetKustomizeName("owner"),
	},
}
