// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

var theArgs struct {
	kustomizationPath string
}

var theFlags struct {
	outputPath string
	enable     struct {
		plugins        bool
		managedByLabel bool
		helm           bool
	}
	helmCommand    string
	loadRestrictor string
	reorderOutput  string
	fnOptions      types.FnPluginLoadingOptions
}

type Help struct {
	Use     string
	Short   string
	Long    string
	Example string
}

func MakeHelp(pgmName, cmdName string) *Help {
	fN := konfig.DefaultKustomizationFileName()
	return &Help{
		Use:   cmdName + " DIR",
		Short: "Build a kustomization target from a directory or URL",
		Long: fmt.Sprintf(`Build a set of KRM resources using a '%s' file.
The DIR argument must be a path to a directory containing
'%s', or a git repository URL with a path suffix
specifying same with respect to the repository root.
If DIR is omitted, '.' is assumed.
`, fN, fN),
		Example: fmt.Sprintf(`# Build the current working directory
  %s %s

# Build some shared configuration directory
  %s %s /home/config/production

# Build from github
  %s %s https://github.com/kubernetes-sigs/kustomize.git/examples/helloWorld?ref=v1.0.6
`, pgmName, cmdName, pgmName, cmdName, pgmName, cmdName),
	}
}

// NewCmdBuild creates a new build command.
func NewCmdBuild(
	fSys filesys.FileSystem, help *Help, writer io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:          help.Use,
		Short:        help.Short,
		Long:         help.Long,
		Example:      help.Example,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := Validate(args); err != nil {
				return err
			}
			k := krusty.MakeKustomizer(
				HonorKustomizeFlags(krusty.MakeDefaultOptions(), cmd.Flags()),
			)
			var keys []string
			m, keys, err := k.Run(fSys, theArgs.kustomizationPath)

			if err != nil {
				return err
			}
			keys = removeEmptyOrNewLine(keys)

			if theFlags.outputPath != "" && fSys.IsDir(theFlags.outputPath) {
				// Ignore writer; write to o.outputPath directly.
				return MakeWriter(fSys).WriteIndividualFiles(
					theFlags.outputPath, m)
			}

			yml, err := m.AsYaml()
			if err != nil {
				return err
			}

			newYml, err := orderKeys(keys, yml)
			if err != nil {
				return err
			}

			if theFlags.outputPath != "" {
				// Ignore writer; write to o.outputPath directly.
				return fSys.WriteFile(theFlags.outputPath, newYml)
			}
			_, err = writer.Write(newYml)
			return err
		},
	}
	AddFlagOutputPath(cmd.Flags())
	AddFunctionBasicsFlags(cmd.Flags())
	AddFlagLoadRestrictor(cmd.Flags())
	AddFlagEnablePlugins(cmd.Flags())
	AddFlagReorderOutput(cmd.Flags())
	AddFlagEnableManagedbyLabel(cmd.Flags())
	msg := "Error marking flag '%s' as deprecated: %v"
	err := cmd.Flags().MarkDeprecated(flagReorderOutputName,
		"use the new 'sortOptions' field in kustomization.yaml instead.")
	if err != nil {
		log.Fatalf(msg, flagReorderOutputName, err)
	}
	err = cmd.Flags().MarkDeprecated(managedByFlag,
		"The flag `enable-managedby-label` has been deprecated. Use the `managedByLabel` option in the `buildMetadata` field instead.")
	if err != nil {
		log.Fatalf(msg, managedByFlag, err)
	}

	AddFlagEnableHelm(cmd.Flags())
	return cmd
}

func removeEmptyOrNewLine(original []string) []string {
	var result []string

	for _, s := range original {
		if s != "\n" && s != "" {
			result = append(result, s)
		}
	}

	return result
}

//TDOD finish this function

func orderKeys(keys []string, yml []byte) ([]byte, error) {
	buffer := bytes.NewBuffer(yml)
	//prevLine := []byte("")
	//_ := -1
	done := false
	var lines []string

	line, err := buffer.ReadBytes('\n')

	for done == false {
		if slices.Contains(keys, strings.TrimSpace(strings.Split(string(line), ":")[0])) {
			println("match")
			//once in here we should be finishing doing all the work. so swap the next len(keys) lines in yml with keys!!
			for i := 0; i < len(keys); i++ {
				lines = append(lines, strings.TrimSpace(string(line)))
				line, err = buffer.ReadBytes('\n')
				done = true
			}
		}
		//prevLine = line
		line, err = buffer.ReadBytes('\n')
	}

	//lines and keys
	//sort lines
	//loop back through yaml
	//repalce yaml lines with lines lines

	/*

		1
		2
		3
		4
		5





	*/

	var newLineIndex = -1
	var oldLineIndex = -1
	for keyIndex, key := range keys {
		for lineIndex, line := range lines {
			if strings.Contains(line, key) {
				newLineIndex = keyIndex
				oldLineIndex = lineIndex
				break
			}
		}
		var tmp = lines[newLineIndex]
		lines[newLineIndex] = lines[oldLineIndex]
		lines[oldLineIndex] = tmp
	}

	println("+++++++ordered+++++++++")
	for _, v := range lines {
		//println(lines[i])
		println(v)
	}
	println("+++++++ordered+++++++++")

	//lines is in order now of the original keys

	ymlLines := bytes.Split(yml, []byte("\n"))

	for i, ymlLine := range ymlLines {
		if slices.Contains(lines, strings.TrimSpace(string(ymlLine))) {
			//replace the nex n lines
			for j := 0; j < len(lines); j++ {
				println("swappinng")
				println(string(ymlLines[i]))
				println(string(lines[j]))
				ymlLines[i] = []byte(lines[j])
				i++
			}
			break
		}
	}

	modifiedData := bytes.Join(ymlLines, []byte("\n"))
	
	if err != nil {
		return nil, err
	}

	println(string(modifiedData))

	return modifiedData, nil
}

// Validate validates build command args and flags.
func Validate(args []string) error {
	if len(args) > 1 {
		return fmt.Errorf(
			"specify one path to " +
				konfig.DefaultKustomizationFileName())
	}
	if len(args) == 0 {
		theArgs.kustomizationPath = filesys.SelfDir
	} else {
		theArgs.kustomizationPath = args[0]
	}
	if err := validateFlagLoadRestrictor(); err != nil {
		return err
	}
	return validateFlagReorderOutput()
}

// HonorKustomizeFlags feeds command line data to the krusty options.
// Flags and such are held in private package variables.
func HonorKustomizeFlags(kOpts *krusty.Options, flags *flag.FlagSet) *krusty.Options {
	kOpts.Reorder = getFlagReorderOutput(flags)
	kOpts.LoadRestrictions = getFlagLoadRestrictorValue()
	if theFlags.enable.plugins {
		c := types.EnabledPluginConfig(types.BploUseStaticallyLinked)
		c.FnpLoadingOptions = theFlags.fnOptions
		kOpts.PluginConfig = c
	} else {
		kOpts.PluginConfig.HelmConfig.Enabled = theFlags.enable.helm
	}
	kOpts.PluginConfig.HelmConfig.Command = theFlags.helmCommand
	kOpts.AddManagedbyLabel = isManagedByLabelEnabled()
	return kOpts
}
