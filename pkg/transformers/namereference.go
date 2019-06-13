// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package transformers

import (
	"fmt"
	"log"

	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/transformers/config"
)

type nameReferenceTransformer struct {
	backRefs []config.NameBackReferences
}

var _ Transformer = &nameReferenceTransformer{}

// NewNameReferenceTransformer constructs a nameReferenceTransformer
// with a given slice of NameBackReferences.
func NewNameReferenceTransformer(br []config.NameBackReferences) Transformer {
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
//     - path: spec/scaleTargetRef/name
//       kind: HorizontalPodAutoscaler
//
// saying that an HPA, via its 'spec/scaleTargetRef/name'
// field, may refer to a Deployment.  This match to HPA
// means we may need to modify the value in its
// 'spec/scaleTargetRef/name' field, by searching for
// the thing it refers to, and getting its new name.
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
	for referrer, res := range m.AsMap() {
		var candidates resmap.ResMap
		for _, target := range o.backRefs {
			for _, fSpec := range target.FieldSpecs {
				if referrer.Gvk().IsSelected(&fSpec.Gvk) {
					if candidates == nil {
						candidates = m.SubsetThatCouldBeReferencedById(referrer)
					}
					err := MutateField(
						res.Map(),
						fSpec.PathSlice(),
						fSpec.CreateIfNotPresent,
						o.getNewName(
							referrer, target.Gvk, candidates))
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (o *nameReferenceTransformer) getNewName(
	referrer resid.ResId,
	target gvk.Gvk,
	referralCandidates resmap.ResMap) func(in interface{}) (interface{}, error) {
	return func(in interface{}) (interface{}, error) {
		switch in.(type) {
		case string:
			oldName, _ := in.(string)
			for id, res := range referralCandidates.AsMap() {
				if id.Gvk().IsSelected(&target) && id.Name() == oldName {
					matchedIds := referralCandidates.GetMatchingIds(id.GvknEquals)
					// If there's more than one match, there's no way
					// to know which one to pick, so emit error.
					if len(matchedIds) > 1 {
						return nil, fmt.Errorf(
							"Multiple matches for name %s:\n  %v", id, matchedIds)
					}
					// In the resource, note that it is referenced
					// by the referrer.
					res.AppendRefBy(referrer)
					// Return transformed name of the object,
					// complete with prefixes, hashes, etc.
					return res.GetName(), nil
				}
			}
			return in, nil
		case []interface{}:
			l, _ := in.([]interface{})
			var names []string
			for _, item := range l {
				name, ok := item.(string)
				if !ok {
					return nil, fmt.Errorf(
						"%#v is expected to be %T", item, name)
				}
				names = append(names, name)
			}
			for id, res := range referralCandidates.AsMap() {
				indexes := indexOf(id.Name(), names)
				if id.Gvk().IsSelected(&target) && len(indexes) > 0 {
					matchedIds := referralCandidates.GetMatchingIds(id.GvknEquals)
					if len(matchedIds) > 1 {
						return nil, fmt.Errorf(
							"Multiple matches for name %s:\n %v", id, matchedIds)
					}
					for _, index := range indexes {
						l[index] = res.GetName()
					}
					res.AppendRefBy(referrer)
					return l, nil
				}
			}
			return in, nil
		default:
			return nil, fmt.Errorf(
				"%#v is expected to be either a string or a []interface{}", in)
		}
	}
}

func indexOf(s string, slice []string) []int {
	var index []int
	for i, item := range slice {
		if item == s {
			index = append(index, i)
		}
	}
	return index
}
