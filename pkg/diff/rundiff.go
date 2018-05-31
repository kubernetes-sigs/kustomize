package diff

import (
	"github.com/ghodss/yaml"

	"io"

	"github.com/kubernetes-sigs/kustomize/pkg/resource"
)

// RunDiff runs system diff program to compare two ResourceCollections.
func RunDiff(raw, transformed resource.ResourceCollection,
	out, errOut io.Writer) error {
	transformedDir, err := writeYamlToNewDir(transformed, "transformed")
	if err != nil {
		return err
	}
	defer transformedDir.delete()

	noopDir, err := writeYamlToNewDir(raw, "noop")
	if err != nil {
		return err
	}
	defer noopDir.delete()

	return newProgram(out, errOut).run(noopDir.name(), transformedDir.name())
}

// writeYamlToNewDir writes each obj in ResourceCollection to a file in a new directory.
// The directory's name will begin with the given prefix.
// Each file is named with GroupVersionKindName.
func writeYamlToNewDir(in resource.ResourceCollection, prefix string) (*directory, error) {
	dir, err := newDirectory(prefix)
	if err != nil {
		return nil, err
	}

	for gvkn, obj := range in {
		f, err := dir.newFile(gvkn.String())
		if err != nil {
			return nil, err
		}
		err = print(obj.Data, f)
		f.Close()
		if err != nil {
			return nil, err
		}
	}
	return dir, nil
}

// Print the object as YAML.
func print(obj interface{}, w io.Writer) error {
	if obj == nil {
		return nil
	}
	data, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}
