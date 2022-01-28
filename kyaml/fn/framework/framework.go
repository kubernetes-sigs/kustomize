// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"bytes"
	goerrors "errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sort"
	"strconv"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// ResourceList is a Kubernetes list type used as the primary data interchange format
// in the Configuration Functions Specification:
// https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md
// This framework facilitates building functions that receive and emit ResourceLists,
// as required by the specification.
type ResourceList struct {
	// Items is the ResourceList.items input and output value.
	//
	// e.g. given the function input:
	//
	//    kind: ResourceList
	//    items:
	//    - kind: Deployment
	//      ...
	//    - kind: Service
	//      ...
	//
	// Items will be a slice containing the Deployment and Service resources
	// Mutating functions will alter this field during processing.
	// This field is required.
	Items []*yaml.RNode `yaml:"items" json:"items"`

	// FunctionConfig is the ResourceList.functionConfig input value.
	//
	// e.g. given the input:
	//
	//    kind: ResourceList
	//    functionConfig:
	//      kind: Example
	//      spec:
	//        foo: var
	//
	// FunctionConfig will contain the RNodes for the Example:
	//      kind: Example
	//      spec:
	//        foo: var
	FunctionConfig *yaml.RNode `yaml:"functionConfig,omitempty" json:"functionConfig,omitempty"`

	// Results is ResourceList.results output value.
	// Validating functions can optionally use this field to communicate structured
	// validation error data to downstream functions.
	Results Results `yaml:"results,omitempty" json:"results,omitempty"`
}

// ParseResourceList parses a ResourceList from the input byte array.
func ParseResourceList(in []byte) (*ResourceList, error) {
	rl := &ResourceList{
		Items: []*yaml.RNode{},
	}

	var nodes []*yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(in))
	for {
		node := &yaml.Node{}
		if err := decoder.Decode(node); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		nodes = append(nodes, node)
	}

	if len(nodes) != 1 {
		return nil, fmt.Errorf("expected exactly one resouceList object, got %d", len(nodes))
	}
	rlRNode := yaml.NewRNode(nodes[0])

	if rlRNode.GetKind() != kio.ResourceListKind {
		return nil, fmt.Errorf("input was of unexpected kind %q; expected ResourceList", rlRNode.GetKind())
	}
	fc := &yaml.RNode{}
	found, err := rlRNode.Get(fc, "functionConfig")
	if err != nil {
		return nil, fmt.Errorf("failed when tried to get functionConfig: %w", err)
	}
	if found {
		rl.FunctionConfig = fc
	}

	itemsRN := &yaml.RNode{}
	found, err = rlRNode.Get(itemsRN, "items")
	if err != nil {
		return nil, fmt.Errorf("failed when tried to get items: %w", err)
	}
	if !found {
		return rl, nil
	}
	itemYNodes := itemsRN.Content()
	var items []*yaml.RNode
	for i := range itemYNodes {
		items = append(items, yaml.NewRNode(itemYNodes[i]))
	}
	rl.Items = items
	return rl, nil
}

// ToYAML convert the ResourceList to its yaml representation.
func (rl *ResourceList) ToYAML() (string, error) {
	rl.SortItems()
	var yml string
	if err := func() error {
		ko, err := rl.toRNode()
		if err != nil {
			return err
		}
		yml, err = ko.String()
		return err
	}(); err != nil {
		return "", fmt.Errorf("failed to convert the resourceList to yaml: %v", err)
	}
	return yml, nil
}

// toRNode converts the ResourceList to a RNode.
func (rl *ResourceList) toRNode() (*yaml.RNode, error) {
	obj, err := yaml.NewEmptyRNode()
	if err != nil {
		return nil, err
	}
	obj.SetApiVersion(kio.ResourceListAPIVersion)
	obj.SetKind(kio.ResourceListKind)

	if rl.Items != nil && len(rl.Items) > 0 {
		if err := obj.Set(rl.Items, "items"); err != nil {
			return nil, err
		}
	}

	if rl.FunctionConfig != nil {
		if err := obj.Set(rl.FunctionConfig, "functionConfig"); err != nil {
			return nil, err
		}
	}

	if rl.Results != nil && len(rl.Results) > 0 {
		if err = obj.Set(rl.Results, "results"); err != nil {
			return nil, err
		}
	}

	return obj, nil
}

// UpsertObjectToItems adds an object to ResourceList.items. The input object can
// be a RNode or any typed object (e.g. corev1.Pod).
func (rl *ResourceList) UpsertObjectToItems(obj interface{}, checkExistence func(obj, another *yaml.RNode) bool, replaceIfAlreadyExist bool) error {
	if checkExistence == nil {
		checkExistence = func(obj, another *yaml.RNode) bool {
			ri1 := obj.GetResourceIdentifier()
			ri2 := another.GetResourceIdentifier()
			return reflect.DeepEqual(ri1, ri2)
		}
	}

	var ko *yaml.RNode
	switch obj := obj.(type) {
	case yaml.RNode:
		ko = &obj
	case *yaml.RNode:
		ko = obj
	case yaml.Node:
		ko = yaml.NewRNode(&obj)
	case *yaml.Node:
		ko = yaml.NewRNode(obj)
	default:
		var err error
		ko, err = yaml.NewRNodeFrom(obj)
		if err != nil {
			return err
		}
	}

	idx := -1
	for i, item := range rl.Items {
		if checkExistence(ko, item) {
			idx = i
			break
		}
	}
	if idx == -1 {
		rl.Items = append(rl.Items, ko)
	} else if replaceIfAlreadyExist {
		rl.Items[idx] = ko
	}
	return nil
}

// ResourceListProcessor is implemented by configuration functions built with this framework
// to conform to the Configuration Functions Specification:
// https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md
// To invoke a processor, pass it to framework.Execute, which will also handle ResourceList IO.
//
// This framework provides several ready-to-use ResourceListProcessors, including
// SimpleProcessor, VersionedAPIProcessor and TemplateProcessor.
// You can also build your own by implementing this interface.
type ResourceListProcessor interface {
	Process(rl *ResourceList) error
}

// ResourceListProcessorFunc converts a compatible function to a ResourceListProcessor.
type ResourceListProcessorFunc func(rl *ResourceList) error

// Process makes ResourceListProcessorFunc implement the ResourceListProcessor interface.
func (p ResourceListProcessorFunc) Process(rl *ResourceList) error {
	return p(rl)
}

// Defaulter is implemented by APIs to have Default invoked.
// The standard application is to create a type to hold your FunctionConfig data, and
// implement Defaulter on that type. All of the framework's processors will invoke Default()
// on your type after unmarshalling the FunctionConfig data into it.
type Defaulter interface {
	Default() error
}

// Validator is implemented by APIs to have Validate invoked.
// The standard application is to create a type to hold your FunctionConfig data, and
// implement Validator on that type. All of the framework's processors will invoke Validate()
// on your type after unmarshalling the FunctionConfig data into it.
type Validator interface {
	Validate() error
}

// Execute is the entrypoint for invoking configuration functions built with this framework
// from code. See framework/command#Build for a Cobra-based command-line equivalent.
// Execute reads a ResourceList from the given source, passes it to a ResourceListProcessor,
// and then writes the result to the target.
// STDIN and STDOUT will be used if no reader or writer respectively is provided.
func Execute(p ResourceListProcessor, rlSource *kio.ByteReadWriter) error {
	// Prepare the resource list source
	if rlSource == nil {
		rlSource = &kio.ByteReadWriter{KeepReaderAnnotations: true}
	}
	if rlSource.Reader == nil {
		rlSource.Reader = os.Stdin
	}
	if rlSource.Writer == nil {
		rlSource.Writer = os.Stdout
	}

	// Read the input
	rl := ResourceList{}
	var err error
	if rl.Items, err = rlSource.Read(); err != nil {
		return errors.WrapPrefixf(err, "failed to read ResourceList input")
	}
	rl.FunctionConfig = rlSource.FunctionConfig

	// We store the original
	nodeAnnos, err := kio.PreprocessResourcesForInternalAnnotationMigration(rl.Items)
	if err != nil {
		return err
	}

	retErr := p.Process(&rl)

	// If either the internal annotations for path, index, and id OR the legacy
	// annotations for path, index, and id are changed, we have to update the other.
	err = kio.ReconcileInternalAnnotations(rl.Items, nodeAnnos)
	if err != nil {
		return err
	}

	// Write the results
	// Set the ResourceList.results for validating functions
	if len(rl.Results) > 0 {
		b, err := yaml.Marshal(rl.Results)
		if err != nil {
			return errors.Wrap(err)
		}
		y, err := yaml.Parse(string(b))
		if err != nil {
			return errors.Wrap(err)
		}
		rlSource.Results = y
	}
	if err := rlSource.Write(rl.Items); err != nil {
		return err
	}

	return retErr
}

// Filter executes the given kio.Filter and replaces the ResourceList's items with the result.
// This can be used to help implement ResourceListProcessors. See SimpleProcessor for example.
//
// Filters that return a Result as error will store the result in the ResourceList
// and continue processing instead of erroring out.
func (rl *ResourceList) Filter(api kio.Filter) error {
	var err error
	rl.Items, err = api.Filter(rl.Items)
	if err != nil {
		var r Results
		if goerrors.As(err, &r) {
			rl.Results = append(rl.Results, r...)
			return nil
		}
		return errors.Wrap(err)
	}
	return nil
}

// SortItems sorts the ResourceList.items by apiVersion, kind, namespace and name.
func (rl *ResourceList) SortItems() {
	sort.Sort(sortRNodeObjects(rl.Items))
}

type sortRNodeObjects []*yaml.RNode

func (o sortRNodeObjects) Len() int      { return len(o) }
func (o sortRNodeObjects) Swap(i, j int) { o[i], o[j] = o[j], o[i] }
func (o sortRNodeObjects) Less(i, j int) bool {
	idi := o[i].GetResourceIdentifier()
	idj := o[j].GetResourceIdentifier()
	idStrI := fmt.Sprintf("%s %s %s %s", idi.GetAPIVersion(), idi.GetKind(), idi.GetNamespace(), idi.GetName())
	idStrJ := fmt.Sprintf("%s %s %s %s", idj.GetAPIVersion(), idj.GetKind(), idj.GetNamespace(), idj.GetName())
	return idStrI < idStrJ
}

// Run evaluates the function. input must be a resourceList in yaml format. An
// updated resourceList will be returned.
func Run(p ResourceListProcessor, input []byte) ([]byte, error) {
	rl, err := ParseResourceList(input)
	if err != nil {
		return nil, err
	}
	err = p.Process(rl)
	if err != nil {
		// If the error is not a Results type, we wrap the error as a Result.
		if results, ok := err.(Results); ok {
			rl.Results = append(rl.Results, results...)
		} else if result, ok := err.(Result); ok {
			rl.Results = append(rl.Results, &result)
		} else if result, ok := err.(*Result); ok {
			rl.Results = append(rl.Results, result)
		} else {
			rl.Results = append(rl.Results, ErrorResult(err))
		}
	}
	yml, er := rl.ToYAML()
	if er != nil {
		return []byte(yml), er
	}
	if len(rl.Results) > 0 {
		return []byte(yml), rl.Results
	}
	return []byte(yml), nil
}

type ErrMissingFnConfig struct{}

func (ErrMissingFnConfig) Error() string {
	return "unable to find the functionConfig in the resourceList"
}

// Chain chains a list of ResourceListProcessor as a single ResourceListProcessor.
func Chain(processors ...ResourceListProcessor) ResourceListProcessor {
	return ResourceListProcessorFunc(func(rl *ResourceList) error {
		for _, processor := range processors {
			if err := processor.Process(rl); err != nil {
				return err
			}
		}
		return nil
	})
}

// ChainFunctions chains a list of ResourceListProcessorFunc as a single ResourceListProcessorFunc.
func ChainFunctions(fns ...ResourceListProcessorFunc) ResourceListProcessorFunc {
	return func(rl *ResourceList) error {
		for _, fn := range fns {
			if err := fn.Process(rl); err != nil {
				return err
			}
		}
		return nil
	}
}

// GetPathAnnotation checks the path annotation first and then the legacy path
// annotation. See: https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md#internalconfigkubernetesiopath
func GetPathAnnotation(rn *yaml.RNode) string {
	anno := rn.GetAnnotation(kioutil.PathAnnotation)
	if anno == "" {
		anno = rn.GetAnnotation(kioutil.LegacyPathAnnotation)
	}
	return anno
}

// GetIndexAnnotation checks the index annotation first and then the legacy index
// annotation. It returns -1 if not found. See: https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md#internalconfigkubernetesioindex
func GetIndexAnnotation(rn *yaml.RNode) int {
	anno := rn.GetAnnotation(kioutil.IndexAnnotation)
	if anno == "" {
		anno = rn.GetAnnotation(kioutil.LegacyIndexAnnotation)
	}

	if anno == "" {
		return -1
	}
	i, _ := strconv.Atoi(anno)
	return i
}

// GetIdAnnotation checks the id annotation first and then the legacy id annotation.
// It returns -1 if not found.
func GetIdAnnotation(rn *yaml.RNode) int {
	anno := rn.GetAnnotation(kioutil.IdAnnotation)
	if anno == "" {
		anno = rn.GetAnnotation(kioutil.LegacyIdAnnotation)
	}

	if anno == "" {
		return -1
	}
	i, _ := strconv.Atoi(anno)
	return i
}
