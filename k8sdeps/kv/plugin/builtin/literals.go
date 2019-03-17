/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package builtin

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kustomize/k8sdeps/kv"
)

// Literals takes a list of literals.
// Each literal source should be a key and literal value,
// e.g. `somekey=somevalue`
// It will be similar to kubectl create configmap|secret --from-literal
type Literals struct{}

// Get implements the interface for kv plugins.
func (p Literals) Get(root string, args []string) ([]kv.Pair, error) {
	var kvs []kv.Pair
	for _, s := range args {
		k, v, err := parseLiteralSource(s)
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, kv.Pair{Key: k, Value: v})
	}
	return kvs, nil
}

// parseLiteralSource parses the source key=val pair into its component pieces.
// This functionality is distinguished from strings.SplitN(source, "=", 2) since
// it returns an error in the case of empty keys, values, or a missing equals sign.
func parseLiteralSource(source string) (keyName, value string, err error) {
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
