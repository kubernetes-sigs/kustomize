// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/cmd/config/internal/commands"
	"sigs.k8s.io/kustomize/kyaml/copyutil"
)

// TestMerge3Command verifies the merge3 correctly applies the diff between 2 sets of resources into another
func TestMerge3Command(t *testing.T) {
	datadir, err := ioutil.TempDir("", "test-data")
	defer os.RemoveAll(datadir)
	if !assert.NoError(t, err) {
		return
	}

	err = ioutil.WriteFile(filepath.Join(datadir, "java-deployment.resource.yaml"), []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  labels:
    app: java
spec:
  replicas: 1
  selector:
    matchLabels:
      app: java
  template:
    metadata:
      labels:
        app: java
    spec:
      restartPolicy: Always
      containers:
      - name: app
        image: gcr.io/project/app:version
        command:
        - java
        - -jar
        - /app.jar
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: app-config
        env:
        - name: JAVA_OPTS
          value: -XX:+UnlockExperimentalVMOptions -XX:+UseCGroupMemoryLimitForHeap
            -Djava.security.egd=file:/dev/./urandom
        imagePullPolicy: Always
  minReadySeconds: 5
`), 0600)
	if !assert.NoError(t, err) {
		return
	}

	expected_dir, err := ioutil.TempDir("", "test-data-expected")
	defer os.RemoveAll(expected_dir)
	if !assert.NoError(t, err) {
		return
	}

	err = ioutil.WriteFile(filepath.Join(expected_dir, "java-deployment.resource.yaml"), []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  labels:
    app: java
    new-local: label
    new-remote: label
spec:
  replicas: 3
  selector:
    matchLabels:
      app: java
  template:
    metadata:
      labels:
        app: java
    spec:
      restartPolicy: Always
      containers:
      - name: app
        image: gcr.io/project/app:version
        command:
        - java
        - -jar
        - /app.jar
        - otherstuff
        args:
        - foo
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: app-config
        env:
        - name: JAVA_OPTS
          value: -XX:+UnlockExperimentalVMOptions -XX:+UseCGroupMemoryLimitForHeap
            -Djava.security.egd=file:/dev/./urandom
        imagePullPolicy: Always
  minReadySeconds: 20
`), 0600)
	if !assert.NoError(t, err) {
		return
	}

	updated_dir, err := ioutil.TempDir("", "test-data-updated")
	defer os.RemoveAll(updated_dir)
	if !assert.NoError(t, err) {
		return
	}

	err = ioutil.WriteFile(filepath.Join(updated_dir, "java-deployment.resource.yaml"), []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  labels:
    app: java
    new-remote: label
spec:
  replicas: 3
  selector:
    matchLabels:
      app: java
  template:
    metadata:
      labels:
        app: java
    spec:
      restartPolicy: Always
      containers:
      - name: app
        image: gcr.io/project/app:version
        command:
        - java
        - -jar
        - /app.jar
        - otherstuff
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: app-config
        env:
        - name: JAVA_OPTS
          value: -XX:+UnlockExperimentalVMOptions -XX:+UseCGroupMemoryLimitForHeap
            -Djava.security.egd=file:/dev/./urandom
        imagePullPolicy: Always
  minReadySeconds: 5
`), 0600)
	if !assert.NoError(t, err) {
		return
	}

	dest_dir, err := ioutil.TempDir("", "test-data-dest")
	defer os.RemoveAll(dest_dir)
	if !assert.NoError(t, err) {
		return
	}

	err = ioutil.WriteFile(filepath.Join(dest_dir, "java-deployment.resource.yaml"), []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  labels:
    app: java
    new-local: label
spec:
  replicas: 2
  selector:
    matchLabels:
      app: java
  template:
    metadata:
      labels:
        app: java
    spec:
      restartPolicy: Always
      containers:
      - name: app
        image: gcr.io/project/app:version
        command:
        - java
        - -jar
        - /app.jar
        args:
        - foo
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: app-config
        env:
        - name: JAVA_OPTS
          value: -XX:+UnlockExperimentalVMOptions -XX:+UseCGroupMemoryLimitForHeap
            -Djava.security.egd=file:/dev/./urandom
        imagePullPolicy: Always
  minReadySeconds: 20
`), 0600)
	if !assert.NoError(t, err) {
		return
	}

	// Perform merge3 with newly created sets
	r := commands.GetMerge3Runner("")
	r.Command.SetArgs([]string{
		"--ancestor",
		datadir,
		"--from",
		updated_dir,
		"--to",
		dest_dir,
	})
	if !assert.NoError(t, r.Command.Execute()) {
		return
	}

	diffs, err := copyutil.Diff(dest_dir, expected_dir)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// Verify there are no diffs
	if !assert.Empty(t, diffs.List()) {
		t.FailNow()
	}
}
