package builtins_qlik

import (
	"fmt"
	"log"
	"regexp"

	"sigs.k8s.io/kustomize/api/builtins_qlik/utils"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/transform"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

type SearchReplacePlugin struct {
	Target    *types.Selector `json:"target,omitempty" yaml:"target,omitempty"`
	Path      string          `json:"path,omitempty" yaml:"path,omitempty"`
	Search    string          `json:"search,omitempty" yaml:"search,omitempty"`
	Replace   string          `json:"replace,omitempty" yaml:"replace,omitempty"`
	logger    *log.Logger
	fieldSpec types.FieldSpec
	re        *regexp.Regexp
}

func (p *SearchReplacePlugin) Config(h *resmap.PluginHelpers, c []byte) (err error) {
	p.Target = nil
	p.Path = ""
	p.Search = ""
	p.Replace = ""
	err = yaml.Unmarshal(c, p)
	if err != nil {
		p.logger.Printf("error unmarshalling config from yaml, error: %v\n", err)
		return err
	}
	if p.Target == nil {
		return fmt.Errorf("must specify a target in the config for the environment variables upsert")
	}

	p.fieldSpec = types.FieldSpec{Path: p.Path}

	p.re, err = regexp.Compile(p.Search)
	if err != nil {
		p.logger.Printf("error compiling regexp from: %v, error: %v\n", p.Search, err)
		return err
	}

	return nil
}

func (p *SearchReplacePlugin) Transform(m resmap.ResMap) error {
	resources, err := m.Select(*p.Target)
	if err != nil {
		p.logger.Printf("error selecting resources based on the target selector, error: %v\n", err)
		return err
	}
	for _, r := range resources {
		err := transform.MutateField(
			r.Map(),
			p.fieldSpec.PathSlice(),
			false,
			p.searchAndReplace)
		if err != nil {
			p.logger.Printf("error executing transformers.MutateField(), error: %v\n", err)
			return err
		}
	}
	return nil
}

func (p *SearchReplacePlugin) searchAndReplace(in interface{}) (interface{}, error) {
	if target, ok := in.(string); !ok {
		return in, nil
	} else {
		return p.re.ReplaceAllString(target, p.Replace), nil
	}
}

func NewSearchReplacePlugin() resmap.TransformerPlugin {
	return &SearchReplacePlugin{logger: utils.GetLogger("SearchReplacePlugin")}
}
