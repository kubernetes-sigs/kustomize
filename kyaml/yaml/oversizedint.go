// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"math/big"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// yamlIntegerLiteral matches YAML integer literals, including ones that do not
// fit in 64 bits. Underscores are the YAML 1.1 digit separator.
var yamlIntegerLiteral = regexp.MustCompile(`(?i)^[+-]?(?:0x[0-9a-f_]+|0o[0-7_]+|0b[01_]+|[0-9_]+)$`)

// IsOversizedYAMLInteger reports whether value is an integer literal whose
// magnitude does not fit in int64 or uint64.
//
// go-yaml quotes strings that resolve as ints, but its resolver gives up past
// 64 bits. Decimal values that overflow uint64 still match the float pattern
// and get quoted; hex, octal, and binary values do not, so they are emitted
// as plain scalars. Other YAML 1.1 parsers then read them as numbers.
func IsOversizedYAMLInteger(value string) bool {
	if !yamlIntegerLiteral.MatchString(value) {
		return false
	}
	negative := false
	s := value
	if s[0] == '+' || s[0] == '-' {
		negative = s[0] == '-'
		s = s[1:]
	}
	s = strings.ReplaceAll(s, "_", "")
	base := 10
	switch {
	case len(s) >= 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X'):
		base = 16
		s = s[2:]
	case len(s) >= 2 && s[0] == '0' && (s[1] == 'o' || s[1] == 'O'):
		base = 8
		s = s[2:]
	case len(s) >= 2 && s[0] == '0' && (s[1] == 'b' || s[1] == 'B'):
		base = 2
		s = s[2:]
	}
	if s == "" {
		return false
	}
	n := new(big.Int)
	if _, ok := n.SetString(s, base); !ok {
		return false
	}
	if negative {
		n.Neg(n)
	}
	if n.IsInt64() || (n.Sign() >= 0 && n.IsUint64()) {
		return false
	}
	return true
}

type quoteEdit struct {
	line   int
	column int
	value  string
}

// QuoteOversizedIntegerScalars double-quotes plain integer scalars that do not
// fit in 64 bits when they appear in ConfigMap data/binaryData or Secret
// data/stringData. Those fields are strings. Quoting keeps YAML 1.1 parsers
// from reading the values as numbers. Other fields and resources are unchanged.
func QuoteOversizedIntegerScalars(doc []byte) ([]byte, error) {
	if len(doc) == 0 {
		return doc, nil
	}
	parsed, err := Parse(string(doc))
	if err != nil || parsed == nil {
		return doc, err
	}
	switch parsed.GetKind() {
	case "ConfigMap", "Secret":
	default:
		return doc, nil
	}

	var edits []quoteEdit
	for _, field := range []string{DataField, BinaryDataField, "stringData"} {
		collectOversizedIntegerEdits(parsed.Field(field), &edits)
	}
	if len(edits) == 0 {
		return doc, nil
	}
	sort.Slice(edits, func(i, j int) bool {
		if edits[i].line != edits[j].line {
			return edits[i].line > edits[j].line
		}
		return edits[i].column > edits[j].column
	})

	lines := strings.Split(string(doc), "\n")
	for _, edit := range edits {
		if edit.line < 1 || edit.line > len(lines) {
			continue
		}
		quoted, ok := quotePlainScalar(lines[edit.line-1], edit.column, edit.value)
		if !ok {
			continue
		}
		lines[edit.line-1] = quoted
	}
	return []byte(strings.Join(lines, "\n")), nil
}

func collectOversizedIntegerEdits(field *MapNode, edits *[]quoteEdit) {
	if field == nil || field.Value == nil || field.Value.YNode() == nil {
		return
	}
	if field.Value.YNode().Kind != MappingNode {
		return
	}
	_ = field.Value.VisitFields(func(n *MapNode) error {
		if n == nil || n.Value == nil || n.Value.YNode() == nil {
			return nil
		}
		yn := n.Value.YNode()
		if yn.Kind != ScalarNode || yn.Line < 1 || yn.Column < 1 {
			return nil
		}
		if yn.Style&(DoubleQuotedStyle|SingleQuotedStyle|LiteralStyle|FoldedStyle) != 0 {
			return nil
		}
		if !IsOversizedYAMLInteger(yn.Value) {
			return nil
		}
		*edits = append(*edits, quoteEdit{
			line:   yn.Line,
			column: yn.Column,
			value:  yn.Value,
		})
		return nil
	})
}

func quotePlainScalar(line string, column int, value string) (string, bool) {
	runes := []rune(line)
	start := column - 1
	val := []rune(value)
	if start < 0 || start+len(val) > len(runes) || string(runes[start:start+len(val)]) != value {
		return line, false
	}
	quoted := []rune(strconv.Quote(value))
	out := make([]rune, 0, len(runes)+2)
	out = append(out, runes[:start]...)
	out = append(out, quoted...)
	out = append(out, runes[start+len(val):]...)
	return string(out), true
}
