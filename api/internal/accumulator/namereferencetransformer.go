// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package accumulator

import (
	"log"

	"sigs.k8s.io/kustomize/api/filters/nameref"
	"sigs.k8s.io/kustomize/api/internal/plugins/builtinconfig"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/kyaml/filtersutil"
)

type nameReferenceTransformer struct {
	backRefs []builtinconfig.NameBackReferences
}

var _ resmap.Transformer = &nameReferenceTransformer{}

// newNameReferenceTransformer constructs a nameReferenceTransformer
// with a given slice of NameBackReferences.
func newNameReferenceTransformer(br []builtinconfig.NameBackReferences) resmap.Transformer {
	if br == nil {
		log.Fatal("backrefs not expected to be nil")
	}
	return &nameReferenceTransformer{backRefs: br}
}

// Transform updates name references in resource A that
// refer to resource B, given that B's name may have
// changed.
//
// For example, a HorizontalPodAutoscaler (HPA)
// necessarily refers to a Deployment, the thing that
// the HPA scales. The Deployment name might change
// (e.g. prefix added), and the reference in the HPA
// has to be fixed.
//
// In the outer loop over the ResMap below, say we
// encounter a specific HPA. Then, in scanning backrefs,
// we encounter an entry like
//
//   - kind: Deployment
//     fieldSpecs:
//     - kind: HorizontalPodAutoscaler
//       path: spec/scaleTargetRef/name
//
// This entry says that an HPA, via its
// 'spec/scaleTargetRef/name' field, may refer to a
// Deployment.  This match to HPA means we may need to
// modify the value in its 'spec/scaleTargetRef/name'
// field, by searching for the thing it refers to,
// and getting its new name.
//
// As a filter, and search optimization, we compute a
// subset of all resources that the HPA could refer to,
// by excluding objects from other namespaces, and
// excluding objects that don't have the same prefix-
// suffix mods as the HPA.
//
// We look in this subset for all Deployment objects
// with a resId that has a Name matching the field value
// present in the HPA.  If no match do nothing; if more
// than one match, it's an error.
//
// We overwrite the HPA name field with the value found
// in the Deployment's name field (the name in the raw
// object - the modified name - not the unmodified name
// in the Deployment's resId).
//
// This process assumes that the name stored in a ResId
// (the ResMap key) isn't modified by name transformers.
// Name transformers should only modify the name in the
// body of the resource object (the value in the ResMap).
//
func (o *nameReferenceTransformer) Transform(m resmap.ResMap) error {
	// TODO: Too much looping, here and in transitive calls.
	for _, referrer := range m.Resources() {
		var candidates resmap.ResMap
		for _, target := range o.backRefs {
			for _, fSpec := range target.FieldSpecs {
				if referrer.OrgId().IsSelected(&fSpec.Gvk) {
					if candidates == nil {
						candidates = m.SubsetThatCouldBeReferencedByResource(referrer)
					}
					err := filtersutil.ApplyToJSON(nameref.Filter{
						FieldSpec:          fSpec,
						Referrer:           referrer,
						Target:             target.Gvk,
						ReferralCandidates: candidates,
					}, referrer)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
