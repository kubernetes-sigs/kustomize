package builtins_qlik

import (
	"log"
	"sigs.k8s.io/kustomize/api/builtins_qlik/utils"

	"sigs.k8s.io/kustomize/api/internal/accumulator"
	"sigs.k8s.io/kustomize/api/internal/plugins/builtinconfig"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

type SuperVarsPlugin struct {
	Vars           []types.Var `json:"vars,omitempty" yaml:"vars,omitempty"`
	Configurations []string    `json:"configurations,omitempty" yaml:"configurations,omitempty"`
	tConfig        *builtinconfig.TransformerConfig
	logger         *log.Logger
}

func (p *SuperVarsPlugin) Config(h *resmap.PluginHelpers, c []byte) (err error) {
	p.Vars = make([]types.Var, 0)
	p.Configurations = make([]string, 0)

	err = yaml.Unmarshal(c, p)
	if err != nil {
		p.logger.Printf("error unmarshelling plugin config yaml, error: %v\n", err)
		return err
	}

	p.tConfig = &builtinconfig.TransformerConfig{}
	p.tConfig, err = builtinconfig.MakeTransformerConfig(h.Loader(), p.Configurations)
	if err != nil {
		p.logger.Printf("error making transformer config, error: %v\n", err)
		return err
	}

	return nil
}

func (p *SuperVarsPlugin) Transform(m resmap.ResMap) error {
	ac := accumulator.MakeEmptyAccumulator()
	if err := ac.AppendAll(m); err != nil {
		return err
	} else if err := ac.MergeConfig(p.tConfig); err != nil {
		return err
	} else if err := ac.MergeVars(p.Vars); err != nil {
		return err
	} else if err := ac.ResolveVars(); err != nil {
		return err
	}
	return nil
}

func NewSuperVarsPlugin() resmap.TransformerPlugin {
	return &SuperVarsPlugin{logger: utils.GetLogger("SuperVarsPlugin")}
}
