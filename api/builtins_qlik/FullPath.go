package builtins_qlik

import (
	"fmt"
	"log"
	"path/filepath"

	"sigs.k8s.io/kustomize/api/builtins_qlik/utils"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/transform"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

type FullPathPlugin struct {
	RootDir    string
	FieldSpecs []types.FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
	logger     *log.Logger
}

func (p *FullPathPlugin) Config(h *resmap.PluginHelpers, c []byte) (err error) {
	p.RootDir = h.Loader().Root()
	p.FieldSpecs = make([]types.FieldSpec, 0)

	return yaml.Unmarshal(c, p)
}

func (p *FullPathPlugin) Transform(m resmap.ResMap) error {
	for _, r := range m.Resources() {
		id := r.OrgId()
		for _, fieldSpec := range p.FieldSpecs {
			if !id.IsSelected(&fieldSpec.Gvk) {
				continue
			}

			err := transform.MutateField(
				r.Map(),
				fieldSpec.PathSlice(),
				fieldSpec.CreateIfNotPresent,
				p.computePath)
			if err != nil {
				p.logger.Printf("error executing transformers.MutateField(), error: %v\n", err)
				return err
			}
		}
	}
	return nil
}

func (p *FullPathPlugin) computePath(in interface{}) (interface{}, error) {
	relativePath, ok := in.(string)
	if !ok {
		return nil, fmt.Errorf("%#v is expected to be %T", in, relativePath)
	}

	if filepath.IsAbs(relativePath) {
		return relativePath, nil
	} else {
		return filepath.Join(p.RootDir, relativePath), nil
	}
}

func NewFullPathPlugin() resmap.TransformerPlugin {
	return &FullPathPlugin{logger: utils.GetLogger("FullPathPlugin")}
}
