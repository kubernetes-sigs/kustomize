// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
)

// GetWrapRunner returns a command runner.
func GetWrapRunner() *WrapRunner {
	r := &WrapRunner{}
	c := &cobra.Command{
		Use:   "wrap CMD...",
		Short: "Wrap an executable so it implements the config fn interface",
		Long: `Wrap an executable so it implements the config fn interface

wrap simplifies writing config functions by:

- invoking an executable command converting an input ResourceList into environment
- merging the output onto the original input as a set of patches
- setting filenames on any Resources missing them

config function authors may use wrap by using it to invoke a command from a container image

The following are equivalent:

	kyaml wrap -- CMD

	kyaml xargs -- CMD | kyaml merge | kyaml fmt --set-filenames

Environment Variables:

  KUST_OVERRIDE_DIR:

    Path to a directory containing patches to apply to after merging.
`,
		Example: `

`,
		RunE:               r.runE,
		SilenceUsage:       true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		Args:               cobra.MinimumNArgs(1),
	}
	r.Command = c
	r.XArgs = GetXArgsRunner()
	c.Flags().BoolVar(&r.XArgs.EnvOnly,
		"env-only", true, "only set env vars, not arguments.")
	c.Flags().StringVar(&r.XArgs.WrapKind,
		"wrap-kind", "List", "wrap the input xargs give to the command in this type.")
	c.Flags().StringVar(&r.XArgs.WrapVersion,
		"wrap-version", "v1", "wrap the input xargs give to the command in this type.")
	return r
}

// WrapRunner contains the run function
type WrapRunner struct {
	Command *cobra.Command
	XArgs   *XArgsRunner
	getEnv  func(key string) string
}

const (
	KustMergeEnv       = "KUST_MERGE"
	KustOverrideDirEnv = "KUST_OVERRIDE_DIR"
)

func WrapCommand() *cobra.Command {
	return GetWrapRunner().Command
}

func (r *WrapRunner) runE(c *cobra.Command, args []string) error {
	if r.getEnv == nil {
		r.getEnv = os.Getenv
	}
	xargsIn := &bytes.Buffer{}
	if _, err := io.Copy(xargsIn, c.InOrStdin()); err != nil {
		return err
	}
	mergeInput := bytes.NewBuffer(xargsIn.Bytes())
	// Run the command
	xargsOut := &bytes.Buffer{}
	r.XArgs.Command.SetArgs(args)
	r.XArgs.Command.SetIn(xargsIn)
	r.XArgs.Command.SetOut(xargsOut)
	r.XArgs.Command.SetErr(os.Stderr)
	if err := r.XArgs.Command.Execute(); err != nil {
		return err
	}

	// merge the results
	buff := &kio.PackageBuffer{}

	var fltrs []kio.Filter
	var inputs []kio.Reader
	if r.getEnv(KustMergeEnv) == "" || r.getEnv(KustMergeEnv) == "true" || r.getEnv(KustMergeEnv) == "1" {
		inputs = append(inputs, &kio.ByteReader{Reader: mergeInput})
		fltrs = append(fltrs, &filters.MergeFilter{})
	}
	inputs = append(inputs, &kio.ByteReader{Reader: xargsOut})

	if err := (kio.Pipeline{Inputs: inputs, Filters: fltrs, Outputs: []kio.Writer{buff}}).
		Execute(); err != nil {
		return err
	}

	inputs, fltrs = []kio.Reader{buff}, nil
	if r.getEnv(KustOverrideDirEnv) != "" {
		// merge the overrides on top of the output
		fltrs = append(fltrs, filters.MergeFilter{})
		inputs = append(inputs,
			kio.LocalPackageReader{
				OmitReaderAnnotations: true, // don't set path annotations, as they would override
				PackagePath:           r.getEnv(KustOverrideDirEnv)})
	}
	fltrs = append(fltrs,
		&filters.FileSetter{
			FilenamePattern: filepath.Join("config", filters.DefaultFilenamePattern)},
		&filters.FormatFilter{})

	err := kio.Pipeline{
		Inputs:  inputs,
		Filters: fltrs,
		Outputs: []kio.Writer{kio.ByteWriter{
			Sort:                  true,
			KeepReaderAnnotations: true,
			Writer:                c.OutOrStdout(),
			WrappingKind:          kio.ResourceListKind,
			WrappingAPIVersion:    kio.ResourceListAPIVersion}}}.Execute()
	if err != nil {
		return err
	}

	return nil
}
