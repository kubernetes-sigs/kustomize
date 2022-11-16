// Copyright 2019 The Kubernetes Authors.
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
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type removeSecretOptions struct {
	secretNames []string
}

// newCmdRemoveSecret remove the name of a file containing a secret to the kustomization file.
func newCmdRemoveSecret(fSys filesys.FileSystem) *cobra.Command {
	var o removeSecretOptions

	cmd := &cobra.Command{
		Use: "secret",
		Short: "Removes specified secret" +
			konfig.DefaultKustomizationFileName(),
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
	if len(args) == 0 {
		return errors.New("must specify a secret name")
	}
	if len(args) > 1 {
		return fmt.Errorf("too many arguments: %s; to provide multiple config map options, please separate options by comma", args)
	}
	o.secretNames = strings.Split(args[0], ",")
	return nil
}

// RunRemoveSecret runs Secret command (do real work).
func (o *removeSecretOptions) RunRemoveSecret(fSys filesys.FileSystem) error {
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err
	}

	m, err := mf.Read()
	if err != nil {
		return err
	}

	var newSecrets []types.SecretArgs
	foundSecrets := make(map[string]bool)
	for _, removeName := range o.secretNames {
		foundSecrets[removeName] = false
	}

	for _, currentSecret := range m.SecretGenerator {
		if kustfile.StringInSlice(currentSecret.Name, o.secretNames) {
			foundSecrets[currentSecret.Name] = true
			continue
		}
		newSecrets = append(newSecrets, currentSecret)
	}

	for name, found := range foundSecrets {
		if !found {
			log.Printf("secret %s doesn't exist in kustomization file", name)
		}
	}

	m.SecretGenerator = newSecrets
	return mf.Write(m)
}
