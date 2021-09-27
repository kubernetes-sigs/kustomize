// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fix

import (
	"bytes"
	"fmt"
	"path"
	"strconv"
	"strings"

	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/resid"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/yaml"
)

func ConvertVarsToReplacements(fSys filesys.FileSystem, k *types.Kustomization) error {
	if k.Vars == nil {
		return nil
	}

	k.Resources = append(k.Resources, k.Bases...)
	k.Replacements = []types.ReplacementField{}

	files, err := filesTouchedByKustomize(k, "", fSys)
	if err != nil {
		return err
	}

	for _, v := range k.Vars {
		repl := &types.Replacement{}
		if err := addTargets(repl, v.Name, files, fSys); err != nil {
			return err
		}
		copySourceFromVars(repl, v)
		if err := setPlaceholderValue(v.Name, files, fSys); err != nil {
			return err
		}
		k.Replacements = append(k.Replacements, types.ReplacementField{Replacement: *repl})
	}
	k.Vars = nil
	return nil
}

var patchTarget = make(map[string]types.Patch)

func filesTouchedByKustomize(k *types.Kustomization, filepath string, fSys filesys.FileSystem) ([]string, error) {
	var result []string
	for _, r := range k.Resources {
		// first, try to read resource as a base/directory
		files, err := fSys.ReadDir(r)
		if err == nil && len(files) > 0 {
			for _, file := range files {
				if !stringInSlice(file, []string{
					"kustomization.yaml",
					"kustomization.yml",
					"Kustomization",
				}) {
					continue
				}

				b, err := fSys.ReadFile(path.Join(r, file))
				if err != nil {
					continue
				}

				subKt := &types.Kustomization{}
				if err := yaml.Unmarshal(b, subKt); err != nil {
					return nil, err
				}
				paths, err := filesTouchedByKustomize(subKt, r, fSys)
				if err != nil {
					return nil, err
				}
				result = append(result, paths...)
			}

		}
		// read the resource as a file
		result = append(result, path.Join(filepath, r))

	}

	// aggregate all of the paths from the `patches` field
	for _, p := range k.Patches {
		if p.Path != "" {
			patchPath := path.Join(filepath, p.Path)
			result = append(result, patchPath)
			patchTarget[patchPath] = p
		}
	}
	return result, nil
}

func copySourceFromVars(repl *types.Replacement, v types.Var) {
	repl.Source = &types.SourceSelector{}
	apiVersion := v.ObjRef.APIVersion
	group, version := resid.ParseGroupVersion(apiVersion)
	repl.Source.Gvk.Group = group
	repl.Source.Gvk.Version = version
	repl.Source.Gvk.Kind = v.ObjRef.Kind
	repl.Source.Name = v.ObjRef.Name
	repl.Source.Namespace = v.ObjRef.Namespace
	repl.Source.FieldPath = v.FieldRef.FieldPath
}

func addTargets(repl *types.Replacement, varName string, files []string, fSys filesys.FileSystem) error {
	for _, file := range files {
		nodes, err := getNodesFromFile(file, fSys)
		if err != nil {
			continue
		}
		for _, n := range nodes {
			fieldPaths, options, err := findVarName(n, varName, []string{})
			if err != nil {
				return fmt.Errorf("error with %s: %s", file, err.Error())
			}
			targets, err := constructTargets(file, n, fieldPaths, options)
			if err != nil {
				return err
			}
			repl.Targets = append(repl.Targets, targets...)
		}
	}
	return nil
}

func getNodesFromFile(fileName string, fSys filesys.FileSystem) ([]*kyaml.RNode, error) {
	b, err := fSys.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	out := &bytes.Buffer{}
	r := kio.ByteReadWriter{
		Reader:                bytes.NewBufferString(string(b)),
		Writer:                out,
		KeepReaderAnnotations: true,
		OmitReaderAnnotations: true,
	}
	return r.Read()
}

func findVarName(node *kyaml.RNode, varName string, path []string) ([]string, []*types.FieldOptions, error) {
	var fieldPaths []string
	var options []*types.FieldOptions

	switch node.YNode().Kind {

	case kyaml.SequenceNode:
		elements, err := node.Elements()
		if err != nil {
			return nil, nil, err
		}
		for i := range elements {
			nextPathItem := strings.TrimSpace(strconv.Itoa(i))
			fieldPathsToAdd, optionsToAdd, err := findVarName(elements[i],
				varName, append(path, nextPathItem))
			if err != nil {
				return nil, nil, err
			}
			fieldPaths = append(fieldPaths, fieldPathsToAdd...)
			options = append(options, optionsToAdd...)
		}

	case kyaml.MappingNode:
		err := node.VisitFields(func(n *kyaml.MapNode) error {
			nextPathItem := strings.TrimSpace(n.Key.MustString())
			if strings.Contains(nextPathItem, ".") {
				nextPathItem = fmt.Sprintf("[%s]", nextPathItem)
			}
			fieldPathsToAdd, optionsToAdd, err := findVarName(n.Value.Copy(),
				varName, append(path, nextPathItem))
			if err != nil {
				return err
			}
			fieldPaths = append(fieldPaths, fieldPathsToAdd...)
			options = append(options, optionsToAdd...)
			return nil
		})
		if err != nil {
			return nil, nil, err
		}

	case kyaml.ScalarNode:
		value := node.YNode().Value
		varString := fmt.Sprintf("$(%s)", varName)
		if strings.Contains(value, varString) {
			fieldPaths = append(fieldPaths, strings.Join(path, "."))
			optionsToAdd, err := constructFieldOptions(value, varString)
			if err != nil {
				return nil, nil, err
			}
			options = append(options, optionsToAdd...)
		}
	}

	return fieldPaths, options, nil
}

func constructFieldOptions(value string, varString string) ([]*types.FieldOptions, error) {
	if value == varString {
		return []*types.FieldOptions{{}}, nil
	}

	var delimiter string
	var index int
	i := strings.Index(value, varString)

	// all array accesses here are safe because we know value != varString and
	// that value contains varString, so len(value) > len(varString)
	switch {
	case i == 0: // prefix
		delimiter = string(value[len(varString)])
		index = 0
	case (i + len(varString)) >= len(value): // suffix
		delimiter = string(value[i-1])
		index = len(strings.Split(value, delimiter)) - 1
	default: // in the middle somewhere
		pre := string(value[i-1])
		post := string(value[i+len(varString)])
		if pre != post {
			return nil, fmt.Errorf("cannot convert all vars to replacements; %s is not delimited", varString)
		}
		delimiter = pre
		index = indexOf(varString, strings.Split(value, delimiter))
		if index == -1 {
			// this should never happen
			return nil, fmt.Errorf("internal error: could not get index of var %s", varString)
		}
	}
	return []*types.FieldOptions{{Delimiter: delimiter, Index: index}}, nil
}

func constructTargets(file string, node *kyaml.RNode, fieldPaths []string,
	options []*types.FieldOptions) ([]*types.TargetSelector, error) {

	if len(fieldPaths) != len(options) {
		// this should never happen
		return nil, fmt.Errorf("internal error: length of fieldPaths != length of fieldOptions")
	}

	if patch, ok := patchTarget[file]; ok {
		if !patch.Options["allowNameChange"] || !patch.Options["allowKindChange"] {
			return writePatchTargets(patch, node, fieldPaths, options)
		}
	}

	var result []*types.TargetSelector
	meta, metaErr := node.GetMeta()

	for i := range fieldPaths {
		target := &types.TargetSelector{
			Select: &types.Selector{
				ResId: resid.ResId{
					Name:      node.GetName(),
					Namespace: node.GetNamespace(),
					Gvk: resid.Gvk{
						Kind: node.GetKind(),
					},
				},
			},
			FieldPaths: []string{fieldPaths[i]},
		}
		if options[i].String() != "" {
			target.Options = options[i]
		}
		if metaErr == nil {
			if meta.TypeMeta.APIVersion != "" {
				group, version := resid.ParseGroupVersion(meta.TypeMeta.APIVersion)
				target.Select.ResId.Gvk.Group = group
				target.Select.ResId.Gvk.Version = version
			}
		}

		result = append(result, target)
	}

	return result, nil
}

// if the var appears in a patch, this must be handled differently than a regular
// resource because a patch may be applied to multiple resources and the resulting
// resources may have different IDs than the patch
func writePatchTargets(patch types.Patch, node *kyaml.RNode, fieldPaths []string,
	options []*types.FieldOptions) ([]*types.TargetSelector, error) {

	var result []*types.TargetSelector
	selector := patch.Target.Copy()

	for i := range fieldPaths {
		target := &types.TargetSelector{
			Select:     &selector,
			FieldPaths: []string{fieldPaths[i]},
		}
		if options[i].String() != "" {
			target.Options = options[i]
		}
		if patch.Options["allowNameChange"] {
			target.Select.ResId.Name = node.GetName()
		}
		if patch.Options["allowKindChange"] {
			target.Select.ResId.Kind = node.GetKind()
		}
		if node.GetNamespace() != "" {
			target.Select.ResId.Namespace = node.GetNamespace()
		}
		result = append(result, target)
	}
	return result, nil
}

func setPlaceholderValue(varName string, files []string, fSys filesys.FileSystem) error {
	for _, filename := range files {
		b, err := fSys.ReadFile(filename)
		if err != nil {
			continue
		}
		newFileContents := strings.ReplaceAll(string(b), fmt.Sprintf("$(%s)", varName),
			fmt.Sprintf("%s_PLACEHOLDER", varName))
		err = fSys.WriteFile(filename, []byte(newFileContents))
		if err != nil {
			return err
		}
	}
	return nil
}

func stringInSlice(elem string, slice []string) bool {
	for i := range slice {
		if slice[i] == elem {
			return true
		}
	}
	return false
}

func indexOf(varName string, slice []string) int {
	for i := range slice {
		if slice[i] == varName {
			return i
		}
	}
	return -1
}
