// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package transformers

import (
	"fmt"
	"log"

	"sigs.k8s.io/kustomize/v3/pkg/resource"

	"sigs.k8s.io/kustomize/v3/pkg/gvk"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/transformers/config"
)

type nameReferenceTransformer struct {
	backRefs []config.NameBackReferences
}

var _ resmap.Transformer = &nameReferenceTransformer{}

// NewNameReferenceTransformer constructs a nameReferenceTransformer
// with a given slice of NameBackReferences.
func NewNameReferenceTransformer(br []config.NameBackReferences) resmap.Transformer {
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
		// Let's select the fields that contain names
		byFieldPath := config.NewFieldPathMapFromSlice(o.backRefs, referrer.OrgId().Gvk)
		if len(byFieldPath) == 0 {
			continue
		}

		candidates := m.SubsetThatCouldBeReferencedByResource(referrer)

		// The original code was doing something like this
		// for _, target := range o.backRefs {
		//   for _, fSpec := range target.FieldSpecs {
		//   ...
		//   }
		// }
		// The issue is if the same fieldPath can be used for refer objects of
		// different kind (for instance ClusterRoleBinding/roleRef/name)
		// the first potential referee in list (which is controlled by Gvk.Sort) will
		// always be picked, the field in the referrer transformed
		// and the second referee silently ignore because fieldPath has already been
		// mutated because the value has now changed.

		for fieldPath, targets := range byFieldPath {
			pathSlice := config.FieldSpec{Path: fieldPath}.PathSlice()
			err := MutateField(
				referrer.Map(),
				pathSlice,
				false,
				o.getNewNameFunc(
					// referrer could be an HPA instance,
					// target could be Gvk for Deployment,
					// candidate a list of resources "reachable"
					// from the HPA.
					referrer, targets, candidates))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// selectReferral picks the referral among a subset of candidates.
// It returns the current name and namespace of the selected candidate.
// Note that the content of the referricalCandidateSubset slice is most of the time
// identical to the referralCandidates resmap. Still in some cases, such
// as ClusterRoleBinding, the subset only contains the resources of a specific
// namespace.
func (o *nameReferenceTransformer) selectReferral(
	oldName string,
	referrer *resource.Resource,
	targets []gvk.Gvk,
	referralCandidates resmap.ResMap,
	referralCandidateSubset []*resource.Resource) (interface{}, interface{}, error) {

	matches := []*resource.Resource{}
	for _, res := range referralCandidateSubset {
		id := res.OrgId()

		for _, target := range targets {
			if id.IsSelected(&target) && res.GetOriginalName() == oldName {
				matches = append(matches, referralCandidates.GetMatchingResourcesByOriginalId(id.Equals)...)
			}
		}
	}

	// We are now really able to detect conflict (unlike in the previous
	// version of the code).
	// Instead of silently ignoring it, output a warning.
	if len(matches) > 1 {
		log.Printf("Warning; multiple matches for name %s in %s:\nSelecting first element of\n%v",
			oldName, referrer.CurId(), getIds(matches))
	}

	// todo(jeb): The 99 is a hack to pass the linter and
	// not have to change the signature of the function until
	// we know how to change the behavior.
	if len(matches) > 99 {
		return nil, nil, fmt.Errorf("multiple matches for %s:\n%v",
			referrer.CurId(), getIds(matches))
	}

	if len(matches) >= 1 {
		// In the resource, note that it is referenced
		// by the referrer.
		matches[0].AppendRefBy(referrer.CurId())

		// Return transformed name of the object,
		// complete with prefixes, hashes, etc.
		return matches[0].GetName(), matches[0].GetNamespace(), nil
	}

	return oldName, nil, nil
}

// utility function to replace a simple string by the new name
func (o *nameReferenceTransformer) getSimpleNameField(
	oldName string,
	referrer *resource.Resource,
	targets []gvk.Gvk,
	referralCandidates resmap.ResMap,
	referralCandidateSubset []*resource.Resource) (interface{}, error) {

	newName, _, err := o.selectReferral(oldName, referrer, targets,
		referralCandidates, referralCandidateSubset)

	return newName, err
}

// utility function to replace name field within a map[string]interface{}
// and leverage the namespace field.
func (o *nameReferenceTransformer) getNameAndNsStruct(
	inMap map[string]interface{},
	referrer *resource.Resource,
	targets []gvk.Gvk,
	referralCandidates resmap.ResMap) (interface{}, error) {

	// Example:
	if _, ok := inMap["name"]; !ok {
		return nil, fmt.Errorf(
			"%#v is expected to contain a name field", inMap)
	}
	oldName, ok := inMap["name"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"%#v is expected to contain a name field of type string", oldName)
	}

	subset := referralCandidates.Resources()
	if namespacevalue, ok := inMap["namespace"]; ok {
		namespace := namespacevalue.(string)
		bynamespace := referralCandidates.GroupedByOriginalNamespace()
		if _, ok := bynamespace[namespace]; !ok {
			return inMap, nil
		}
		subset = bynamespace[namespace]
	}

	newname, newnamespace, err := o.selectReferral(oldName, referrer, targets,
		referralCandidates, subset)
	if err != nil {
		return nil, err
	}

	if (newname == oldName) && (newnamespace == nil) {
		// no candidate found.
		return inMap, nil
	}

	inMap["name"] = newname
	if newnamespace != "" {
		// We don't want value "" to replace value "default" since
		// the empty string is handled as a wild card here not default namespace
		// by kubernetes.
		inMap["namespace"] = newnamespace
	}
	return inMap, nil

}

func (o *nameReferenceTransformer) getNewNameFunc(
	referrer *resource.Resource,
	targets []gvk.Gvk,
	referralCandidates resmap.ResMap) func(in interface{}) (interface{}, error) {
	return func(in interface{}) (interface{}, error) {
		switch in.(type) {
		case string:
			oldName, _ := in.(string)
			return o.getSimpleNameField(oldName, referrer, targets,
				referralCandidates, referralCandidates.Resources())
		case map[string]interface{}:
			// Kind: ValidatingWebhookConfiguration
			// FieldSpec is webhooks/clientConfig/service
			oldMap, _ := in.(map[string]interface{})
			return o.getNameAndNsStruct(oldMap, referrer, targets,
				referralCandidates)
		case []interface{}:
			l, _ := in.([]interface{})
			for idx, item := range l {
				switch item.(type) {
				case string:
					// Kind: Role/ClusterRole
					// FieldSpec is rules.resourceNames
					oldName, _ := item.(string)
					newName, err := o.getSimpleNameField(oldName, referrer, targets,
						referralCandidates, referralCandidates.Resources())
					if err != nil {
						return nil, err
					}
					l[idx] = newName
				case map[string]interface{}:
					// Kind: RoleBinding/ClusterRoleBinding
					// FieldSpec is subjects
					// Note: The corresponding fieldSpec had been changed from
					// from path: subjects/name to just path: subjects. This is
					// what get mutatefield to request the mapping of the whole
					// map containing namespace and name instead of just a simple
					// string field containing the name
					oldMap, _ := item.(map[string]interface{})
					newMap, err := o.getNameAndNsStruct(oldMap, referrer, targets,
						referralCandidates)
					if err != nil {
						return nil, err
					}
					l[idx] = newMap
				default:
					return nil, fmt.Errorf(
						"%#v is expected to be either a []string or a []map[string]interface{}", in)
				}
			}
			return in, nil
		default:
			return nil, fmt.Errorf(
				"%#v is expected to be either a string or a []interface{}", in)
		}
	}
}

func getIds(rs []*resource.Resource) []string {
	var result []string
	for _, r := range rs {
		result = append(result, r.CurId().String()+"\n")
	}
	return result
}
