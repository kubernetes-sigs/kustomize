// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kustfile

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"reflect"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/comments"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/kyaml/order"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var fieldMarshallingOrder = determineFieldOrder()

// determineFieldOrder returns a slice of Kustomization field
// names in the preferred order for serialization to a file.
// The field list is checked against the actual struct type
// to confirm that all fields are present, and no unknown
// fields are specified. Deprecated fields are removed from
// the list, meaning they will drop to the bottom on output
// (if present). The ordering and/or deprecation of fields
// in nested structs is not determined or considered.
func determineFieldOrder() []string {
	m := make(map[string]bool)
	s := reflect.ValueOf(&types.Kustomization{}).Elem()
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		m[typeOfT.Field(i).Name] = false
	}

	ordered := []string{
		"MetaData",
		"Resources",
		"Bases",
		"NamePrefix",
		"NameSuffix",
		"Namespace",
		"Crds",
		"CommonLabels",
		"Labels",
		"CommonAnnotations",
		"PatchesStrategicMerge",
		"PatchesJson6902",
		"Patches",
		"ConfigMapGenerator",
		"SecretGenerator",
		"HelmCharts",
		"HelmChartInflationGenerator",
		"HelmGlobals",
		"GeneratorOptions",
		"Vars",
		"Images",
		"Replacements",
		"Replicas",
		"Configurations",
		"Generators",
		"Transformers",
		"Inventory",
		"Components",
		"OpenAPI",
		"BuildMetadata",
	}

	// Add deprecated fields here.
	deprecated := map[string]bool{}

	// Account for the inlined TypeMeta fields.
	var result []string
	result = append(result, "APIVersion", "Kind")
	m["TypeMeta"] = true

	// Make sure all these fields are recognized.
	for _, n := range ordered {
		if _, ok := m[n]; ok {
			m[n] = true
		} else {
			log.Fatalf("%s is not a recognized field.", n)
		}
		// Keep if not deprecated.
		if _, f := deprecated[n]; !f {
			result = append(result, n)
		}
	}
	return result
}

type kustomizationFile struct {
	path          string
	fSys          filesys.FileSystem
	originalRNode *yaml.RNode
}

// NewKustomizationFile returns a new instance.
func NewKustomizationFile(fSys filesys.FileSystem) (*kustomizationFile, error) {
	mf := &kustomizationFile{fSys: fSys}
	err := mf.validate()
	if err != nil {
		return nil, err
	}
	return mf, nil
}

func (mf *kustomizationFile) GetPath() string {
	if mf == nil {
		return ""
	}
	return mf.path
}

func (mf *kustomizationFile) validate() error {
	match := 0
	var path []string
	for _, kfilename := range konfig.RecognizedKustomizationFileNames() {
		if mf.fSys.Exists(kfilename) {
			match += 1
			path = append(path, kfilename)
		}
	}

	switch match {
	case 0:
		return fmt.Errorf(
			"Missing kustomization file '%s'.\n",
			konfig.DefaultKustomizationFileName())
	case 1:
		mf.path = path[0]
	default:
		return fmt.Errorf("Found multiple kustomization file: %v\n", path)
	}

	if mf.fSys.IsDir(mf.path) {
		return fmt.Errorf("%s should be a file", mf.path)
	}
	return nil
}

func (mf *kustomizationFile) Read() (*types.Kustomization, error) {
	data, err := mf.fSys.ReadFile(mf.path)
	if err != nil {
		return nil, err
	}

	var k types.Kustomization
	if err := k.Unmarshal(data); err != nil {
		return nil, err
	}

	k.FixKustomization()

	if err := mf.parseCommentedFields(data); err != nil {
		return nil, err
	}
	return &k, nil
}

func (mf *kustomizationFile) Write(kustomization *types.Kustomization) error {
	if kustomization == nil {
		return errors.New("util: kustomization file arg is nil")
	}
	data, err := mf.marshal(kustomization)
	if err != nil {
		return err
	}
	return mf.fSys.WriteFile(mf.path, data)
}

// StringInSlice returns true if the string is in the slice.
func StringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func (mf *kustomizationFile) parseCommentedFields(content []byte) error {
	var nodes *yaml.RNode

	nodes, err := yaml.Parse(string(content))
	if err != nil {
		return err
	}
	mf.originalRNode = nodes
	return nil
}

// marshal converts a kustomization to a byte stream.
func (mf *kustomizationFile) marshal(kustomization *types.Kustomization) ([]byte, error) {
	var buffer []byte
	var output string
	var toRNode *yaml.RNode

	for _, field := range fieldMarshallingOrder {
		if mf.hasField(field) {
			continue
		}
		content, err := marshalField(field, kustomization)
		if err != nil {
			return content, nil
		}
		buffer = append(buffer, content...)
	}
	toRNode, err := yaml.Parse(string(buffer))
	if err != nil {
		return nil, err
	}
	comments.CopyComments(mf.originalRNode, toRNode)
	order.SyncOrder(mf.originalRNode, toRNode)

	output, err = toRNode.String()
	if err != nil {
		return nil, err
	}

	return []byte(output), nil
}

/*
isCommentOrBlankLine determines if a line is a comment or blank line
Return true for following lines
# This line is a comment

	# This line is also a comment with several leading white spaces

(The line above is a blank line)
*/
func isCommentOrBlankLine(line []byte) bool {
	s := bytes.TrimRight(bytes.TrimLeft(line, " "), "\n")
	return len(s) == 0 || bytes.HasPrefix(s, []byte(`#`))
}

func (mf *kustomizationFile) hasField(name string) bool {
	if mf.originalRNode == nil {
		return false
	}
	if mf.originalRNode.Field(name) != nil {
		return true
	}
	return false
}

// marshalField marshal a given field of a kustomization object into yaml format.
// If the field wasn't in the original kustomization.yaml file or wasn't added,
// an empty []byte is returned.
func marshalField(field string, kustomization *types.Kustomization) ([]byte, error) {
	r := reflect.ValueOf(*kustomization)
	titleCaser := cases.Title(language.English, cases.NoLower)
	v := r.FieldByName(titleCaser.String(field))

	if !v.IsValid() || isEmpty(v) {
		return []byte{}, nil
	}

	k := &types.Kustomization{}
	kr := reflect.ValueOf(k)
	kv := kr.Elem().FieldByName(titleCaser.String(field))
	kv.Set(v)

	return yaml.Marshal(k)
}

func isEmpty(v reflect.Value) bool {
	// If v is a pointer type
	if v.Type().Kind() == reflect.Ptr {
		return v.IsNil()
	}
	return v.Len() == 0
}
