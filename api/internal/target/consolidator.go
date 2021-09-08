package target

import (
	"bytes"
	"path/filepath"

	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/kustomize/kyaml/yaml/merge2"
)

type CompositionConsolidator struct {
	ldr         ifc.Loader
	content     []byte
	composition *types.Composition
}

func (cc *CompositionConsolidator) Read() error {
	c := &types.Composition{}
	d := yaml.NewDecoder(bytes.NewReader(cc.content))
	d.KnownFields(true)
	if err := d.Decode(c); err != nil {
		return errors.WrapPrefixf(err, "failed to decode Composition at path %s", cc.ldr.Root())
	}
	if err := c.Default(); err != nil {
		return errors.WrapPrefixf(err, "failed to apply defaults to Composition at path %s", cc.ldr.Root())
	}
	if err := c.Validate(); err != nil {
		return errors.WrapPrefixf(err, "invalid Composition at path %s", cc.ldr.Root())
	}
	cc.composition = c
	return nil
}

// New creates a consolidator for a Composition being imported by this consolidator's Composition
// The new consolidator uses a loader created from the parent's loader, to enforce root restrictions.
func (cc *CompositionConsolidator) New(path string) (*CompositionConsolidator, error) {
	childLdr, err := cc.ldr.New(path)
	if err != nil {
		return nil, err
	}
	childContent, kind, err := loadKustFile(childLdr)
	if err != nil {
		return nil, err
	}
	if kind != types.CompositionKind {
		return nil, errors.Errorf("cannot import transformers from a %s", kind)
	}
	return &CompositionConsolidator{ldr: childLdr, content: childContent}, nil
}

// Consolidate prepares the composition for execution by resolving imported transformers.
func (cc *CompositionConsolidator) Consolidate() (*types.Composition, error) {
	c := cc.composition
	transformers := make([]types.Transformer, len(c.Transformers))
	ids := make([]resid.ResId, len(c.Transformers))
	for i, transformer := range c.Transformers {
		transformers[i] = transformer
		ids[i] = transformer.ID()
	}
	transformerMap, err := newTransformerMap(ids)
	if err != nil {
		return nil, errors.WrapPrefixf(err, "transformers field is invalid")
	}

	for _, transformerSource := range c.TransformersFrom {
		childConsolidator, err := cc.New(transformerSource.Path)
		if err != nil {
			return nil, err
		}
		if err := childConsolidator.Read(); err != nil {
			return nil, err
		}
		childComp, err := childConsolidator.Consolidate()
		if err != nil {
			return nil, errors.WrapPrefixf(err, "failed to import transformers from %s at path %s", types.CompositionKind, childConsolidator.Root())
		}
		for _, imported := range childComp.Transformers {
			// TODO: find a better approach?
			if err := fixFilePaths(&imported.TransformerConfig, childConsolidator.Root()); err != nil {
				return nil, errors.WrapPrefixf(err, "fixing file paths in %s", imported.ID())
			}
		}

		importedTransformers, err := applyOverrides(c.TransformerOverrides, childComp.Transformers)
		if err != nil {
			return nil, errors.Wrap(err)
		}

		// Make sure this import isn't generating duplicates
		if err := ensureNoDuplicates(importedTransformers, transformerMap); err != nil {
			return nil, err
		}

		// Append the new transformers
		switch transformerSource.ImportMode {
		case "append":
			transformers = append(transformers, importedTransformers...)
		case "prepend", "":
			transformers = append(importedTransformers, transformers...)
		default:
			return nil, errors.Errorf("invalid import mode %q specified for transformers from %s", transformerSource.ImportMode, childConsolidator.Root())
		}
	}

	sorted, err := transformerSorter{ordering: c.TransformerOrder}.sortTransformers(transformers)
	if err != nil {
		return nil, err
	}
	return &types.Composition{Transformers: sorted}, nil
}

func ensureNoDuplicates(importedTransformers []types.Transformer, transformerMap transformerOrderMap) error {
	for _, imported := range importedTransformers {
		// We don't use `transformerOrderMap#getUnderspecified` here because we want exact matches only
		if _, ok := transformerMap[imported.ID()]; ok {
			return errors.Errorf("importing composition resulted in duplicate %q transformer", imported)
		}
	}
	return nil
}

// Root returns the root directory containing this consolidator's Composition
func (cc *CompositionConsolidator) Root() string {
	return cc.ldr.Root()
}

func fixFilePaths(config *yaml.Node, rootPath string) error {
	fields, err := yaml.Lookup("files").Filter(yaml.NewRNode(config))
	if err != nil {
		return err
	}
	if fields == nil || yaml.ErrorIfInvalid(fields, yaml.SequenceNode) != nil {
		return nil
	}
	return fields.VisitElements(func(node *yaml.RNode) error {
		// only target scalar nodes
		if err := yaml.ErrorIfInvalid(node, yaml.ScalarNode); err != nil {
			return nil
		}
		// don't touch paths that are already absolute
		originalPath := node.YNode().Value
		if filepath.IsAbs(originalPath) {
			return nil
		}
		node.YNode().Value = filepath.Join(rootPath, originalPath)
		return nil
	})
}

func applyOverrides(overrides []types.Transformer, importedTransformers []types.Transformer) ([]types.Transformer, error) {
	if len(overrides) == 0 {
		return importedTransformers, nil
	}
	ids := make([]resid.ResId, len(importedTransformers))
	for i := range importedTransformers {
		ids[i] = importedTransformers[i].ID()
	}
	transformerMap, err := newTransformerMap(ids)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	for _, override := range overrides {
		idx, ok := transformerMap[override.ID()]
		if !ok {
			return nil, errors.Errorf("no transformer found for override %q", override)
		}

		merged, err := merge2.Merge(
			yaml.NewRNode(&override.TransformerConfig),
			yaml.NewRNode(&importedTransformers[*idx].TransformerConfig),
			yaml.MergeOptions{})
		if err != nil {
			return nil, errors.WrapPrefixf(err, "failed to merge TransformerConfig for transformer %q",
				override)
		}
		importedTransformers[*idx].TransformerConfig = *merged.YNode()

		// If runtime fields are being overridden, replace the fields and the runtime itself
		// TODO: This makes big assumptions about the fields in RuntimeConfig. Make the merge filter handle it?
		if override.RuntimeConfig.Container.Image != "" || override.RuntimeConfig.Starlark.Path != "" || override.RuntimeConfig.Exec.Path != "" {
			importedTransformers[*idx].RuntimeConfig = override.RuntimeConfig
		}
	}
	return importedTransformers, nil
}

// transformerSorter efficiently sorts transformers into an ordering specified by a list of transformer metadata.
type transformerSorter struct {
	ordering      []resid.ResId
	orderMapCache transformerOrderMap
}

// transformerOrderMap is a map of transformer identifiers to positions.
// It is used to look up the position of a related transformer in a secondary list,
// for example to find the transformer targeted by an override,
// or to find the desired position for a transformer based on an ordering list.
type transformerOrderMap map[resid.ResId]*int

// set inserts a transformer into the order map.
// It returns an error if the transformer already exists in the map.
func (m transformerOrderMap) set(id resid.ResId, position int) error {
	if id.Name == "" {
		return errors.Errorf("transformer identifier %q must include a name", id)
	}
	if _, ok := m[id]; !ok {
		m[id] = &position
		return nil
	}
	return errors.Errorf("list contains multiple %q transformers", id)
}

// getUnderspecified retrieves a transformer from the transformer map.
// The identifier used for retrieval can be underspecified, but must include a name.
// For example `{ Name: foo }` can retrieve `{ Name: foo, Kind: bar, APIVersion: baz }` as long as
// that is the only entry in the transformer map by that name.
// TODO: We're defaulting the name field to lowercase kind. Does this need to be changed to Kind?
func (m transformerOrderMap) getUnderspecified(id resid.ResId) (int, error) {
	var positions []int
	for ordered, i := range m {
		if id.IsSelectedBy(ordered) {
			positions = append(positions, *i)
		}
	}

	if len(positions) > 1 {
		return -1, errors.Errorf("multiple entries matched")
	}
	if len(positions) == 0 {
		return -1, errors.Errorf("no match found")
	}
	return positions[0], nil
}

// newTransformerMap creates a mapping of identifiers to positions.
// It returns an error if the identifier list contains duplicates.
func newTransformerMap(ordering []resid.ResId) (transformerOrderMap, error) {
	transformerMap := transformerOrderMap{}
	for i, id := range ordering {
		if err := transformerMap.set(id, i); err != nil {
			return nil, err
		}
	}
	return transformerMap, nil
}

// sortOrder creates and caches an indexable mapping of the object's identity to its desired position.
// Since the transformerOrderMap rejects duplicate insertions, this also validates the uniqueness of the identifier list.
func (s *transformerSorter) sortOrder() (transformerOrderMap, error) {
	if s.orderMapCache != nil {
		return s.orderMapCache, nil
	}
	order, err := newTransformerMap(s.ordering)
	if err != nil {
		return nil, errors.WrapPrefixf(err, "failed to sort transformers")
	}
	s.orderMapCache = order
	return s.orderMapCache, nil
}

// sortTransformers sorts the input transformers to match transformerSorter's ordering field.
// It returns an error if the ordering identifiers contain duplicates or do not
// match up perfectly with the transformers to be sorted.
func (s transformerSorter) sortTransformers(transformers []types.Transformer) ([]types.Transformer, error) {
	orderMap, err := s.sortOrder()
	if err != nil {
		return nil, errors.Wrap(err)
	}
	if len(orderMap) == 0 {
		return transformers, nil
	}
	if len(orderMap) > len(transformers) {
		return nil, errors.Errorf("transformer order list contains too many entries")
	}
	if len(orderMap) < len(transformers) {
		return nil, errors.Errorf("transformer order list contains too few entries")
	}

	reordered := make([]types.Transformer, len(transformers))
	for _, transformer := range transformers {
		insertIdx, err := orderMap.getUnderspecified(transformer.ID())
		if err != nil {
			return nil, errors.WrapPrefixf(err, "unable to find position for transformer %q", transformer)
		}
		reordered[insertIdx] = transformer
	}

	return reordered, nil
}
