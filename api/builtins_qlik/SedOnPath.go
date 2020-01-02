package builtins_qlik

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"sigs.k8s.io/kustomize/api/builtins_qlik/utils"

	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/transform"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

type SedOnPathPlugin struct {
	Path      string   `json:"path,omitempty" yaml:"path,omitempty"`
	Regex     []string `json:"regex,omitempty" yaml:"regex,omitempty"`
	logger    *log.Logger
	fieldSpec types.FieldSpec
}

func (p *SedOnPathPlugin) Config(h *resmap.PluginHelpers, c []byte) (err error) {
	p.Path = ""
	p.Regex = make([]string, 0)
	err = yaml.Unmarshal(c, p)
	if err != nil {
		p.logger.Printf("error unmarshalling config from yaml, error: %v\n", err)
		return err
	}
	p.fieldSpec = types.FieldSpec{Path: p.Path}
	return nil
}

func (p *SedOnPathPlugin) Transform(m resmap.ResMap) error {
	for _, r := range m.Resources() {
		err := transform.MutateField(
			r.Map(),
			p.fieldSpec.PathSlice(),
			false,
			p.executeSedOnValue)
		if err != nil {
			p.logger.Printf("error executing transformers.MutateField(), error: %v\n", err)
			return err
		}
	}
	return nil
}

func (p *SedOnPathPlugin) executeSedOnValue(in interface{}) (interface{}, error) {
	zString, ok := in.(string)
	if ok {
		return p.executeSed(zString)
	}

	zArray, ok := in.([]interface{})
	if ok {
		zNewArray := zArray[:0]
		for _, zValue := range zArray {
			zString, ok := zValue.(string)
			if !ok {
				return nil, fmt.Errorf("%#v is expected to be a string or []string", in)
			}
			zNewValue, err := p.executeSed(zString)
			if err != nil {
				return nil, err
			}
			zNewArray = append(zNewArray, zNewValue)
		}
		return zNewArray, nil
	}

	return nil, fmt.Errorf("%#v is expected to be a string or []string", in)
}

func (p *SedOnPathPlugin) executeSed(zString string) (string, error) {
	for _, regex := range p.Regex {
		cmd := exec.Command("sed", "-e", regex)
		cmd.Stdin = bytes.NewBuffer([]byte(zString))
		output, err := cmd.Output()
		if err != nil {
			return "", err
		}
		zString = string(output)
	}
	return zString, nil
}

func NewSedOnPathPlugin() resmap.TransformerPlugin {
	return &SedOnPathPlugin{logger: utils.GetLogger("SedOnPathPlugin")}
}
