package main

import (
	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

//go:generate pluginator

type plugin struct {
	h                           *resmap.PluginHelpers
	types.ObjectMeta            `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	types.ResourceGeneratorArgs `yaml:",inline" json:",inline"`
}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(h *resmap.PluginHelpers, config []byte) (err error) {
	p.ResourceGeneratorArgs = types.ResourceGeneratorArgs{}
	err = yaml.Unmarshal(config, p)
	p.h = h
	return
}

// TODO: This is implementing Transform instead instead of Generate because it gets used in the
// transformers field, and it is not currently possible for a plugin to implement both.
// We should consider moving plugins to BE filters and using the framework's dispatcher for them.
func (p *plugin) Transform(ra resmap.ResMap) error {
	// TODO: this only works with files and is duplicated from the "real" code in api
	// Importing api/kusty creates an import cycle.
	// Besides being sketchy, shelling out requires reconstructing the possibly virtual filesystem
	// in a tempdir on the real filesystem. It might even invoke a different version, and
	// it won't work well in CI.
	// Ideally the normal kustomize build path for the resources field will use the same generator
	// code, so try extracting it all to be plugin-internal instead of api-internal.
	for _, path := range p.Files {
		if err := p.accumulateFile(ra, path, &resource.Origin{}); err != nil {
			return errors.Wrapf(
				err, "accumulation err='%s'", err.Error())
		}
	}
	return nil
}

func (p *plugin) accumulateFile(
	resMap resmap.ResMap, path string, origin *resource.Origin) error {
	resources, err := p.h.ResmapFactory().FromFile(p.h.Loader(), path)
	if err != nil {
		return errors.Wrapf(err, "accumulating resources from '%s'", path)
	}
	if p.OriginAnnotations {
		origin = origin.Append(path)
		err = resources.AnnotateAll(resource.OriginAnnotation, origin.String())
		if err != nil {
			return errors.Wrapf(err, "cannot add path annotation for '%s'", path)
		}
	}
	err = resMap.AppendAll(resources)
	if err != nil {
		return errors.Wrapf(err, "merging resources from '%s'", path)
	}
	return nil
}
