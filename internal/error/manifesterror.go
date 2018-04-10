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
type ManifestError struct {
	ManifestFilepath string
	ErrorMsg         string
}

func (me ManifestError) Error() string {
	return fmt.Sprintf("Manifest File [%s]: %s\n", me.ManifestFilepath, me.ErrorMsg)
}

type ManifestErrors struct {
	merrors []error
}

func (me *ManifestErrors) Error() string {
	errormsg := ""
	for _, e := range me.merrors {
		errormsg += e.Error() + "\n"
	}
	return errormsg
}

func (me *ManifestErrors) Append(e error) {
	me.merrors = append(me.merrors, e)
}

func (me *ManifestErrors) Get() []error {
	return me.merrors
}

func (me *ManifestErrors) BatchAppend(e ManifestErrors) {
	for _, err := range e.Get() {
		me.merrors = append(me.merrors, err)
	}
}
