package krmfunction

import (
	"github.com/spf13/cobra"
)

// NewKrmFunctionCmd returns a pointer to a command
func NewKrmFunctionCmd() *cobra.Command {
	var outputDir string
	var inputFile string

	cmd := &cobra.Command{
		Use:   "krm -i FILE -o DIR",
		Short: "Convert the plugin to KRM function instead of builtin function",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := NewConverter(outputDir, inputFile)
			return c.Convert()
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "",
		"Path to the directory which will contain the KRM function")
	cmd.Flags().StringVarP(&inputFile, "input", "i", "",
		"Path to the input file")

	cmd.MarkFlagRequired("output")
	cmd.MarkFlagRequired("input")

	return cmd
}
