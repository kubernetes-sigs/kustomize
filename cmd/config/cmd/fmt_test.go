// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/cmd/config/cmd"
	"sigs.k8s.io/kustomize/kyaml/kio/filters/testyaml"
)

// TestCmd_files verifies the fmt command formats the files
func TestFmtCommand_files(t *testing.T) {
	f1, err := ioutil.TempFile("", "cmdfmt*.yaml")
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(f1.Name())
	err = ioutil.WriteFile(f1.Name(), testyaml.UnformattedYaml1, 0600)
	if !assert.NoError(t, err) {
		return
	}

	f2, err := ioutil.TempFile("", "cmdfmt*.yaml")
	if !assert.NoError(t, err) {
		return
	}
	defer os.RemoveAll(f2.Name())
	err = ioutil.WriteFile(f2.Name(), testyaml.UnformattedYaml2, 0600)
	if !assert.NoError(t, err) {
		return
	}

	// fmt the files
	r := cmd.GetFmtRunner("")
	r.Command.SetArgs([]string{f1.Name(), f2.Name()})
	err = r.Command.Execute()
	if !assert.NoError(t, err) {
		return
	}

	// verify the contents
	b, err := ioutil.ReadFile(f1.Name())
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Equal(t, string(testyaml.FormattedYaml1), string(b)) {
		return
	}

	b, err = ioutil.ReadFile(f2.Name())
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Equal(t, string(testyaml.FormattedYaml2), string(b)) {
		return
	}
}

func TestFmtCommand_stdin(t *testing.T) {
	out := &bytes.Buffer{}
	r := cmd.GetFmtRunner("")
	r.Command.SetOut(out)
	r.Command.SetIn(bytes.NewReader(testyaml.UnformattedYaml1))

	// fmt the input
	err := r.Command.Execute()
	assert.NoError(t, err)

	// verify the output
	assert.Equal(t, string(testyaml.FormattedYaml1), out.String())
}

// TestCmd_filesAndstdin verifies that if both files and stdin input are provided, only
// the files are formatted and the input is ignored
func TestFmtCmd_filesAndStdin(t *testing.T) {
	f1, err := ioutil.TempFile("", "cmdfmt*.yaml")
	if !assert.NoError(t, err) {
		return
	}
	err = ioutil.WriteFile(f1.Name(), testyaml.UnformattedYaml1, 0600)
	if !assert.NoError(t, err) {
		return
	}

	f2, err := ioutil.TempFile("", "cmdfmt*.yaml")
	if !assert.NoError(t, err) {
		return
	}
	err = ioutil.WriteFile(f2.Name(), testyaml.UnformattedYaml2, 0600)
	if !assert.NoError(t, err) {
		return
	}

	out := &bytes.Buffer{}
	in := &bytes.Buffer{}
	r := cmd.GetFmtRunner("")
	r.Command.SetOut(out)
	r.Command.SetIn(in)

	// fmt the files
	r = cmd.GetFmtRunner("")
	r.Command.SetArgs([]string{f1.Name(), f2.Name()})
	err = r.Command.Execute()
	if !assert.NoError(t, err) {
		return
	}

	// verify the output
	b, err := ioutil.ReadFile(f1.Name())
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Equal(t, string(testyaml.FormattedYaml1), string(b)) {
		return
	}

	b, err = ioutil.ReadFile(f2.Name())
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Equal(t, string(testyaml.FormattedYaml2), string(b)) {
		return
	}
	err = r.Command.Execute()
	if !assert.NoError(t, err) {
		return
	}

	if !assert.Equal(t, "", out.String()) {
		return
	}
}

// TestCmd_files verifies the fmt command formats the files
func TestCmd_failFiles(t *testing.T) {
	// fmt the files
	r := cmd.GetFmtRunner("")
	r.Command.SetArgs([]string{"notrealfile"})
	err := r.Command.Execute()
	assert.EqualError(t, err, "lstat notrealfile: no such file or directory")
}

// TestCmd_files verifies the fmt command formats the files
func TestCmd_failFileContents(t *testing.T) {
	out := &bytes.Buffer{}
	r := cmd.GetFmtRunner("")
	r.Command.SetOut(out)
	r.Command.SetIn(strings.NewReader(`{`))

	// fmt the input
	err := r.Command.Execute()

	// expect an error
	assert.EqualError(t, err, "yaml: line 1: did not find expected node content")
}
