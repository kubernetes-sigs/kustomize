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

package resource

// RenamingBehavior specifies renaming behavior of configmaps, secrets and maybe other resources.
type RenamingBehavior int

const (
	// RenamingBehaviorUnspecified is an Unspecified behavior; typically treated as the default.
	RenamingBehaviorUnspecified RenamingBehavior = iota
	// RenamingBehaviorNone suppresses the addition of a hash suffix on the end of the resource
	RenamingBehaviorNone
)

// String converts a RenamingBehavior to a string.
func (b RenamingBehavior) String() string {
	switch b {
	case RenamingBehaviorNone:
		return "none"
	default:
		return "unspecified"
	}
}

// NewRenamingBehavior converts a string to a RenamingBehavior.
func NewRenamingBehavior(s string) RenamingBehavior {
	switch s {
	case "none":
		return RenamingBehaviorNone
	default:
		return RenamingBehaviorUnspecified
	}
}
