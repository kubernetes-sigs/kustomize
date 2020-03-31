package builtins_qlik

import (
	"fmt"
	"log"
	"sigs.k8s.io/kustomize/api/builtins_qlik/utils"

	"sigs.k8s.io/kustomize/api/builtins"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/yaml"
)

type SuperConfigMapPlugin struct {
	Data   map[string]string `json:"data,omitempty" yaml:"data,omitempty"`
	logger *log.Logger
	builtins.ConfigMapGeneratorPlugin
	SuperMapPluginBase
}

func (p *SuperConfigMapPlugin) Config(h *resmap.PluginHelpers, c []byte) (err error) {
	p.SuperMapPluginBase = NewBase(h.ResmapFactory(), p)
	p.Data = make(map[string]string)
	err = yaml.Unmarshal(c, p)
	if err != nil {
		p.logger.Printf("error unmarshalling yaml, error: %v\n", err)
		return err
	}
	err = p.SuperMapPluginBase.SetupTransformerConfig(h.Loader())
	if err != nil {
		p.logger.Printf("error setting up transformer config, error: %v\n", err)
		return err
	}
	return p.ConfigMapGeneratorPlugin.Config(h, c)
}

func (p *SuperConfigMapPlugin) Generate() (resmap.ResMap, error) {
	for k, v := range p.Data {
		p.LiteralSources = append(p.LiteralSources, fmt.Sprintf("%v=%v", k, v))
	}
	return p.ConfigMapGeneratorPlugin.Generate()
}

func (p *SuperConfigMapPlugin) Transform(m resmap.ResMap) error {
	return p.SuperMapPluginBase.Transform(m)
}

func (p *SuperConfigMapPlugin) GetLogger() *log.Logger {
	return p.logger
}

func (p *SuperConfigMapPlugin) GetName() string {
	return p.ConfigMapGeneratorPlugin.Name
}

func (p *SuperConfigMapPlugin) GetNamespace() string {
	return p.ConfigMapGeneratorPlugin.Namespace
}

func (p *SuperConfigMapPlugin) SetNamespace(namespace string) {
	p.ConfigMapGeneratorPlugin.Namespace = namespace
	p.ConfigMapGeneratorPlugin.GeneratorArgs.Namespace = namespace
}

func (p *SuperConfigMapPlugin) GetType() string {
	return "ConfigMap"
}

func (p *SuperConfigMapPlugin) GetConfigData() map[string]string {
	return p.Data
}

func (p *SuperConfigMapPlugin) ShouldBase64EncodeConfigData() bool {
	return false
}

func (p *SuperConfigMapPlugin) GetDisableNameSuffixHash() bool {
	return p.ConfigMapGeneratorPlugin.DisableNameSuffixHash
}

func NewSuperConfigMapTransformerPlugin() resmap.TransformerPlugin {
	return &SuperConfigMapPlugin{logger: utils.GetLogger("SuperConfigMapTransformerPlugin")}
}

func NewSuperConfigMapGeneratorPlugin() resmap.GeneratorPlugin {
	return &SuperConfigMapPlugin{logger: utils.GetLogger("SuperConfigMapGeneratorPlugin")}
}
