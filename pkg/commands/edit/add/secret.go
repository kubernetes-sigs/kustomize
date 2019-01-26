/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package add

import (
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/pkg/commands/kustfile"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/loader"
	"sigs.k8s.io/kustomize/pkg/types"
)

// newCmdAddSecret returns a new command.
func newCmdAddSecret(fSys fs.FileSystem, kf ifc.KunstructuredFactory) *cobra.Command {
	var flags flagsAndArgs
	cmd := &cobra.Command{
		Use:   "secret NAME [--from-file=[key=]source] [--from-literal=key1=value1] [--type=Opaque|kubernetes.io/tls]",
		Short: "Adds a secret to the kustomization file.",
		Long:  "",
		Example: `
	# Adds a secret to the kustomization file (with a specified key)
	kustomize edit add secret my-secret --from-file=my-key=file/path --from-literal=my-literal=12345

	# Adds a secret to the kustomization file (key is the filename)
	kustomize edit add secret my-secret --from-file=file/path

	# Adds a secret from env-file
	kustomize edit add secret my-secret --from-env-file=env/path.env
`,
		RunE: func(_ *cobra.Command, args []string) error {
			err := flags.ExpandFileSource(fSys)
			if err != nil {
				return err
			}

			err = flags.Validate(args)
			if err != nil {
				return err
			}

			// Load the kustomization file.
			mf, err := kustfile.NewKustomizationFile(fSys)
			if err != nil {
				return err
			}

			kustomization, err := mf.Read()
			if err != nil {
				return err
			}

			// Add the flagsAndArgs map to the kustomization file.
			kf.Set(loader.NewFileLoaderAtCwd(fSys))
			err = addSecret(kustomization, flags, kf)
			if err != nil {
				return err
			}

			// Write out the kustomization file with added secret.
			return mf.Write(kustomization)
		},
	}

	cmd.Flags().StringSliceVar(
		&flags.FileSources,
		"from-file",
		[]string{},
		"Key file can be specified using its file path, in which case file basename will be used as secret "+
			"key, or optionally with a key and file path, in which case the given key will be used.  Specifying a "+
			"directory will iterate each named file in the directory whose basename is a valid secret key.")
	cmd.Flags().StringArrayVar(
		&flags.LiteralSources,
		"from-literal",
		[]string{},
		"Specify a key and literal value to insert in secret (i.e. mykey=somevalue)")
	cmd.Flags().StringVar(
		&flags.EnvFileSource,
		"from-env-file",
		"",
		"Specify the path to a file to read lines of key=val pairs to create a secret (i.e. a Docker .env file).")
	cmd.Flags().StringVar(
		&flags.Type,
		"type",
		"Opaque",
		"Specify the secret type this can be 'Opaque' (default), or 'kubernetes.io/tls'")

	return cmd
}

// addSecret adds a secret to a kustomization file.
// Note: error may leave kustomization file in an undefined state.
// Suggest passing a copy of kustomization file.
func addSecret(
	k *types.Kustomization,
	flags flagsAndArgs, kf ifc.KunstructuredFactory) error {
	secretArgs := makeSecretArgs(k, flags.Name, flags.Type)
	err := mergeFlagsIntoSecretArgs(&secretArgs.DataSources, flags)
	if err != nil {
		return err
	}
	// Validate by trying to create corev1.secret.
	_, err = kf.MakeSecret(secretArgs, k.GeneratorOptions)
	if err != nil {
		return err
	}
	return nil
}

func makeSecretArgs(m *types.Kustomization, name, secretType string) *types.SecretArgs {
	for i, v := range m.SecretGenerator {
		if name == v.Name {
			return &m.SecretGenerator[i]
		}
	}
	// secret not found, create new one and add it to the kustomization file.
	secret := &types.SecretArgs{GeneratorArgs: types.GeneratorArgs{Name: name}, Type: secretType}
	m.SecretGenerator = append(m.SecretGenerator, *secret)
	return &m.SecretGenerator[len(m.SecretGenerator)-1]
}

func mergeFlagsIntoSecretArgs(src *types.DataSources, flags flagsAndArgs) error {
	src.LiteralSources = append(src.LiteralSources, flags.LiteralSources...)
	src.FileSources = append(src.FileSources, flags.FileSources...)
	if src.EnvSource != "" && src.EnvSource != flags.EnvFileSource {
		return fmt.Errorf("updating existing env source '%s' not allowed", src.EnvSource)
	}
	src.EnvSource = flags.EnvFileSource
	return nil
}
