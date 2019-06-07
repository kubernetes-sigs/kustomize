// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package resmap implements a map from ResId to Resource that
// tracks all resources in a kustomization.
package resmap

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/types"
	"sigs.k8s.io/yaml"
)

// ResMap is an interface describing operations on the
// core kustomize data structure.
//
// TODO: delete the commentary below when/if the issues
// discussed are addressed.
//
// It's a temporary(?) interface used during a refactoring
// from a bare map (map[resid.ResId]*resource.Resource) to a
// pointer to struct (currently named *resWrangler).
// Replacing it with a ptr to struct will ease click-thrus
// to implementation during development.
// OTOH, hackery in a PR might be easier to see if the
// interface is left in place.
//
// The old (bare map) ResMap had pervasive problems:
//
//  * It was mutated inside loops over itself.
//
//    Bugs introduced this way were hard to find since the
//    bare map was recursively passed everywhere, sometimes
//    mid loop.
//
//  * Its keys (ResId) aren't opaque, and are effectively
//    mutated (via copy and replace) for data storage reasons
//    as a hack.
//
//    ResId was modified a long time ago as a hack to
//    store name transformation data (prefix and suffix),
//    destabilizing the basic map concept and resulting
//    in the need for silly ResId functions like
//    NewResIdWithPrefixSuffixNamespace, NsGvknEquals,
//    HasSameLeftmostPrefix, CopyWithNewPrefixSuffix, etc.
//    plus logic to use them, and overly complex tests.
//
//    If this data were stored in the Resource object
//    (not in Kunstructured, but as a sibling to it next to
//    GenArgs, references, etc.) then much code could be
//    deleted and the remainder simplified.
//
//  * It doesn't preserve (by definition) value order.
//
//    Preserving order is now needed to support
//    transformer plugins (they aren't commutative).
//
// One way to fix this is deprecate use of ResId as the
// key in favor of ItemId.  See use of the resmap.Remove
// function to spot the places that need fixing to allow
// this.
type ResMap interface {
	// Size reports the number of resources.
	Size() int

	// Resources provides a discardable slice
	// of resource pointers, returned in the order
	// as appended.
	Resources() []*resource.Resource

	// Append adds a Resource, automatically computing its
	// associated Id.
	// Error on Id collision.
	Append(*resource.Resource) error

	// AppendWithId adds a Resource with the given Id.
	// Error on Id collision.
	AppendWithId(resid.ResId, *resource.Resource) error

	// AppendAll appends another ResMap to self,
	// failing on any Id collision.
	AppendAll(ResMap) error

	// AbsorbAll appends, replaces or merges the contents
	// of another ResMap into self,
	// allowing and sometimes demanding ID collisions.
	// A collision would be demanded, say, when a generated
	// ConfigMap has the "replace" option in its generation
	// instructions, meaning it _must_ replace
	// something in the known set of resources.
	// If a resource id for resource X is found to already
	// be in self, then the behavior field for X must
	// be BehaviorMerge or BehaviorReplace. If X is not in
	// self, then its behavior _cannot_ be merge or replace.
	AbsorbAll(ResMap) error

	// AsMap returns (ResId, *Resource) pairs in
	// arbitrary order via a generated map.
	// The map is discardable, and edits to map structure
	// have no impact on the ResMap.
	// The Ids are copies, but the resources are pointers,
	// so the resources themselves can be modified.
	AsMap() map[resid.ResId]*resource.Resource

	// AsYaml returns the yaml form of resources.
	AsYaml() ([]byte, error)

	// Gets the resource with the given Id, else nil.
	GetById(resid.ResId) *resource.Resource

	// ReplaceResource associates a new resource with
	// an _existing_ Id.
	// Error if Id unknown, or if some other Id points
	// to the same resource object.
	ReplaceResource(resid.ResId, *resource.Resource) error

	// AllIds returns all known Ids.
	// Result order is arbitrary.
	AllIds() []resid.ResId

	// GetMatchingIds returns a slice of Ids that
	// satisfy the given matcher function.
	// Result order is arbitrary.
	GetMatchingIds(IdMatcher) []resid.ResId

	// Remove removes the Id and the resource it points to.
	Remove(resid.ResId) error

	// Clear removes all resources and Ids.
	Clear()

	// ResourcesThatCouldReference returns a new ResMap with
	// resources that _might_ reference the resource represented
	// by the argument Id, excluding resources that should
	// _never_ reference the Id.  E.g., if the Id
	// refers to a ConfigMap, the returned set may include a
	// Deployment from the same namespace and exclude Deployments
	// from other namespaces.  Cluster wide objects are
	// never excluded.
	ResourcesThatCouldReference(resid.ResId) ResMap

	// DeepCopy copies the ResMap and underlying resources.
	DeepCopy() ResMap

	// ShallowCopy copies the ResMap but
	// not the underlying resources.
	ShallowCopy() ResMap

	// ErrorIfNotEqualSets returns an error if the
	// argument doesn't have the same Ids and resource
	// data as self. Ordering is _not_ taken into account,
	// as this function was solely used in tests written
	// before internal resource order was maintained,
	// and those tests are initialized with maps which
	// by definition have random ordering, and will
	// fail spuriously.
	// TODO: modify tests to not use resmap.FromMap,
	// TODO: - and replace this with a stricter equals.
	ErrorIfNotEqualSets(ResMap) error

	// Debug prints the ResMap.
	Debug(title string)
}

// resWrangler holds the content manipulated by kustomize.
type resWrangler struct {
	// Resource list maintained in load (append) order.
	// This is important for transformers, which must
	// be performed in a specific order, and for users
	// who for whatever reasons wish the order they
	// specify in kustomizations to be maintained and
	// available as an option for final YAML rendering.
	rList []*resource.Resource

	// A map from id to an index into rList.
	// At the time of writing, the ids used as keys in
	// this map cannot be assumed to match the id
	// generated from the resource.Id() method pointed
	// to by the map's value (via rList).  These keys
	// have been hacked to store prefix/suffix data.
	rIndex map[resid.ResId]int
}

func newOne() *resWrangler {
	result := &resWrangler{}
	result.Clear()
	return result
}

// Clear implements ResMap.
func (m *resWrangler) Clear() {
	m.rList = nil
	m.rIndex = make(map[resid.ResId]int)
}

// Size implements ResMap.
func (m *resWrangler) Size() int {
	if len(m.rList) != len(m.rIndex) {
		panic("class size invariant violation")
	}
	return len(m.rList)
}

func (m *resWrangler) indexOfResource(other *resource.Resource) int {
	for i, r := range m.rList {
		if r == other {
			return i
		}
	}
	return -1
}

// Resources implements ResMap.
func (m *resWrangler) Resources() []*resource.Resource {
	tmp := make([]*resource.Resource, len(m.rList))
	copy(tmp, m.rList)
	return tmp
}

// GetById implements ResMap.
func (m *resWrangler) GetById(id resid.ResId) *resource.Resource {
	if i, ok := m.rIndex[id]; ok {
		return m.rList[i]
	}
	return nil
}

// Append implements ResMap.
func (m *resWrangler) Append(res *resource.Resource) error {
	return m.AppendWithId(res.Id(), res)
}

// Remove implements ResMap.
func (m *resWrangler) Remove(adios resid.ResId) error {
	tmp := newOne()
	for i, r := range m.rList {
		id, err := m.idMappingToIndex(i)
		if err != nil {
			return errors.Wrap(err, "assumption failure in remove")
		}
		if id != adios {
			tmp.AppendWithId(id, r)
		}
	}
	if tmp.Size() != m.Size()-1 {
		return fmt.Errorf("id %s not found in removal", adios)
	}
	m.rIndex = tmp.rIndex
	m.rList = tmp.rList
	return nil
}

// AppendWithId implements ResMap.
func (m *resWrangler) AppendWithId(id resid.ResId, res *resource.Resource) error {
	if already, ok := m.rIndex[id]; ok {
		return fmt.Errorf(
			"attempt to add res %s at id %s; that id already maps to %d",
			res, id, already)
	}
	i := m.indexOfResource(res)
	if i >= 0 {
		return fmt.Errorf(
			"attempt to add res %s that is already held",
			res)
	}
	m.rList = append(m.rList, res)
	m.rIndex[id] = len(m.rList) - 1
	return nil
}

// ReplaceResource implements ResMap.
func (m *resWrangler) ReplaceResource(
	id resid.ResId, newGuy *resource.Resource) error {
	insertAt, ok := m.rIndex[id]
	if !ok {
		return fmt.Errorf(
			"attempt to reset resource at id %s; that id not used", id)
	}
	existingSpot := m.indexOfResource(newGuy)
	if insertAt == existingSpot {
		// Be idempotent.
		return nil
	}
	if existingSpot >= 0 {
		return fmt.Errorf(
			"the new resource %s is already present", newGuy.Id())
	}
	m.rList[insertAt] = newGuy
	return nil
}

// AsMap implements ResMap.
func (m *resWrangler) AsMap() map[resid.ResId]*resource.Resource {
	result := make(map[resid.ResId]*resource.Resource, m.Size())
	for id, i := range m.rIndex {
		result[id] = m.rList[i]
	}
	return result
}

// AllIds implements ResMap.
func (m *resWrangler) AllIds() (ids []resid.ResId) {
	ids = make([]resid.ResId, m.Size())
	i := 0
	for id := range m.rIndex {
		ids[i] = id
		i++
	}
	return
}

// Debug implements ResMap.
func (m *resWrangler) Debug(title string) {
	fmt.Println("--------------------------- " + title)
	firstObj := true
	for i, r := range m.rList {
		if firstObj {
			firstObj = false
		} else {
			fmt.Println("---")
		}
		fmt.Printf("# %d  %s\n", i, m.debugIdMappingToIndex(i))
		blob, err := yaml.Marshal(r.Map())
		if err != nil {
			panic(err)
		}
		fmt.Println(string(blob))
	}
}

func (m *resWrangler) debugIdMappingToIndex(i int) string {
	id, err := m.idMappingToIndex(i)
	if err != nil {
		return err.Error()
	}
	return id.String()
}

func (m *resWrangler) idMappingToIndex(i int) (resid.ResId, error) {
	var foundId resid.ResId
	found := false
	for id, index := range m.rIndex {
		if index == i {
			if found {
				return foundId, fmt.Errorf("found multiple")
			}
			found = true
			foundId = id
		}
	}
	if !found {
		return foundId, fmt.Errorf("cannot find index %d", i)
	}
	return foundId, nil
}

type IdMatcher func(resid.ResId) bool

// GetMatchingIds implements ResMap.
func (m *resWrangler) GetMatchingIds(matches IdMatcher) []resid.ResId {
	var result []resid.ResId
	for id := range m.rIndex {
		if matches(id) {
			result = append(result, id)
		}
	}
	return result
}

// AsYaml implements ResMap.
func (m *resWrangler) AsYaml() ([]byte, error) {
	firstObj := true
	var b []byte
	buf := bytes.NewBuffer(b)
	for _, res := range m.Resources() {
		out, err := yaml.Marshal(res.Map())
		if err != nil {
			return nil, err
		}
		if firstObj {
			firstObj = false
		} else {
			if _, err = buf.WriteString("---\n"); err != nil {
				return nil, err
			}
		}
		if _, err = buf.Write(out); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// ErrorIfNotEqualSets implements ResMap.
func (m *resWrangler) ErrorIfNotEqualSets(other ResMap) error {
	m2, ok := other.(*resWrangler)
	if !ok {
		panic("bad cast")
	}
	if m.Size() != m2.Size() {
		return fmt.Errorf(
			"lists have different number of entries: %#v doesn't equal %#v",
			m.rList, m2.rList)
	}
	for id, i := range m.rIndex {
		r1 := m.rList[i]
		r2 := m2.GetById(id)
		if r2 == nil {
			return fmt.Errorf("id in self missing from other; id: %s", id)
		}
		if !r1.KunstructEqual(r2) {
			return fmt.Errorf(
				"kuns equal mismatch: \n -- %s,\n -- %s\n\n--\n%#v\n------\n%#v\n",
				r1, r2, r1, r2)
		}
	}
	return nil
}

type resCopier func(r *resource.Resource) *resource.Resource

// ShallowCopy implements ResMap.
func (m *resWrangler) ShallowCopy() ResMap {
	return m.makeCopy(
		func(r *resource.Resource) *resource.Resource {
			return r
		})
}

// DeepCopy implements ResMap.
func (m *resWrangler) DeepCopy() ResMap {
	return m.makeCopy(
		func(r *resource.Resource) *resource.Resource {
			return r.DeepCopy()
		})
}

// makeCopy copies the ResMap.
func (m *resWrangler) makeCopy(copier resCopier) ResMap {
	result := &resWrangler{}
	result.rIndex = make(map[resid.ResId]int, m.Size())
	result.rList = make([]*resource.Resource, m.Size())
	for i, r := range m.rList {
		result.rList[i] = copier(r)
		id, err := m.idMappingToIndex(i)
		if err != nil {
			panic("corrupt index map")
		}
		result.rIndex[id] = i
	}
	return result
}

// ResourcesThatCouldReference implements ResMap.
func (m *resWrangler) ResourcesThatCouldReference(inputId resid.ResId) ResMap {
	if inputId.Gvk().IsClusterKind() {
		return m
	}
	result := New()
	for id, i := range m.rIndex {
		if id.Gvk().IsClusterKind() || id.Namespace() == inputId.Namespace() &&
			id.HasSameLeftmostPrefix(inputId) &&
			id.HasSameRightmostSuffix(inputId) {
			err := result.AppendWithId(id, m.rList[i])
			if err != nil {
				panic(err)
			}
		}
	}
	return result
}

// AppendAll implements ResMap.
func (m *resWrangler) AppendAll(other ResMap) error {
	if other == nil {
		return nil
	}
	w2, ok := other.(*resWrangler)
	if !ok {
		panic("bad cast")
	}
	for i, res := range w2.Resources() {
		id, err := w2.idMappingToIndex(i)
		if err != nil {
			panic("map is irrecoverably corrupted; " + err.Error())
		}
		err = m.AppendWithId(id, res)
		if err != nil {
			return err
		}
	}
	return nil
}

// AbsorbAll implements ResMap.
func (m *resWrangler) AbsorbAll(other ResMap) error {
	if other == nil {
		return nil
	}
	w2, ok := other.(*resWrangler)
	if !ok {
		panic("bad cast")
	}
	for i, r := range w2.Resources() {
		id, err := w2.idMappingToIndex(i)
		if err != nil {
			panic("map is irrecoverably corrupted; " + err.Error())
		}
		err = m.appendReplaceOrMerge(id, r)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *resWrangler) appendReplaceOrMerge(
	idForRes resid.ResId, res *resource.Resource) error {
	matchedId := m.GetMatchingIds(idForRes.GvknEquals)
	switch len(matchedId) {
	case 0:
		switch res.Behavior() {
		case types.BehaviorMerge, types.BehaviorReplace:
			return fmt.Errorf(
				"id %#v does not exist; cannot merge or replace", idForRes)
		default:
			// presumably types.BehaviorCreate
			err := m.AppendWithId(idForRes, res)
			if err != nil {
				return err
			}
		}
	case 1:
		mId := matchedId[0]
		old := m.GetById(mId)
		if old == nil {
			return fmt.Errorf("id lookup failure")
		}
		switch res.Behavior() {
		case types.BehaviorReplace:
			res.Replace(old)
			err := m.ReplaceResource(mId, res)
			if err != nil {
				return err
			}
		case types.BehaviorMerge:
			res.Merge(old)
			err := m.ReplaceResource(mId, res)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf(
				"id %#v exists; must merge or replace", idForRes)
		}
	default:
		return fmt.Errorf(
			"found multiple objects %v that could accept merge of %v",
			matchedId, idForRes)
	}
	return nil
}
