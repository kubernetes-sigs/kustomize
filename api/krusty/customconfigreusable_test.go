// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

// Demo custom configuration as a base.
// This test shows usage of three custom configurations sitting
// in a base, allowing them to be reused in any number of
// kustomizations.
func TestReusableCustomTransformers(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("PrefixSuffixTransformer").
		PrepBuiltin("AnnotationsTransformer").
		PrepBuiltin("LabelTransformer")
	defer th.Reset()

	// First write three custom configurations for builtin plugins.

	// A custom name prefixer that only touches Deployments and Services.
	th.WriteF("/app/mytransformers/deploymentServicePrefixer.yaml", `
apiVersion: builtin
kind: PrefixSuffixTransformer
metadata:
  name: myPrefixer
prefix: bob-
fieldSpecs:
- kind: Deployment
  path: metadata/name
- kind: Service
  path: metadata/name
`)

	// A custom annotator exclusively annotating Roles.
	th.WriteF("/app/mytransformers/roleAnnotator.yaml", `
apiVersion: builtin
kind: AnnotationsTransformer
metadata:
  name: myAnnotator
annotations:
  # Imagine that SRE has not approved this role yet.
  status: probationary
fieldSpecs:
- path: metadata/annotations
  create: true
  kind: Role
`)

	// A custom labeller that only labels Deployments,
	// and only labels them at their top metadata level
	// exclusively.  It does not modify selectors or
	// add labels to pods in the template.
	th.WriteF("/app/mytransformers/deploymentLabeller.yaml", `
apiVersion: builtin
kind: LabelTransformer
metadata:
  name: myLabeller
labels:
  pager: 867-5301
fieldSpecs:
- path: metadata/labels
  create: true
  kind: Deployment
`)

	// Combine these custom transformers in one kustomization file.
	// This kustomization file contains only resources that
	// all happen to be plugin configurations. This makes
	// these plugins re-usable as a group in any number of other
	// kustomizations.
	th.WriteK("/app/mytransformers", `
resources:
- deploymentServicePrefixer.yaml
- roleAnnotator.yaml
- deploymentLabeller.yaml
`)

	// Finally, define the kustomization for the (arbitrarily named)
	// staging environment.
	th.WriteK("/app/staging", `

# Bring in the custom transformers.
transformers:
- ../mytransformers

# Also use the "classic" labeller, which behind the scenes uses
# the LabelTransformer, but with a broad configuration targeting
# many resources and fields (including selector fields).
# It's a big hammer - probably too big; the output shows all the
# places 'acmeCorp' now appears as a result.  To avoid this,
# define your own config for your own LabelTransformer instance
# as shown above.
commonLabels:
  company: acmeCorp  

# Specify the resources to modify.
resources:
- deployment.yaml
- role.yaml
- service.yaml
`)
	th.WriteF("/app/staging/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDeployment
spec:
  template:
    metadata:
      labels:
        backend: flawless
    spec:
      containers:
      - name: whatever
        image: whatever
`)
	th.WriteF("/app/staging/role.yaml", `
apiVersion: v1
kind: Role
metadata:
  name: myRole
`)
	th.WriteF("/app/staging/service.yaml", `
apiVersion: v1
kind: Service
metadata:
  name: myService
`)

	m := th.Run("/app/staging", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    company: acmeCorp
    pager: 867-5301
  name: bob-myDeployment
spec:
  selector:
    matchLabels:
      company: acmeCorp
  template:
    metadata:
      labels:
        backend: flawless
        company: acmeCorp
    spec:
      containers:
      - image: whatever
        name: whatever
---
apiVersion: v1
kind: Role
metadata:
  annotations:
    status: probationary
  labels:
    company: acmeCorp
  name: myRole
---
apiVersion: v1
kind: Service
metadata:
  labels:
    company: acmeCorp
  name: bob-myService
spec:
  selector:
    company: acmeCorp
`)
}
