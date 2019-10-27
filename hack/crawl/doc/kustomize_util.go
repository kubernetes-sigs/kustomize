package doc

import (
	"fmt"
	"regexp"

	"sigs.k8s.io/yaml"
)

func FixKustomizationPreUnmarshallingNonFatal(data []byte) ([]byte, error) {
	deprecateFieldsMap := map[string]string{
		"imageTags:": "images:",
	}
	for oldname, newname := range deprecateFieldsMap {
		pattern := regexp.MustCompile(oldname)
		data = pattern.ReplaceAll(data, []byte(newname))
	}

	found, err := useLegacyPatch(data)
	if err == nil && found {
		pattern := regexp.MustCompile("patches:")
		data = pattern.ReplaceAll(data, []byte("patchesStrategicMerge:"))
	}

	return data, err
}

func useLegacyPatch(data []byte) (bool, error) {
	found := false

	var object map[string]interface{}
	err := yaml.Unmarshal(data, &object)
	if err != nil {
		return false, fmt.Errorf("invalid content from %s",
			string(data))
	}
	if rawPatches, ok := object["patches"]; ok {
		patches, ok := rawPatches.([]interface{})
		if !ok {
			return false, fmt.Errorf("invalid patches from %v",
				rawPatches)
		}
		for _, p := range patches {
			_, ok := p.(string)
			if ok {
				found = true
			}
		}
	}
	return found, nil
}
