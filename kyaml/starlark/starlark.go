// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package starlark

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/qri-io/starlib/util"
	"go.starlark.net/starlark"
	"sigs.k8s.io/kustomize/kyaml/comments"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Filter transforms a set of resources through the provided program
type Filter struct {
	Name string

	// Program is a starlark script which will be run against the resources
	Program string

	// URL is the url of a starlark program to fetch and run
	URL string

	// Path is the path to a starlark program to read and run
	Path string

	// FunctionConfig is the value to be provided for resourceList.functionConfig as specified by
	// https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md.
	FunctionConfig *yaml.RNode
}

func (sf *Filter) Filter(input []*yaml.RNode) ([]*yaml.RNode, error) {
	if sf.URL != "" && sf.Path != "" ||
		sf.URL != "" && sf.Program != "" ||
		sf.Path != "" && sf.Program != "" {
		return nil, errors.Errorf("Filter Path, Program and URL are mutually exclusive")
	}

	// read the program from a file
	if sf.Path != "" {
		b, err := ioutil.ReadFile(sf.Path)
		if err != nil {
			return nil, err
		}
		sf.Program = string(b)
	}

	// read the program from a URL
	if sf.URL != "" {
		err := func() error {
			resp, err := http.Get(sf.URL)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			sf.Program = string(b)
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}

	// retain map of inputs to outputs by id so if the name is changed by the
	// starlark program, we are able to match the same resources
	value, ids, err := sf.inputToResourceList(input)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	// run the starlark as program as transformation function
	thread := &starlark.Thread{Name: sf.Name}
	predeclared := starlark.StringDict{"resourceList": value}
	_, err = starlark.ExecFile(thread, sf.Name, sf.Program, predeclared)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	results, err := sf.resourceListToOutput(value, ids)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	// starlark will serialize the resources sorting the fields alphabetically,
	// format them to have a better ordering
	return filters.FormatFilter{}.Filter(results)
}

// tuple maps an input resource to the output resource
type tuple struct {
	// in is the RNode provided to the starlark program
	in *yaml.RNode
	// out is the RNode emitted by the starlark program with the id matching in
	out *yaml.RNode
}

// inputToResourceList transforms input into a starlark.Value
func (sf *Filter) inputToResourceList(
	input []*yaml.RNode) (starlark.Value, map[int]*tuple, error) {
	var id int
	ids := map[int]*tuple{}

	// convert into a ResourceList which will be converted to a starlark dictionary
	// create the ResourceList
	resourceList, err := yaml.Parse(`kind: ResourceList`)
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}

	// set the functionConfig
	if sf.FunctionConfig != nil {
		if err := resourceList.PipeE(
			yaml.FieldSetter{Name: "functionConfig", Value: sf.FunctionConfig}); err != nil {
			return nil, nil, err
		}
	}

	// the inputs should be provided as the list "items"
	items, err := resourceList.Pipe(yaml.LookupCreate(yaml.SequenceNode, "items"))
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}
	// add the input as items, give each resource an id
	for i := range input {
		item := input[i]

		// create an id for tracking the resource through the program
		err := item.PipeE(yaml.SetAnnotation("config.k8s.io/id", fmt.Sprintf("%d", id)))
		if err != nil {
			return nil, nil, errors.Wrap(err)
		}
		ids[id] = &tuple{in: item}
		id++

		items.YNode().Content = append(items.YNode().Content, item.YNode())
	}

	// convert the ResourceList into a starlark dictionary by
	// first converting it into a map[string]interface{}
	s, err := resourceList.String()
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}
	var in map[string]interface{}
	if err := yaml.Unmarshal([]byte(s), &in); err != nil {
		return nil, nil, errors.Wrap(err)
	}
	value, err := util.Marshal(in)
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}
	return value, ids, err
}

// resourceListToOutput converts the output of the starlark program to the filter output
func (sf *Filter) resourceListToOutput(
	value starlark.Value, ids map[int]*tuple) ([]*yaml.RNode, error) {
	// convert the modified resourceList back into a slice of RNodes
	// by first converting to a map[string]interface{}
	out, err := util.Unmarshal(value)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	o := out.(map[string]interface{})

	// parse the function config
	if _, found := o["functionConfig"]; found {
		fc := (o["functionConfig"].(map[string]interface{}))
		b, err := yaml.Marshal(fc)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		sf.FunctionConfig, err = yaml.Parse(string(b))
		if err != nil {
			return nil, errors.Wrap(err)
		}
	}

	// parse the items
	// copy the items out of the ResourceList, and into the Filter output
	var results []*yaml.RNode
	it := (o["items"].([]interface{}))
	for i := range it {
		// convert the resource back to the native yaml form
		b, err := yaml.Marshal(it[i])
		if err != nil {
			return nil, errors.Wrap(err)
		}
		node, err := yaml.Parse(string(b))
		if err != nil {
			return nil, errors.Wrap(err)
		}

		// match it to an input
		idS, err := node.Pipe(yaml.GetAnnotation("config.k8s.io/id"))
		if err != nil {
			return nil, errors.Wrap(err)
		}
		if idS == nil {
			// no matching input -- new resource
			results = append(results, node)
			continue
		}

		id, err := strconv.Atoi(idS.YNode().Value)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		if match, found := ids[id]; found {
			// matching resources
			match.out = node
		} else {
			// no matching input with the same id -- new resource
			// this may be an error case, the outputs probably shouldn't have ids
			// assigned by the starlark program
			results = append(results, node)
		}
	}

	// retain the comments instead of dropping them by copying them from the original inputs
	for i := 0; i < len(ids); i++ {
		v := ids[i]
		if v.out == nil {
			continue
		}
		if err := comments.CopyComments(v.in, v.out); err != nil {
			return nil, errors.Wrap(err)
		}
		results = append(results, v.out)
	}

	// delete the ids from resources, these were only to track through the starlark program
	// and that is finished.
	for i := range results {
		err := results[i].PipeE(yaml.ClearAnnotation("config.k8s.io/id"))
		if err != nil {
			return nil, err
		}
	}
	return results, nil
}
