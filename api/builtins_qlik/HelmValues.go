package builtins_qlik

import (
	"log"

	"sigs.k8s.io/kustomize/api/builtins_qlik/utils"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/transform"
	"sigs.k8s.io/kustomize/api/types"

	"github.com/imdario/mergo"
	"sigs.k8s.io/yaml"
)

type HelmValuesPlugin struct {
	Overwrite        bool                   `json:"overwrite,omitempty" yaml:"overwrite,omitempty"`
	Chart            string                 `json:"chartName,omitempty" yaml:"chartName,omitempty"`
	ReleaseName      string                 `json:"releaseName,omitempty" yaml:"releaseName,omitempty"`
	ReleaseNamespace string                 `json:"releaseNamespace,omitempty" yaml:"releaseNamespace,omitempty"`
	FieldSpecs       []types.FieldSpec      `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
	Values           map[string]interface{} `json:"values,omitempty" yaml:"values,omitempty"`
	ValuesName       string
	logger           *log.Logger
}

func (p *HelmValuesPlugin) Config(h *resmap.PluginHelpers, c []byte) (err error) {
	return yaml.Unmarshal(c, p)
}

func (p *HelmValuesPlugin) mutateReleaseNameSpace(in interface{}) (interface{}, error) {
	return p.ReleaseNamespace, nil
}

func (p *HelmValuesPlugin) mutateReleaseName(in interface{}) (interface{}, error) {
	return p.ReleaseName, nil
}

func (p *HelmValuesPlugin) mutateValues(in interface{}) (interface{}, error) {
	var mergedData map[interface{}]interface{}

	// first merge in whats already in the document stream
	var mergeFrom = make(map[interface{}]interface{})
	mergeFrom["root"] = in
	err := mergeValues(&mergedData, mergeFrom, p.Overwrite)
	if err != nil {
		p.logger.Printf("error executing mergeValues(), error: %v\n", err)
		return nil, err
	}

	// second merge in new values then output
	if p.ValuesName != "" {
		mergeFrom["root"] = p.Values[p.ValuesName]
	} else {
		mergeFrom["root"] = p.Values
	}
	err = mergeValues(&mergedData, mergeFrom, p.Overwrite)
	if err != nil {
		p.logger.Printf("error executing mergeValues(), error: %v\n", err)
		return nil, err
	}
	return mergedData["root"], nil
}

func (p *HelmValuesPlugin) Transform(m resmap.ResMap) error {
	for _, r := range m.Resources() {
		if isHelmChart(r) {
			if applyResources(r, p.Chart) {
				pathToField := []string{"values"}
				err := transform.MutateField(
					r.Map(),
					pathToField,
					true,
					p.mutateValues)
				if err != nil {
					p.logger.Printf("error executing MutateField for chart: %v, pathToField: %v, error: %v\n", p.Chart, pathToField, err)
					return err
				}
			}
		}
		name, err := r.GetString("chartName")
		if err != nil {
			p.logger.Printf("error extracting chartName attribute for chart: %v, error: %v\n", p.Chart, err)
		}

		if p.Values[name] != nil && p.Values[name] != "null" {
			p.ValuesName = name
			pathToField := []string{"values", name}
			err := transform.MutateField(
				r.Map(),
				pathToField,
				true,
				p.mutateValues)
			if err != nil {
				p.logger.Printf("error executing MutateField for chart: %v, pathToField: %v, error: %v\n", p.Chart, pathToField, err)
				return err
			}
			p.ValuesName = ""
		}
		if len(p.ReleaseNamespace) > 0 && p.ReleaseNamespace != "null" {
			pathToField := []string{"releaseNamespace"}
			err := transform.MutateField(
				r.Map(),
				pathToField,
				true,
				p.mutateReleaseNameSpace)
			if err != nil {
				p.logger.Printf("error executing MutateField for chart: %v, pathToField: %v, error: %v\n", p.Chart, pathToField, err)
				return err
			}
		}
		if len(p.ReleaseName) > 0 && p.ReleaseName != "null" {
			pathToField := []string{"releaseName"}
			err := transform.MutateField(
				r.Map(),
				pathToField,
				true,
				p.mutateReleaseName)
			if err != nil {
				p.logger.Printf("error executing MutateField for chart: %v, pathToField: %v, error: %v\n", p.Chart, pathToField, err)
				return err
			}
		}
	}
	return nil
}

func isHelmChart(obj ifc.Kunstructured) bool {
	kind := obj.GetKind()
	if kind == "HelmChart" {
		return true
	}
	return false
}

func applyResources(obj ifc.Kunstructured, chart string) bool {
	name, _ := obj.GetString("chartName")
	if name == chart || chart == "" || chart == "null" {
		return true
	}
	return false
}

func mergeValues(values1 interface{}, values2 interface{}, overwrite bool) error {
	if overwrite {
		return mergo.Merge(values1, values2, mergo.WithOverride)
	}
	return mergo.Merge(values1, values2)
}

func NewHelmValuesPlugin() resmap.TransformerPlugin {
	return &HelmValuesPlugin{logger: utils.GetLogger("HelmValuesPlugin")}
}
