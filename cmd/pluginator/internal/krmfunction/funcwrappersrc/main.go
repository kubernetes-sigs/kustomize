// This file will be processed and embedded to pluginator.

package funcwrappersrc

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/yaml"
)

//nolint
func main() {
	var plugin resmap.Configurable
	p := provider.NewDefaultDepProvider()
	resmapFactory := resmap.NewFactory(
		p.GetResourceFactory(), p.GetConflictDetectorFactory())
	pluginHelpers := resmap.NewPluginHelpers(
		nil, p.GetFieldValidator(), resmapFactory)

	resourceList := &framework.ResourceList{}
	resourceList.FunctionConfig = map[string]interface{}{}

	cmd := framework.Command(resourceList, func() error {
		resMap, err := resmapFactory.NewResMapFromRNodeSlice(resourceList.Items)
		if err != nil {
			return err
		}
		dataField, err := getDataFromFunctionConfig(resourceList.FunctionConfig)
		if err != nil {
			return err
		}
		dataValue, err := yaml.Marshal(dataField)
		if err != nil {
			return err
		}

		err = plugin.Config(pluginHelpers, dataValue)
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
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

//nolint
func getDataFromFunctionConfig(fc interface{}) (interface{}, error) {
	f, ok := fc.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("function config %#v is not valid", fc)
	}
	return f["data"], nil
}
