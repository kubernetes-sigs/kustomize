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

package k8sdeps

import (
	"bytes"
	"errors"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// KustDecoder unmarshalls bytes to objects.
type KustDecoder struct {
	d *yaml.YAMLOrJSONDecoder
}

// NewKustDecoder returns a new KustDecoder.
func NewKustDecoder() *KustDecoder {
	return &KustDecoder{}
}

// SetInput initializes an apimachinery decoder.
func (k *KustDecoder) SetInput(in []byte) {
	k.d = yaml.NewYAMLOrJSONDecoder(
		bytes.NewReader(in), 1024)
}

// Decode delegates to the apimachinery decoder.
func (k *KustDecoder) Decode(into interface{}) error {
	if k.d == nil {
		return errors.New("no decoder")
	}
	return k.d.Decode(into)
}
