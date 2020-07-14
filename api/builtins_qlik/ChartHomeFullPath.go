package builtins_qlik

import (
	"log"
	"os"
	"path/filepath"

	"sigs.k8s.io/kustomize/api/builtins_qlik/utils"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/konfig"
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
	dir, err := konfig.DefaultAbsPluginHome(filesys.MakeFsOnDisk())
	if err != nil {
		dir = filepath.Join(konfig.HomeDir(), konfig.XdgConfigHomeEnvDefault, konfig.ProgramName, konfig.RelPluginHome)
		p.logger.Printf("No kustomize plugin directory, will create default: %v\n", dir)
	}
	dir = filepath.Join(dir, "qlik", "v1", "charts")
	err = os.MkdirAll(dir, 0777)
	if err != nil {
		p.logger.Printf("error creating directory: %v, error: %v\n", dir, err)
		return nil, err
	}
	if p.Kind == "HelmChart" {
		err := utils.CopyDir(p.ChartHome, dir, p.logger)
		if err != nil {
			p.logger.Printf("error copying directory from: %v, to: %v, error: %v\n", p.ChartHome, dir, err)
			return nil, err
		}
	}
	// The chart name should be a subdirectory, so chartHome should be the parent
	return dir, nil
}

func (p *ChartHomeFullPathPlugin) Transform(m resmap.ResMap) error {

	for _, r := range m.Resources() {
		p.ChartName = p.GetFieldValue(r, "chartName")
		p.Kind = p.GetFieldValue(r, "kind")
		if len(p.ChartHome) > 0 {
			//join the root(root of kustomize file) + location to chartHome
			p.ChartHome = filepath.Join(p.Root, p.ChartHome)
		} else {
			p.ChartHome = filepath.Join(p.Root, p.GetFieldValue(r, "chartHome"))
		}
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
