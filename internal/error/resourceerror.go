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

package error

import "fmt"

// First pass to encapsulate fields for more informative error messages.
type ResourceError struct {
	ManifestFilepath string
	ResourceFilepath string
	ErrorMsg         string
}

func (e ResourceError) Error() string {
	return fmt.Sprintf("Manifest file [%s] encounters a resource error for [%s]: %s\n", e.ManifestFilepath, e.ResourceFilepath, e.ErrorMsg)
}
