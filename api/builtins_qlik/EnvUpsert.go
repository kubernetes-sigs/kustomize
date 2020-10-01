package builtins_qlik

import (
	"encoding/json"
	"fmt"
	"log"

	"sigs.k8s.io/kustomize/api/builtins_qlik/utils"
	"sigs.k8s.io/kustomize/api/filters/fieldspec"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filtersutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/yaml"
)

type EnvVarType struct {
	Name      *string                `json:"name,omitempty" yaml:"name,omitempty"`
	Value     *string                `json:"value,omitempty" yaml:"value,omitempty"`
	ValueFrom map[string]interface{} `json:"valueFrom,omitempty" yaml:"valueFrom,omitempty"`
	Delete    bool                   `json:"delete,omitempty" yaml:"delete,omitempty"`
}

type EnvUpsertPlugin struct {
	Enabled   bool            `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Target    *types.Selector `json:"target,omitempty" yaml:"target,omitempty"`
	Path      string          `json:"path,omitempty" yaml:"path,omitempty"`
	EnvVars   []EnvVarType    `json:"env,omitempty" yaml:"env,omitempty"`
	logger    *log.Logger
	fieldSpec types.FieldSpec
}

func (p *EnvUpsertPlugin) Config(h *resmap.PluginHelpers, c []byte) (err error) {
	p.Enabled = false
	p.Target = nil
	p.Path = ""
	p.EnvVars = make([]EnvVarType, 0)
	err = yaml.Unmarshal(c, p)
	if err != nil {
		p.logger.Printf("error unmarshalling config from yaml, error: %v\n", err)
		return err
	}
	if p.Target == nil {
		return fmt.Errorf("must specify a target in the config for the environment variables upsert")
	}
	for _, envVar := range p.EnvVars {
		if envVar.Name == nil {
			err = fmt.Errorf("env var config has no name: %v", envVar)
			p.logger.Printf("config error: %v\n", err)
			return err
		}
		if envVar.Value == nil && envVar.ValueFrom == nil && !envVar.Delete {
			err = fmt.Errorf("env var config has no value or valueFrom: %v", envVar)
			p.logger.Printf("config error: %v\n", err)
			return err
		}
	}
	p.fieldSpec = types.FieldSpec{Path: p.Path}
	return nil
}

func (p *EnvUpsertPlugin) Transform(m resmap.ResMap) error {
	if p.Enabled {
		resources, err := m.Select(*p.Target)

		if err != nil {
			p.logger.Printf("error selecting resources based on the target selector, error: %v\n", err)
			return err
		}
		for _, r := range resources {
			err := filtersutil.ApplyToJSON(envUpsertFilter{
				envVars:   p.EnvVars,
				fieldSpec: p.fieldSpec,
			}, r)
			if err != nil {
				p.logger.Printf("error upserting env vars: %+v, error: %v\n", p.EnvVars, err)
				return err
			}
		}
	}
	return nil
}

func NewEnvUpsertPlugin() resmap.TransformerPlugin {
	return &EnvUpsertPlugin{logger: utils.GetLogger("EnvUpsertPlugin")}
}

type envUpsertFilter struct {
	fieldSpec types.FieldSpec
	envVars   []EnvVarType
}

func (f envUpsertFilter) Filter(nodes []*kyaml.RNode) ([]*kyaml.RNode, error) {
	_, err := kio.FilterAll(kyaml.FilterFunc(
		func(node *kyaml.RNode) (*kyaml.RNode, error) {
			if err := node.PipeE(fieldspec.Filter{
				FieldSpec: f.fieldSpec,
				SetValue:  f.set,
			}); err != nil {
				return nil, err
			}
			return node, nil
		})).Filter(nodes)
	return nodes, err
}

func (f *envUpsertFilter) set(node *kyaml.RNode) error {
	var a []interface{}
	if jsonBytes, err := node.MarshalJSON(); err != nil {
		return err
	} else if err := json.Unmarshal(jsonBytes, &a); err != nil {
		return err
	} else {
		changed := f.upsertEnvironmentVariables(a)
		//we need this because rnode.UnmarshalJSON() cannot unmarshal JSON arrays:
		tempMap := map[string]interface{}{"tmp": changed}
		if tempMapRNode, err := utils.NewKyamlRNode(tempMap); err != nil {
			return err
		} else {
			node.SetYNode(tempMapRNode.Field("tmp").Value.YNode())
		}
	}
	return nil
}

func (f *envUpsertFilter) upsertEnvironmentVariables(in interface{}) interface{} {
	presentEnvVars, ok := in.([]interface{})
	if ok {
		for _, envVar := range f.envVars {
			foundMatching := false
			for i := 0; i < len(presentEnvVars); i++ {
				presentEnvVar, ok := presentEnvVars[i].(map[string]interface{})
				if ok {
					name, ok := presentEnvVar["name"].(string)
					if ok {
						if name == *envVar.Name {
							foundMatching = true
							if envVar.Delete {
								//delete:
								presentEnvVars = append(presentEnvVars[:i], presentEnvVars[i+1:]...)
								i--
							} else {
								//update:
								f.setEnvVar(presentEnvVar, envVar)
							}
							break
						}
					}
				}
			}
			if !foundMatching && !envVar.Delete {
				//insert:
				newEnvVar := map[string]interface{}{
					"name": *envVar.Name,
				}
				f.setEnvVar(newEnvVar, envVar)
				presentEnvVars = append(presentEnvVars, newEnvVar)
			}
		}
		return presentEnvVars
	}
	return in
}

func (f *envUpsertFilter) setEnvVar(setEnvVar map[string]interface{}, fromEnvVar EnvVarType) {
	if fromEnvVar.Value != nil {
		setEnvVar["value"] = *fromEnvVar.Value
	} else if fromEnvVar.ValueFrom != nil {
		setEnvVar["valueFrom"] = fromEnvVar.ValueFrom
	}
}
