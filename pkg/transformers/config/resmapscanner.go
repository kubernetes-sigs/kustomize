package config

import (
	"fmt"
	"regexp"
	"strings"

	"sigs.k8s.io/kustomize/v3/pkg/expansion"
	"sigs.k8s.io/kustomize/v3/pkg/gvk"
	"sigs.k8s.io/kustomize/v3/pkg/resid"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/types"
)

type ResMapScanner struct {
	manualVars types.VarSet
	manualRefs fsSlice
	autoVars   types.VarSet
	autoRefs   fsSlice
}

type detectedRef struct {
	pathSlice     []string
	detectedNames []string
}

// KindRegistry contains a cache of the GVK used in the resources.
var KindRegistry = map[string]gvk.Gvk{}
var IndexRegex *regexp.Regexp = regexp.MustCompile(`<[0-9]+>`)

const (
	ParentInline string = "parent-inline"
	Dot          string = "."
	Slash        string = "/"
)

// NewResMapScanner returns a new ResMapScanner
// that detects $(Kind.name.path) style variables with values.
func NewResMapScanner(userVars types.VarSet, userRefs fsSlice) *ResMapScanner {
	return &ResMapScanner{
		manualVars: userVars.Copy(),
		manualRefs: userRefs,
		autoVars:   types.NewVarSet(),
		autoRefs:   fsSlice{},
	}
}

// DiscoveredVars returns the list of Var to add to the
// consolidated var section of the kustomization.yaml(s)
// This allows the user to not have to do that manually.
func (rv *ResMapScanner) DiscoveredVars() types.VarSet {
	return rv.autoVars
}

// DiscoveredConfig returns a TransformerConfig containing
// a consolidated VarReference sections.
// This allows the user to not have to do that manually.
func (rv *ResMapScanner) DiscoveredConfig() *TransformerConfig {
	return &TransformerConfig{
		VarReference: rv.autoRefs,
	}
}

func collectReferences(in interface{}) []detectedRef {
	return collectReferencesFromPath(in, []string{})
}

// Walk the path (curPath) of a resource (in) and collect
// the detected Var in the refMap object
func collectReferencesFromPath(in interface{}, curPath []string) []detectedRef {
	switch typedIn := in.(type) {
	case []interface{}:
		// Check each member of the slice for variable references.
		var allRefs []detectedRef
		for _, v := range typedIn {
			childRefs := collectReferencesFromPath(v, curPath)
			allRefs = append(allRefs, childRefs...)
		}
		return allRefs
	case map[string]interface{}:
		// Check each member of the map for variable references.
		var allRefs []detectedRef
		for key, v := range typedIn {
			childRefs := collectReferencesFromPath(v, append(curPath, key))
			allRefs = append(allRefs, childRefs...)
		}
		return allRefs
	case string:
		// Look for potential variable references within the string
		detectedNames := expansion.Detect(typedIn)
		if len(detectedNames) == 0 {
			return nil
		}
		pathSlice := curPath
		if len(detectedNames) == 1 && pathSlice[len(pathSlice)-1] == ParentInline {
			pathSlice = pathSlice[:len(pathSlice)-1]
		}

		normalizedPathSlice := normalizePathSlice(pathSlice)

		return []detectedRef{{
			pathSlice:     normalizedPathSlice,
			detectedNames: detectedNames,
		}}
	default:
		return nil
	}
}

// buildVar checks the syntax of the detected var and assert
// that it matches the $(Kind.name.path) pattern.
func buildVar(detectedName string) (*types.Var, error) {
	// Note:
	// $(Ingress.name.metadata.annotations['ingress.auth-secretkubernetes.io/auth-secret'])
	// will be split in s in a strange way, but the fieldpath is rebuilt
	// properly a few line further.
	s := strings.Split(detectedName, Dot)

	if len(s) < 3 {
		return nil, fmt.Errorf("var %s does not match expected "+
			"pattern $(Kind.name.path)", detectedName)
	}

	kind := s[0]
	name := s[1]
	fieldPath := strings.Join(s[2:], Dot)

	if _, ok := KindRegistry[kind]; !ok {
		// We don't have a entry for that kind.
		return nil, fmt.Errorf("var $(%s) referencing an unknown  "+
			"or conflicting Kind %s", detectedName, kind)
	}

	group := KindRegistry[kind].Group
	version := KindRegistry[kind].Version
	objref := types.Target{
		Gvk: gvk.Gvk{
			Group:   group,
			Version: version,
			Kind:    kind,
		},
		APIVersion: group + Slash + version,
		Name:       name,
	}
	fieldref := types.FieldSelector{
		FieldPath: fieldPath,
	}
	tVar := &types.Var{
		Name:     detectedName,
		ObjRef:   objref,
		FieldRef: fieldref,
	}

	return tVar, nil
}

// normalizedPathSlice escapes all slashes and removes the indices
func normalizePathSlice(pathSlice []string) []string {
	normalizedPath := []string{}
	for _, elt := range pathSlice {
		// According to fieldspec.PathSlice, need to escape the slash
		normalizedElt := strings.ReplaceAll(elt, Slash, escapedForwardSlash)
		if normalizedElt != "" {
			normalizedPath = append(normalizedPath, normalizedElt)
		}
	}
	return normalizedPath
}

// buildFieldSpec builds FieldSpec to add to the VarReference.
func buildFieldSpec(id resid.ResId, pathSlice []string) *FieldSpec {
	// varReference paths use / as separator.
	path := strings.Join(pathSlice, Slash)

	return &FieldSpec{
		Gvk:  gvk.FromKind(id.Kind),
		Path: path,
	}
}

// BuildAutoConfig scans the ResMap and detects the $(Kind.name.path) pattern
func (rv *ResMapScanner) BuildAutoConfig(m resmap.ResMap) {
	// Cache the GVK used by the project
	for _, res := range m.Resources() {
		// TODO(jeb): We need to check for conflicting information
		// in t.manualVars and in the resources.
		// TODO(jeb): We also need to check that for a specific
		// kind, the apiversion is used everywhere.
		groupversionkind := res.GetGvk()
		KindRegistry[groupversionkind.Kind] = groupversionkind
	}

	for _, res := range m.Resources() {
		// Walk the resource to collect the variables
		// and where they're referenced
		referenceMap := collectReferences(res.Map())

		for _, detected := range referenceMap {
			varReference := buildFieldSpec(res.OrgId(), detected.pathSlice)
			for _, detectedName := range detected.detectedNames {
				tVar, err := buildVar(detectedName)
				if err != nil {
					// Algorithm can't deal with the detected variable name,
					// probably because it is not actually a variable.
					continue
				}

				// Check if this could be a variable by looking for the corresponding resource.
				// First try in the same namespace as the current resource.
				targetId := resid.NewResIdWithNamespace(tVar.ObjRef.GVK(), tVar.ObjRef.Name, res.CurId().Namespace)
				idMatcher := targetId.Equals
				matched := m.GetMatchingResourcesByCurrentId(idMatcher)

				if len(matched) == 0 {
					// Look using original namespace and name
					targetId = resid.NewResIdWithNamespace(tVar.ObjRef.GVK(), tVar.ObjRef.Name, res.OrgId().Namespace)
					idMatcher = targetId.Equals
					matched = m.GetMatchingResourcesByOriginalId(idMatcher)
				}

				if len(matched) == 0 {
					// Look using current name but no namespace
					targetId = resid.NewResId(tVar.ObjRef.GVK(), tVar.ObjRef.Name)
					idMatcher = targetId.GvknEquals
					matched = m.GetMatchingResourcesByCurrentId(idMatcher)
				}

				if len(matched) == 0 {
					// Look using original name but no namespace
					targetId = resid.NewResId(tVar.ObjRef.GVK(), tVar.ObjRef.Name)
					idMatcher = targetId.GvknEquals
					matched = m.GetMatchingResourcesByOriginalId(idMatcher)
				}

				// If not, this is probably not a variable.
				if len(matched) == 1 {
					_, err := matched[0].GetFieldValue(tVar.FieldRef.FieldPath)
					if err != nil {
						// We detected $(validkind.validname.invalidfieldspec)
						// This is probably not a variable.
						continue
					}
					tVar.ObjRef.Name = matched[0].OrgId().Name
					tVar.ObjRef.Namespace = matched[0].OrgId().Namespace

					// If either the var or the varReference were previously added
					// by the user, let's trust the user and ignore that variable.
					// if rv.manualVars.Contains(*tVar) || rv.manualRefs.contains(*varReference) {
					err = rv.manualVars.Absorb(*tVar)
					if err != nil {
						// Detected a GVK for that var which conflicts with the manual entered one.
						// Let's trust the user and ignore that potential variable.
						// TODO(jeb): Probably won't detect potential duplicate
						// between if fieldSpec have differnet format: spec.foo[bar] and spec.foo.bar`
						continue
					}

					// TODO(jeb): Probably won't detect potential duplicates
					// if fieldSpecs have different formats,
					// e.g. spec.foo[bar] and spec.foo.bar
					rv.autoVars.Merge(*tVar)
					// we can safely ignore this error since we've already
					// checked for collisions
					rv.autoRefs, _ = rv.autoRefs.mergeOne(*varReference)
					matched[0].AbsorbRefVarName(*tVar)
				}
			}
		}
	}
}
