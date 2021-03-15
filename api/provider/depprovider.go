// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"sigs.k8s.io/kustomize/api/hasher"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/internal/conflict"
	"sigs.k8s.io/kustomize/api/internal/validate"
	"sigs.k8s.io/kustomize/api/resource"
)

// DepProvider is a dependency provider, injecting different
// implementations depending on the context.
//
// Notes on interfaces:
//
// - resource.ConflictDetector
//
//   implemented by api/internal/conflict.smPatchMergeOnlyDetector
//
//           At time of writing, this doesn't report conflicts,
//           but it does know how to merge patches. Conflict
//           reporting isn't vital to kustomize function.  It's
//           rare that a person would configure one transformer
//           with many patches, much less so many that it became
//           hard to spot conflicts.  In the case of an undetected
//           conflict, the last patch applied wins, likely what
//           the user wants anyway.  Regardless, the effect of this
//           is plainly visible and usable in the output, even if
//           a conflict happened but wasn't reported as an error.
//
// - ifc.Validator
//
//   implemented by api/internal/validate.FieldValidator
//
//            See TODO inside the validator for status.
//            At time of writing, this is a do-nothing
//            validator as it's not critical to kustomize function.
//
type DepProvider struct {
	resourceFactory          *resource.Factory
	conflictDectectorFactory resource.ConflictDetectorFactory
	fieldValidator           ifc.Validator
}

func NewDepProvider() *DepProvider {
	rf := resource.NewFactory(&hasher.Hasher{})
	return &DepProvider{
		resourceFactory:          rf,
		conflictDectectorFactory: conflict.NewFactory(),
		fieldValidator:           validate.NewFieldValidator(),
	}
}

func NewDefaultDepProvider() *DepProvider {
	return NewDepProvider()
}

func (dp *DepProvider) GetResourceFactory() *resource.Factory {
	return dp.resourceFactory
}

func (dp *DepProvider) GetConflictDetectorFactory() resource.ConflictDetectorFactory {
	return dp.conflictDectectorFactory
}

func (dp *DepProvider) GetFieldValidator() ifc.Validator {
	return dp.fieldValidator
}
