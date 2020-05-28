package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

type module struct {
	// Module name
	name string
	// Module path
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

func (m *module) CheckModReplace() error {
	goModPath := filepath.Join(m.path, m.name, "go.mod")
	info, err := os.Stat(goModPath)
	if os.IsNotExist(err) || info.IsDir() {
		return nil
	}

	goModContent, err := ioutil.ReadFile(goModPath)
	if err != nil {
		return err
	}
	return checkModReplace(goModPath, goModContent)
}

func checkModReplace(path string, data []byte) error {
	f, err := modfile.Parse(path, data, nil)
	if err != nil {
		return err
	}
	if len(f.Replace) > 0 {
		var msg strings.Builder
		for _, replace := range f.Replace {
			fmt.Fprintf(&msg, " - Please update go.mod to pin a specific version of %s\n", replace.Old.Path)
		}
		return fmt.Errorf("Found replace in %s\n%s", path, msg.String())
	}
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
