// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package frameworktestutil contains utilities for testing functions written using the framework.
package frameworktestutil

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

const (
	DefaultTestDataDirectory   = "testdata"
	DefaultConfigInputFilename = "config.yaml"
	DefaultInputFilename       = "input.yaml"
	DefaultInputFilenameGlob   = "input*.yaml"
	DefaultOutputFilename      = "expected.yaml"
	DefaultErrorFilename       = "errors.txt"
)

// CommandResultsChecker tests a command-wrapped function by running it with predefined inputs
// and comparing the outputs to expected results.
type CommandResultsChecker struct {
	// TestDataDirectory is the directory containing the testdata subdirectories.
	// CommandResultsChecker will recurse into each test directory and run the Command
	// if the directory contains at least one of ExpectedOutputFilename or ExpectedErrorFilename.
	// Defaults to "testdata"
	TestDataDirectory string

	// ExpectedOutputFilename is the file with the expected output of the function
	// Defaults to "expected.yaml".  Directories containing neither this file
	// nor ExpectedErrorFilename will be skipped.
	ExpectedOutputFilename string

	// ExpectedErrorFilename is the file containing elements of an expected error message.
	// Each line of the file will be treated as a regex that must match the actual error.
	// Defaults to "errors.txt".  Directories containing neither this file
	// nor ExpectedOutputFilename will be skipped.
	ExpectedErrorFilename string

	// UpdateExpectedFromActual if set to true will write the actual results to the
	// expected testdata files.  This is useful for updating test data.
	UpdateExpectedFromActual bool

	// OutputAssertionFunc allows you to swap out the logic used to compare the expected output
	// from the fixture file to the actual output.
	// By default, it performs a string comparison after normalizing whitespace.
	OutputAssertionFunc AssertionFunc

	// ErrorAssertionFunc allows you to swap out the logic used to compare the expected error
	// message from the fixture file to the actual error message.
	// By default, it interprets each line of the fixture as a regex that the actual error must match.
	ErrorAssertionFunc AssertionFunc

	// ConfigInputFilename is the name of the config file provided as the first
	// argument to the function.  Directories without this file will be skipped.
	// Defaults to "config.yaml"
	ConfigInputFilename string

	// InputFilenameGlob matches function inputs
	// Defaults to "input*.yaml"
	InputFilenameGlob string

	// Command provides the function to run.
	Command func() *cobra.Command
}

// Assert runs the command with the input provided in each valid test directory
// and verifies that the actual output and error match the fixtures in the directory.
func (rc *CommandResultsChecker) Assert(t *testing.T) bool {
	if rc.ConfigInputFilename == "" {
		rc.ConfigInputFilename = DefaultConfigInputFilename
	}
	if rc.InputFilenameGlob == "" {
		rc.InputFilenameGlob = DefaultInputFilenameGlob
	}

	checker := newResultsChecker(
		rc.TestDataDirectory, rc.ExpectedOutputFilename, rc.ExpectedErrorFilename,
		rc.OutputAssertionFunc, rc.ErrorAssertionFunc,
		rc.UpdateExpectedFromActual,
	)
	checker.assert(t, func() (string, string) {
		_, err := os.Stat(rc.ConfigInputFilename)
		if os.IsNotExist(err) {
			t.Errorf("Test case is missing FunctionConfig input file (default: %s)", DefaultConfigInputFilename)
		}
		require.NoError(t, err)
		args := []string{rc.ConfigInputFilename}

		if rc.InputFilenameGlob != "" {
			inputs, err := filepath.Glob(rc.InputFilenameGlob)
			require.NoError(t, err)
			args = append(args, inputs...)
		}

		var stdOut, stdErr bytes.Buffer
		cmd := rc.Command()
		cmd.SetArgs(args)
		cmd.SetOut(&stdOut)
		cmd.SetErr(&stdErr)

		err = cmd.Execute()
		return stdOut.String(), stdErr.String()
	})

	return true
}

// ProcessorResultsChecker tests a processor function by running it with predefined inputs
// and comparing the outputs to expected results.
type ProcessorResultsChecker struct {
	// TestDataDirectory is the directory containing the testdata subdirectories.
	// ProcessorResultsChecker will recurse into each test directory and run the Command
	// if the directory contains at least one of ExpectedOutputFilename or ExpectedErrorFilename.
	// Defaults to "testdata"
	TestDataDirectory string

	// ExpectedOutputFilename is the file with the expected output of the function
	// Defaults to "expected.yaml".  Directories containing neither this file
	// nor ExpectedErrorFilename will be skipped.
	ExpectedOutputFilename string

	// ExpectedErrorFilename is the file containing elements of an expected error message.
	// Each line of the file will be treated as a regex that must match the actual error.
	// Defaults to "errors.txt".  Directories containing neither this file
	// nor ExpectedOutputFilename will be skipped.
	ExpectedErrorFilename string

	// UpdateExpectedFromActual if set to true will write the actual results to the
	// expected testdata files.  This is useful for updating test data.
	UpdateExpectedFromActual bool

	// InputFilename is the name of the file containing the ResourceList input.
	// Directories without this file will be skipped.
	// Defaults to "input.yaml"
	InputFilename string

	// OutputAssertionFunc allows you to swap out the logic used to compare the expected output
	// from the fixture file to the actual output.
	// By default, it performs a string comparison after normalizing whitespace.
	OutputAssertionFunc AssertionFunc

	// ErrorAssertionFunc allows you to swap out the logic used to compare the expected error
	// message from the fixture file to the actual error message.
	// By default, it interprets each line of the fixture as a regex that the actual error must match.
	ErrorAssertionFunc AssertionFunc

	// Processor returns a ResourceListProcessor to run.
	Processor func() framework.ResourceListProcessor
}

// Assert runs the processor with the input provided in each valid test directory
// and verifies that the actual output and error match the fixtures in the directory.
func (rc *ProcessorResultsChecker) Assert(t *testing.T) bool {
	if rc.InputFilename == "" {
		rc.InputFilename = DefaultInputFilename
	}

	checker := newResultsChecker(
		rc.TestDataDirectory, rc.ExpectedOutputFilename, rc.ExpectedErrorFilename,
		rc.OutputAssertionFunc, rc.ErrorAssertionFunc,
		rc.UpdateExpectedFromActual,
	)
	checker.assert(t, func() (string, string) {
		_, err := os.Stat(rc.InputFilename)
		if os.IsNotExist(err) {
			t.Error("Test case is missing input file")
		}
		require.NoError(t, err)

		actualOutput := bytes.NewBuffer([]byte{})
		rlBytes, err := ioutil.ReadFile(rc.InputFilename)
		require.NoError(t, err)

		rw := kio.ByteReadWriter{
			Reader: bytes.NewBuffer(rlBytes),
			Writer: actualOutput,
		}
		err = framework.Execute(rc.Processor(), &rw)
		if err != nil {
			require.NotEmptyf(t, err.Error(), "processor returned error with empty message")
			return actualOutput.String(), err.Error()
		}
		return actualOutput.String(), ""
	})
	return true
}

type AssertionFunc func(t *testing.T, expected string, actual string)

// RequireEachLineMatches is an AssertionFunc that treats each line of expected string
// as a regex that must match the actual string.
func RequireEachLineMatches(t *testing.T, expected, actual string) {
	// Check that each expected line matches the output
	actual = standardizeSpacing(actual)
	for _, msg := range strings.Split(standardizeSpacing(expected), "\n") {
		require.Regexp(t, regexp.MustCompile(msg), actual)
	}
}

// RequireStrippedStringsEqual is an AssertionFunc that does a simple string comparison
// of expected and actual after normalizing whitespace.
func RequireStrippedStringsEqual(t *testing.T, expected, actual string) {
	require.Equal(t,
		standardizeSpacing(expected),
		standardizeSpacing(actual))
}

func standardizeSpacing(s string) string {
	// remove extra whitespace and convert Windows line endings
	return strings.ReplaceAll(strings.TrimSpace(s), "\r\n", "\n")
}

// resultsChecker implements the core logic shared by all results checking types.
type resultsChecker struct {
	testDataDirectory        string
	expectedOutputFilename   string
	expectedErrorFilename    string
	updateExpectedFromActual bool
	outputAssertionFunc      AssertionFunc
	errorAssertionFunc       AssertionFunc

	testsCasesRun int
}

func newResultsChecker(testDataDir string, outputFilename string, errorFilename string,
	outputAsserter AssertionFunc, errorAsserter AssertionFunc,
	updateFixtures bool) *resultsChecker {
	rc := resultsChecker{
		testDataDirectory:        testDataDir,
		expectedOutputFilename:   outputFilename,
		expectedErrorFilename:    errorFilename,
		updateExpectedFromActual: updateFixtures,
		outputAssertionFunc:      outputAsserter,
		errorAssertionFunc:       errorAsserter,
	}
	if rc.testDataDirectory == "" {
		rc.testDataDirectory = DefaultTestDataDirectory
	}
	if rc.expectedOutputFilename == "" {
		rc.expectedOutputFilename = DefaultOutputFilename
	}
	if rc.expectedErrorFilename == "" {
		rc.expectedErrorFilename = DefaultErrorFilename
	}
	if rc.outputAssertionFunc == nil {
		rc.outputAssertionFunc = RequireStrippedStringsEqual
	}
	if rc.errorAssertionFunc == nil {
		rc.errorAssertionFunc = RequireEachLineMatches
	}
	return &rc
}

// assert traverses TestDataDirectory to find test cases, calls getResult to invoke the function
// under test in each directory, and then runs assertions on the returned output and error results.
// It triggers a test failure if no valid test directories were found.
func (rc *resultsChecker) assert(t *testing.T, getResult func() (string, string)) {
	err := filepath.Walk(rc.testDataDirectory,
		func(path string, info os.FileInfo, err error) error {
			require.NoError(t, err)
			if !info.IsDir() {
				// skip non-directories
				return nil
			}
			rc.runDirectoryTestCase(t, path, getResult)
			return nil
		})
	require.NoError(t, err)
	require.NotZero(t, rc.testsCasesRun, "No complete test cases found in %s", rc.testDataDirectory)
}

func (rc *resultsChecker) runDirectoryTestCase(t *testing.T, path string, getResult func() (string, string)) {
	// cd into the directory so we can test functions that refer
	// local files by relative paths
	d, err := os.Getwd()
	require.NoError(t, err)

	defer func() { require.NoError(t, os.Chdir(d)) }()
	require.NoError(t, os.Chdir(path))

	expectedOutput, expectedError := rc.readAssertionFiles(t)
	if expectedError == "" && expectedOutput == "" && !rc.updateExpectedFromActual {
		// not a test directory: missing expectations and updateExpectedFromActual == false
		return
	}

	// run the test
	t.Run(path, func(t *testing.T) {
		rc.testsCasesRun += 1
		actualOutput, actualError := getResult()

		// Configured to update the assertion files instead of comparing them
		if rc.updateExpectedFromActual {
			if actualError != "" {
				require.NoError(t, ioutil.WriteFile(rc.expectedErrorFilename, []byte(actualError), 0600))
			}
			if len(actualOutput) > 0 {
				require.NoError(t, ioutil.WriteFile(rc.expectedOutputFilename, []byte(actualOutput), 0600))
			}
			t.Skip("Updated fixtures for test case")
		}

		// Compare the results to the assertion files
		if expectedError != "" {
			// We expected an error, so make sure there was one
			require.NotEmptyf(t, actualError, "test expected an error but message was empty, and output was:\n%s", actualOutput)
			rc.errorAssertionFunc(t, expectedError, actualError)
		} else {
			// We didn't expect an error, and the output should match
			require.Emptyf(t, actualError, "test expected no error but got an error message, and output was:\n%s", actualOutput)
			rc.outputAssertionFunc(t, expectedOutput, actualOutput)
		}
	})
}

// readAssertionFiles reads the expected results and error files
func (rc *resultsChecker) readAssertionFiles(t *testing.T) (string, string) {
	// read the expected results
	var expectedOutput, expectedError string
	if rc.expectedOutputFilename != "" {
		_, err := os.Stat(rc.expectedOutputFilename)
		if !os.IsNotExist(err) && err != nil {
			t.FailNow()
		}
		if err == nil {
			// only read the file if it exists
			b, err := ioutil.ReadFile(rc.expectedOutputFilename)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			expectedOutput = string(b)
		}
	}
	if rc.expectedErrorFilename != "" {
		_, err := os.Stat(rc.expectedErrorFilename)
		if !os.IsNotExist(err) && err != nil {
			t.FailNow()
		}
		if err == nil {
			// only read the file if it exists
			b, err := ioutil.ReadFile(rc.expectedErrorFilename)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			expectedError = string(b)
		}
	}
	return expectedOutput, expectedError
}
