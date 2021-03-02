// This file will be processed and embedded to pluginator.

package funcwrappersrc

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
)

//nolint
func main() {
	var plugin resmap.Configurable
	p := provider.NewDefaultDepProvider()
	resmapFactory := resmap.NewFactory(
		p.GetResourceFactory(), p.GetConflictDetectorFactory())
	pluginHelpers := resmap.NewPluginHelpers(
		nil, p.GetFieldValidator(), resmapFactory)

	processor := framework.ResourceListProcessorFunc(func(resourceList *framework.ResourceList) error {
		resMap, err := resmapFactory.NewResMapFromRNodeSlice(resourceList.Items)
		if err != nil {
			return err
		}
		dataValue, err := resourceList.FunctionConfig.Field("data").Value.String()
		if err != nil {
			return err
		}

		err = plugin.Config(pluginHelpers, []byte(dataValue))
		if err != nil {
			return err
		}
		if t, ok := plugin.(resmap.TransformerPlugin); ok {
			err = t.Transform(resMap)
			if err != nil {
				return err
			}
		} else if g, ok := plugin.(resmap.GeneratorPlugin); ok {
			resMap, err = g.Generate()
			if err != nil {
				return err
			}
		}

		resourceList.Items, err = resMap.ToRNodeSlice()
		if err != nil {
			return err
		}
		return nil
	})
	if err := framework.Execute(&processor, nil); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
