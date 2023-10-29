// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"strings"

	"golang.org/x/exp/slices"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const (
	FromFileFlag              = "from-file"
	FromLiteralFlag           = "from-literal"
	FromEnvFileFlag           = "from-env-file"
	DisableNameSuffixHashFlag = "disableNameSuffixHash"
	BehaviorFlag              = "behavior"
	NamespaceFlag             = "namespace"
	NewNamespaceFlag          = "new-namespace"
	FlagFormat                = "--%s=%s"
)

// ConfigMapSecretFlagsAndArgs encapsulates the options for add secret/configmap commands.
type ConfigMapSecretFlagsAndArgs struct {
	// Name of ConfigMap/Secret (required)
	Name string
	// FileSources to derive the ConfigMap/Secret from (optional)
	FileSources []string
	// LiteralSources to derive the ConfigMap/Secret from (optional)
	LiteralSources []string
	// EnvFileSource to derive the ConfigMap/Secret from (optional)
	// TODO: Rationalize this name with Generic.EnvSource
	EnvFileSource string
	// Resource generation behavior (optional)
	Behavior string
	// Type of secret to create
	Type string
	// Namespace of ConfigMap/Secret (optional) -- if unspecified, default is assumed
	Namespace string
	// Disable name suffix
	DisableNameSuffixHash bool
	// NewNamespace for ConfigMap/Secret (optional) -- only for 'edit set' command
	NewNamespace string
}

// ValidateAdd validates required fields are set to support structured generation for the
// edit add command.
func (a *ConfigMapSecretFlagsAndArgs) ValidateAdd(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("name must be specified once")
	}
	a.Name = args[0]

	if len(a.EnvFileSource) == 0 && len(a.FileSources) == 0 && len(a.LiteralSources) == 0 {
		return fmt.Errorf("at least from-env-file, or from-file or from-literal must be set")
	}
	if len(a.EnvFileSource) > 0 && (len(a.FileSources) > 0 || len(a.LiteralSources) > 0) {
		return fmt.Errorf("from-env-file cannot be combined with from-file or from-literal")
	}
	if len(a.Behavior) > 0 && types.NewGenerationBehavior(a.Behavior) == types.BehaviorUnspecified {
		return fmt.Errorf(`invalid behavior: must be one of "%s", "%s", or "%s"`,
			types.BehaviorCreate, types.BehaviorMerge, types.BehaviorReplace)
	}
	// TODO: Should we check if the path exists? if it's valid, if it's within the same (sub-)directory?
	return nil
}

// ValidateSet validates required fields are set to support structured generation for the
// edit set command.
func (a *ConfigMapSecretFlagsAndArgs) ValidateSet(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("name must be specified once")
	}
	a.Name = args[0]

	if len(a.LiteralSources) == 0 && a.NewNamespace == "" {
		return fmt.Errorf("at least one of [--from-literal, --new-namespace] must be specified")
	}

	return nil
}

// ExpandFileSource normalizes a string list, possibly
// containing globs, into a validated, globless list.
// For example, this list:
// - some/path
// - some/dir/a*
// - bfile=some/dir/b*
// becomes:
// - some/path
// - some/dir/airplane
// - some/dir/ant
// - some/dir/apple
// - bfile=some/dir/banana
// i.e. everything is converted to a key=value pair,
// where the value is always a relative file path,
// and the key, if missing, is the same as the value.
// In the case where the key is explicitly declared,
// the globbing, if present, must have exactly one match.
func (a *ConfigMapSecretFlagsAndArgs) ExpandFileSource(fSys filesys.FileSystem) error {
	const NumberComponentsWithKey = 2

	var results []string
	for _, pattern := range a.FileSources {
		var patterns []string
		key := ""
		// check if the pattern is in `--from-file=[key=]source` format
		// and if so split it to send only the file-pattern to glob function
		s := strings.Split(pattern, "=")
		if len(s) == NumberComponentsWithKey {
			patterns = append(patterns, s[1])
			key = s[0]
		} else {
			patterns = append(patterns, s[0])
		}
		result, err := GlobPatterns(fSys, patterns)
		if err != nil {
			return err
		}
		// if the format is `--from-file=[key=]source` accept only one result
		// and extend it with the `key=` prefix
		if key != "" {
			if len(result) != 1 {
				return fmt.Errorf(
					"'pattern '%s' catches files %v, should catch only one", pattern, result)
			}
			fileSource := fmt.Sprintf("%s=%s", key, result[0])
			results = append(results, fileSource)
		} else {
			results = append(results, result...)
		}
	}
	a.FileSources = results
	return nil
}

// UpdateLiteralSources looks for literal sources that already exist and tries
// to replace their values with new values.
// The key specified must exist in the target resource (ConfigMap or Secret).
func UpdateLiteralSources(
	args *types.GeneratorArgs,
	flags ConfigMapSecretFlagsAndArgs,
) error {
	sources := make(map[string]any)
	for _, val := range args.LiteralSources {
		key, value, found := strings.Cut(val, "=")
		if !found { // this is unlikely to happen since these are already validated during the add process
			return fmt.Errorf("invalid format: literal values must be specified in the key=value format")
		}
		sources[key] = value
	}

	for _, val := range flags.LiteralSources {
		key, value, found := strings.Cut(val, "=")
		if !found {
			return fmt.Errorf("invalid format: literal values must be specified in the key=value format")
		}

		if _, ok := sources[key]; !ok {
			return fmt.Errorf("key '%s' not found in resource", key)
		}

		sources[key] = value
	}

	// re-assemble key-pairs
	newLiteralSources := make([]string, 0)
	for key, val := range sources {
		newLiteralSources = append(newLiteralSources, fmt.Sprintf("%s=%s", key, val))
	}

	// guarantee order is predictable
	slices.Sort(newLiteralSources)

	args.LiteralSources = newLiteralSources

	return nil
}

func MergeFlagsIntoGeneratorArgs(args *types.GeneratorArgs, flags ConfigMapSecretFlagsAndArgs) {
	if len(flags.LiteralSources) > 0 {
		args.LiteralSources = append(
			args.LiteralSources, flags.LiteralSources...)
	}
	if len(flags.FileSources) > 0 {
		args.FileSources = append(
			args.FileSources, flags.FileSources...)
	}
	if flags.EnvFileSource != "" {
		args.EnvSources = append(
			args.EnvSources, flags.EnvFileSource)
	}
	if flags.DisableNameSuffixHash {
		args.Options = &types.GeneratorOptions{
			DisableNameSuffixHash: true,
		}
	}
	if flags.Behavior != "" {
		args.Behavior = flags.Behavior
	}
}
