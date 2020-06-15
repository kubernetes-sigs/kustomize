// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fnplugin

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/pkg/errors"

	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/kustomize/kyaml/kio"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"

	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
	"sigs.k8s.io/kustomize/kyaml/runfn"
)

const (
	idAnnotation       = "kustomize.config.k8s.io/id"
	HashAnnotation     = "kustomize.config.k8s.io/needs-hash"
	BehaviorAnnotation = "kustomize.config.k8s.io/behavior"
)

type FnPlugin struct {
	// Function runner
	runFns runfn.RunFns

	// Plugin configuration data.
	cfg []byte

	// PluginHelpers
	h *resmap.PluginHelpers
}

func bytesToRNode(yml []byte) (*kyaml.RNode, error) {
	rnode, err := kyaml.Parse(string(yml))
	if err != nil {
		return nil, err
	}
	return rnode, nil
}

func resourceToRNode(res *resource.Resource) (*kyaml.RNode, error) {
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
			Functions:      []*kyaml.RNode{},
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
	return p.UpdateResourceOptions(rm)
}

func (p *FnPlugin) Transform(rm resmap.ResMap) error {
	// add ResIds as annotations to all objects so that we can add them back
	inputRM, err := p.getResMapWithIdAnnotation(rm)
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
	return p.updateResMapValues(output, rm)
}

// invokePlugin uses Function runner to run function as plugin
func (p *FnPlugin) invokePlugin(input []byte) ([]byte, error) {
	// get config rnode
	rnode, err := bytesToRNode(p.cfg)
	if err != nil {
		return nil, err
	}
	err = rnode.PipeE(kyaml.SetAnnotation("config.kubernetes.io/local-config", "true"))
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

// Returns a new copy of the given ResMap with the ResIds annotated in each Resource
func (p *FnPlugin) getResMapWithIdAnnotation(rm resmap.ResMap) (resmap.ResMap, error) {
	inputRM := rm.DeepCopy()
	for _, r := range inputRM.Resources() {
		idString, err := yaml.Marshal(r.CurId())
		if err != nil {
			return nil, err
		}
		annotations := r.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		annotations[idAnnotation] = string(idString)
		r.SetAnnotations(annotations)
	}
	return inputRM, nil
}

// updateResMapValues updates the Resource value in the given ResMap
// with the emitted Resource values in output.
func (p *FnPlugin) updateResMapValues(output []byte, rm resmap.ResMap) error {
	outputRM, err := p.h.ResmapFactory().NewResMapFromBytes(output)
	if err != nil {
		return err
	}
	for _, r := range outputRM.Resources() {
		// for each emitted Resource, find the matching Resource in the original ResMap
		// using its id
		annotations := r.GetAnnotations()
		idString, ok := annotations[idAnnotation]
		if !ok {
			return fmt.Errorf("the transformer should not remove annotation %s",
				idAnnotation)
		}
		id := resid.ResId{}
		err := yaml.Unmarshal([]byte(idString), &id)
		if err != nil {
			return err
		}
		res, err := rm.GetByCurrentId(id)
		if err != nil {
			return fmt.Errorf("unable to find unique match to %s", id.String())
		}
		// remove the annotation set by Kustomize to track the resource
		delete(annotations, idAnnotation)
		if len(annotations) == 0 {
			annotations = nil
		}
		r.SetAnnotations(annotations)

		// update the ResMap resource value with the transformed object
		res.Kunstructured = r.Kunstructured
	}
	return nil
}

// updateResourceOptions updates the generator options for each resource in the
// given ResMap based on plugin provided annotations.
func (p *FnPlugin) UpdateResourceOptions(rm resmap.ResMap) (resmap.ResMap, error) {
	for _, r := range rm.Resources() {
		// Disable name hashing by default and require plugin to explicitly
		// request it for each resource.
		annotations := r.GetAnnotations()
		behavior := annotations[BehaviorAnnotation]
		var needsHash bool
		if val, ok := annotations[HashAnnotation]; ok {
			b, err := strconv.ParseBool(val)
			if err != nil {
				return nil, fmt.Errorf(
					"the annotation %q contains an invalid value (%q)",
					HashAnnotation, val)
			}
			needsHash = b
		}
		delete(annotations, HashAnnotation)
		delete(annotations, BehaviorAnnotation)
		if len(annotations) == 0 {
			annotations = nil
		}
		r.SetAnnotations(annotations)
		r.SetOptions(types.NewGenArgs(
			&types.GeneratorArgs{
				Behavior: behavior,
				Options:  &types.GeneratorOptions{DisableNameSuffixHash: !needsHash}}))
	}
	return rm, nil
}
