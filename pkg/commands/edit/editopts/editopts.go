package editopts

import (
	"github.com/spf13/cobra"
)

type Options struct {
	KustomizationDir string
}

func (o *Options) ValidateCommon(cmd *cobra.Command, args []string) error {
	var err error

	o.KustomizationDir, err = cmd.Flags().GetString("kustomization-dir")
	if err != nil {
		return err
	}

	return nil
}
