// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// openapi-bundle compiles a Kubernetes OpenAPI v2 protobuf document into the
// compact, deterministic bundle embedded by kyaml.
package main

import (
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	openapi_v2 "github.com/google/gnostic-models/openapiv2"
	"google.golang.org/protobuf/proto"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/kyaml/openapi/internal/builtinopenapi"
)

const gvkExtension = "x-kubernetes-group-version-kind"

const maxInputSize = 64 << 20

type options struct {
	input             string
	output            string
	legacyProtoOutput string
	kubernetesVersion string
}

func main() {
	var opts options
	flag.StringVar(&opts.input, "input", "", "path to a Kubernetes OpenAPI v2 protobuf document (optionally gzip-compressed)")
	flag.StringVar(&opts.output, "output", "", "path to the generated .json.gz bundle")
	flag.StringVar(&opts.legacyProtoOutput, "legacy-proto-output", "", "optional path to a deterministic gzip archive of the input protobuf")
	flag.StringVar(&opts.kubernetesVersion, "kubernetes-version", "", "Kubernetes version represented by the input")
	flag.Parse()

	if err := run(opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(opts options) error {
	if opts.input == "" || opts.output == "" || opts.kubernetesVersion == "" {
		return errors.New("-input, -output, and -kubernetes-version are required")
	}

	input, err := readInput(opts.input)
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}
	bundle, err := compile(input, opts.kubernetesVersion)
	if err != nil {
		return fmt.Errorf("compile bundle: %w", err)
	}
	if err := writeBundle(opts.output, bundle); err != nil {
		return fmt.Errorf("write bundle: %w", err)
	}
	if opts.legacyProtoOutput != "" {
		if err := writeGzip(opts.legacyProtoOutput, input); err != nil {
			return fmt.Errorf("write legacy protobuf archive: %w", err)
		}
	}
	return nil
}

func readInput(path string) ([]byte, error) {
	return readInputWithLimit(path, maxInputSize)
}

func readInputWithLimit(path string, limit int64) (result []byte, resultErr error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("close input %q: %w", path, err))
		}
	}()

	reader := bufio.NewReader(file)
	magic, err := reader.Peek(2)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("inspect input %q: %w", path, err)
	}
	if len(magic) < 2 || magic[0] != 0x1f || magic[1] != 0x8b {
		return readLimitedInput(reader, limit, "input", path)
	}

	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("open gzip input %q: %w", path, err)
	}
	uncompressed, readErr := readLimitedInput(gzipReader, limit, "decompressed input", path)
	closeErr := gzipReader.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close gzip input %q: %w", path, closeErr)
	}
	return uncompressed, nil
}

func readLimitedInput(reader io.Reader, limit int64, description, path string) ([]byte, error) {
	contents, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, fmt.Errorf("read %s %q: %w", description, path, err)
	}
	if int64(len(contents)) > limit {
		return nil, fmt.Errorf("%s %q exceeds %d bytes", description, path, limit)
	}
	return contents, nil
}

func compile(input []byte, kubernetesVersion string) (*builtinopenapi.Bundle, error) {
	document := &openapi_v2.Document{}
	if err := proto.Unmarshal(input, document); err != nil {
		return nil, fmt.Errorf("unmarshal OpenAPI protobuf: %w", err)
	}

	var swagger spec.Swagger
	ok, err := swagger.FromGnostic(document)
	if err != nil {
		return nil, fmt.Errorf("convert gnostic document: %w", err)
	}
	if !ok {
		return nil, errors.New("gnostic document cannot be converted without data loss")
	}

	resources, err := collectResources(&swagger)
	if err != nil {
		return nil, err
	}
	if err := validateDefinitionReferences(swagger.Definitions); err != nil {
		return nil, err
	}

	digest := sha256.Sum256(input)
	bundle := &builtinopenapi.Bundle{
		FormatVersion: builtinopenapi.FormatVersion,
		Coverage: builtinopenapi.Coverage{
			Floor:   kubernetesVersion,
			Ceiling: kubernetesVersion,
		},
		SelectionPolicy: builtinopenapi.SelectionPolicy,
		Sources: []builtinopenapi.Source{{
			KubernetesVersion: kubernetesVersion,
			SHA256:            hex.EncodeToString(digest[:]),
		}},
		Definitions: swagger.Definitions,
		Resources:   resources,
	}
	if err := bundle.Validate(); err != nil {
		return nil, fmt.Errorf("validate compiled bundle: %w", err)
	}
	return bundle, nil
}

func collectResources(swagger *spec.Swagger) ([]builtinopenapi.Resource, error) {
	resources := map[string]builtinopenapi.Resource{}
	for definitionName, definition := range swagger.Definitions {
		extension, found := definition.Extensions[gvkExtension]
		if !found {
			continue
		}
		entries, ok := extension.([]interface{})
		if !ok {
			return nil, fmt.Errorf("definition %q has a malformed %s extension", definitionName, gvkExtension)
		}
		for _, entry := range entries {
			apiVersion, kind, err := parseGVK(entry)
			if err != nil {
				return nil, fmt.Errorf("definition %q: %w", definitionName, err)
			}
			key := resourceKey(apiVersion, kind)
			resource := resources[key]
			if resource.Definition != "" && resource.Definition != definitionName {
				return nil, fmt.Errorf("GVK %s/%s is advertised by definitions %q and %q",
					apiVersion, kind, resource.Definition, definitionName)
			}
			resource.APIVersion = apiVersion
			resource.Kind = kind
			resource.Definition = definitionName
			resources[key] = resource
		}
	}

	if err := collectPathResources(swagger.Paths, resources); err != nil {
		return nil, err
	}

	result := make([]builtinopenapi.Resource, 0, len(resources))
	for _, resource := range resources {
		result = append(result, resource)
	}
	builtinopenapi.SortResources(result)
	return result, nil
}

func collectPathResources(paths *spec.Paths, resources map[string]builtinopenapi.Resource) error {
	if paths == nil {
		return nil
	}
	for path, pathInfo := range paths.Paths {
		if pathInfo.Get == nil {
			continue
		}
		extension, found := pathInfo.Get.Extensions[gvkExtension]
		if !found {
			continue
		}
		apiVersion, kind, err := parseGVK(extension)
		if err != nil {
			return fmt.Errorf("path %q: %w", path, err)
		}
		key := resourceKey(apiVersion, kind)
		resource := resources[key]
		resource.APIVersion = apiVersion
		resource.Kind = kind
		if strings.Contains(path, "namespaces/{namespace}") {
			resource.Scope = builtinopenapi.ScopeNamespaced
		} else if resource.Scope == builtinopenapi.ScopeUnknown {
			resource.Scope = builtinopenapi.ScopeCluster
		}
		resources[key] = resource
	}
	return nil
}

func parseGVK(value interface{}) (string, string, error) {
	entry, ok := value.(map[string]interface{})
	if !ok {
		return "", "", fmt.Errorf("malformed %s extension entry", gvkExtension)
	}
	version, versionOK := entry["version"].(string)
	kind, kindOK := entry["kind"].(string)
	if !versionOK || version == "" || !kindOK || kind == "" {
		return "", "", fmt.Errorf("incomplete %s extension entry", gvkExtension)
	}
	group, groupOK := entry["group"].(string)
	if groupOK && group != "" {
		return group + "/" + version, kind, nil
	}
	return version, kind, nil
}

func resourceKey(apiVersion, kind string) string {
	return apiVersion + "\x00" + kind
}

func validateDefinitionReferences(definitions spec.Definitions) error {
	b, err := json.Marshal(definitions)
	if err != nil {
		return fmt.Errorf("marshal definitions for reference validation: %w", err)
	}
	var value interface{}
	if err := json.Unmarshal(b, &value); err != nil {
		return fmt.Errorf("unmarshal definitions for reference validation: %w", err)
	}
	return walkReferences(value, definitions)
}

func walkReferences(value interface{}, definitions spec.Definitions) error {
	switch typed := value.(type) {
	case []interface{}:
		for _, item := range typed {
			if err := walkReferences(item, definitions); err != nil {
				return err
			}
		}
	case map[string]interface{}:
		for key, item := range typed {
			if key == "$ref" {
				if err := validateReference(item, definitions); err != nil {
					return err
				}
			}
			if err := walkReferences(item, definitions); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateReference(value interface{}, definitions spec.Definitions) error {
	ref, ok := value.(string)
	// A schema may itself describe an object with a property named "$ref".
	// Such a property has a schema object as its value and is not an OpenAPI
	// reference keyword.
	if !ok {
		return nil
	}
	const prefix = "#/definitions/"
	if !strings.HasPrefix(ref, prefix) {
		return fmt.Errorf("OpenAPI definition contains unsupported reference %q", ref)
	}
	name := strings.TrimPrefix(ref, prefix)
	name = strings.ReplaceAll(strings.ReplaceAll(name, "~1", "/"), "~0", "~")
	if _, found := definitions[name]; !found {
		return fmt.Errorf("OpenAPI definition references missing definition %q", name)
	}
	return nil
}

func writeBundle(path string, bundle *builtinopenapi.Bundle) error {
	jsonBytes, err := json.Marshal(bundle)
	if err != nil {
		return fmt.Errorf("marshal bundle: %w", err)
	}
	return writeGzip(path, jsonBytes)
}

func writeGzip(path string, contents []byte) (resultErr error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create output directory %q: %w", dir, err)
	}
	tmp, err := os.CreateTemp(dir, ".openapi-bundle-*")
	if err != nil {
		return fmt.Errorf("create temporary output: %w", err)
	}
	tmpName := tmp.Name()
	tmpClosed := false
	defer func() {
		if !tmpClosed {
			if err := tmp.Close(); err != nil {
				resultErr = errors.Join(resultErr, fmt.Errorf("close temporary output: %w", err))
			}
		}
		if err := os.Remove(tmpName); err != nil && !errors.Is(err, os.ErrNotExist) {
			resultErr = errors.Join(resultErr, fmt.Errorf("remove temporary output: %w", err))
		}
	}()

	writer, err := gzip.NewWriterLevel(tmp, gzip.BestCompression)
	if err != nil {
		return fmt.Errorf("create gzip writer: %w", err)
	}
	writerClosed := false
	defer func() {
		if !writerClosed {
			if err := writer.Close(); err != nil {
				resultErr = errors.Join(resultErr, fmt.Errorf("close gzip writer: %w", err))
			}
		}
	}()
	writer.Header.ModTime = time.Time{}
	writer.Header.Name = ""
	writer.Header.Comment = ""
	writer.Header.Extra = nil
	writer.Header.OS = 255

	if _, err := writer.Write(contents); err != nil {
		return fmt.Errorf("write compressed output: %w", err)
	}
	closeWriterErr := writer.Close()
	writerClosed = true
	if closeWriterErr != nil {
		return fmt.Errorf("close gzip writer: %w", closeWriterErr)
	}
	if err := tmp.Chmod(0o644); err != nil {
		return fmt.Errorf("set output permissions: %w", err)
	}
	closeTempErr := tmp.Close()
	tmpClosed = true
	if closeTempErr != nil {
		return fmt.Errorf("close temporary output: %w", closeTempErr)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("replace output %q: %w", path, err)
	}
	return nil
}
