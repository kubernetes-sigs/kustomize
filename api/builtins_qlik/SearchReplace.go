package builtins_qlik

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"sigs.k8s.io/kustomize/api/builtins_qlik/utils"
	"sigs.k8s.io/kustomize/api/filters/fieldspec"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filtersutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/yaml"
)

type ReplaceWithGitDescribeTagT struct {
	Default string `json:"default,omitempty" yaml:"default,omitempty"`
}

type SearchReplacePlugin struct {
	Target                    *types.Selector             `json:"target,omitempty" yaml:"target,omitempty"`
	Path                      string                      `json:"path,omitempty" yaml:"path,omitempty"`
	Search                    string                      `json:"search,omitempty" yaml:"search,omitempty"`
	Replace                   string                      `json:"replace,omitempty" yaml:"replace,omitempty"`
	ReplaceWithEnvVar         string                      `json:"replaceWithEnvVar,omitempty" yaml:"replaceWithEnvVar,omitempty"`
	ReplaceWithObjRef         *types.Var                  `json:"replaceWithObjRef,omitempty" yaml:"replaceWithObjRef,omitempty"`
	ReplaceWithGitDescribeTag *ReplaceWithGitDescribeTagT `json:"replaceWithGitDescribeTag,omitempty" yaml:"replaceWithGitDescribeTag,omitempty"`
	logger                    *log.Logger
	fieldSpec                 types.FieldSpec
	re                        *regexp.Regexp
	pwd                       string
}

func (p *SearchReplacePlugin) Config(h *resmap.PluginHelpers, c []byte) (err error) {
	p.Target = nil
	p.Path = ""
	p.Search = ""
	p.Replace = ""
	p.ReplaceWithEnvVar = ""
	p.ReplaceWithObjRef = nil
	p.ReplaceWithGitDescribeTag = nil
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

	p.pwd = h.Loader().Root()

	return nil
}

func (p *SearchReplacePlugin) Transform(m resmap.ResMap) error {
	resources, err := m.Select(*p.Target)
	if err != nil {
		p.logger.Printf("error selecting resources based on the target selector, error: %v\n", err)
		return err
	}

	if p.Replace == "" {
		if p.ReplaceWithObjRef != nil {
			var replaceEmpty bool
			for _, res := range m.Resources() {
				if p.matchesObjRef(res) {
					if replacementValue, err := getReplacementValue(res, p.ReplaceWithObjRef.FieldRef.FieldPath); err != nil {
						p.logger.Printf("error getting replacement value: %v\n", err)
					} else {
						p.Replace = replacementValue
						replaceEmpty = true
						break
					}
				}
			}
			if p.Replace == "" && !replaceEmpty {
				p.logger.Printf("Object Reference could not be found")
				return nil
			}
		} else if p.ReplaceWithGitDescribeTag != nil {
			if gitDescribeTag, err := utils.GetGitDescribeForHead(p.pwd, p.ReplaceWithGitDescribeTag.Default); err != nil {
				return err
			} else {
				p.Replace = strings.TrimPrefix(gitDescribeTag, "v")
			}
		} else if len(p.ReplaceWithEnvVar) > 0 {
			p.Replace = os.Getenv(p.ReplaceWithEnvVar)
		}
	}
	for _, r := range resources {
		if p.fieldSpec.Path == "/" {
			if newRoot, err := p.searchAndReplace(r.Map(), false); err != nil {
				p.logger.Printf("error executing transformers.MutateField(), error: %v\n", err)
				return err
			} else if newRootMap, newRootIsMap := newRoot.(map[string]interface{}); !newRootIsMap {
				return errors.New("search/replace on root did not return a map[string]interface{}")
			} else if jsonBytes, err := json.Marshal(newRootMap); err != nil {
				return err
			} else if err := r.UnmarshalJSON(jsonBytes); err != nil {
				return err
			}
		} else if err := filtersutil.ApplyToJSON(kio.FilterFunc(func(nodes []*kyaml.RNode) ([]*kyaml.RNode, error) {
			return kio.FilterAll(kyaml.FilterFunc(func(rn *kyaml.RNode) (*kyaml.RNode, error) {
				if err := rn.PipeE(fieldspec.Filter{
					FieldSpec: p.fieldSpec,
					SetValue: func(n *kyaml.RNode) error {
						return p.searchAndReplaceRNode(n, isSecretDataTarget(r, p.fieldSpec.PathSlice()))
					},
				}); err != nil {
					return nil, err
				}
				return rn, nil
			})).Filter(nodes)
		}), r); err != nil {
			return err
		}
	}
	return nil
}

func getReplacementValue(res *resource.Resource, fieldPath string) (string, error) {
	if val, err := res.GetFieldValue(fieldPath); err != nil {
		return "", err
	} else if strVal, ok := val.(string); !ok {
		return "", errors.New("FieldRef for the ReplaceWithObjRef must point to a value of string type")
	} else if isSecretDataReplacement(res, fieldPath) {
		if decodedStrVal, err := base64.StdEncoding.DecodeString(strVal); err != nil {
			return "", err
		} else {
			return string(decodedStrVal), nil
		}
	} else {
		return strVal, nil
	}
}

func isSecretDataReplacement(res *resource.Resource, fieldPath string) bool {
	return res.GetGvk().Kind == "Secret" &&
		(strings.HasPrefix(fieldPath, "data.") || strings.HasPrefix(fieldPath, "data["))
}

func isSecretDataTarget(r *resource.Resource, pathSlice []string) bool {
	return r.GetGvk().Kind == "Secret" && len(pathSlice) > 0 && pathSlice[0] == "data"
}

func (p *SearchReplacePlugin) matchesObjRef(res *resource.Resource) bool {
	if res.GetGvk().IsSelected(&p.ReplaceWithObjRef.ObjRef.Gvk) {
		if len(p.ReplaceWithObjRef.ObjRef.Name) > 0 {
			return res.GetName() == p.ReplaceWithObjRef.ObjRef.Name
		}
		return true
	}
	return false
}

func (p *SearchReplacePlugin) searchAndReplaceRNode(node *kyaml.RNode, base64Encoded bool) error {
	var in interface{}
	if node.YNode().Kind == kyaml.ScalarNode {
		in = node.YNode().Value
	} else {
		if jsonBytes, err := node.MarshalJSON(); err != nil {
			return err
		} else if err := json.Unmarshal(jsonBytes, &in); err != nil {
			return err
		}
	}

	changed, err := p.searchAndReplace(in, base64Encoded)
	if err != nil {
		return err
	}

	if _, ok := changed.(string); ok {
		node.YNode().Value = changed.(string)
	} else {
		tempMap := map[string]interface{}{"tmp": changed}
		if tempMapRNode, err := utils.NewKyamlRNode(tempMap); err != nil {
			return err
		} else {
			node.SetYNode(tempMapRNode.Field("tmp").Value.YNode())
		}
	}
	return nil
}

func (p *SearchReplacePlugin) searchAndReplace(in interface{}, base64Encoded bool) (interface{}, error) {
	if target, ok := in.(string); ok {
		if base64Encoded {
			if decodedValue, err := base64.StdEncoding.DecodeString(target); err != nil {
				return nil, err
			} else {
				replacedDecodedValue := p.re.ReplaceAllString(string(decodedValue), p.Replace)
				return base64.StdEncoding.EncodeToString([]byte(replacedDecodedValue)), nil
			}
		} else {
			return p.re.ReplaceAllString(target, p.Replace), nil
		}
	} else if target, ok := in.(map[string]interface{}); ok {
		return p.marshallToJsonAndReplace(target)
	} else if target, ok := in.([]interface{}); ok {
		return p.marshallToJsonAndReplace(target)
	}
	return in, nil
}

func (p *SearchReplacePlugin) marshallToJsonAndReplace(in interface{}) (interface{}, error) {
	if marshalledTarget, err := json.Marshal(in); err != nil {
		p.logger.Printf("error marshalling interface to JSON, error: %v\n", err)
		return nil, err
	} else {
		replaced := p.re.ReplaceAllString(string(marshalledTarget), p.Replace)
		if err := json.Unmarshal([]byte(replaced), &in); err != nil {
			p.logger.Printf("error unmarshalling JSON string after replacements back to interface, error: %v\n", err)
			return nil, err
		} else {
			return in, err
		}
	}
}

func NewSearchReplacePlugin() resmap.TransformerPlugin {
	return &SearchReplacePlugin{logger: utils.GetLogger("SearchReplacePlugin")}
}
