// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"fmt"

	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/util"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func runEditSetGenerator(
	flags util.ConfigMapSecretFlagsAndArgs,
	fSys filesys.FileSystem,
	args []string,
	ldr ifc.KvLoader,
	rf *resource.Factory,
	kind string,
	set func(ifc.KvLoader, *types.Kustomization, util.ConfigMapSecretFlagsAndArgs, *resource.Factory) error,
) error {
	if err := flags.ExpandFileSource(fSys); err != nil {
		return fmt.Errorf("failed to expand file source: %w", err)
	}
	if err := flags.ValidateSet(args); err != nil {
		return fmt.Errorf("failed to validate flags: %w", err)
	}
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return fmt.Errorf("failed to load kustomization file: %w", err)
	}
	kustomization, err := mf.Read()
	if err != nil {
		return fmt.Errorf("failed to read kustomization file: %w", err)
	}
	if err := set(ldr, kustomization, flags, rf); err != nil {
		return fmt.Errorf("failed to create %s: %w", kind, err)
	}
	if err := mf.Write(kustomization); err != nil {
		return fmt.Errorf("failed to write kustomization file: %w", err)
	}
	return nil
}

func updateGeneratorArgs(
	args *types.GeneratorArgs,
	flags util.ConfigMapSecretFlagsAndArgs,
	options *types.GeneratorOptions,
) error {
	if len(flags.LiteralSources) > 0 {
		if err := util.UpdateLiteralSources(args, flags); err != nil {
			return fmt.Errorf("failed to update literal sources: %w", err)
		}
	}
	if flags.NewNamespace != "" {
		args.Namespace = flags.NewNamespace
	}
	args.Options = types.MergeGlobalOptionsIntoLocal(args.Options, options)
	return nil
}
