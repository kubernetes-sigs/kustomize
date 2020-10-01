package utils

import (
	"encoding/json"

	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

func NewKyamlRNode(from map[string]interface{}) (*kyaml.RNode, error) {
	node := &kyaml.RNode{}
	if jsonBytes, err := json.Marshal(from); err != nil {
		return nil, err
	} else if err := node.UnmarshalJSON(jsonBytes); err != nil {
		return nil, err
	} else {
		return node, nil
	}
}
