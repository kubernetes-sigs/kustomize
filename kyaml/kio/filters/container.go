// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filters

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// ContainerFilter filters Resources using a container image.
// The container must start a process that reads the list of
// input Resources from stdin, reads the Configuration from the env
// API_CONFIG, and writes the filtered Resources to stdout.
// If there is a error or validation failure, the process must exit
// non-zero.
// The full set of environment variables from the parent process
// are passed to the container.
//
// Function Scoping:
// ContainerFilter applies the function only to Resources to which it is scoped.
//
// Resources are scoped to a function if any of the following are true:
// - the Resource were read from the same directory as the function config
// - the Resource were read from a subdirectory of the function config directory
// - the function config is in a directory named "functions" and
//   they were read from a subdirectory of "functions" parent
// - the function config doesn't have a path annotation (considered globally scoped)
// - the ContainerFilter has GlobalScope == true
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
type ContainerFilter struct {

	// Image is the container image to use to create a container.
	Image string `yaml:"image,omitempty"`

	// Network is the container network to use.
	Network string `yaml:"network,omitempty"`

	// StorageMounts is a list of storage options that the container will have mounted.
	StorageMounts []StorageMount

	// Config is the API configuration for the container and passed through the
	// API_CONFIG env var to the container.
	// Typically a Kubernetes style Resource Config.
	Config *yaml.RNode `yaml:"config,omitempty"`

	// GlobalScope will cause the function to be run against all input
	// nodes instead of only nodes scoped under the function.
	GlobalScope bool

	// args may be specified by tests to override how a container is spawned
	args []string

	checkInput func(string)
}

func (c ContainerFilter) String() string {
	return c.Image
}

// StorageMount represents a container's mounted storage option(s)
type StorageMount struct {
	// Type of mount e.g. bind mount, local volume, etc.
	MountType string

	// Source for the storage to be mounted.
	// For named volumes, this is the name of the volume.
	// For anonymous volumes, this field is omitted (empty string).
	// For bind mounts, this is the path to the file or directory on the host.
	Src string

	// The path where the file or directory is mounted in the container.
	DstPath string
}

func (s *StorageMount) String() string {
	return fmt.Sprintf("type=%s,src=%s,dst=%s:ro", s.MountType, s.Src, s.DstPath)
}

// functionsDirectoryName is keyword directory name for functions scoped 1 directory higher
const functionsDirectoryName = "functions"

// getFunctionScope returns the path of the directory containing the function config,
// or its parent directory if the base directory is named "functions"
func (c *ContainerFilter) getFunctionScope() (string, error) {
	m, err := c.Config.GetMeta()
	if err != nil {
		return "", errors.Wrap(err)
	}
	p, found := m.Annotations[kioutil.PathAnnotation]
	if !found {
		return "", nil
	}

	functionDir := path.Clean(path.Dir(p))

	if path.Base(functionDir) == functionsDirectoryName {
		// the scope of functions in a directory called "functions" is 1 level higher
		// this is similar to how the golang "internal" directory scoping works
		functionDir = path.Dir(functionDir)
	}
	return functionDir, nil
}

// scope partitions the input nodes into 2 slices.  The first slice contains only Resources
// which are scoped under dir, and the second slice contains the Resources which are not.
func (c *ContainerFilter) scope(dir string, nodes []*yaml.RNode) ([]*yaml.RNode, []*yaml.RNode, error) {
	// scope container filtered Resources to Resources under that directory
	var input, saved []*yaml.RNode
	if c.GlobalScope {
		return nodes, nil, nil
	}

	if dir == "" {
		// global function
		return nodes, nil, nil
	}

	// identify Resources read from directories under the function configuration
	for i := range nodes {
		m, err := nodes[i].GetMeta()
		if err != nil {
			return nil, nil, err
		}
		p, found := m.Annotations[kioutil.PathAnnotation]
		if !found {
			// this Resource isn't scoped under the function -- don't know where it came from
			// consider it out of scope
			saved = append(saved, nodes[i])
			continue
		}

		resourceDir := path.Clean(path.Dir(p))
		if path.Base(resourceDir) == functionsDirectoryName {
			// Functions in the `functions` directory are scoped to
			// themselves, and should see themselves as input
			resourceDir = path.Dir(resourceDir)
		}
		if !strings.HasPrefix(resourceDir, dir) {
			// this Resource doesn't fall under the function scope if it
			// isn't in a subdirectory of where the function lives
			saved = append(saved, nodes[i])
			continue
		}

		// this input is scoped under the function
		input = append(input, nodes[i])
	}

	return input, saved, nil
}

// GrepFilter implements kio.GrepFilter
func (c *ContainerFilter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	// get the command to filter the Resources
	cmd, err := c.getCommand()
	if err != nil {
		return nil, err
	}

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}

	// only process Resources scoped to this function, save the others
	functionDir, err := c.getFunctionScope()
	if err != nil {
		return nil, err
	}
	input, saved, err := c.scope(functionDir, nodes)
	if err != nil {
		return nil, err
	}

	// write the input
	err = kio.ByteWriter{
		WrappingAPIVersion:    kio.ResourceListAPIVersion,
		WrappingKind:          kio.ResourceListKind,
		Writer:                in,
		KeepReaderAnnotations: true,
		FunctionConfig:        c.Config}.Write(input)
	if err != nil {
		return nil, err
	}

	// capture the command stdout for the return value
	r := &kio.ByteReader{Reader: out}

	// do the filtering
	if c.checkInput != nil {
		c.checkInput(in.String())
	}
	cmd.Stdin = in
	cmd.Stdout = out
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	output, err := r.Read()
	if err != nil {
		return nil, err
	}

	// annotate any generated Resources with a path and index if they don't already have one
	if err := kioutil.DefaultPathAnnotation(functionDir, output); err != nil {
		return nil, err
	}

	// emit both the Resources output from the function, and the out-of-scope Resources
	// which were not provided to the function
	return append(output, saved...), nil
}

// getArgs returns the command + args to run to spawn the container
func (c *ContainerFilter) getArgs() []string {
	// run the container using docker.  this is simpler than using the docker
	// libraries, and ensures things like auth work the same as if the container
	// was run from the cli.

	network := "none"
	if c.Network != "" {
		network = c.Network
	}

	args := []string{"docker", "run",
		"--rm",                                              // delete the container afterward
		"-i", "-a", "STDIN", "-a", "STDOUT", "-a", "STDERR", // attach stdin, stdout, stderr

		// added security options
		"--network", network,
		"--user", "nobody", // run as nobody
		// don't make fs readonly because things like heredoc rely on writing tmp files
		"--security-opt=no-new-privileges", // don't allow the user to escalate privileges
	}

	// TODO(joncwong): Allow StorageMount fields to have default values.
	for _, storageMount := range c.StorageMounts {
		args = append(args, "--mount", storageMount.String())
	}

	// export the local environment vars to the container
	for _, pair := range os.Environ() {
		tokens := strings.Split(pair, "=")
		if tokens[0] == "" {
			continue
		}
		args = append(args, "-e", tokens[0])
	}
	return append(args, c.Image)
}

// getCommand returns a command which will apply the Filter using the container image
func (c *ContainerFilter) getCommand() (*exec.Cmd, error) {
	// encode the filter command API configuration
	cfg := &bytes.Buffer{}
	if err := func() error {
		e := yaml.NewEncoder(cfg)
		defer e.Close()
		// make it fit on a single line
		c.Config.YNode().Style = yaml.FlowStyle
		return e.Encode(c.Config.YNode())
	}(); err != nil {
		return nil, err
	}

	if len(c.args) == 0 {
		c.args = c.getArgs()
	}

	cmd := exec.Command(c.args[0], c.args[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	// set stderr for err messaging
	return cmd, nil
}

// IsReconcilerFilter filters Resources based on whether or not they are Reconciler Resource.
// Resources with an apiVersion starting with '*.gcr.io', 'gcr.io' or 'docker.io' are considered
// Reconciler Resources.
type IsReconcilerFilter struct {
	// ExcludeReconcilers if set to true, then Reconcilers will be excluded -- e.g.
	// Resources with a reconcile container through the apiVersion (gcr.io prefix) or
	// through the annotations
	ExcludeReconcilers bool `yaml:"excludeReconcilers,omitempty"`

	// IncludeNonReconcilers if set to true, the NonReconciler will be included.
	IncludeNonReconcilers bool `yaml:"includeNonReconcilers,omitempty"`
}

// Filter implements kio.Filter
func (c *IsReconcilerFilter) Filter(inputs []*yaml.RNode) ([]*yaml.RNode, error) {
	var out []*yaml.RNode
	for i := range inputs {
		img, _ := GetContainerName(inputs[i])
		isContainerResource := img != ""
		if isContainerResource && !c.ExcludeReconcilers {
			out = append(out, inputs[i])
		}
		if !isContainerResource && c.IncludeNonReconcilers {
			out = append(out, inputs[i])
		}
	}
	return out, nil
}

const (
	FunctionAnnotationKey    = "config.kubernetes.io/function"
	oldFunctionAnnotationKey = "config.k8s.io/function"
)

var functionAnnotationKeys = []string{FunctionAnnotationKey, oldFunctionAnnotationKey}

// GetContainerName returns the container image for an API if one exists
func GetContainerName(n *yaml.RNode) (string, string) {
	meta, _ := n.GetMeta()

	// path to the function, this will be mounted into the container
	path := meta.Annotations[kioutil.PathAnnotation]

	// check previous keys for backwards compatibility
	for _, s := range functionAnnotationKeys {
		functionAnnotation := meta.Annotations[s]
		if functionAnnotation != "" {
			annotationContent, _ := yaml.Parse(functionAnnotation)
			image, _ := annotationContent.Pipe(yaml.Lookup("container", "image"))
			return image.YNode().Value, path
		}
	}

	container := meta.Annotations["config.kubernetes.io/container"]
	if container != "" {
		return container, path
	}

	image, err := n.Pipe(yaml.Lookup("metadata", "configFn", "container", "image"))
	if err != nil || yaml.IsMissingOrNull(image) {
		return "", path
	}
	return image.YNode().Value, path
}
