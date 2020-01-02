package builtins_qlik

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"sigs.k8s.io/kustomize/api/builtins_qlik/utils"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/transform"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

type ChartHomeFullPathPlugin struct {
	ChartHome  string            `json:"chartHome,omitempty" yaml:"chartHome,omitempty"`
	FieldSpecs []types.FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
	Root       string
	ChartName  string
	Kind       string
	logger     *log.Logger
}

func (p *ChartHomeFullPathPlugin) Config(h *resmap.PluginHelpers, c []byte) (err error) {
	p.Root = h.Loader().Root()
	return yaml.Unmarshal(c, p)
}

func (p *ChartHomeFullPathPlugin) mutate(in interface{}) (interface{}, error) {
	dir, err := ioutil.TempDir("", "temp")
	if err != nil {
		p.logger.Printf("error creating temporaty directory: %v\n", err)
		return nil, err
	}
	directory := fmt.Sprintf("%s/%s", dir, p.ChartName)
	err = os.Mkdir(directory, 0777)
	if err != nil {
		p.logger.Printf("error creating directory: %v, error: %v\n", directory, err)
		return nil, err
	}
	if p.Kind == "HelmChart" {
		err := utils.CopyDir(p.ChartHome, directory, p.logger)
		if err != nil {
			p.logger.Printf("error copying directory from: %v, to: %v, error: %v\n", p.ChartHome, directory, err)
			return nil, err
		}
	}
	return directory, nil
}

func (p *ChartHomeFullPathPlugin) Transform(m resmap.ResMap) error {
	//join the root(root of kustomize file) + location to chartHome
	p.ChartHome = path.Join(p.Root, p.ChartHome)

	for _, r := range m.Resources() {
		p.ChartName = p.GetFieldValue(r, "chartName")
		p.Kind = p.GetFieldValue(r, "kind")
		pathToField := []string{"chartHome"}
		err := transform.MutateField(
			r.Map(),
			pathToField,
			true,
			p.mutate)
		if err != nil {
			p.logger.Printf("error executing MutateField for chart: %v, pathToField: %v, error: %v\n", p.ChartName, pathToField, err)
			return err
		}
	}
	return nil
}

func (p *ChartHomeFullPathPlugin) GetFieldValue(obj ifc.Kunstructured, fieldName string) string {
	v, err := obj.GetString(fieldName)
	if err != nil {
		p.logger.Printf("error extracting fieldName: %v (will return empty string), error: %v\n", fieldName, err)
		return ""
	}
	return v
}

func NewChartHomeFullPathPlugin() resmap.TransformerPlugin {
	return &ChartHomeFullPathPlugin{logger: utils.GetLogger("ChartHomeFullPathPlugin")}
}