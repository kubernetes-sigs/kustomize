// Copyright 2023 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package remove

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v5/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type removeSecretOptions struct {
	secretNamesToRemove []string
}

// newCmdRemoveSecret removes secretGenerator(s) with the specified name(s).
func newCmdRemoveSecret(fSys filesys.FileSystem) *cobra.Command {
	var o removeSecretOptions

	cmd := &cobra.Command{
		Use: "secret",
		Short: "Removes the specified secret(s) from " +
			konfig.DefaultKustomizationFileName(),
		Long: "",
		Example: `
		remove secret my-secret
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			return o.RunRemoveSecret(fSys)
		},
	}
	return cmd
}

// Validate validates removeSecret command.
func (o *removeSecretOptions) Validate(args []string) error {
	switch {
	case len(args) == 0:
		return errors.New("at least one secret name must be specified")
	case len(args) > 1:
		return fmt.Errorf("too many arguments: %s; to provide multiple secrets to remove, please separate secret names by commas", args)
	}

	o.secretNamesToRemove = strings.Split(args[0], ",")
	return nil
}

// RunRemoveSecret runs Secret command (do real work).
func (o *removeSecretOptions) RunRemoveSecret(fSys filesys.FileSystem) error {
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return fmt.Errorf("could not read kustomization file: %w", err)
	}

	m, err := mf.Read()
	if err != nil {
		return fmt.Errorf("could not read kustomization file contents: %w", err)
	}

	foundSecrets := make(map[string]struct{})

	newSecrets := make([]types.SecretArgs, 0, len(m.SecretGenerator))
	for _, currentSecret := range m.SecretGenerator {
		if kustfile.StringInSlice(currentSecret.Name, o.secretNamesToRemove) {
			foundSecrets[currentSecret.Name] = struct{}{}
			continue
		}
		newSecrets = append(newSecrets, currentSecret)
	}

	if len(foundSecrets) == 0 {
		return fmt.Errorf("no specified secret(s) were found in the %s file",
			konfig.DefaultKustomizationFileName())
	}

	for _, name := range o.secretNamesToRemove {
		if _, found := foundSecrets[name]; !found {
			log.Printf("secret %s doesn't exist in kustomization file", name)
		}
	}
	m.SecretGenerator = newSecrets

	err = mf.Write(m)
	if err != nil {
		return fmt.Errorf("failed to write kustomization file: %w", err)
	}
	return nil
}
