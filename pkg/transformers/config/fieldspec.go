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

package config

import (
	"fmt"
	"sort"
	"strings"

	"sigs.k8s.io/kustomize/v3/pkg/gvk"
)

// FieldSpec completely specifies a kustomizable field in
// an unstructured representation of a k8s API object.
// It helps define the operands of transformations.
//
// For example, a directive to add a common label to objects
// will need to know that a 'Deployment' object (in API group
// 'apps', any version) can have labels at field path
// 'spec/template/metadata/labels', and further that it is OK
// (or not OK) to add that field path to the object if the
// field path doesn't exist already.
//
// This would look like
// {
//   group: apps
//   kind: Deployment
//   path: spec/template/metadata/labels
//   create: true
// }
type FieldSpec struct {
	gvk.Gvk            `json:",inline,omitempty" yaml:",inline,omitempty"`
	Path               string `json:"path,omitempty" yaml:"path,omitempty"`
	CreateIfNotPresent bool   `json:"create,omitempty" yaml:"create,omitempty"`
	SkipTransformation bool   `json:"skip,omitempty" yaml:"skip,omitempty"`
}

const (
	escapedForwardSlash  = "\\/"
	tempSlashReplacement = "???"
)

func (fs FieldSpec) String() string {
	return fmt.Sprintf(
		"%s:%v:%s", fs.Gvk.String(), fs.CreateIfNotPresent, fs.Path)
}

// TODO(jeb): Method needs to be improve deal with multiple
// formats of a path: foo.bar is equivalent to foo[bar]
func (fs FieldSpec) ArePathEquals(other FieldSpec) bool {
	return fs.Path == other.Path
}

// If true, the primary key is the same, but other fields might not be.
func (fs FieldSpec) effectivelyEquals(other FieldSpec) bool {
	return fs.IsSelected(&other.Gvk) && fs.ArePathEquals(other)
}

// PathSlice converts the path string to a slice of strings,
// separated by a '/'. Forward slash can be contained in a
// fieldname. such as ingress.kubernetes.io/auth-secret in
// Ingress annotations. To deal with this special case, the
// path to this field should be formatted as
//
//   metadata/annotations/ingress.kubernetes.io\/auth-secret
//
// Then PathSlice will return
//
//   []string{
//      "metadata",
//      "annotations",
//      "ingress.auth-secretkubernetes.io/auth-secret"
//   }
func (fs FieldSpec) PathSlice() []string {
	if !strings.Contains(fs.Path, escapedForwardSlash) {
		return strings.Split(fs.Path, "/")
	}
	s := strings.Replace(fs.Path, escapedForwardSlash, tempSlashReplacement, -1)
	paths := strings.Split(s, "/")
	var result []string
	for _, path := range paths {
		result = append(result, strings.Replace(path, tempSlashReplacement, "/", -1))
	}
	return result
}

type fsSlice []FieldSpec

func (s fsSlice) Len() int      { return len(s) }
func (s fsSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s fsSlice) Less(i, j int) bool {
	if !s[i].Gvk.Equals(s[j].Gvk) {
		return s[i].Gvk.IsLessThan2(s[j].Gvk)
	}
	return s[i].Path < s[j].Path
}

// mergeAll merges the argument into this, returning the result.
// Items already present are ignored.
// Items that conflict (primary key matches, but remain data differs)
// result in an error.
func (s fsSlice) mergeAll(incoming fsSlice) (result fsSlice, err error) {
	result = s
	for _, x := range incoming {
		result, err = result.mergeOne(x)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// mergeOne merges the argument into this, returning the result.
// If the item's primary key is already present, and there are no
// conflicts, it is ignored (we don't want duplicates).
// If there is a conflict, the merge fails.
func (s fsSlice) mergeOne(x FieldSpec) (fsSlice, error) {
	i := s.intersect(x)
	if i > -1 {
		// It's already there.
		if (s[i].SkipTransformation == x.SkipTransformation) && (s[i].CreateIfNotPresent != x.CreateIfNotPresent) {
			return nil, fmt.Errorf("conflicting fieldspecs")
		}
		return s, nil
	}
	return append(s, x), nil
}

func (s fsSlice) index(fs FieldSpec) int {
	for i, x := range s {
		// Allow more specific versions to coexist
		if x.effectivelyEquals(fs) && (x.Version == "" || fs.Version != "") {
			return i
		}
	}
	return -1
}

// todo(jeb): This should most likely be updated to return
// an array instead of just an index.
func (s fsSlice) intersect(fs FieldSpec) int {
	for i, x := range s {
		if (x.Gvk.Kind == fs.Gvk.Kind) &&
			x.effectivelyEquals(fs) &&
			(x.Version == "" || fs.Version != "") {
			return i
		}
	}
	return -1
}

// FieldSpecs wraps a FieldSpec slice in order to add
// utility method.
type FieldSpecs []FieldSpec

var _ sort.Interface = FieldSpecs{}

func (s FieldSpecs) Len() int      { return len(s) }
func (s FieldSpecs) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s FieldSpecs) Less(i, j int) bool {
	if !s[i].Gvk.Equals(s[j].Gvk) {
		return s[i].Gvk.IsLessThan2(s[j].Gvk)
	}
	return s[i].Path < s[j].Path
}

// Create a new FieldSpecs out of []FieldSpec
func NewFieldSpecs(selected fsSlice) FieldSpecs {
	s := FieldSpecs{}
	for _, x := range selected {
		s = append(s, x)
	}
	return s
}

func NewFieldSpecsFromSlice(other []FieldSpec) FieldSpecs {
	s := FieldSpecs{}
	for _, x := range other {
		s = append(s, x)
	}
	return s
}

// Normalize detects the conflict in the FieldSpec Slice
// and compress the slice a much as possible
// todo(jeb): Implement the function
func (s FieldSpecs) Normalize() FieldSpecs {
	return s
}

// This method either adds a new FieldSpec to the list
// or remove a global/generic one with one which is more specific
// because the Kind is specified.
// todo(jeb): Check if we can not reuse fsSlice.mergeOne(add)
// todo(jeb): This method should deals with version and apiGroup.
func (s FieldSpecs) squashFieldSpecs(fs FieldSpec) FieldSpecs {
	for idx, already := range s {
		if !fs.ArePathEquals(already) {
			continue
		}

		if fs.IsSelected(&already.Gvk) {
			// There is already a more global fieldspec definition
			// Let's replace it with a more narrow one
			s[idx] = fs
		}

		return s
	}

	s = append(s, fs)
	return s
}

// This method remove from the FieldSpecs the existing FieldSpec
// which are matching the Gvk. Mainly used to trim the FieldSpec
// slice in order to prevent a transformation from behing applied
// on a specific Gvk.
// todo(jeb): Check if we can not reuse fsSlice.mergeOne(remove)
// todo(jeb): This method can only remove one element at the time.
func (s FieldSpecs) pruneFieldSpecs(fs FieldSpec) FieldSpecs {
	for idx, already := range s {
		if !fs.ArePathEquals(already) {
			continue
		}

		if fs.IsSelected(&already.Gvk) {
			// There is already a more global fieldspec definition
			// Let's remove it
			copy(s[idx:], s[idx+1:])
			s[len(s)-1] = FieldSpec{}
			s = s[:len(s)-1]
		}

		return s
	}

	return s
}

// ApplicableFieldsSpecs extract out of the Transformer Config
// the FieldSpec which are applicable for that particular Gvk
func (s FieldSpecs) ApplicableFieldSpecs(x gvk.Gvk) FieldSpecs {
	selected := FieldSpecs{}
	for _, fs := range s {
		if !fs.SkipTransformation && x.IsSelected(&fs.Gvk) {
			selected = selected.squashFieldSpecs(fs)
		}
	}
	for _, fs := range s {
		if fs.SkipTransformation && x.IsSelected(&fs.Gvk) {
			selected = selected.pruneFieldSpecs(fs)
		}
	}
	sort.Sort(selected)
	return selected
}
