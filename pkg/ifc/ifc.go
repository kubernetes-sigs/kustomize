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

// Package ifc holds miscellaneous interfaces used by kustomize.
package ifc

// Decoder unmarshalls byte input into an object.
type Decoder interface {
	// SetInput accepts new input.
	SetInput([]byte)
	// Decode yields the next object from the input, else io.EOF
	Decode(interface{}) error
}

// Validator provides functions to validate annotations and labels
type Validator interface {
	MakeAnnotationValidator() func(map[string]string) error
	MakeLabelValidator() func(map[string]string) error
	ValidateNamespace(string) []string
}
