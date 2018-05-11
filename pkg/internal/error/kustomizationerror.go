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

import (
	"fmt"
)

// First pass to encapsulate fields for more informative error messages.
type KustomizationError struct {
	KustomizationPath string
	ErrorMsg          string
}

func (ke KustomizationError) Error() string {
	return fmt.Sprintf("Kustomization File [%s]: %s\n", ke.KustomizationPath, ke.ErrorMsg)
}

type KustomizationErrors struct {
	kErrors []error
}

func (ke *KustomizationErrors) Error() string {
	errormsg := ""
	for _, e := range ke.kErrors {
		errormsg += e.Error() + "\n"
	}
	return errormsg
}

func (ke *KustomizationErrors) Append(e error) {
	ke.kErrors = append(ke.kErrors, e)
}

func (ke *KustomizationErrors) Get() []error {
	return ke.kErrors
}

func (ke *KustomizationErrors) BatchAppend(e KustomizationErrors) {
	for _, err := range e.Get() {
		ke.kErrors = append(ke.kErrors, err)
	}
}
