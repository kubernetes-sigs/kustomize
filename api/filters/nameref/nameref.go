package nameref

import (
	"encoding/json"
	"fmt"
	"strings"

	"sigs.k8s.io/kustomize/api/filters/fieldspec"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	kyaml_filtersutil "sigs.k8s.io/kustomize/kyaml/filtersutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Filter updates a name references.
type Filter struct {
	// Referrer is the object that refers to something else by a name,
	// a name that this filter seeks to update.
	Referrer *resource.Resource

	// NameFieldToUpdate is the field in the Referrer that holds the
	// name requiring an update.
	NameFieldToUpdate types.FieldSpec `json:"nameFieldToUpdate,omitempty" yaml:"nameFieldToUpdate,omitempty"`

	// Source of the new value for the name (in its name field).
	ReferralTarget resid.Gvk

	// Set of resources to hunt through to find the ReferralTarget.
	ReferralCandidates resmap.ResMap
}

// At time of writing, in practice this is called with a slice with only
// one entry, the node also referred to be the resource in the Referrer field.
func (f Filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	return kio.FilterAll(yaml.FilterFunc(f.run)).Filter(nodes)
}

// The node passed in here is the same node as held in Referrer, and
// that's how the referrer's name field is updated.
// However, this filter still needs the extra methods on Referrer
// to consult things like the resource Id, its namespace, etc.
func (f Filter) run(node *yaml.RNode) (*yaml.RNode, error) {
	err := node.PipeE(fieldspec.Filter{
		FieldSpec: f.NameFieldToUpdate,
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
		return f.setMapping(node)
	case yaml.SequenceNode:
		return applyFilterToSeq(seqFilter{
			setScalarFn:  f.setScalar,
			setMappingFn: f.setMapping,
		}, node)
	default:
		return fmt.Errorf(
			"node is expected to be either a string or a slice of string or a map of string")
	}
}

// Replace name field within a map RNode and leverage the namespace field.
func (f Filter) setMapping(node *yaml.RNode) error {
	if node.YNode().Kind != yaml.MappingNode {
		return fmt.Errorf("expect a mapping node")
	}
	nameNode, err := node.Pipe(yaml.FieldMatcher{Name: "name"})
	if err != nil || nameNode == nil {
		return fmt.Errorf("cannot find field 'name' in node")
	}
	namespaceNode, err := node.Pipe(yaml.FieldMatcher{Name: "namespace"})
	if err != nil {
		return fmt.Errorf("error when find field 'namespace'")
	}

	// name will not be updated if the namespace doesn't match
	subset := f.ReferralCandidates.Resources()
	if namespaceNode != nil {
		namespace := namespaceNode.YNode().Value
		bynamespace := f.ReferralCandidates.GroupedByOriginalNamespace()
		if _, ok := bynamespace[namespace]; !ok {
			return nil
		}
		subset = bynamespace[namespace]
	}

	oldName := nameNode.YNode().Value
	res, err := f.selectReferral(oldName, subset)
	if err != nil || res == nil {
		// Nil res means nothing to do.
		return err
	}
	f.recordTheReferral(res)
	if res.GetName() == oldName && res.GetNamespace() == "" {
		// The name has not changed, nothing to do.
		return nil
	}
	err = node.PipeE(yaml.FieldSetter{
		Name:        "name",
		StringValue: res.GetName(),
	})
	if err != nil {
		return err
	}
	if res.GetNamespace() != "" {
		// We don't want value "" to replace value "default" since
		// the empty string is handled as a wild card here not default namespace
		// by kubernetes.
		err = node.PipeE(yaml.FieldSetter{
			Name:        "namespace",
			StringValue: res.GetNamespace(),
		})
	}
	return err
}

func (f Filter) setScalar(node *yaml.RNode) error {
	res, err := f.selectReferral(
		node.YNode().Value, f.ReferralCandidates.Resources())
	if err != nil || res == nil {
		// Nil res means nothing to do.
		return err
	}
	f.recordTheReferral(res)
	if res.GetName() == node.YNode().Value {
		// The name has not changed, nothing to do.
		return nil
	}
	return node.PipeE(yaml.FieldSetter{StringValue: res.GetName()})
}

// In the resource, make a note that it is referred to by the referrer.
func (f Filter) recordTheReferral(res *resource.Resource) {
	res.AppendRefBy(f.Referrer.CurId())
}

func (f Filter) isRoleRef() bool {
	return strings.HasSuffix(f.NameFieldToUpdate.Path, "roleRef/name")
}

// getRoleRefGvk returns a Gvk in the roleRef field. Return error
// if the roleRef, roleRef/apiGroup or roleRef/kind is missing.
func getRoleRefGvk(res json.Marshaler) (*resid.Gvk, error) {
	n, err := kyaml_filtersutil.GetRNode(res)
	if err != nil {
		return nil, err
	}
	roleRef, err := n.Pipe(yaml.Lookup("roleRef"))
	if err != nil {
		return nil, err
	}
	if roleRef.IsNil() {
		return nil, fmt.Errorf("roleRef cannot be found in %s", n.MustString())
	}
	apiGroup, err := roleRef.Pipe(yaml.Lookup("apiGroup"))
	if err != nil {
		return nil, err
	}
	if apiGroup.IsNil() {
		return nil, fmt.Errorf(
			"apiGroup cannot be found in roleRef %s", roleRef.MustString())
	}
	kind, err := roleRef.Pipe(yaml.Lookup("kind"))
	if err != nil {
		return nil, err
	}
	if kind.IsNil() {
		return nil, fmt.Errorf(
			"kind cannot be found in roleRef %s", roleRef.MustString())
	}
	return &resid.Gvk{
		Group: apiGroup.YNode().Value,
		Kind:  kind.YNode().Value,
	}, nil
}

func (f Filter) filterReferralCandidates(
	matches []*resource.Resource) []*resource.Resource {
	var ret []*resource.Resource
	for _, m := range matches {
		// If target kind is not ServiceAccount, we shouldn't consider condidates which
		// doesn't have same namespace.
		if f.ReferralTarget.Kind != "ServiceAccount" &&
			m.GetNamespace() != f.Referrer.GetNamespace() {
			continue
		}
		if !f.Referrer.PrefixesSuffixesEquals(m) {
			continue
		}
		ret = append(ret, m)
	}
	return ret
}

// selectReferral picks the referral among a subset of candidates.
// The content of the candidateSubset slice is most of the time
// identical to the ReferralCandidates resmap. Still in some cases, such
// as ClusterRoleBinding, the subset only contains the resources of a specific
// namespace.
func (f Filter) selectReferral(
	oldName string,
	candidateSubset []*resource.Resource) (*resource.Resource, error) {
	var roleRefGvk *resid.Gvk
	if f.isRoleRef() {
		var err error
		roleRefGvk, err = getRoleRefGvk(f.Referrer)
		if err != nil {
			return nil, err
		}
	}
	for _, res := range candidateSubset {
		if res.GetOriginalName() != oldName {
			continue
		}
		id := res.OrgId()
		if !id.IsSelected(&f.ReferralTarget) {
			continue
		}
		// If the we are processing a roleRef, the apiGroup and Kind in the
		// roleRef are needed to be considered.
		if f.isRoleRef() && !id.IsSelected(roleRefGvk) {
			continue
		}
		matches := f.ReferralCandidates.GetMatchingResourcesByOriginalId(id.Equals)
		// If there's more than one match,
		// filter the matches by prefix and suffix
		if len(matches) > 1 {
			filteredMatches := f.filterReferralCandidates(matches)
			if len(filteredMatches) > 1 {
				return nil, fmt.Errorf(
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
		res.AppendRefBy(f.Referrer.CurId())
		// Return transformed name of the object,
		// complete with prefixes, hashes, etc.
		return res, nil
	}
	return nil, nil
}

func getIds(rs []*resource.Resource) []string {
	var result []string
	for _, r := range rs {
		result = append(result, r.CurId().String()+"\n")
	}
	return result
}
