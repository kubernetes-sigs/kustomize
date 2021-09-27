// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package setters2

import (
	"sort"
	"strings"

	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// List lists the setters specified in the OpenAPI
// excludes the subpackages which contain file with
// name OpenAPIFileName in them
type List struct {
	Name string

	OpenAPIFileName string

	Setters []SetterDefinition

	Substitutions []SubstitutionDefinition

	SettersSchema *spec.Schema
}

// ListSetters initializes l.Setters with the setters from the OpenAPI definitions in the file
func (l *List) ListSetters(openAPIPath, resourcePath string) error {
	y, err := yaml.ReadFile(openAPIPath)
	if err != nil {
		return err
	}
	return l.listSetters(y, resourcePath)
}

// ListSubst initializes l.Substitutions with the substitutions from the OpenAPI definitions in the file
func (l *List) ListSubst(openAPIPath string) error {
	y, err := yaml.ReadFile(openAPIPath)
	if err != nil {
		return err
	}
	return l.listSubst(y)
}

func (l *List) listSetters(object *yaml.RNode, resourcePath string) error {
	// read the OpenAPI definitions
	def, err := object.Pipe(yaml.LookupCreate(yaml.MappingNode, "openAPI", "definitions"))
	if err != nil {
		return err
	}
	if yaml.IsMissingOrNull(def) {
		return nil
	}

	// iterate over definitions -- find those that are setters
	err = def.VisitFields(func(node *yaml.MapNode) error {
		setter := SetterDefinition{}

		// the definition key -- contains the setter name
		key := node.Key.YNode().Value

		if !strings.HasPrefix(key, fieldmeta.SetterDefinitionPrefix) {
			// not a setter -- doesn't have the right prefix
			return nil
		}

		setterNode, err := node.Value.Pipe(yaml.Lookup(K8sCliExtensionKey, "setter"))
		if err != nil {
			return err
		}
		if yaml.IsMissingOrNull(setterNode) {
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

func (l *List) listSubst(object *yaml.RNode) error {
	// read the OpenAPI definitions
	def, err := object.Pipe(yaml.LookupCreate(yaml.MappingNode, "openAPI", "definitions"))
	if err != nil {
		return err
	}
	if yaml.IsMissingOrNull(def) {
		return nil
	}

	// iterate over definitions -- find those that are substitutions
	err = def.VisitFields(func(node *yaml.MapNode) error {
		subst := SubstitutionDefinition{}

		// the definition key -- contains the substitution name
		key := node.Key.YNode().Value

		if !strings.HasPrefix(key, fieldmeta.SubstitutionDefinitionPrefix) {
			// not a substitution -- doesn't have the right prefix
			return nil
		}

		substNode, err := node.Value.Pipe(yaml.Lookup(K8sCliExtensionKey, "substitution"))
		if err != nil {
			return err
		}
		if yaml.IsMissingOrNull(substNode) {
			// has the substitution prefix, but missing the setter extension
			return errors.Errorf("missing x-k8s-cli.substitution for %s", key)
		}

		// unmarshal the yaml for the substitution extension into the definition struct
		b, err := substNode.String()
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal([]byte(b), &subst); err != nil {
			return err
		}

		if l.Name != "" && l.Name != subst.Name {
			// not the substitution that was requested by list
			return nil
		}

		l.Substitutions = append(l.Substitutions, subst)
		return nil
	})
	if err != nil {
		return err
	}

	// sort the substitutions by their name
	sort.Slice(l.Substitutions, func(i, j int) bool {
		return l.Substitutions[i].Name < l.Substitutions[j].Name
	})

	return nil
}

// count returns the number of fields set by the setter with name
// this excludes all the subpackages with openAPI file in them
// set filter is leveraged for this but the resources are not written
// back to files as only LocalPackageReader is invoked and not writer
func (l *List) count(path, name string) (int, error) {
	s := &Set{Name: name, SettersSchema: l.SettersSchema}
	err := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.LocalPackageReader{PackagePath: path, PackageFileName: l.OpenAPIFileName}},
		Filters: []kio.Filter{kio.FilterAll(s)},
	}.Execute()

	return s.Count, err
}
