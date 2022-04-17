// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/cmd/config/internal/commands"
)

func TestInit_args(t *testing.T) {
	d := t.TempDir()

	// fmt the files
	b := &bytes.Buffer{}
	r := commands.GetInitRunner("")
	r.Command.SetArgs([]string{d})
	r.Command.SetOut(b)
	if !assert.NoError(t, r.Command.Execute()) {
		t.FailNow()
	}

	actual, err := ioutil.ReadFile(filepath.Join(d, "Krmfile"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.Equal(t, strings.TrimSpace(`
apiVersion: config.k8s.io/v1alpha1
kind: Krmfile
`), strings.TrimSpace(string(actual))) {
		t.FailNow()
	}

	if !assert.Equal(t, `Command "init" is deprecated, setter commands and substitutions will no longer be available in kustomize v5.
See discussion in https://github.com/kubernetes-sigs/kustomize/issues/3953.
`, b.String()) {
		t.FailNow()
	}
}

func TestInit_noargs(t *testing.T) {
	d := t.TempDir()

	cwd, err := os.Getwd()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.NoError(t, os.Chdir(d)) {
		t.FailNow()
	}

	t.Cleanup(func() {
		if !assert.NoError(t, os.Chdir(cwd)) {
			t.FailNow()
		}
	})

	b := &bytes.Buffer{}
	r := commands.GetInitRunner("")
	r.Command.SetOut(b)
	if !assert.NoError(t, r.Command.Execute()) {
		t.FailNow()
	}

	actual, err := ioutil.ReadFile(filepath.Join(d, "Krmfile"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.Equal(t, strings.TrimSpace(`
apiVersion: config.k8s.io/v1alpha1
kind: Krmfile
`), strings.TrimSpace(string(actual))) {
		t.FailNow()
	}

	if !assert.Equal(t, `Command "init" is deprecated, setter commands and substitutions will no longer be available in kustomize v5.
See discussion in https://github.com/kubernetes-sigs/kustomize/issues/3953.
`, b.String()) {
		t.FailNow()
	}
}
