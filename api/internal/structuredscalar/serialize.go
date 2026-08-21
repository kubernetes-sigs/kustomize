// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package structuredscalar serializes YAML RNodes parsed from embedded JSON/YAML
// strings, preserving the original outer format where possible.
package structuredscalar

import (
	"encoding/json"
	"fmt"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Serialize writes structured data back to a string, preserving JSON vs YAML and
// basic JSON pretty-printing when the original used newlines and indentation.
func Serialize(structuredData *yaml.RNode, originalValue string) (string, error) {
	trimmed := strings.TrimSpace(originalValue)
	if trimmed == "" {
		return "", fmt.Errorf("empty structured scalar")
	}
	firstChar := rune(trimmed[0])
	if firstChar == '{' || firstChar == '[' {
		return serializeAsJSON(structuredData, originalValue)
	}
	return serializeAsYAML(structuredData)
}

func serializeAsJSON(structuredData *yaml.RNode, originalValue string) (string, error) {
	modifiedData, err := structuredData.String()
	if err != nil {
		return "", fmt.Errorf("failed to serialize structured data: %w", err)
	}

	var jsonData interface{}
	if err := yaml.Unmarshal([]byte(modifiedData), &jsonData); err != nil {
		return "", fmt.Errorf("failed to unmarshal YAML data: %w", err)
	}

	if strings.Contains(originalValue, "\n") && strings.Contains(originalValue, "  ") {
		if prettyJSON, err := json.MarshalIndent(jsonData, "", "  "); err == nil {
			return string(prettyJSON), nil
		}
	}

	if compactJSON, err := json.Marshal(jsonData); err == nil {
		return string(compactJSON), nil
	}

	return "", fmt.Errorf("failed to marshal JSON data")
}

func serializeAsYAML(structuredData *yaml.RNode) (string, error) {
	modifiedData, err := structuredData.String()
	if err != nil {
		return "", fmt.Errorf("failed to serialize YAML data: %w", err)
	}

	return strings.TrimSpace(modifiedData), nil
}
