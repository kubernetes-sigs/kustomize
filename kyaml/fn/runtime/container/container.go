// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"fmt"
	"os"
	"strings"

	runtimeexec "sigs.k8s.io/kustomize/kyaml/fn/runtime/exec"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Filter filters Resources using a container image.
// The container must start a process that reads the list of
// input Resources from stdin, reads the Configuration from the env
// API_CONFIG, and writes the filtered Resources to stdout.
// If there is a error or validation failure, the process must exit
// non-zero.
// The full set of environment variables from the parent process
// are passed to the container.
//
// Function Scoping:
// Filter applies the function only to Resources to which it is scoped.
//
// Resources are scoped to a function if any of the following are true:
// - the Resource were read from the same directory as the function config
// - the Resource were read from a subdirectory of the function config directory
// - the function config is in a directory named "functions" and
//   they were read from a subdirectory of "functions" parent
// - the function config doesn't have a path annotation (considered globally scoped)
// - the Filter has GlobalScope == true
//
// In Scope Examples:
//
// Example 1: deployment.yaml and service.yaml in function.yaml scope
//            same directory as the function config directory
//     .
//     ├── function.yaml
//     ├── deployment.yaml
//     └── service.yaml
//
// Example 2: apps/deployment.yaml and apps/service.yaml in function.yaml scope
//            subdirectory of the function config directory
//     .
//     ├── function.yaml
//     └── apps
//         ├── deployment.yaml
//         └── service.yaml
//
// Example 3: apps/deployment.yaml and apps/service.yaml in functions/function.yaml scope
//            function config is in a directory named "functions"
//     .
//     ├── functions
//     │   └── function.yaml
//     └── apps
//         ├── deployment.yaml
//         └── service.yaml
//
// Out of Scope Examples:
//
// Example 1: apps/deployment.yaml and apps/service.yaml NOT in stuff/function.yaml scope
//     .
//     ├── stuff
//     │   └── function.yaml
//     └── apps
//         ├── deployment.yaml
//         └── service.yaml
//
// Example 2: apps/deployment.yaml and apps/service.yaml NOT in stuff/functions/function.yaml scope
//     .
//     ├── stuff
//     │   └── functions
//     │       └── function.yaml
//     └── apps
//         ├── deployment.yaml
//         └── service.yaml
//
// Default Paths:
// Resources emitted by functions will have default path applied as annotations
// if none is present.
// The default path will be the function-dir/ (or parent directory in the case of "functions")
// + function-file-name/ + namespace/ + kind_name.yaml
//
// Example 1: Given a function in fn.yaml that produces a Deployment name foo and a Service named bar
//     dir
//     └── fn.yaml
//
// Would default newly generated Resources to:
//
//     dir
//     ├── fn.yaml
//     └── fn
//         ├── deployment_foo.yaml
//         └── service_bar.yaml
//
// Example 2: Given a function in functions/fn.yaml that produces a Deployment name foo and a Service named bar
//     dir
//     └── fn.yaml
//
// Would default newly generated Resources to:
//
//     dir
//     ├── functions
//     │   └── fn.yaml
//     └── fn
//         ├── deployment_foo.yaml
//         └── service_bar.yaml
//
// Example 3: Given a function in fn.yaml that produces a Deployment name foo, namespace baz and a Service named bar namespace baz
//     dir
//     └── fn.yaml
//
// Would default newly generated Resources to:
//
//     dir
//     ├── fn.yaml
//     └── fn
//         └── baz
//             ├── deployment_foo.yaml
//             └── service_bar.yaml
type Filter struct {

	// Image is the container image to use to create a container.
	Image string `yaml:"image,omitempty"`

	// Network is the container network to use.
	Network string `yaml:"network,omitempty"`

	// StorageMounts is a list of storage options that the container will have mounted.
	StorageMounts []runtimeutil.StorageMount `yaml:"mounts,omitempty"`

	Exec runtimeexec.Filter
}

func (c Filter) String() string {
	if c.Exec.DeferFailure {
		return fmt.Sprintf("%s deferFailure: %v", c.Image, c.Exec.DeferFailure)
	}
	return c.Image
}
func (c Filter) GetExit() error {
	return c.Exec.GetExit()
}

func (c *Filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	c.setupExec()
	return c.Exec.Filter(nodes)
}

func (c *Filter) setupExec() {
	// don't init 2x
	if c.Exec.Path != "" {
		return
	}

	path, args := c.getCommand()
	c.Exec.Path = path
	c.Exec.Args = args
}

// getArgs returns the command + args to run to spawn the container
func (c *Filter) getCommand() (string, []string) {
	// run the container using docker.  this is simpler than using the docker
	// libraries, and ensures things like auth work the same as if the container
	// was run from the cli.

	network := "none"
	if c.Network != "" {
		network = c.Network
	}

	args := []string{"run",
		"--rm",                                              // delete the container afterward
		"-i", "-a", "STDIN", "-a", "STDOUT", "-a", "STDERR", // attach stdin, stdout, stderr
		"--network", network,

		// added security options
		"--user", "nobody", // run as nobody
		"--security-opt=no-new-privileges", // don't allow the user to escalate privileges
		// note: don't make fs readonly because things like heredoc rely on writing tmp files
	}

	// TODO(joncwong): Allow StorageMount fields to have default values.
	for _, storageMount := range c.StorageMounts {
		args = append(args, "--mount", storageMount.String())
	}

	os.Setenv("LOG_TO_STDERR", "true")
	os.Setenv("STRUCTURED_RESULTS", "true")

	// export the local environment vars to the container
	for _, pair := range os.Environ() {
		items := strings.Split(pair, "=")
		if items[0] == "" || items[1] == "" {
			continue
		}
		args = append(args, "-e", items[0])
	}
	a := append(args, c.Image)
	return "docker", a
}
