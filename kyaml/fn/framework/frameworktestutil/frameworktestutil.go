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

	*checkerCore
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
	rc.checkerCore = &checkerCore{
		testDataDirectory:        rc.TestDataDirectory,
		expectedOutputFilename:   rc.ExpectedOutputFilename,
		expectedErrorFilename:    rc.ExpectedErrorFilename,
		updateExpectedFromActual: rc.UpdateExpectedFromActual,
		outputAssertionFunc:      rc.OutputAssertionFunc,
		errorAssertionFunc:       rc.ErrorAssertionFunc,
	}
	rc.checkerCore.setDefaults()
	runAllTestCases(t, rc)
	return true
}

func (rc *CommandResultsChecker) isTestDir(path string) bool {
	return atLeastOneFileExists(
		filepath.Join(path, rc.ConfigInputFilename),
		filepath.Join(path, rc.checkerCore.expectedOutputFilename),
		filepath.Join(path, rc.checkerCore.expectedErrorFilename),
	)
}

func (rc *CommandResultsChecker) runInCurrentDir(t *testing.T) (string, string) {
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

	_ = cmd.Execute()
	return stdOut.String(), stdErr.String()
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

	*checkerCore
}

// Assert runs the processor with the input provided in each valid test directory
// and verifies that the actual output and error match the fixtures in the directory.
func (rc *ProcessorResultsChecker) Assert(t *testing.T) bool {
	if rc.InputFilename == "" {
		rc.InputFilename = DefaultInputFilename
	}
	rc.checkerCore = &checkerCore{
		testDataDirectory:        rc.TestDataDirectory,
		expectedOutputFilename:   rc.ExpectedOutputFilename,
		expectedErrorFilename:    rc.ExpectedErrorFilename,
		updateExpectedFromActual: rc.UpdateExpectedFromActual,
		outputAssertionFunc:      rc.OutputAssertionFunc,
		errorAssertionFunc:       rc.ErrorAssertionFunc,
	}
	rc.checkerCore.setDefaults()
	runAllTestCases(t, rc)
	return true
}

func (rc *ProcessorResultsChecker) isTestDir(path string) bool {
	return atLeastOneFileExists(
		filepath.Join(path, rc.InputFilename),
		filepath.Join(path, rc.checkerCore.expectedOutputFilename),
		filepath.Join(path, rc.checkerCore.expectedErrorFilename),
	)
}

func (rc *ProcessorResultsChecker) runInCurrentDir(t *testing.T) (string, string) {
	_, err := os.Stat(rc.InputFilename)
	if os.IsNotExist(err) {
		t.Errorf("Test case is missing input file (default: %s)", DefaultInputFilename)
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

// resultsChecker is implemented by ProcessorResultsChecker and CommandResultsChecker, partially via checkerCore
type resultsChecker interface {
	// TestCasesRun returns a list of the test case directories that have been seen.
	TestCasesRun() (paths []string)

	// rootDir is the root of the tree of test directories to be searched for fixtures.
	rootDir() string
	// isTestDir takes the name of directory and returns whether or not it contains the files required to be a test case.
	isTestDir(dir string) bool
	// runInCurrentDir executes the code the checker is checking in the context of the current working directory.
	runInCurrentDir(t *testing.T) (stdOut, stdErr string)
	// resetTestCasesRun resets the list of test case directories seen.
	resetTestCasesRun()
	// recordTestCase adds to the list of test case directories seen.
	recordTestCase(path string)
	// readAssertionFiles returns the contents of the output and error fixtures
	readAssertionFiles(t *testing.T) (expectedOutput string, expectedError string)
	// shouldUpdateFixtures indicates whether or not this checker is currently being used to update fixtures instead of run them.
	shouldUpdateFixtures() bool
	// updateFixtures modifies the test fixture files to match the given content
	updateFixtures(t *testing.T, actualOutput string, actualError string)
	// assertOutputMatches compares the expected output to the output received.
	assertOutputMatches(t *testing.T, expected string, actual string)
	// assertErrorMatches compares the expected error to the error received.
	assertErrorMatches(t *testing.T, expected string, actual string)
}

// checkerCore implements the resultsChecker methods that are shared between the two checker types
type checkerCore struct {
	testDataDirectory        string
	expectedOutputFilename   string
	expectedErrorFilename    string
	updateExpectedFromActual bool
	outputAssertionFunc      AssertionFunc
	errorAssertionFunc       AssertionFunc

	testsCasesRun []string
}

func (rc *checkerCore) setDefaults() {
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
}

func (rc *checkerCore) rootDir() string {
	return rc.testDataDirectory
}

func (rc *checkerCore) TestCasesRun() []string {
	return rc.testsCasesRun
}

func (rc *checkerCore) resetTestCasesRun() {
	rc.testsCasesRun = []string{}
}

func (rc *checkerCore) recordTestCase(s string) {
	rc.testsCasesRun = append(rc.testsCasesRun, s)
}

func (rc *checkerCore) shouldUpdateFixtures() bool {
	return rc.updateExpectedFromActual
}

func (rc *checkerCore) updateFixtures(t *testing.T, actualOutput string, actualError string) {
	if actualError != "" {
		require.NoError(t, ioutil.WriteFile(rc.expectedErrorFilename, []byte(actualError), 0600))
	}
	if len(actualOutput) > 0 {
		require.NoError(t, ioutil.WriteFile(rc.expectedOutputFilename, []byte(actualOutput), 0600))
	}
	t.Skip("Updated fixtures for test case")
}

func (rc *checkerCore) assertOutputMatches(t *testing.T, expected string, actual string) {
	rc.outputAssertionFunc(t, expected, actual)
}

func (rc *checkerCore) assertErrorMatches(t *testing.T, expected string, actual string) {
	rc.errorAssertionFunc(t, expected, actual)
}

func (rc *checkerCore) readAssertionFiles(t *testing.T) (string, string) {
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

// runAllTestCases traverses rootDir to find test cases, calls getResult to invoke the function
// under test in each directory, and then runs assertions on the returned output and error results.
// It triggers a test failure if no valid test directories were found.
func runAllTestCases(t *testing.T, checker resultsChecker) {
	checker.resetTestCasesRun()
	err := filepath.Walk(checker.rootDir(),
		func(path string, info os.FileInfo, err error) error {
			require.NoError(t, err)
			if info.IsDir() && checker.isTestDir(path) {
				runDirectoryTestCase(t, path, checker)
			}
			return nil
		})
	require.NoError(t, err)
	require.NotZero(t, len(checker.TestCasesRun()), "No complete test cases found in %s", checker.rootDir())
}

func runDirectoryTestCase(t *testing.T, path string, checker resultsChecker) {
	// cd into the directory so we can test functions that refer to local files by relative paths
	d, err := os.Getwd()
	require.NoError(t, err)

	defer func() { require.NoError(t, os.Chdir(d)) }()
	require.NoError(t, os.Chdir(path))

	expectedOutput, expectedError := checker.readAssertionFiles(t)
	if expectedError == "" && expectedOutput == "" && !checker.shouldUpdateFixtures() {
		t.Fatalf("test directory %s must include either expected output or expected error fixture", path)
	}

	// run the test
	t.Run(path, func(t *testing.T) {
		checker.recordTestCase(path)
		actualOutput, actualError := checker.runInCurrentDir(t)

		// Configured to update the assertion files instead of comparing them
		if checker.shouldUpdateFixtures() {
			checker.updateFixtures(t, actualOutput, actualError)
		}

		// Compare the results to the assertion files
		if expectedError != "" {
			// We expected an error, so make sure there was one
			require.NotEmptyf(t, actualError, "test expected an error but message was empty, and output was:\n%s", actualOutput)
			checker.assertErrorMatches(t, expectedError, actualError)
		} else {
			// We didn't expect an error, and the output should match
			require.Emptyf(t, actualError, "test expected no error but got an error message, and output was:\n%s", actualOutput)
			checker.assertOutputMatches(t, expectedOutput, actualOutput)
		}
	})
}

func atLeastOneFileExists(files ...string) bool {
	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			return true
		}
	}
	return false
}
