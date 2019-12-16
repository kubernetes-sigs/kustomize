// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package generators

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"sigs.k8s.io/kustomize/functions/config.kubernetes.io/kustomize/fieldreference"
	"sigs.k8s.io/kustomize/kyaml/inpututil"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	GeneratedAnnotation = "kustomize.io/generated-resource/"
)

type ConfigMapGeneratorFilter struct {
	Root                   string
	GeneratorKustomizeName string

	Args []GeneratorArgs `yaml:"configMapGenerators,omitempty"`

	index int
}

// GeneratorArgs contains arguments common to ConfigMap and Secret generators.
type GeneratorArgs struct {
	// Name - actually the partial name - of the generated resource.
	// The full name ends up being something like
	// NamePrefix + this.Name + NameSuffix + hash(content of generated resource).
	Name string `yaml:"name,omitempty"`

	// KvPairSources for the generator.
	KvPairSources `yaml:",inline,omitempty"`

	done bool
}

// KvPairSources defines places to obtain key value pairs.
type KvPairSources struct {
	// literals is a list of literal
	// pair sources. Each literal source should
	// be a key and literal value, e.g. `key=value`
	LiteralSources []string `yaml:"literals,omitempty"`

	// files is a list of file "sources" to
	// use in creating a list of key, value pairs.
	// A source takes the form:  [{key}=]{path}
	// If the "key=" part is missing, the key is the
	// path's basename. If they "key=" part is present,
	// it becomes the key (replacing the basename).
	// In either case, the value is the file contents.
	// Specifying a directory will iterate each named
	// file in the directory whose basename is a
	// valid configmap key.
	FileSources []string `yaml:"files,omitempty"`

	// envs is a list of file paths.
	// The contents of each file should be one
	// key=value pair per line, e.g. a Docker
	// or npm ".env" file or a ".ini" file
	// (wikipedia.org/wiki/INI_file)
	EnvSources []string `yaml:"envs,omitempty"`
}

// Pair is a key value pair.
type Pair struct {
	Key   string
	Value string
}

func (knf *ConfigMapGeneratorFilter) Filter(input []*yaml.RNode) ([]*yaml.RNode, error) {
	output, err := inpututil.MapInputs(input, knf.mapFn)
	if err != nil {
		return nil, err
	}

	for i := range knf.Args {
		if knf.Args[i].done {
			continue
		}
		cm, err := knf.doConfigMap(nil, knf.Args[i])
		if err != nil {
			return nil, err
		}
		output = append(output, cm)
	}
	return output, nil
}

func (knf *ConfigMapGeneratorFilter) mapFn(node *yaml.RNode, meta yaml.ResourceMeta) (
	[]*yaml.RNode, error) {
	if meta.Annotations[GeneratedAnnotation+knf.GeneratorKustomizeName] != "ConfigMap" {
		// not a generated configmap
		return []*yaml.RNode{node}, nil
	}

	refName := meta.Annotations[fieldreference.ReferenceNameAnnotation+knf.GeneratorKustomizeName]

	// generated configmap find the matching GeneratorArgs
	var arg *GeneratorArgs
	for i := range knf.Args {
		if knf.Args[i].Name == refName {
			arg = &knf.Args[i]
			arg.done = true
			break
		}
	}
	if arg == nil {
		// generated configmap was deleted
		return []*yaml.RNode{}, nil
	}

	cm, err := knf.doConfigMap(node, *arg)
	if err != nil {
		return nil, err
	}

	// return the configmap
	return []*yaml.RNode{cm}, nil
}

func (knf *ConfigMapGeneratorFilter) doConfigMap(input *yaml.RNode, args GeneratorArgs) (*yaml.RNode, error) {
	// RNode to struct
	cm := &configMap{}
	if input != nil {
		b, err := yaml.Marshal(input)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(b, cm)
		if err != nil {
			return nil, err
		}
	}
	cm.APIVersion = configMapAPIVersion
	cm.Kind = configMapKind

	// add the data to the configmap
	pairs, err := knf.load(args.KvPairSources)
	if err != nil {
		return nil, err
	}
	for i := range pairs {
		if err := knf.addKvToConfigMap(cm, pairs[i]); err != nil {
			return nil, err
		}
	}
	h := sha256.New()
	for i := range pairs {
		if _, err := h.Write([]byte(pairs[i].Key)); err != nil {
			return nil, err
		}
		if _, err := h.Write([]byte(pairs[i].Value)); err != nil {
			return nil, err
		}
	}
	sum := fmt.Sprintf("%x", h.Sum([]byte{}))[0:10]
	cm.Name = args.Name + "-kust" + sum

	if cm.Annotations == nil {
		cm.Annotations = map[string]string{}
	}
	cm.Annotations[fieldreference.ReferenceNameAnnotation+knf.GeneratorKustomizeName] = args.Name
	cm.Annotations[fieldreference.OriginalNameAnnotation+knf.GeneratorKustomizeName] = cm.Name
	cm.Annotations[GeneratedAnnotation+knf.GeneratorKustomizeName] = "ConfigMap"
	cm.Annotations[kioutil.PathAnnotation] = filepath.Join(
		"generated-resources", "configmap.yaml")
	cm.Annotations[kioutil.IndexAnnotation] = fmt.Sprintf("%d", knf.index)
	knf.index = knf.index + 1

	// struct to RNode
	b, err := yaml.Marshal(cm)
	if err != nil {
		return nil, err
	}
	return yaml.Parse(string(b))
}

const keyExistsErrorMsg = "cannot add key %s, another key by that name already exists: %v"

// addKvToConfigMap adds the given key and data to the given config map.
// Error if key invalid, or already exists.
func (knf *ConfigMapGeneratorFilter) addKvToConfigMap(configMap *configMap, p Pair) error {
	// If the configmap data contains byte sequences that are all in the UTF-8
	// range, we will write it to .Data
	if utf8.Valid([]byte(p.Value)) {
		if _, entryExists := configMap.Data[p.Key]; entryExists {
			return fmt.Errorf(keyExistsErrorMsg, p.Key, configMap.Data)
		}
		if configMap.Data == nil {
			configMap.Data = map[string]string{}
		}
		configMap.Data[p.Key] = p.Value
		return nil
	}
	// otherwise, it's BinaryData
	if configMap.BinaryData == nil {
		configMap.BinaryData = map[string][]byte{}
	}
	if _, entryExists := configMap.BinaryData[p.Key]; entryExists {
		return fmt.Errorf(keyExistsErrorMsg, p.Key, configMap.BinaryData)
	}
	configMap.BinaryData[p.Key] = []byte(p.Value)
	return nil
}

func (knf *ConfigMapGeneratorFilter) load(args KvPairSources) (all []Pair, err error) {
	pairs, err := knf.keyValuesFromEnvFiles(args.EnvSources)
	if err != nil {
		return nil, err
	}
	all = append(all, pairs...)

	pairs, err = knf.keyValuesFromLiteralSources(args.LiteralSources)
	if err != nil {
		return nil, err
	}
	all = append(all, pairs...)

	pairs, err = knf.keyValuesFromFileSources(args.FileSources)
	if err != nil {
		return nil, err
	}
	return append(all, pairs...), nil
}

func (knf *ConfigMapGeneratorFilter) keyValuesFromLiteralSources(sources []string) ([]Pair, error) {
	var kvs []Pair
	for _, s := range sources {
		k, v, err := knf.parseLiteralSource(s)
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, Pair{Key: k, Value: v})
	}
	return kvs, nil
}

func (knf *ConfigMapGeneratorFilter) keyValuesFromFileSources(sources []string) ([]Pair, error) {
	var kvs []Pair
	for _, s := range sources {
		k, fPath, err := knf.parseFileSource(s)
		if err != nil {
			return nil, err
		}
		// read file
		content, err := ioutil.ReadFile(filepath.Join(knf.Root, fPath))
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, Pair{Key: k, Value: string(content)})
	}
	return kvs, nil
}

func (knf *ConfigMapGeneratorFilter) keyValuesFromEnvFiles(paths []string) ([]Pair, error) {
	var kvs []Pair
	for _, p := range paths {
		content, err := ioutil.ReadFile(filepath.Join(knf.Root, p))
		if err != nil {
			return nil, err
		}
		more, err := knf.keyValuesFromLines(content)
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, more...)
	}
	return kvs, nil
}

// keyValuesFromLines parses given content in to a list of key-value pairs.
func (knf *ConfigMapGeneratorFilter) keyValuesFromLines(content []byte) ([]Pair, error) {
	var kvs []Pair

	scanner := bufio.NewScanner(bytes.NewReader(content))
	currentLine := 0
	for scanner.Scan() {
		// Process the current line, retrieving a key/value pair if
		// possible.
		scannedBytes := scanner.Bytes()
		kv, err := knf.keyValuesFromLine(scannedBytes, currentLine)
		if err != nil {
			return nil, err
		}
		currentLine++

		if len(kv.Key) == 0 {
			// no key means line was empty or a comment
			continue
		}

		kvs = append(kvs, kv)
	}
	return kvs, nil
}

var utf8bom = []byte{0xEF, 0xBB, 0xBF}

// KeyValuesFromLine returns a kv with blank key if the line is empty or a comment.
// The value will be retrieved from the environment if necessary.
func (knf *ConfigMapGeneratorFilter) keyValuesFromLine(line []byte, currentLine int) (Pair, error) {
	kv := Pair{}

	if !utf8.Valid(line) {
		return kv, fmt.Errorf("line %d has invalid utf8 bytes : %v", line, string(line))
	}

	// We trim UTF8 BOM from the first line of the file but no others
	if currentLine == 0 {
		line = bytes.TrimPrefix(line, utf8bom)
	}

	// trim the line from all leading whitespace first
	line = bytes.TrimLeftFunc(line, unicode.IsSpace)

	// If the line is empty or a comment, we return a blank key/value pair.
	if len(line) == 0 || line[0] == '#' {
		return kv, nil
	}

	data := strings.SplitN(string(line), "=", 2)
	key := data[0]

	if len(data) == 2 {
		kv.Value = data[1]
	} else {
		// No value (no `=` in the line) is a signal to obtain the value
		// from the environment.
		kv.Value = os.Getenv(key)
	}
	kv.Key = key
	return kv, nil
}

// ParseFileSource parses the source given.
//
//  Acceptable formats include:
//   1.  source-path: the basename will become the key name
//   2.  source-name=source-path: the source-name will become the key name and
//       source-path is the path to the key file.
//
// Key names cannot include '='.
func (knf *ConfigMapGeneratorFilter) parseFileSource(source string) (keyName, filePath string, err error) {
	numSeparators := strings.Count(source, "=")
	switch {
	case numSeparators == 0:
		return path.Base(source), source, nil
	case numSeparators == 1 && strings.HasPrefix(source, "="):
		return "", "", fmt.Errorf("key name for file path %v missing", strings.TrimPrefix(source, "="))
	case numSeparators == 1 && strings.HasSuffix(source, "="):
		return "", "", fmt.Errorf("file path for key name %v missing", strings.TrimSuffix(source, "="))
	case numSeparators > 1:
		return "", "", fmt.Errorf("key names or file paths cannot contain '='")
	default:
		components := strings.Split(source, "=")
		return components[0], components[1], nil
	}
}

// parseLiteralSource parses the source key=val pair into its component pieces.
// This functionality is distinguished from strings.SplitN(source, "=", 2) since
// it returns an error in the case of empty keys, values, or a missing equals sign.
func (knf *ConfigMapGeneratorFilter) parseLiteralSource(source string) (keyName, value string, err error) {
	// leading equal is invalid
	if strings.Index(source, "=") == 0 {
		return "", "", fmt.Errorf("invalid literal source %v, expected key=value", source)
	}
	// split after the first equal (so values can have the = character)
	items := strings.SplitN(source, "=", 2)
	if len(items) != 2 {
		return "", "", fmt.Errorf("invalid literal source %v, expected key=value", source)
	}
	return items[0], strings.Trim(items[1], "\"'"), nil
}

// configMap defines a Kubernetes ConfigMap -- don't use the kubernetes/api definition directly
// we don't want to take the dependency
type configMap struct {
	yaml.ResourceMeta `yaml:",inline,omitempty" json:",inline,omitempty"`

	// Data contains the configuration data.
	// Each key must consist of alphanumeric characters, '-', '_' or '.'.
	// Values with non-UTF-8 byte sequences must use the BinaryData field.
	// The keys stored in Data must not overlap with the keys in
	// the BinaryData field, this is enforced during validation process.
	// +optional
	Data map[string]string `yaml:"data,omitempty" json:"data,omitempty"`

	// BinaryData contains the binary data.
	// Each key must consist of alphanumeric characters, '-', '_' or '.'.
	// BinaryData can contain byte sequences that are not in the UTF-8 range.
	// The keys stored in BinaryData must not overlap with the ones in
	// the Data field, this is enforced during validation process.
	// Using this field will require 1.10+ apiserver and
	// kubelet.
	// +optional
	BinaryData map[string][]byte `yaml:"binaryData,omitempty" json:"binaryData,omitempty"`
}

const (
	configMapAPIVersion = "v1"
	configMapKind       = "ConfigMap"
)
