package set

import (
	"errors"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/util"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type setGeneratorOptions struct {
	GeneratorOptions types.GeneratorOptions
	Labels           []string
	Annotations      []string
}

func newCmdSetGeneratorOptions(fSys filesys.FileSystem) *cobra.Command {
	var o setGeneratorOptions
	cmd := &cobra.Command{
		Use:   "generatoroptions",
		Short: "Set a generatoroptions to the kustomization file",
		Long:  "",
		Example: `
	# Set a generatoroptions with a "immutable = true" to the kustomization file.
	kustomize edit set generatoroptions --immutable

	# Set a generatoroptions with a "disableNameSuffixHash = true" to the kustomization file.
	kustomize edit set generatoroptions --disableNameSuffixHash

	# Set generatoroptions with labels to the kustomization file.
	kustomize edit set generatoroptions --labels=app:hazelcast --labels=env:prod

	# Set generatoroptions with annotations to the kustomization file.
	kustomize edit set generatoroptions --annotations=app:hazelcast --annotations=env:prod
`,
		RunE: func(_ *cobra.Command, args []string) error {
			err := o.Validate()
			if err != nil {
				return err
			}
			o.GeneratorOptions.Labels, err = util.ConvertSliceToMap(o.Labels, "label")
			if err != nil {
				return err
			}
			o.GeneratorOptions.Annotations, err = util.ConvertSliceToMap(o.Annotations, "annotation")
			if err != nil {
				return err
			}
			return o.RunSetGeneratorOptions(fSys)
		},
	}
	cmd.Flags().BoolVar(
		&o.GeneratorOptions.Immutable,
		"immutable",
		false,
		"Enable immutable of the generatorOptions.")
	cmd.Flags().BoolVar(
		&o.GeneratorOptions.DisableNameSuffixHash,
		"disableNameSuffixHash",
		false,
		"Disable the name suffix of the generatorOptions.")
	cmd.Flags().StringArrayVar(
		&o.Labels,
		"labels",
		[]string{},
		"Specify labels to be added to the generatorOptions.")
	cmd.Flags().StringArrayVar(
		&o.Annotations,
		"annotations",
		[]string{},
		"Specify annotations to be added to the generatorOptions.")

	return cmd
}

func (o *setGeneratorOptions) Validate() error {
	if !o.GeneratorOptions.Immutable &&
		!o.GeneratorOptions.DisableNameSuffixHash &&
		len(o.Labels) == 0 &&
		len(o.Annotations) == 0 {
		return errors.New("at least immutable, or disableNameSuffixHash or labels or annotations must be set.")
	}
	return nil
}

func (o *setGeneratorOptions) RunSetGeneratorOptions(fSys filesys.FileSystem) error {
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err
	}

	m, err := mf.Read()
	if err != nil {
		return err
	}

	m.GeneratorOptions = types.MergeGlobalOptionsIntoLocal(&o.GeneratorOptions, m.GeneratorOptions)
	return mf.Write(m)
}
