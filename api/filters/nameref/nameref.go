package nameref

import (
	"fmt"

	"sigs.k8s.io/kustomize/api/filters/fieldspec"
	"sigs.k8s.io/kustomize/api/filters/filtersutil"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Filter will update the name reference
type Filter struct {
	FieldSpec          types.FieldSpec `json:"fieldSpec,omitempty" yaml:"fieldSpec,omitempty"`
	Referrer           *resource.Resource
	Target             resid.Gvk
	ReferralCandidates resmap.ResMap
}

func (f Filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	return kio.FilterAll(yaml.FilterFunc(f.run)).Filter(nodes)
}

func (f Filter) run(node *yaml.RNode) (*yaml.RNode, error) {
	err := node.PipeE(fieldspec.Filter{
		FieldSpec: f.FieldSpec,
		SetValue:  f.set,
	})
	return node, err
}

func (f Filter) set(node *yaml.RNode) error {
	if yaml.IsMissingOrNull(node) {
		return nil
	}
	switch node.YNode().Kind {
	case yaml.ScalarNode:
		return f.setScalar(node)
	case yaml.MappingNode:
		// Kind: ValidatingWebhookConfiguration
		// FieldSpec is webhooks/clientConfig/service
		return f.setMapping(node)
	case yaml.SequenceNode:
		return f.setSequence(node)
	default:
		return fmt.Errorf(
			"node is expected to be either a string or a slice of string or a map of string")
	}
}

func (f Filter) setSequence(node *yaml.RNode) error {
	return applyFilterToSeq(seqFilter{
		setScalarFn:  f.setScalar,
		setMappingFn: f.setMapping,
	}, node)
}

func (f Filter) setMapping(node *yaml.RNode) error {
	return setNameAndNs(
		node,
		f.Referrer,
		f.Target,
		f.ReferralCandidates,
	)
}

func (f Filter) setScalar(node *yaml.RNode) error {
	newValue, err := getSimpleNameField(
		node.YNode().Value,
		f.Referrer,
		f.Target,
		f.ReferralCandidates,
		f.ReferralCandidates.Resources(),
	)
	if err != nil {
		return err
	}
	err = filtersutil.SetScalar(newValue)(node)
	if err != nil {
		return err
	}
	return nil
}

func filterReferralCandidates(
	referrer *resource.Resource,
	matches []*resource.Resource) []*resource.Resource {
	var ret []*resource.Resource
	for _, m := range matches {
		if referrer.PrefixesSuffixesEquals(m) {
			ret = append(ret, m)
		}
	}
	return ret
}

// selectReferral picks the referral among a subset of candidates.
// It returns the current name and namespace of the selected candidate.
// Note that the content of the referricalCandidateSubset slice is most of the time
// identical to the referralCandidates resmap. Still in some cases, such
// as ClusterRoleBinding, the subset only contains the resources of a specific
// namespace.
func selectReferral(
	oldName string,
	referrer *resource.Resource,
	target resid.Gvk,
	referralCandidates resmap.ResMap,
	referralCandidateSubset []*resource.Resource) (string, string, error) {

	for _, res := range referralCandidateSubset {
		id := res.OrgId()
		if id.IsSelected(&target) && res.GetOriginalName() == oldName {
			matches := referralCandidates.GetMatchingResourcesByOriginalId(id.Equals)
			// If there's more than one match,
			// filter the matches by prefix and suffix
			if len(matches) > 1 {
				filteredMatches := filterReferralCandidates(referrer, matches)
				if len(filteredMatches) > 1 {
					return "", "", fmt.Errorf(
						"multiple matches for %s:\n  %v",
						id, getIds(filteredMatches))
				}
				// Check is the match the resource we are working on
				if len(filteredMatches) == 0 || res != filteredMatches[0] {
					continue
				}
			}
			// In the resource, note that it is referenced
			// by the referrer.
			res.AppendRefBy(referrer.CurId())
			// Return transformed name of the object,
			// complete with prefixes, hashes, etc.
			return res.GetName(), res.GetNamespace(), nil
		}
	}

	return oldName, "", nil
}

// utility function to replace a simple string by the new name
func getSimpleNameField(
	oldName string,
	referrer *resource.Resource,
	target resid.Gvk,
	referralCandidates resmap.ResMap,
	referralCandidateSubset []*resource.Resource) (string, error) {

	newName, _, err := selectReferral(oldName, referrer, target,
		referralCandidates, referralCandidateSubset)

	return newName, err
}

func getIds(rs []*resource.Resource) []string {
	var result []string
	for _, r := range rs {
		result = append(result, r.CurId().String()+"\n")
	}
	return result
}

// utility function to replace name field within a map RNode
// and leverage the namespace field.
func setNameAndNs(
	in *yaml.RNode,
	referrer *resource.Resource,
	target resid.Gvk,
	referralCandidates resmap.ResMap) error {

	if in.YNode().Kind != yaml.MappingNode {
		return fmt.Errorf("expect a mapping node")
	}

	// Get name field
	nameNode, err := in.Pipe(yaml.FieldMatcher{
		Name: "name",
	})
	if err != nil || nameNode == nil {
		return fmt.Errorf("cannot find field 'name' in node")
	}

	// Get namespace field
	namespaceNode, err := in.Pipe(yaml.FieldMatcher{
		Name: "namespace",
	})
	if err != nil {
		return fmt.Errorf("error when find field 'namespace'")
	}

	// check is namespace matched
	// name will bot be updated if the namespace doesn't match
	subset := referralCandidates.Resources()
	if namespaceNode != nil {
		namespace := namespaceNode.YNode().Value
		bynamespace := referralCandidates.GroupedByOriginalNamespace()
		if _, ok := bynamespace[namespace]; !ok {
			return nil
		}
		subset = bynamespace[namespace]
	}

	oldName := nameNode.YNode().Value
	newname, newnamespace, err := selectReferral(oldName, referrer, target,
		referralCandidates, subset)
	if err != nil {
		return err
	}

	if (newname == oldName) && (newnamespace == "") {
		// no candidate found.
		return nil
	}

	// set name
	in.Pipe(yaml.FieldSetter{
		Name:        "name",
		StringValue: newname,
	})
	if newnamespace != "" {
		// We don't want value "" to replace value "default" since
		// the empty string is handled as a wild card here not default namespace
		// by kubernetes.
		in.Pipe(yaml.FieldSetter{
			Name:        "namespace",
			StringValue: newnamespace,
		})
	}
	return nil
}
