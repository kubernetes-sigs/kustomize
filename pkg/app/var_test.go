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

package app

import (
	"strings"
	"testing"
)

func TestGetFieldAsString(t *testing.T) {
	m := map[string]interface{}{
		"Kind": "Service",
		"metadata": map[string]interface{}{
			"labels": map[string]string{
				"app": "application-name",
			},
			"name": "service-name",
		},
	}
	p := []string{"Kind"}
	s, _ := getFieldAsString(m, p)
	if s != "Service" {
		t.Errorf("Expected to get Service, but actually got %s", s)
	}

	p = []string{"metadata", "name"}
	s, _ = getFieldAsString(m, p)
	if s != "service-name" {
		t.Errorf("Expected to get service-name, but actually got %s", s)
	}

	p = []string{"metadata", "non-existing-field"}
	s, err := getFieldAsString(m, p)
	if !strings.HasSuffix(err.Error(), "field at given fieldpath does not exist") {
		t.Errorf("Unexpected failure due to incorrect error message %s", err.Error())
	}
}
