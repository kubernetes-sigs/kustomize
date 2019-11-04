// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package walk

import (
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type ListKind int32

const (
	AssociativeList ListKind = 1 + iota
	NonAssociateList
)

// Visitor is invoked by walk with source and destination node pairs
type Visitor interface {
	VisitMap(nodes Sources) (*yaml.RNode, error)

	VisitScalar(nodes Sources) (*yaml.RNode, error)

	VisitList(nodes Sources, kind ListKind) (*yaml.RNode, error)
}

// NoOp is returned if GrepFilter should do nothing after calling Set
var ClearNode *yaml.RNode = nil
