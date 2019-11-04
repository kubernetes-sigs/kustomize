// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathMatcher_Filter(t *testing.T) {
	node := MustParse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        ports:
        - containerPort: 80
      - name: sidecar
        image: sidecar:1.0.0
        ports:
        - containerPort: 8081
        - containerPort: 9090
`)

	updates := []struct {
		path  []string
		value string
	}{
		{[]string{
			"spec", "template", "spec", "containers", "[name=.*]"},
			"- name: nginx\n  image: nginx:1.7.9\n  ports:\n  - containerPort: 80\n" +
				"- name: sidecar\n  image: sidecar:1.0.0\n  ports:\n  - containerPort: 8081\n  - containerPort: 9090\n"},
		{[]string{
			"spec", "template", "spec", "containers", "[name=.*]", "image"},
			"- nginx:1.7.9\n- sidecar:1.0.0\n"},
		{[]string{
			"spec", "template", "spec", "containers", "[name=n.*]", "image"},
			"- nginx:1.7.9\n"},
		{[]string{
			"spec", "template", "spec", "containers", "[name=s.*]", "image"},
			"- sidecar:1.0.0\n"},
		{[]string{
			"spec", "template", "spec", "containers", "[name=.*x]", "image"},
			"- nginx:1.7.9\n"},
		{[]string{
			"spec", "template", "spec", "containers", "[name=.*]", "ports"},
			"- - containerPort: 80\n- - containerPort: 8081\n  - containerPort: 9090\n"},
		{[]string{
			"spec", "template", "spec", "containers", "[name=.*]", "ports", "[containerPort=8.*]"},
			"- containerPort: 80\n- containerPort: 8081\n"},
		{[]string{
			"spec", "template", "spec", "containers", "[name=.*]", "ports", "[containerPort=.*1]"},
			"- containerPort: 8081\n"},
		{[]string{
			"spec", "template", "spec", "containers", "[name=.*]", "ports", "[containerPort=9.*]"},
			"- containerPort: 9090\n"},
		{[]string{
			"spec", "template", "spec", "containers", "[name=s.*]", "ports", "[containerPort=8.*]"},
			"- containerPort: 8081\n"},
		{[]string{
			"spec", "template", "spec", "containers", "[name=s.*]", "ports", "[containerPort=.*2]"},
			""},
	}
	for i, u := range updates {
		result, err := node.Pipe(&PathMatcher{Path: u.path})
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, u.value, result.MustString(), fmt.Sprintf("%d", i))
	}
}
