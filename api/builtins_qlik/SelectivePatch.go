package builtins_qlik

import (
	"log"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/api/builtins"
	"sigs.k8s.io/kustomize/api/builtins_qlik/utils"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

type SelectivePatchPlugin struct {
	Enabled    bool          `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Default    bool          `json:"default,omitempty" yaml:"default,omitempty"`
	Patches    []types.Patch `json:"patches,omitempty" yaml:"patches,omitempty"`
	Defaults   []types.Patch `json:"defaults,omitempty" yaml:"defaults,omitempty"`
	ts         []resmap.Transformer
	tsDefaults []resmap.Transformer
	logger     *log.Logger
}

func (p *SelectivePatchPlugin) makeIndividualPatches(pat types.Patch) ([]byte, error) {
	var s struct {
		types.Patch
	}
	s.Patch = pat
	return yaml.Marshal(s)
}

func (p *SelectivePatchPlugin) Config(h *resmap.PluginHelpers, c []byte) (err error) {
	// To avoid https://github.com/kubernetes-sigs/kustomize/blob/master/docs/FAQ.md#security-file-foo-is-not-in-or-below-bar
	// start of work around
	fSys := filesys.MakeFsOnDisk()
	newLdr, err := loader.NewLoader(loader.RestrictionNone, h.Loader().Root(), fSys)
	if err != nil {
		p.logger.Printf("error creating a new loader from default loader, error: %v\n", err)
		return errors.Wrapf(err, "Cannot create new loader from default loader")
	}
	// End of work around
	if err := yaml.Unmarshal(c, p); err != nil {
		p.logger.Printf("error unmarshalling bytes: %v, error: %v\n", string(c), err)
		return errors.Wrapf(err, "Inside unmarshal "+string(c))
	}
	for _, v := range p.Patches {
		//fmt.Println(v.Path)
		b, _ := p.makeIndividualPatches(v)
		prefixer := builtins.NewPatchTransformerPlugin()
		err = prefixer.Config(resmap.NewPluginHelpers(newLdr, h.Validator(), h.ResmapFactory()), b)
		if err != nil {
			p.logger.Printf("error executing PatchTransformerPlugin.Config(), error: %v\n", err)
			return errors.Wrapf(err, "stringprefixer configure")
		}
		p.ts = append(p.ts, prefixer)

	}
	for _, v := range p.Defaults {
		//fmt.Println(v.Path)
		b, _ := p.makeIndividualPatches(v)
		prefixer := builtins.NewPatchTransformerPlugin()
		err = prefixer.Config(resmap.NewPluginHelpers(newLdr, h.Validator(), h.ResmapFactory()), b)
		if err != nil {
			p.logger.Printf("error executing PatchTransformerPlugin.Config(), error: %v\n", err)
			return errors.Wrapf(err, "stringprefixer configure")
		}
		p.tsDefaults = append(p.tsDefaults, prefixer)

	}
	return nil
}

func (p *SelectivePatchPlugin) Transform(m resmap.ResMap) error {
	if p.Enabled {
		for _, t := range p.ts {
			err := t.Transform(m)
			if err != nil {
				p.logger.Printf("error executing Transform(), error: %v\n", err)
				return err
			}
		}
	}
	if p.Default {
		for _, t := range p.tsDefaults {
			err := t.Transform(m)
			if err != nil {
				p.logger.Printf("error executing Transform(), error: %v\n", err)
				return err
			}
		}
	}
	return nil
}

func NewSelectivePatchPlugin() resmap.TransformerPlugin {
	return &SelectivePatchPlugin{logger: utils.GetLogger("SelectivePatchPlugin")}
}
