/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"

	"k8s.io/utils/exec"
)

// DiffProgram finds and run the diff program. The value of
// KUBERNETES_EXTERNAL_DIFF environment variable will be used a diff
// program. By default, `diff(1)` will be used.
type DiffProgram struct {
	Exec   exec.Interface
	Stdout io.Writer
	Stderr io.Writer
}

func (d *DiffProgram) getCommand(args ...string) exec.Cmd {
	diff := ""
	if envDiff := os.Getenv("KUBERNETES_EXTERNAL_DIFF"); envDiff != "" {
		diff = envDiff
	} else {
		diff = "diff"
		args = append([]string{"-u", "-N"}, args...)
	}

	cmd := d.Exec.Command(diff, args...)
	cmd.SetStdout(d.Stdout)
	cmd.SetStderr(d.Stderr)

	return cmd
}

// Run runs the detected diff program. `from` and `to` are the directory to diff.
func (d *DiffProgram) Run(from, to string) error {
	d.getCommand(from, to).Run() // Ignore diff return code
	return nil
}

// Printer is used to print an object.
type Printer struct{}

// Print the object inside the writer w.
func (p *Printer) Print(obj interface{}, w io.Writer) error {
	if obj == nil {
		return nil
	}
	data, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err

}

// Directory creates a new temp directory, and allows to easily create new files.
type Directory struct {
	Name string
}

// CreateDirectory does create the actual disk directory, and return a
// new representation of it.
func CreateDirectory(prefix string) (*Directory, error) {
	name, err := ioutil.TempDir("", prefix+"-")
	if err != nil {
		return nil, err
	}

	return &Directory{
		Name: name,
	}, nil
}

// NewFile creates a new file in the directory.
func (d *Directory) NewFile(name string) (*os.File, error) {
	return os.OpenFile(filepath.Join(d.Name, name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0700)
}

// Delete removes the directory recursively.
func (d *Directory) Delete() error {
	return os.RemoveAll(d.Name)
}
