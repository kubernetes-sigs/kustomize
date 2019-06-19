// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package kunstruct provides unstructured from api machinery and factory for creating unstructured
package kunstruct

import (
	"fmt"
	"strconv"
	"strings"
)

// A PathSection contains a list of nested fields, which may end with an
// indexable value. For instance, foo.bar resolves to a PathSection with 2
// fields and no index, while foo[0].bar resolves to two path sections, the
// first containing the field foo and the index 0, and the second containing
// the field bar, with no index. The latter PathSection references the bar
// field of the first item in the foo list
type PathSection struct {
	fields      []string
	idx         int
	searchName  string
	searchValue string
}

func newPathSection() PathSection {
	return PathSection{idx: -1, searchName: "", searchValue: ""}
}

func (ps *PathSection) appendNonEmpty(field string) {
	if len(field) != 0 {
		ps.fields = append(ps.fields, field)
	}
}

func (ps *PathSection) NotIndexed() bool {
	return ps.idx == -1 && ps.searchName == ""
}

func (ps *PathSection) ResolveIndex(s []interface{}) (int, bool, error) {
	if ps.idx >= len(s) {
		return ps.idx, false, fmt.Errorf("index %d is out of bounds", ps.idx)
	}

	if ps.idx != -1 {
		return ps.idx, true, nil
	}

	for curId, subField := range s {
		subMap, ok1 := subField.(map[string]interface{})
		if !ok1 {
			return ps.idx, false,
				fmt.Errorf("%v is of the type %T, expected map[string]interface{}",
					subField, subField)
		}
		if foundValue, ok2 := subMap[ps.searchName]; ok2 {
			if stringValue, ok3 := foundValue.(string); ok3 {
				if stringValue == ps.searchValue {
					return curId, true, nil
				}
			}
		}

	}

	return ps.idx, false, nil
}

func (ps *PathSection) parseIndex(pathElement string) {
	// Assign this index to the current
	// PathSection, save it to the result, then begin
	// a new PathSection
	tmpIdx, err := strconv.Atoi(pathElement)
	if err == nil {
		// We have detected an integer so an array.
		ps.idx = tmpIdx
		ps.searchName = ""
		ps.searchValue = ""
		return
	}

	if strings.Contains(pathElement, "=") {
		// We have detected an searchKey so an array
		keyPart := strings.Split(pathElement, "=")
		ps.searchName = keyPart[0]
		ps.searchValue = keyPart[1]
		return
	}

	// We have detected the downwardapi syntax
	ps.appendNonEmpty(pathElement)
}

func parseFields(path string) (result []PathSection, err error) {
	section := newPathSection()
	if !strings.Contains(path, "[") {
		section.fields = strings.Split(path, ".")
		result = append(result, section)
		return result, nil
	}

	start := 0
	insideParentheses := false
	for i, c := range path {
		switch c {
		case '.':
			if !insideParentheses {
				section.appendNonEmpty(path[start:i])
				start = i + 1
			}
		case '[':
			if !insideParentheses {
				section.appendNonEmpty(path[start:i])
				start = i + 1
				insideParentheses = true
			} else {
				return nil, fmt.Errorf("nested parentheses are not allowed: %s", path)
			}
		case ']':
			if insideParentheses {
				section.parseIndex(path[start:i])
				result = append(result, section)
				section = newPathSection()

				start = i + 1
				insideParentheses = false
			} else {
				return nil, fmt.Errorf("invalid field path %s", path)
			}
		}
	}
	if start < len(path)-1 {
		section.appendNonEmpty(path[start:])
		result = append(result, section)
	}

	for _, section := range result {
		for i, f := range section.fields {
			if strings.HasPrefix(f, "\"") || strings.HasPrefix(f, "'") {
				section.fields[i] = strings.Trim(f, "\"'")
			}
		}
	}
	return result, nil
}
