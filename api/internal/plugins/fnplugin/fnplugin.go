package fnplugin

import (
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
	"sigs.k8s.io/kustomize/kyaml/runfn"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// FnPlugin holds the information about thie function
type FnPlugin struct {
	// Plugin configuration data.
	cfg []byte

	// PluginHelpers
	h *resmap.PluginHelpers

	// Function runner
	RunFns runfn.RunFns
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

func getFunctionSpec(res *resource.Resource) *runtimeutil.FunctionSpec {
	rnode, err := resourceToRNode(res)
	if err != nil {
		return nil
	}

	return runtimeutil.GetFunctionSpec(rnode)
}

// IsFnPlugin returns a resource is a function plugin spec or not
func IsFnPlugin(res *resource.Resource) bool {
	return getFunctionSpec(res) != nil
}

// NewFnPlugin returns a FnPlugin struct
func NewFnPlugin(res *resource.Resource) *FnPlugin {
	return &FnPlugin{}
}

// Config accepts the plugin helper and plugin config
func (f *FnPlugin) Config(h *resmap.PluginHelpers, config []byte) error {
	f.h = h
	f.cfg = config
	// config is the content of the config file for the functions.
	// If there are multiple functions in on config file, they will
	// be passed in one by one.
	fn, err := bytesToRNode(config)
	if err != nil {
		return err
	}

	f.RunFns.Functions = append(f.RunFns.Functions, fn)

	return nil
}

// Transform does the transformation when the plugin is a transformer
func (f *FnPlugin) Transform(rm resmap.ResMap) error {
	// convert input to ResourceList
	// add functionConfig
	// invoke function
	// convert back to ResMap
	return nil
}

// invoke call the actual function and send the input to it. It captures
// and returns the output
func (f *FnPlugin) invoke(input []byte) ([]byte, error) {
	// setup input and output
	var output []byte
	return output, nil
}
