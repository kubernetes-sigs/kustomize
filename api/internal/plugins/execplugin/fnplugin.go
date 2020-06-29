// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package execplugin

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"

	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
	"sigs.k8s.io/kustomize/kyaml/runfn"
)

type FnPlugin struct {
	// Function runner
	runFns runfn.RunFns

	// Plugin configuration data.
	cfg []byte

	// Plugin name cache for error output
	pluginName string

	// PluginHelpers
	h *resmap.PluginHelpers
}

func bytesToRNode(yml []byte) (*yaml.RNode, error) {
	rnode, err := yaml.Parse(string(yml))
	if err != nil {
		return nil, err
	}
	return rnode, nil
}

func resourceToRNode(res *resource.Resource) (*yaml.RNode, error) {
	yml, err := res.AsYAML()
	if err != nil {
		return nil, err
	}

	return bytesToRNode(yml)
}

func GetFunctionSpec(res *resource.Resource) (*runtimeutil.FunctionSpec, error) {
	rnode, err := resourceToRNode(res)
	if err != nil {
		return nil, err
	}

	fSpec := runtimeutil.GetFunctionSpec(rnode)
	if fSpec == nil {
		return nil, fmt.Errorf("resource %v doesn't contain function spec", res.GetGvk())
	}

	return fSpec, nil
}

func toStorageMounts(mounts []string) []runtimeutil.StorageMount {
	var sms []runtimeutil.StorageMount
	for _, mount := range mounts {
		sms = append(sms, runtimeutil.StringToStorageMount(mount))
	}
	return sms
}

func NewFnPlugin(o *types.FnPluginLoadingOptions) *FnPlugin {
	//log.Printf("options: %v\n", o)
	return &FnPlugin{
		runFns: runfn.RunFns{
			Functions:      []*yaml.RNode{},
			Network:        o.Network,
			NetworkName:    o.NetworkName,
			EnableStarlark: o.EnableStar,
			EnableExec:     o.EnableExec,
			StorageMounts:  toStorageMounts(o.Mounts),
		},
	}
}

func (p *FnPlugin) Cfg() []byte {
	return p.cfg
}

func (p *FnPlugin) Config(h *resmap.PluginHelpers, config []byte) error {
	p.h = h
	p.cfg = config

	rnode, err := bytesToRNode(p.cfg)
	if err != nil {
		return err
	}

	meta, err := rnode.GetMeta()
	if err != nil {
		return err
	}

	p.pluginName = fmt.Sprintf("api: %s, kind: %s, name: %s",
		meta.APIVersion, meta.Kind, meta.Name)
	//log.Printf("config based pluginName: %s", p.pluginName)

	return nil
}

func (p *FnPlugin) Generate() (resmap.ResMap, error) {
	output, err := p.invokePlugin(nil)
	if err != nil {
		return nil, err
	}
	rm, err := p.h.ResmapFactory().NewResMapFromBytes(output)
	if err != nil {
		return nil, err
	}
	return UpdateResourceOptions(rm)
}

func (p *FnPlugin) Transform(rm resmap.ResMap) error {
	// add ResIds as annotations to all objects so that we can add them back
	inputRM, err := getResMapWithIdAnnotation(rm)
	if err != nil {
		return err
	}

	// encode the ResMap so it can be fed to the plugin
	resources, err := inputRM.AsYaml()
	if err != nil {
		return err
	}

	// invoke the plugin with resources as the input
	output, err := p.invokePlugin(resources)
	if err != nil {
		return fmt.Errorf("%v %s", err, string(output))
	}

	// update the original ResMap based on the output
	return updateResMapValues(p.pluginName, p.h, output, rm)
}

// invokePlugin uses Function runner to run function as plugin
func (p *FnPlugin) invokePlugin(input []byte) ([]byte, error) {
	// get config rnode
	rnode, err := bytesToRNode(p.cfg)
	if err != nil {
		return nil, err
	}
	err = rnode.PipeE(yaml.SetAnnotation("config.kubernetes.io/local-config", "true"))
	if err != nil {
		return nil, err
	}

	// we need to add config as input for generators. Some of them don't work with FunctionConfig
	// and in addition kio.Pipeline won't create anything if there are no objects
	// see https://github.com/kubernetes-sigs/kustomize/blob/master/kyaml/kio/kio.go#L93
	if input == nil {
		yaml, err := rnode.String()
		if err != nil {
			return nil, err
		}
		input = []byte(yaml)
	}

	// Transform to ResourceList
	var inOut bytes.Buffer
	inIn := bytes.NewReader(input)
	err = kio.Pipeline{
		Inputs: []kio.Reader{&kio.ByteReader{Reader: inIn}},
		Outputs: []kio.Writer{kio.ByteWriter{
			Writer:             &inOut,
			WrappingKind:       kio.ResourceListKind,
			WrappingAPIVersion: kio.ResourceListAPIVersion}},
	}.Execute()
	if err != nil {
		return nil, errors.Wrap(
			err, "couldn't transform to ResourceList")
	}
	//log.Printf("converted to:\n%s\n", inOut.String())

	// Configure and Execute Fn
	var runFnsOut bytes.Buffer
	p.runFns.Input = bytes.NewReader(inOut.Bytes())
	p.runFns.Functions = append(p.runFns.Functions, rnode)
	p.runFns.Output = &runFnsOut

	err = p.runFns.Execute()
	if err != nil {
		return nil, errors.Wrap(
			err, "couldn't execute function")
	}

	//log.Printf("fn returned:\n%s\n", runFnsOut.String())

	// Convert back to a single multi-yaml doc
	var outOut bytes.Buffer
	outIn := bytes.NewReader(runFnsOut.Bytes())

	err = kio.Pipeline{
		Inputs:  []kio.Reader{&kio.ByteReader{Reader: outIn}},
		Outputs: []kio.Writer{kio.ByteWriter{Writer: &outOut}},
	}.Execute()
	if err != nil {
		return nil, errors.Wrap(
			err, "couldn't transform from ResourceList")
	}

	//log.Printf("converted back to:\n%s\n", outOut.String())

	return outOut.Bytes(), nil
}
