// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fieldspec

import (
	"testing"

	"sigs.k8s.io/kustomize/functions/config.kubernetes.io/kustomize/testutil"
)

func TestSelectorsFilter(t *testing.T) {
	testutil.RunTestCases(t, selectorsTestCases)
}

var selectorsTestCases = []testutil.FieldSpecTestCase{

	// Test Case
	{
		Name: "crd-spec-selector-matchLabels",
		Input: `
apiVersion: batch/v1
kind: Foo
metadata:
  Name: Instance-1
spec:
  selector:
    matchLabels:
      a: b
---
apiVersion: batch/v1
kind: Foo
metadata:
  Name: Instance-2
spec:
  selector:
    matchLabels:
      e: d
---
apiVersion: batch/v1
kind: Foo
metadata:
  Name: Instance-3
`,
		Expected: `
apiVersion: batch/v1
kind: Foo
metadata:
  Name: Instance-1
spec:
  selector:
    matchLabels:
      a: b
      e: f
---
apiVersion: batch/v1
kind: Foo
metadata:
  Name: Instance-2
spec:
  selector:
    matchLabels:
      e: f
---
apiVersion: batch/v1
kind: Foo
metadata:
  Name: Instance-3
`,
		Instance: KustomizeSelectorsFilter{Selectors: map[string]*string{"e": &f}},
	},

	// Test Case
	{
		Name: "builtin-spec-selector-matchLabels",
		Input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  Name: Instance-1
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  Name: Instance-2
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  Name: Instance-3
---
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  Name: Instance-4
---
apiVersion: apps/v1
kind: Service
metadata:
  Name: Instance-5
---
apiVersion: apps/v1
kind: ReplicationController
metadata:
  Name: Instance-6
`,
		Expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  Name: Instance-1
spec:
  selector:
    matchLabels:
      e: f
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  Name: Instance-2
spec:
  selector:
    matchLabels:
      e: f
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  Name: Instance-3
spec:
  selector:
    matchLabels:
      e: f
---
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  Name: Instance-4
spec:
  selector:
    matchLabels:
      e: f
---
apiVersion: apps/v1
kind: Service
metadata:
  Name: Instance-5
spec:
  selector:
    e: f
---
apiVersion: apps/v1
kind: ReplicationController
metadata:
  Name: Instance-6
spec:
  selector:
    e: f
`,
		Instance: KustomizeSelectorsFilter{Selectors: map[string]*string{"e": &f}},
	},

	// Test Case
	{
		Name: "cronjob",
		Input: `
apiVersion: batch/v1
kind: CronJob
metadata:
  Name: Instance-1
spec:
  jobTemplate:
    spec:
      selector:
        matchLabels:
          a: b
---
apiVersion: batch/v1
kind: CronJob
metadata:
  Name: Instance-2
spec:
  jobTemplate:
    spec:
      selector:
        matchLabels:
          e: b
---
apiVersion: batch/v1
kind: CronJob
metadata:
  Name: Instance-3
`,
		Expected: `
apiVersion: batch/v1
kind: CronJob
metadata:
  Name: Instance-1
spec:
  jobTemplate:
    spec:
      selector:
        matchLabels:
          a: b
          e: f
---
apiVersion: batch/v1
kind: CronJob
metadata:
  Name: Instance-2
spec:
  jobTemplate:
    spec:
      selector:
        matchLabels:
          e: f
---
apiVersion: batch/v1
kind: CronJob
metadata:
  Name: Instance-3
`,
		Instance: KustomizeSelectorsFilter{Selectors: map[string]*string{"e": &f}},
	},

	// Test Case
	{
		Name: "network-policy",
		Input: `
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  Name: Instance-1
spec:
  podSelector:
    matchLabels: {}
  ingress:
    from:
      podSelector:
        matchLabels: {}
  egress:
    to:
      podSelector:
        matchLabels: {}
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  Name: Instance-2
`,
		Expected: `
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  Name: Instance-1
spec:
  podSelector:
    matchLabels:
      e: f
  ingress:
    from:
      podSelector:
        matchLabels:
          e: f
  egress:
    to:
      podSelector:
        matchLabels:
          e: f
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  Name: Instance-2
`,
		Instance: KustomizeSelectorsFilter{Selectors: map[string]*string{"e": &f}},
	},

	// Test Case
	{
		Name: "owner",
		Input: `
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  Name: Instance-1
spec:
  podSelector:
    matchLabels: {}
  ingress:
    from:
      podSelector:
        matchLabels: {}
  egress:
    to:
      podSelector:
        matchLabels: {}
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  Name: Instance-2
`,
		Expected: `
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  Name: Instance-1
spec:
  podSelector:
    matchLabels:
      e: f # {"ownedBy":"owner"}
  ingress:
    from:
      podSelector:
        matchLabels:
          e: f # {"ownedBy":"owner"}
  egress:
    to:
      podSelector:
        matchLabels:
          e: f # {"ownedBy":"owner"}
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  Name: Instance-2
`,
		Instance: KustomizeSelectorsFilter{Selectors: map[string]*string{"e": &f}, kustomizeName: "owner"},
	},
}
