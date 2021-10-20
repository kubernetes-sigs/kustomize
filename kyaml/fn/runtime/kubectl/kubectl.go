// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kubectl

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	osexec "os/exec"
	"time"

	runtimeexec "sigs.k8s.io/kustomize/kyaml/fn/runtime/exec"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// We're running krm via kubectl here:
// 1. If PodTemplateName is set - taking it from kubectl (the same namespace), and using defaultPodTemplate overwise. See some mandatory fields values there, e.g. stdin.
// 2. Creating pod using PodTemplate spec and applying imageUrl and set of envs
// 3. Waiting maximum PodStartTimeout for the pod up and running
// 4. using kubectl attach -q to send input and to get output
// 5. deleting the created pod

const (
	defaultPodTemplate = `
{
    "apiVersion": "v1",
    "kind": "PodTemplate",
    "metadata": {
        "name": "not-important",
        "labels": {
                "app": "krm-pod"
        }
    },
    "template": {
        "spec": {
            "containers": [
                {
                    "name": "default",
                    "stdin": true,
                    "stdinOnce": true
                }
            ],
            "restartPolicy": "Never"
	}
    }
}
`
	alphanums = "bcdfghjklmnpqrstvwxz2456789"
)

var (
	podTemplateCache        = ""
	podTemplateCacheForName = ""
)

type Filter struct {
	runtimeutil.ContainerSpec `json:",inline" yaml:",inline"`

	PodTemplateName   string
	KubectlGlobalArgs []string

	PodStartTimeout *time.Duration

	Exec runtimeexec.Filter
}

var rng *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = alphanums[rng.Intn(len(alphanums))]
	}
	return string(b)
}

func mergeYamlStrings(y1, y2 string) (string, error) {
	var data1 []map[string]interface{}
	err := yaml.Unmarshal([]byte(y1), &data1)
	if err != nil {
		return "", fmt.Errorf("couldn't unmarshal %s: %w", y1, err)
	}
	var data2 []map[string]interface{}
	err = yaml.Unmarshal([]byte(y2), &data2)
	if err != nil {
		return "", fmt.Errorf("couldn't unmarshal %s: %w", y1, err)
	}

	for i := range data2 {
		for j := range data1 {
			if data1[j]["name"] == data2[i]["name"] {
				// TODO - maybe first will override second?
				return "", fmt.Errorf("key %v exists in both yamls", data1[j]["name"])
			}
		}
	}

	data1 = append(data1, data2...)

	r, err := yaml.Marshal(data1)
	if err != nil {
		return "", fmt.Errorf("couldn't marshal %v: %w", data1, err)
	}
	return string(r), nil
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

func (c Filter) getPodTemplate() (*yaml.RNode, error) {
	if podTemplateCache == "" || podTemplateCacheForName != c.PodTemplateName {
		if c.PodTemplateName != "" {
			args := append(c.GetKubectlGlobalArgs(),
				"get",
				fmt.Sprintf("podTemplate/%s", c.PodTemplateName),
				"-o",
				"json")
			out, err := osexec.Command("kubectl", args...).Output()
			if err != nil {
				return nil, fmt.Errorf("kubectl %v failed: %w", args, err)
			}
			podTemplateCache = string(out)
		} else {
			podTemplateCache = defaultPodTemplate
		}
		podTemplateCacheForName = c.PodTemplateName
	}

	return yaml.ConvertJSONToYamlNode(podTemplateCache)
}

func (c Filter) getPodConfig(podName string) (string, error) {
	template, err := c.getPodTemplate()
	if err != nil {
		return "", fmt.Errorf("getPodTemplate failed: %w", err)
	}

	res, err := yaml.Parse(
		fmt.Sprintf(`
apiVersion: v1
kind: Pod
metadata:
  name: %s
`,
			podName,
		),
	)

	spec, err := template.Pipe(yaml.PathGetter{Path: []string{"template", "spec"}})
	if err != nil {
		return "", fmt.Errorf("PathGetter for spec failed: %w", err)
	}
	if spec == nil {
		return "", fmt.Errorf("specified template doesn't contain spec")
	}
	err = res.PipeE(yaml.FieldSetter{Name: "spec", Value: spec})
	if err != nil {
		return "", fmt.Errorf("FieldSetter for spec failed: %w", err)
	}

	// set remaining fields
	// labels:
	labels, err := template.Pipe(yaml.PathGetter{Path: []string{"metadata", "labels"}})
	if err != nil {
		return "", fmt.Errorf("PathGetter for spec failed: %w", err)
	}
	if labels != nil {
		metadata, err := res.Pipe(yaml.PathGetter{Path: []string{"metadata"}})
		if err != nil {
			return "", fmt.Errorf("PathGetter for metadata failed: %w", err)
		}
		err = metadata.PipeE(yaml.FieldSetter{Name: "labels", Value: labels})
		if err != nil {
			return "", fmt.Errorf("FieldSetter for metadata failed: %w", err)
		}
	}

	// image:
	container, err := res.Pipe(yaml.PathGetter{Path: []string{"spec", "containers", "0"}})
	if err != nil {
		return "", fmt.Errorf("PathGetter for default container failed: %w", err)
	}
	err = container.PipeE(yaml.FieldSetter{Name: "image", StringValue: c.Image})
	if err != nil {
		return "", fmt.Errorf("FieldSetter for image failed: %w", err)
	}
	// env:
	tmpltEnvData, err := res.Pipe(yaml.PathGetter{Path: []string{"spec", "containers", "0", "env"}})
	if err != nil {
		return "", fmt.Errorf("PathGetter for default container env failed: %w", err)
	}
	envDataStr := runtimeutil.NewContainerEnvFromStringSlice(c.Env).GetPodEnvConfig()
	if tmpltEnvData != nil {
		// update envDataStr with data from tmpltEnvData merged with envDataStr
		tmpltEnvDataStr, err := tmpltEnvData.String()
		if err != nil {
			return "", fmt.Errorf("failed to convert tmpltEnvData to string: %w", err)
		}
		envDataStr, err = mergeYamlStrings(envDataStr, tmpltEnvDataStr)
		if err != nil {
			return "", fmt.Errorf("failed to merge %s and %s: %w", envDataStr, tmpltEnvDataStr, err)
		}
	}
	envData, err := yaml.Parse(envDataStr)
	if err != nil {
		return "", fmt.Errorf("Wasn't able to parse env string: %w", err)
	}
	err = container.PipeE(yaml.FieldSetter{Name: "env", Value: envData})
	if err != nil {
		return "", fmt.Errorf("FieldSetter for env failed: %w", err)
	}

	str, err := res.String()
	if err != nil {
		return "", fmt.Errorf("couldn't convert to string: %w", err)
	}
	return str, nil
}

func (c *Filter) GetKubectlGlobalArgs() []string {
	if c.KubectlGlobalArgs == nil {
		return []string{}
	}
	return c.KubectlGlobalArgs
}

func (c *Filter) runKubctlPodCmd(podName, cmdLine string) error {
	podConfig, err := c.getPodConfig(podName)
	if err != nil {
		return err
	}

	args := append(c.GetKubectlGlobalArgs(),
		cmdLine,
		"-f",
		"-")

	cmd := osexec.Command("kubectl", args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, podConfig)
	}()
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (c *Filter) getPodStartTimeout() time.Duration {
	if c.PodStartTimeout != nil {
		return *c.PodStartTimeout
	}

	return 60 * time.Second
}

func (c *Filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	podName := "krm-" + randString(5)

	// prepare attach command
	if err := c.setupExec(podName); err != nil {
		return nil, fmt.Errorf("setup exec filter returned error: %w", err)
	}
	// start pod
	if err := c.runKubctlPodCmd(podName, "create"); err != nil {
		return nil, fmt.Errorf("apply returned error: %w", err)
	}
	// we we'll need to clean that up
	defer c.runKubctlPodCmd(podName, "delete")

	// wait till pod is up
	args := append(c.GetKubectlGlobalArgs(),
		"wait",
		"--for=condition=Ready",
		fmt.Sprintf("--timeout=%s", c.getPodStartTimeout().String()),
		fmt.Sprintf("pod/%s", podName))

	cmd := osexec.Command("kubectl", args...)
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("kubectl with args %v returned error: %w", args, err)
	}

	return c.Exec.Filter(nodes)
}

func (c *Filter) setupExec(podName string) error {
	// don't init 2x
	if c.Exec.Path != "" {
		return nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	c.Exec.WorkingDir = wd

	path, args := c.getCommand(podName)
	c.Exec.Path = path
	c.Exec.Args = args
	return nil
}

// getArgs returns the command + args to run to spawn the container
func (c *Filter) getCommand(podName string) (string, []string) {
	args := append(c.GetKubectlGlobalArgs(),
		"attach",
		"-iq",
		podName,
	)
	return "kubectl", args
}

// NewContainer returns a new container filter
func NewContainer(spec runtimeutil.ContainerSpec, podTemplateName string, kubectlGlobalArgs []string, podStartTimeout *time.Duration) Filter {
	f := Filter{ContainerSpec: spec, PodTemplateName: podTemplateName, KubectlGlobalArgs: kubectlGlobalArgs, PodStartTimeout: podStartTimeout}

	return f
}
