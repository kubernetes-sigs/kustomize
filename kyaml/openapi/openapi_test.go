// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package openapi_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestSchemaForResourceType(t *testing.T) {
	s := openapi.SchemaForResourceType(
		yaml.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"})
	if !assert.NotNil(t, s) {
		t.FailNow()
	}

	f := s.Field("spec")
	if !assert.NotNil(t, f) {
		t.FailNow()
	}
	if !assert.Equal(t, "DeploymentSpec is the specification of the desired behavior of the Deployment.",
		f.Schema.Description) {
		t.FailNow()
	}

	replicas := f.Field("replicas")
	if !assert.NotNil(t, replicas) {
		t.FailNow()
	}
	if !assert.Equal(t, "Number of desired pods. This is a pointer to distinguish between explicit zero and not specified. Defaults to 1.",
		replicas.Schema.Description) {
		t.FailNow()
	}

	temp := f.Field("template")
	if !assert.NotNil(t, temp) {
		t.FailNow()
	}
	if !assert.Equal(t, "PodTemplateSpec describes the data a pod should have when created from a template",
		temp.Schema.Description) {
		t.FailNow()
	}

	containers := temp.Field("spec").Field("containers").Elements()
	if !assert.NotNil(t, containers) {
		t.FailNow()
	}

	targetPort := containers.Field("ports").Elements().Field("containerPort")
	if !assert.NotNil(t, targetPort) {
		t.FailNow()
	}
	if !assert.Equal(t, "Number of port to expose on the pod's IP address. This must be a valid port number, 0 < x < 65536.",
		targetPort.Schema.Description) {
		t.FailNow()
	}

	arg := containers.Field("args").Elements()
	if !assert.NotNil(t, arg) {
		t.FailNow()
	}
	if !assert.Equal(t, "string", arg.Schema.Type[0]) {
		t.FailNow()
	}
}
