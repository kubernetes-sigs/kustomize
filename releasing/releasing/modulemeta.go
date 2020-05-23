package main

import (
	"os/exec"
	"path/filepath"
)

type module struct {
	// Module name
	name string
	// Module path. This path is only used for running test.
	path string
	// Module version
	version moduleVersion
}

func (m *module) UpdateVersion(tags string) error {
	v, err := newModuleVersionFromGitTags(tags, m.name)
	if err != nil {
		return err
	}
	m.version = v
	return nil
}

func (m *module) Tag() string {
	return m.name + "/" + m.version.String()
}

func (m *module) RunTest() (string, error) {
	if !doTest {
		logInfo("Tests disabled.")
		return "", nil
	}
	testPath := filepath.Join(m.path, m.name)
	logInfo("Running tests in %s...", testPath)
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = testPath
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return string(stdoutStderr), err
	}
	logInfo("Tests are successfully finished")
	return "", nil
}
