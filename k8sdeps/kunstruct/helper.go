/*
Copyright 2018 The Kubernetes Authors.

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
	fields []string
	idx    *int
}

func appendNonEmpty(section *PathSection, field string) {
	if len(field) != 0 {
		section.fields = append(section.fields, field)
	}
}

func parseFields(path string) ([]PathSection, error) {
	section := PathSection{}
	sectionset := []PathSection{}
	if !strings.Contains(path, "[") {
		section.fields = strings.Split(path, ".")
		sectionset = append(sectionset, section)
		return sectionset, nil
	}

	start := 0
	insideParentheses := false
	for i, c := range path {
		switch c {
		case '.':
			if !insideParentheses {
				appendNonEmpty(&section, path[start:i])
				start = i + 1
			}
		case '[':
			if !insideParentheses {
				appendNonEmpty(&section, path[start:i])
				start = i + 1
				insideParentheses = true
			} else {
				return nil, fmt.Errorf("nested parentheses are not allowed: %s", path)
			}
		case ']':
			if insideParentheses {
				// Assign this index to the current
				// PathSection, save it to the set, then begin
				// a new PathSection
				tmpIdx, err := strconv.Atoi(path[start:i])
				if err != nil {
					return nil, fmt.Errorf("invalid index %s", path)
				}
				section.idx = &tmpIdx
				sectionset = append(sectionset, section)
				section = PathSection{}

				start = i + 1
				insideParentheses = false
			} else {
				return nil, fmt.Errorf("invalid field path %s", path)
			}
		}
	}
	if start < len(path)-1 {
		appendNonEmpty(&section, path[start:])
		sectionset = append(sectionset, section)
	}

	for _, section := range sectionset {
		for i, f := range section.fields {
			if strings.HasPrefix(f, "\"") || strings.HasPrefix(f, "'") {
				section.fields[i] = strings.Trim(f, "\"'")
			}
		}
	}
	return sectionset, nil
}
