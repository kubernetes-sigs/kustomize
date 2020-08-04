// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge3_test

var kustomizationTestCases = []testCase{
	// Kustomization Test Cases

	{
		description: `ConfigMapGenerator merge`,
		origin: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: a-configmap1
  files:
  - configs/configfile1
  - configkey=configs/another_configfile1`,
		update: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- files:
  - configs/configfile2
  - configkey=configs/another_configfile2
  name: a-configmap2`,
		local: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: a-configmap1
  files:
  - configs/configfile1
  - configkey=configs/another_configfile1
- name: a-configmap3
  files:
  - configs/configfile3
  - configkey=configs/another_configfile3`,
		expected: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: a-configmap3
  files:
  - configs/configfile3
  - configkey=configs/another_configfile3
- files:
  - configs/configfile2
  - configkey=configs/another_configfile2
  name: a-configmap2`},

	{
		description: `SecretGenerator merge`,
		origin: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- name: a-secret1
  files:
  - configs/configfile1
  - configkey=configs/another_configfile1`,
		update: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- files:
  - configs/configfile2
  - configkey=configs/another_configfile2
  name: a-secret2`,
		local: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- name: a-secret1
  files:
  - configs/configfile1
  - configkey=configs/another_configfile1
- name: a-secret3
  files:
  - configs/configfile3
  - configkey=configs/another_configfile3`,
		expected: `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
secretGenerator:
- name: a-secret3
  files:
  - configs/configfile3
  - configkey=configs/another_configfile3
- files:
  - configs/configfile2
  - configkey=configs/another_configfile2
  name: a-secret2`},
}
