// Code generated by pluginator on SecretGenerator; DO NOT EDIT.
// pluginator {(devel)  unknown   }

package builtins

import (
	"sigs.k8s.io/kustomize/api/kv"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

type SecretGeneratorPlugin struct {
	h                *resmap.PluginHelpers
	types.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	types.SecretArgs
}

func (p *SecretGeneratorPlugin) Config(h *resmap.PluginHelpers, config []byte) (err error) {
	p.SecretArgs = types.SecretArgs{}
	err = yaml.Unmarshal(config, p)
	if p.SecretArgs.Name == "" {
		p.SecretArgs.Name = p.Name
	}
	if p.SecretArgs.Namespace == "" {
		p.SecretArgs.Namespace = p.Namespace
	}
	p.h = h
	return
}

func (p *SecretGeneratorPlugin) Generate() (resmap.ResMap, []string, error) {
	return p.h.ResmapFactory().FromSecretArgs(
		kv.NewLoader(p.h.Loader(), p.h.Validator()), p.SecretArgs)
}

func NewSecretGeneratorPlugin() resmap.GeneratorPlugin {
	return &SecretGeneratorPlugin{}
}
