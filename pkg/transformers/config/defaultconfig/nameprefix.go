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

package defaultconfig

const (
	namePrefixFieldSpecs = `
namePrefix:
- path: metadata/name
- path: metadata/name
  kind: CustomResourceDefinition
  skip: true

# Following merge PR broke backward compatility
# https://github.com/kubernetes-sigs/kustomize/pull/1526
- path: metadata/name
  kind: APIService
  group: apiregistration.k8s.io
  skip: true

# Would make sense to skip those
# by default but would break backward
# compatility
#
# - path: metadata/name
#   kind: Namespace
#   skip: true
# - path: metadata/name
#   group: storage.k8s.io
#   kind: StorageClass
#   skip: true
`
)
