/*
Copyright 2017 The Kubernetes Authors.

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

package commands

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"k8s.io/kubectl/pkg/kustomize/configmapandsecret"
	"k8s.io/kubectl/pkg/kustomize/constants"
	"k8s.io/kubectl/pkg/kustomize/types"
	"k8s.io/kubectl/pkg/kustomize/util/fs"
)

func newCmdAddConfigMap(errOut io.Writer, fsys fs.FileSystem) *cobra.Command {
	var config dataConfig
	cmd := &cobra.Command{
		Use:   "configmap NAME [--from-file=[key=]source] [--from-literal=key1=value1]",
		Short: "Adds a configmap to the kustomization file.",
		Long:  "",
		Example: `
	# Adds a configmap to the kustomization file (with a specified key)
	kustomize edit add configmap my-configmap --from-file=my-key=file/path --from-literal=my-literal=12345

	# Adds a configmap to the kustomization file (key is the filename)
	kustomize edit add configmap my-configmap --from-file=file/path

	# Adds a configmap from env-file
	kustomize edit add configmap my-configmap --from-env-file=env/path.env
`,
		RunE: func(_ *cobra.Command, args []string) error {
			err := config.Validate(args)
			if err != nil {
				return err
			}

			// Load in the kustomization file.
			mf, err := newKustomizationFile(constants.KustomizationFileName, fsys)
			if err != nil {
				return err
			}

			m, err := mf.read()
			if err != nil {
				return err
			}

			// Add the config map to the kustomization file.
			err = addConfigMap(m, config)
			if err != nil {
				return err
			}

			// Write out the kustomization file with added configmap.
			return mf.write(m)
		},
	}

	cmd.Flags().StringSliceVar(
		&config.FileSources,
		"from-file",
		[]string{},
		"Key file can be specified using its file path, in which case file basename will be used as configmap key, or optionally with a key and file path, in which case the given key will be used.  Specifying a directory will iterate each named file in the directory whose basename is a valid configmap key.")
	cmd.Flags().StringArrayVar(
		&config.LiteralSources,
		"from-literal",
		[]string{},
		"Specify a key and literal value to insert in configmap (i.e. mykey=somevalue)")
	cmd.Flags().StringVar(
		&config.EnvFileSource,
		"from-env-file",
		"",
		"Specify the path to a file to read lines of key=val pairs to create a configmap (i.e. a Docker .env file).")

	return cmd
}

// addConfigMap updates a configmap within a kustomization file, using the data in config.
// Note: error may leave kustomization file in an undefined state. Suggest passing a copy
// of kustomization file.
func addConfigMap(m *types.Kustomization, config dataConfig) error {
	cm := getOrCreateConfigMap(m, config.Name)

	err := mergeData(&cm.DataSources, config)
	if err != nil {
		return err
	}

	// Validate by trying to create corev1.configmap.
	_, _, err = configmapandsecret.MakeConfigmapAndGenerateName(*cm)
	if err != nil {
		return err
	}

	return nil
}

func getOrCreateConfigMap(m *types.Kustomization, name string) *types.ConfigMapArgs {
	for i, v := range m.ConfigMapGenerator {
		if name == v.Name {
			return &m.ConfigMapGenerator[i]
		}
	}
	// config map not found, create new one and add it to the kustomization file.
	cm := &types.ConfigMapArgs{Name: name}
	m.ConfigMapGenerator = append(m.ConfigMapGenerator, *cm)
	return &m.ConfigMapGenerator[len(m.ConfigMapGenerator)-1]
}

func mergeData(src *types.DataSources, config dataConfig) error {
	src.LiteralSources = append(src.LiteralSources, config.LiteralSources...)
	src.FileSources = append(src.FileSources, config.FileSources...)
	if src.EnvSource != "" && src.EnvSource != config.EnvFileSource {
		return fmt.Errorf("updating existing env source '%s' not allowed.", src.EnvSource)
	}
	src.EnvSource = config.EnvFileSource

	return nil
}
