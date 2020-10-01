package builtins_qlik

import (
	"encoding/json"
	"log"

	"github.com/imdario/mergo"
	"sigs.k8s.io/kustomize/api/builtins_qlik/utils"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filtersutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/yaml"
)

type HelmValuesPlugin struct {
	Overwrite        bool                   `json:"overwrite,omitempty" yaml:"overwrite,omitempty"`
	Chart            string                 `json:"chartName,omitempty" yaml:"chartName,omitempty"`
	ReleaseName      string                 `json:"releaseName,omitempty" yaml:"releaseName,omitempty"`
	ReleaseNamespace string                 `json:"releaseNamespace,omitempty" yaml:"releaseNamespace,omitempty"`
	FieldSpecs       []types.FieldSpec      `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
	Values           map[string]interface{} `json:"values,omitempty" yaml:"values,omitempty"`
	logger           *log.Logger
}

func (p *HelmValuesPlugin) Config(h *resmap.PluginHelpers, c []byte) (err error) {
	return yaml.Unmarshal(c, p)
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
	mergeFrom["root"] = p.Values
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
				if err := filtersutil.ApplyToJSON(kio.FilterFunc(func(nodes []*kyaml.RNode) ([]*kyaml.RNode, error) {
					return kio.FilterAll(kyaml.FilterFunc(func(rn *kyaml.RNode) (*kyaml.RNode, error) {
						if valuesRn, err := rn.Pipe(kyaml.FieldMatcher{Name: "values"}); err != nil {
							return nil, err
						} else {
							valuesRnMap := make(map[string]interface{})

							if valuesRn != nil {
								if jsonBytes, err := valuesRn.MarshalJSON(); err != nil {
									return nil, err
								} else if err := json.Unmarshal(jsonBytes, &valuesRnMap); err != nil {
									return nil, err
								}
							} else {
								valuesRn = &kyaml.RNode{}
							}

							if newValuesRnMap, err := p.mutateValues(valuesRnMap); err != nil {
								return nil, err
							} else if newJsonBytes, err := json.Marshal(newValuesRnMap); err != nil {
								return nil, err
							} else if err := valuesRn.UnmarshalJSON(newJsonBytes); err != nil {
								return nil, err
							} else if err := rn.PipeE(kyaml.FieldSetter{Name: "values", Value: valuesRn}); err != nil {
								return nil, err
							}
						}
						return rn, nil
					})).Filter(nodes)
				}), r); err != nil {
					return err
				}
			}
		}
		if len(p.ReleaseNamespace) > 0 && p.ReleaseNamespace != "null" {
			if err := filtersutil.ApplyToJSON(kio.FilterFunc(func(nodes []*kyaml.RNode) ([]*kyaml.RNode, error) {
				return kio.FilterAll(kyaml.FilterFunc(func(rn *kyaml.RNode) (*kyaml.RNode, error) {
					return rn.Pipe(kyaml.FieldSetter{Name: "releaseNamespace", StringValue: p.ReleaseNamespace})
				})).Filter(nodes)
			}), r); err != nil {
				return err
			}
		}
		if len(p.ReleaseName) > 0 && p.ReleaseName != "null" {
			if err := filtersutil.ApplyToJSON(kio.FilterFunc(func(nodes []*kyaml.RNode) ([]*kyaml.RNode, error) {
				return kio.FilterAll(kyaml.FilterFunc(func(rn *kyaml.RNode) (*kyaml.RNode, error) {
					return rn.Pipe(kyaml.FieldSetter{Name: "releaseName", StringValue: p.ReleaseName})
				})).Filter(nodes)
			}), r); err != nil {
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
