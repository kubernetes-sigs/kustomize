package register

import (
	"banno.com/api/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/kustomize/api/resmap"
)

type plugin struct{}

func (p *plugin) Config(
	h *resmap.PluginHelpers, config []byte) error {
	v1alpha1.AddToScheme(scheme.Scheme)
	return nil
}

func (p *plugin) Transform(m resmap.ResMap) error {
	return nil
}

func NewAppCRDRegisterPlugin() resmap.TransformerPlugin {
	return &plugin{}
}
