//nolint
package funcwrappersrc

import (
	"sigs.k8s.io/kustomize/api/resmap"
)

type plugin struct{}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(
	_ *resmap.PluginHelpers, _ []byte) (err error) {
	return nil
}

func (p *plugin) Transform(_ resmap.ResMap) error {
	return nil
}
