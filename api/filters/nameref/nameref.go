package nameref

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/api/filters/fieldspec"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filtersutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Filter updates a name references.
type Filter struct {
	// Referrer refers to another resource X by X's name.
	// E.g. A Deployment can refer to a ConfigMap.
	// The Deployment is the Referrer,
	// the ConfigMap is the ReferralTarget.
	// This filter seeks to repair the reference in Deployment, given
	// that the ConfigMap's name may have changed.
	Referrer *resource.Resource

	// NameFieldToUpdate is the field in the Referrer
	// that holds the name requiring an update.
	// This is the field to write.
	NameFieldToUpdate types.FieldSpec

	// ReferralTarget is the source of the new value for
	// the name, always in the 'metadata/name' field.
	// This is the field to read.
	ReferralTarget resid.Gvk

	// Set of resources to scan to find the ReferralTarget.
	ReferralCandidates resmap.ResMap
}

// At time of writing, in practice this is called with a slice with only
// one entry, the node also referred to be the resource in the Referrer field.
func (f Filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	return kio.FilterAll(yaml.FilterFunc(f.run)).Filter(nodes)
}

// The node passed in here is the same node as held in Referrer;
// that's how the referrer's name field is updated.
// Currently, however, this filter still needs the extra methods on Referrer
// to consult things like the resource Id, its namespace, etc.
// TODO(3455): No filter should use the Resource api; all information
// about names should come from annotations, with helper methods
// on the RNode object.  Resource should get stupider, RNode smarter.
func (f Filter) run(node *yaml.RNode) (*yaml.RNode, error) {
	if err := f.confirmNodeMatchesReferrer(node); err != nil {
		// sanity check.
		return nil, err
	}
	if err := node.PipeE(fieldspec.Filter{
		FieldSpec: f.NameFieldToUpdate,
		SetValue:  f.set,
	}); err != nil {
		return nil, errors.Wrapf(
			err, "updating name reference in '%s' field of '%s'",
			f.NameFieldToUpdate.Path, f.Referrer.CurId().String())
	}
	return node, nil
}

// This function is called at many nodes in the YAML doc tree.
// Only on first entry can one expect the argument to match the
// top-level node backing the Referrer.
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
		return fmt.Errorf("node must be a scalar, sequence or map")
	}
}

// Replace name field within a map RNode and leverage the namespace field.
func (f Filter) setMapping(node *yaml.RNode) error {
	if node.YNode().Kind != yaml.MappingNode {
		return fmt.Errorf("expect a mapping node")
	}
	nameNode, err := node.Pipe(yaml.FieldMatcher{Name: "name"})
	if err != nil {
		return errors.Wrap(err, "trying to match 'name' field")
	}
	if nameNode == nil {
		return fmt.Errorf("path config error; no 'name' field in node")
	}
	namespaceNode, err := node.Pipe(yaml.FieldMatcher{Name: "namespace"})
	if err != nil {
		return errors.Wrap(err, "trying to match 'namespace' field")
	}

	// name will not be updated if the namespace doesn't match
	candidates := f.ReferralCandidates.Resources()
	if namespaceNode != nil {
		namespace := namespaceNode.YNode().Value
		bynamespace := f.ReferralCandidates.GroupedByOriginalNamespace()
		if _, ok := bynamespace[namespace]; !ok {
			bynamespace = f.ReferralCandidates.GroupedByCurrentNamespace()
			if _, ok := bynamespace[namespace]; !ok {
				return nil
			}
		}
		candidates = bynamespace[namespace]
	}

	oldName := nameNode.YNode().Value
	referral, err := f.selectReferral(oldName, candidates)
	if err != nil || referral == nil {
		// Nil referral means nothing to do.
		return err
	}
	f.recordTheReferral(referral)
	if referral.GetName() == oldName && referral.GetNamespace() == "" {
		// The name has not changed, nothing to do.
		return nil
	}
	err = node.PipeE(yaml.FieldSetter{
		Name:        "name",
		StringValue: referral.GetName(),
	})
	if err != nil {
		return err
	}
	if referral.GetNamespace() != "" {
		// We don't want value "" to replace value "default" since
		// the empty string is handled as a wild card here not default namespace
		// by kubernetes.
		err = node.PipeE(yaml.FieldSetter{
			Name:        "namespace",
			StringValue: referral.GetNamespace(),
		})
	}
	return err
}

func (f Filter) setScalar(node *yaml.RNode) error {
	referral, err := f.selectReferral(
		node.YNode().Value, f.ReferralCandidates.Resources())
	if err != nil || referral == nil {
		// Nil referral means nothing to do.
		return err
	}
	f.recordTheReferral(referral)
	if referral.GetName() == node.YNode().Value {
		// The name has not changed, nothing to do.
		return nil
	}
	return node.PipeE(yaml.FieldSetter{StringValue: referral.GetName()})
}

// In the resource, make a note that it is referred to by the Referrer.
func (f Filter) recordTheReferral(referral *resource.Resource) {
	referral.AppendRefBy(f.Referrer.CurId())
}

func (f Filter) isRoleRef() bool {
	return strings.HasSuffix(f.NameFieldToUpdate.Path, "roleRef/name")
}

// getRoleRefGvk returns a Gvk in the roleRef field. Return error
// if the roleRef, roleRef/apiGroup or roleRef/kind is missing.
func getRoleRefGvk(res json.Marshaler) (*resid.Gvk, error) {
	n, err := filtersutil.GetRNode(res)
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
// identical to the ReferralCandidates ResMap. Still in some cases, such
// as ClusterRoleBinding, the subset only contains the resources of a specific
// namespace.
func (f Filter) selectReferral(
	oldName string,
	referralCandidates []*resource.Resource) (*resource.Resource, error) {
	var roleRefGvk *resid.Gvk
	if f.isRoleRef() {
		var err error
		roleRefGvk, err = getRoleRefGvk(f.Referrer)
		if err != nil {
			return nil, err
		}
	}
	for _, candidate := range referralCandidates {
		if candidate.GetOriginalName() != oldName {
			continue
		}
		id := candidate.OrgId()
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
					"cannot fix name in '%s' field of referrer '%s';"+
						" found multiple possible referrals: %v",
					f.NameFieldToUpdate.Path,
					f.Referrer.CurId(),
					getIds(filteredMatches))
			}
			// Check is the match the resource we are working on
			if len(filteredMatches) == 0 || candidate != filteredMatches[0] {
				continue
			}
		}
		return candidate, nil
	}
	return nil, nil
}

func getIds(rs []*resource.Resource) string {
	var result []string
	for _, r := range rs {
		result = append(result, r.CurId().String())
	}
	return strings.Join(result, ", ")
}

func checkEqual(k, a, b string) error {
	if a != b {
		return fmt.Errorf(
			"node-referrerOriginal '%s' mismatch '%s' != '%s'",
			k, a, b)
	}
	return nil
}

func (f Filter) confirmNodeMatchesReferrer(node *yaml.RNode) error {
	meta, err := node.GetMeta()
	if err != nil {
		return err
	}
	gvk := f.Referrer.GetGvk()
	if err = checkEqual(
		"APIVersion", meta.APIVersion, gvk.ApiVersion()); err != nil {
		return err
	}
	if err = checkEqual(
		"Kind", meta.Kind, gvk.Kind); err != nil {
		return err
	}
	if err = checkEqual(
		"Name", meta.Name, f.Referrer.GetName()); err != nil {
		return err
	}
	if err = checkEqual(
		"Namespace", meta.Namespace, f.Referrer.GetNamespace()); err != nil {
		return err
	}
	return nil
}
