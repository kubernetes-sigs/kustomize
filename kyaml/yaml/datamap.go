// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"encoding/base64"
	"sort"
	"strings"
	"unicode/utf8"
)

type Pair struct {
	Key   string
	Value string
}

// SortedMapKeys returns a sorted slice of keys to the given map.
func SortedMapKeys(m map[string]string) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

// Old map-based methods for compatibility with other kyaml files (like rnode.go)
func (rn *RNode) LoadMapIntoConfigMapData(m map[string]string) error {
	pairs := make([]Pair, 0, len(m))
	for _, k := range SortedMapKeys(m) {
		pairs = append(pairs, Pair{Key: k, Value: m[k]})
	}
	return rn.LoadPairsIntoConfigMapData(pairs)
}

func (rn *RNode) LoadMapIntoConfigMapBinaryData(m map[string]string) error {
	pairs := make([]Pair, 0, len(m))
	for _, k := range SortedMapKeys(m) {
		pairs = append(pairs, Pair{Key: k, Value: m[k]})
	}
	return rn.LoadPairsIntoConfigMapBinaryData(pairs)
}

func (rn *RNode) LoadMapIntoSecretData(m map[string]string) error {
	pairs := make([]Pair, 0, len(m))
	for _, k := range SortedMapKeys(m) {
		pairs = append(pairs, Pair{Key: k, Value: m[k]})
	}
	return rn.LoadPairsIntoSecretData(pairs)
}

// New pair-based methods to preserve input sequence order
func (rn *RNode) LoadPairsIntoConfigMapData(pairs []Pair) error {
	for _, p := range pairs {
		fldName, vrN := makeConfigMapValueRNode(p.Value)
		if _, err := rn.Pipe(
			LookupCreate(MappingNode, fldName),
			SetField(p.Key, vrN)); err != nil {
			return err
		}
	}
	return nil
}

func (rn *RNode) LoadPairsIntoConfigMapBinaryData(pairs []Pair) error {
	for _, p := range pairs {
		_, vrN := makeConfigMapValueRNode(p.Value)
		// we know this is binary data
		fldName := BinaryDataField
		if _, err := rn.Pipe(
			LookupCreate(MappingNode, fldName),
			SetField(p.Key, vrN)); err != nil {
			return err
		}
	}
	return nil
}

func makeConfigMapValueRNode(s string) (field string, rN *RNode) {
	yN := &Node{Kind: ScalarNode}
	yN.Tag = NodeTagString
	if utf8.ValidString(s) {
		field = DataField
		yN.Value = s
	} else {
		field = BinaryDataField
		yN.Value = encodeBase64(s)
	}
	if strings.Contains(yN.Value, "\n") {
		yN.Style = LiteralStyle
	}
	return field, NewRNode(yN)
}

func (rn *RNode) LoadPairsIntoSecretData(pairs []Pair) error {
	mapNode, err := rn.Pipe(LookupCreate(MappingNode, DataField))
	if err != nil {
		return err
	}
	for _, p := range pairs {
		vrN := makeSecretValueRNode(p.Value)
		if _, err := mapNode.Pipe(SetField(p.Key, vrN)); err != nil {
			return err
		}
	}
	return nil
}

// In a secret, all data is base64 encoded, regardless of its conformance
// or lack thereof to UTF-8.
func makeSecretValueRNode(s string) *RNode {
	yN := &Node{Kind: ScalarNode}
	yN.Tag = NodeTagString
	yN.Value = encodeBase64(s)
	if strings.Contains(yN.Value, "\n") {
		yN.Style = LiteralStyle
	}
	return NewRNode(yN)
}

// encodeBase64 encodes s as base64 that is broken up into multiple lines
// as appropriate for the resulting length.
func encodeBase64(s string) string {
	const lineLen = 70
	encLen := base64.StdEncoding.EncodedLen(len(s))
	lines := encLen/lineLen + 1
	buf := make([]byte, encLen*2+lines)
	in := buf[0:encLen]
	out := buf[encLen:]
	base64.StdEncoding.Encode(in, []byte(s))
	k := 0
	for i := 0; i < len(in); i += lineLen {
		j := i + lineLen
		if j > len(in) {
			j = len(in)
		}
		k += copy(out[k:], in[i:j])
		if lines > 1 {
			out[k] = '\n'
			k++
		}
	}
	return string(out[:k])
}