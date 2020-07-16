package builtins_qlik

import (
	"encoding/base64"
	"fmt"
	"log"

	"sigs.k8s.io/kustomize/api/builtins_qlik/utils"

	"sigs.k8s.io/kustomize/api/builtins"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/yaml"
)

type SuperSecretPlugin struct {
	StringData          map[string]string `json:"stringData,omitempty" yaml:"stringData,omitempty"`
	Data                map[string]string `json:"data,omitempty" yaml:"data,omitempty"`
	aggregateConfigData map[string]string
	logger              *log.Logger
	builtins.SecretGeneratorPlugin
	SuperMapPluginBase
}

func (p *SuperSecretPlugin) Config(h *resmap.PluginHelpers, c []byte) (err error) {
	p.SuperMapPluginBase = NewBase(h.ResmapFactory(), p)
	p.Data = make(map[string]string)
	p.StringData = make(map[string]string)
	err = yaml.Unmarshal(c, p)
	if err != nil {
		p.logger.Printf("error unmarshalling yaml, error: %v\n", err)
		return err
	}
	p.aggregateConfigData, err = p.getAggregateConfigData()
	if err != nil {
		p.logger.Printf("error accumulating config data: %v\n", err)
		return err
	}
	err = p.SuperMapPluginBase.SetupTransformerConfig(h.Loader())
	if err != nil {
		p.logger.Printf("error setting up transformer config, error: %v\n", err)
		return err
	}
	return p.SecretGeneratorPlugin.Config(h, c)
}

func (p *SuperSecretPlugin) getAggregateConfigData() (map[string]string, error) {
	aggregateConfigData := make(map[string]string)
	for k, v := range p.StringData {
		aggregateConfigData[k] = v
	}
	for k, v := range p.Data {
		if decodedValue, err := base64.StdEncoding.DecodeString(v); err != nil {
			p.logger.Printf("error base64 decoding value: %v for key: %v, error: %v\n", v, k, err)
			aggregateConfigData[k] = ""
		} else {
			aggregateConfigData[k] = string(decodedValue)
		}
	}
	return aggregateConfigData, nil
}

func (p *SuperSecretPlugin) Generate() (resmap.ResMap, error) {
	for k, v := range p.aggregateConfigData {
		p.LiteralSources = append(p.LiteralSources, fmt.Sprintf("%v=%v", k, v))
	}
	return p.SecretGeneratorPlugin.Generate()
}

func (p *SuperSecretPlugin) Transform(m resmap.ResMap) error {
	return p.SuperMapPluginBase.Transform(m)
}

func (p *SuperSecretPlugin) GetLogger() *log.Logger {
	return p.logger
}

func (p *SuperSecretPlugin) GetName() string {
	return p.SecretGeneratorPlugin.Name
}

func (p *SuperSecretPlugin) GetNamespace() string {
	return p.SecretGeneratorPlugin.Namespace
}

func (p *SuperSecretPlugin) SetNamespace(namespace string) {
	p.SecretGeneratorPlugin.Namespace = namespace
	p.SecretGeneratorPlugin.GeneratorArgs.Namespace = namespace
}

func (p *SuperSecretPlugin) GetType() string {
	return "Secret"
}

func (p *SuperSecretPlugin) GetConfigData() map[string]string {
	return p.aggregateConfigData
}

func (p *SuperSecretPlugin) ShouldBase64EncodeConfigData() bool {
	return true
}

func (p *SuperSecretPlugin) GetDisableNameSuffixHash() bool {
	if p.SecretGeneratorPlugin.Options != nil {
		return p.SecretGeneratorPlugin.Options.DisableNameSuffixHash
	}
	return false
}

func NewSuperSecretTransformerPlugin() resmap.TransformerPlugin {
	return &SuperSecretPlugin{logger: utils.GetLogger("SuperSecretTransformerPlugin")}
}

func NewSuperSecretGeneratorPlugin() resmap.GeneratorPlugin {
	return &SuperSecretPlugin{logger: utils.GetLogger("SuperSecretGeneratorPlugin")}
}
