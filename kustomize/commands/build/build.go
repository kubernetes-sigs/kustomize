// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"io"
	"k8s.io/utils/strings/slices"
	"log"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"strings"
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
			var keys [][]string
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

func removeEmptyOrNewLine(original [][]string) [][]string {
	var result [][]string

	for _, keySet := range original {
		var innerResult []string
		for _, key := range keySet {
			if key != "\n" && key != "" {
				innerResult = append(innerResult, key)
			}
		}
		if len(innerResult) > 0 {
			result = append(result, innerResult)
		}

	}

	return result
}

// sort keys in the order they were passed in from kustomization file
func orderKeys(keys [][]string, yml []byte) ([]byte, error) {

	buffer := bytes.NewBuffer(yml)
	var lines [][]string

	line, err := buffer.ReadBytes('\n')

	//reading input yml line by line
	for err == nil {
		//loop through each set of keys
		for _, keys := range keys {
			var innerLines []string
			if slices.Contains(keys, strings.TrimSpace(strings.Split(string(line), ":")[0])) {
				for i := 0; i < len(keys); i++ {
					innerLines = append(innerLines, strings.TrimSuffix(string(line), "\n"))
					line, _ = buffer.ReadBytes('\n')
				}
			}
			if len(innerLines) > 0 {
				lines = append(lines, innerLines)
			}
		}

		line, err = buffer.ReadBytes('\n')
	}

	if err != io.EOF {
		return nil, err
	}

	for _, keySet := range keys {
		for _, lineSet := range lines {
			orderLinesByKeys(keySet, lineSet)
		}
	}

	for _, key := range lines {
		println("*")
		for _, s := range key {
			println(s)
		}
	}

	println("==================New Debugging=======================")

	/*
		keys - [][]string
		lines - [][]string

	*/

	//for _, keys := range keys {
	//	for _, lines := range lines {
	//		println(key)
	//		//orderLinesByKeys(keys, lines)
	//	}
	//}

	//for _, keys := range keys {
	//	for _, s := range keys {
	//		println(s)
	//		//orderLinesByKeys(keys, lines)
	//	}
	//}
	//
	//for _, lines := range lines {
	//	for _, s := range lines {
	//		println(s)
	//	}
	//}

	//
	//// lines is now in order of the original keys
	//
	ymlLines := bytes.Split(yml, []byte("\n"))

	// loops are out of order  outside loop should be lines [][]string
	//

	for _, lineSet := range lines {
		for i := 0; i < len(ymlLines)-1; i++ {

			if slices.Contains(lineSet, string(ymlLines[i])) {
				for _, s := range lineSet {
					println("replacing")
					println(string(ymlLines[i]))
					println(s)
					ymlLines[i] = []byte(s)
					i++
				}
				break
			}
		}

	}

	//for i, ymlLine := range ymlLines {
	//	//println(string(ymlLine))
	//	for _, line1 := range lines {
	//		//for _, s := range line1 {
	//		if slices.Contains(line1, string(ymlLine)) {
	//			for _, line := range line1 {
	//				println("replacing")
	//				println(string(ymlLines[i-1]))
	//				println(string([]byte(line)))
	//				ymlLines[i-1] = []byte(line)
	//				i++
	//			}
	//			break
	//		}
	//		//println(s)
	//		//	println(strings.Contains(string(ymlLine), s))
	//
	//		// replace next n lines
	//
	//		//for j := 0; j < len(line1)-1; j++ {
	//		//	println("replacing")
	//		//	println(string(ymlLines[i]))
	//		//	println(string([]byte(line1[j])))
	//		//	ymlLines[i] = []byte(line1[j])
	//		//	i++
	//		//}
	//		break
	//
	//		break
	//	}
	//	//break
	//
	//	//println("\n")
	//}

	//for i, ymlLine := range ymlLines {
	//	if slices.Contains(lines, strings.TrimSpace(string(ymlLine))) {
	//		//replace the nex n lines
	//		for j := 0; j < len(lines); j++ {
	//			ymlLines[i] = []byte(lines[j])
	//			i++
	//		}
	//		break
	//	}
	//}
	//
	modifiedData := bytes.Join(ymlLines, []byte("\n"))
	//return []byte(""), nil
	return modifiedData, nil
}

func orderLinesByKeys(keys []string, linesToOrder []string) {
	var lines = &linesToOrder

	var newLineIndex = -1
	var oldLineIndex = -1
	for keyIndex, key := range keys {
		for lineIndex, line := range *lines {
			if strings.Contains(line, key) {
				newLineIndex = keyIndex
				oldLineIndex = lineIndex
				break
			}
		}
		var tmp = linesToOrder[newLineIndex]
		linesToOrder[newLineIndex] = linesToOrder[oldLineIndex]
		linesToOrder[oldLineIndex] = tmp
	}
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
