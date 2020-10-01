package builtins_qlik

import (
	"log"
	"path/filepath"

	"sigs.k8s.io/kustomize/api/builtins_qlik/utils"
	"sigs.k8s.io/kustomize/api/filters/fsslice"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filtersutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
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

			err := filtersutil.ApplyToJSON(fullPathFilter{
				rootDir: p.RootDir,
				fsSlice: p.FieldSpecs,
			}, r)
			if err != nil {
				p.logger.Printf("error updating path for root dir: %v, error: %v\n", p.RootDir, err)
				return err
			}
		}
	}
	return nil
}

func NewFullPathPlugin() resmap.TransformerPlugin {
	return &FullPathPlugin{logger: utils.GetLogger("FullPathPlugin")}
}

type fullPathFilter struct {
	fsSlice types.FsSlice
	rootDir string
}

func (f fullPathFilter) Filter(nodes []*kyaml.RNode) ([]*kyaml.RNode, error) {
	_, err := kio.FilterAll(kyaml.FilterFunc(
		func(node *kyaml.RNode) (*kyaml.RNode, error) {
			if err := node.PipeE(fsslice.Filter{
				FsSlice: f.fsSlice,
				SetValue: func(n *kyaml.RNode) error {
					if !filepath.IsAbs(n.YNode().Value) {
						n.YNode().Value = filepath.Join(f.rootDir, n.YNode().Value)
					}
					return nil
				},
			}); err != nil {
				return nil, err
			}
			return node, nil
		})).Filter(nodes)
	return nodes, err
}
