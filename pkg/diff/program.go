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

package diff

import (
	"io"
	"os"

	"github.com/kubernetes-sigs/kustomize/pkg/exec"
)

// program wraps the system diff program.
// If specified, the value of KUBERNETES_EXTERNAL_DIFF environment variable
// will be used instead of simply `diff(1)`.
type program struct {
	stdout io.Writer
	stderr io.Writer
}

func newProgram(out, errOut io.Writer) *program {
	return &program{
		stdout: out,
		stderr: errOut,
	}
}

func (d *program) makeCommand(args ...string) exec.Cmd {
	diff := ""
	if envDiff := os.Getenv("KUBERNETES_EXTERNAL_DIFF"); envDiff != "" {
		diff = envDiff
	} else {
		diff = "diff"
		args = append([]string{"-u", "-N"}, args...)
	}
	cmd := exec.New().Command(diff, args...)
	cmd.SetStdout(d.stdout)
	cmd.SetStderr(d.stderr)
	return cmd
}

// run runs the detected diff program. `from` and `to` are the directory to diff.
func (d *program) run(from, to string) error {
	d.makeCommand(from, to).Run() // Ignore diff return code
	return nil
}
