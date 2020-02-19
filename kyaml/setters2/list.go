// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"sort"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// List lists the setters specified in the OpenAPI
type List struct {
	Name string

	Setters []SetterDefinition
}

// List initializes l.Setters with the setters from the OpenAPI definitions in the file
func (l *List) List(openAPIPath, resourcePath string) error {
	if err := openapi.AddSchemaFromFile(openAPIPath); err != nil {
		return err
	}
	y, err := yaml.ReadFile(openAPIPath)
	if err != nil {
		return err
	}
	return l.list(y, resourcePath)
}

func (l *List) list(object *yaml.RNode, resourcePath string) error {
	// read the OpenAPI definitions
	def, err := object.Pipe(yaml.LookupCreate(yaml.MappingNode, "openAPI", "definitions"))
	if err != nil {
		return err
	}
	if yaml.IsEmpty(def) {
		return nil
	}

	// iterate over definitions -- find those that are setters
	err = def.VisitFields(func(node *yaml.MapNode) error {
		setter := SetterDefinition{}

		// the definition key -- contains the setter name
		key := node.Key.YNode().Value

		if !strings.HasPrefix(key, SetterDefinitionPrefix) {
			// not a setter -- doesn't have the right prefix
			return nil
		}

		setterNode, err := node.Value.Pipe(yaml.Lookup(K8sCliExtensionKey, "setter"))
		if err != nil {
			return err
		}
		if yaml.IsEmpty(setterNode) {
			// has the setter prefix, but missing the setter extension
			return errors.Errorf("missing x-k8s-cli.setter for %s", key)
		}

		// unmarshal the yaml for the setter extension into the definition struct
		b, err := setterNode.String()
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal([]byte(b), &setter); err != nil {
			return err
		}

		if l.Name != "" && l.Name != setter.Name {
			// not the setter that was requested by list
			return nil
		}

		// the description is not part of the extension, and should be pulled out
		// separately from the extension values.
		description := node.Value.Field("description")
		if description != nil {
			setter.Description = description.Value.YNode().Value
		}

		// count the number of fields set by this setter
		setter.Count, err = l.count(resourcePath, setter.Name)
		if err != nil {
			return err
		}

		l.Setters = append(l.Setters, setter)
		return nil
	})
	if err != nil {
		return err
	}

	// sort the setters by their name
	sort.Slice(l.Setters, func(i, j int) bool {
		return l.Setters[i].Name < l.Setters[j].Name
	})

	return nil
}

// count returns the number of fields set by the setter with name
func (l *List) count(path, name string) (int, error) {
	s := &Set{Name: name}
	err := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.LocalPackageReader{PackagePath: path}},
		Filters: []kio.Filter{kio.FilterAll(s)},
	}.Execute()

	return s.Count, err
}
