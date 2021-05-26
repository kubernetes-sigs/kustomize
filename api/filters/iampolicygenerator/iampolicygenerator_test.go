// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package iampolicygenerator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	filtertest "sigs.k8s.io/kustomize/api/testutils/filtertest"
	"sigs.k8s.io/kustomize/api/types"
)

func TestFilter(t *testing.T) {
	testCases := map[string]struct {
		args     types.IAMPolicyGeneratorArgs
		expected string
	}{
		"with namespace": {
			args: types.IAMPolicyGeneratorArgs{
				Cloud: types.GKE,
				KubernetesService: types.KubernetesService{
					Namespace: "k8s-namespace",
					Name:      "k8s-sa-name",
				},
				ServiceAccount: types.ServiceAccount{
					Name:      "gsa-name",
					ProjectId: "project-id",
				},
			},
			expected: `
apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    iam.gke.io/gcp-service-account: gsa-name@project-id.iam.gserviceaccount.com
  name: k8s-sa-name
  namespace: k8s-namespace
`,
		},
		"without namespace": {
			args: types.IAMPolicyGeneratorArgs{
				Cloud: types.GKE,
				KubernetesService: types.KubernetesService{
					Name: "k8s-sa-name",
				},
				ServiceAccount: types.ServiceAccount{
					Name:      "gsa-name",
					ProjectId: "project-id",
				},
			},
			expected: `
apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    iam.gke.io/gcp-service-account: gsa-name@project-id.iam.gserviceaccount.com
  name: k8s-sa-name
`,
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			f := Filter{
				IAMPolicyGenerator: tc.args,
			}
			actual := filtertest.RunFilter(t, "", f)
			if !assert.Equal(t, strings.TrimSpace(tc.expected), strings.TrimSpace(actual)) {
				t.FailNow()
			}
		})
	}
}
