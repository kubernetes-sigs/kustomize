// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/distribution/reference"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
	"oras.land/oras-go/v2/registry/remote/retry"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type PushOptions struct {
	fSys          filesys.FileSystem
	root          filesys.ConfirmedDir
	kustomization *types.Kustomization
	targets       []reference.NamedTagged
	annotations   map[string]string
}

func validatePath(path string, elementType string) error {
	if path == "" {
		return nil
	}

	u, err := url.Parse(path)
	if err != nil {
		return err
	}
	if u.Scheme == "file" && u.Host == "" {
		return errors.Errorf("Path %s in element %s is a file url relative to localhost", path, elementType)
	} else if u.IsAbs() || u.Host != "" {
		// Other schemes or host-rooted URLs are assumed to be valid....
		return nil
	}

	path = u.Path

	if !filepath.IsLocal(path) {
		return errors.Errorf("Path '%s' in element %s is not local", path, elementType)
	}

	return nil
}

func iteratePathElementsSimple(elements []string, elementType string, errors []error) []error {
	return iteratePathElements(elements, func(x string) string { return x }, elementType, errors)
}

func iteratePathElements[T any](elements []T, fn func(T) string, elementType string, errors []error) []error {
	for _, element := range elements {
		path := fn(element)

		if err := validatePath(path, elementType); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func validateFilePaths(k *types.Kustomization) *[]error {
	errors := []error{}

	errors = iteratePathElementsSimple(k.Components, "Components", errors)
	errors = iteratePathElements(k.Patches, func(x types.Patch) string { return x.Path }, "Patches", errors)
	errors = iteratePathElements(k.Replacements, func(x types.ReplacementField) string { return x.Path }, "Replacements", errors)
	errors = iteratePathElementsSimple(k.Resources, "Resources", errors)
	errors = iteratePathElementsSimple(k.Crds, "Crds", errors)

	for _, generator := range k.ConfigMapGenerator {
		errors = iteratePathElementsSimple(generator.EnvSources, "ConfigMapGenerator", errors)
		errors = iteratePathElementsSimple(generator.FileSources, "ConfigMapGenerator", errors)
	}
	for _, generator := range k.SecretGenerator {
		errors = iteratePathElementsSimple(generator.EnvSources, "SecretGenerator", errors)
		errors = iteratePathElementsSimple(generator.FileSources, "SecretGenerator", errors)
	}

	for _, charts := range k.HelmCharts {
		errors = iteratePathElementsSimple(append(charts.AdditionalValuesFiles, charts.ValuesFile), "HelmCharts", errors)
	}

	errors = iteratePathElementsSimple(k.Configurations, "Configurations", errors)
	errors = iteratePathElementsSimple(k.Generators, "Generators", errors)
	errors = iteratePathElementsSimple(k.Transformers, "Transformers", errors)
	errors = iteratePathElementsSimple(k.Validators, "Validators", errors)

	return &errors
}

func PushToOciRegistries(options *PushOptions) error {
	ctx := context.Background()

	if len(options.targets) == 0 {
		return errors.Errorf("At least one target is required.")
	}

	if options.kustomization == nil {
		return errors.Errorf("kustomization cannot be null")
	}

	if err := options.kustomization.CheckEmpty(); err != nil {
		return err
	}

	if err := options.kustomization.EnforceFields(); err != nil || len(err) > 0 {
		return errors.Errorf("kustomization has field errors: %v", err)
	}

	if deprecated := options.kustomization.CheckDeprecatedFields(); deprecated != nil && len(*deprecated) > 0 {
		for _, field := range *deprecated {
			log.Println(field)
		}
	}

	options.kustomization.FixKustomization()

	// We attempt to perform validation that the paths are either remote URLs or in the root of the kustomization file.
	// There are limitations:
	//  - This doesn't prevent someone manually creating an invalid artifact, so the reader still has to perform its own validations
	//  - We can only examine the root kustomization.  If there are nested kustomization definitions, we (currently) have to skip those.
	//  - We can only examine the kustomization elements.  If there are resource definitions that reference an invalid file path, we will never see it.
	//  - The paths discovered here are only for validation - the actual list of files added to the artifact will be gathered by walking the directory
	if pathErrors := validateFilePaths(options.kustomization); pathErrors != nil && len(*pathErrors) > 0 {
		return errors.Errorf("kustomization includes non-local file paths: %v", pathErrors)
	}

	dir := options.root.String()
	local, err := file.New(dir)
	if err != nil {
		return err
	}
	defer local.Close()

	descriptor, err := local.Add(ctx, ".", "mediaType", dir)
	if err != nil {
		return err
	}

	manifestOptions := oras.PackManifestOptions{
		Layers: []v1.Descriptor{
			descriptor,
		},

		// ManifestAnnotations: map[string]string{
		// 	"org.opencontainers.image.created": o.createdAt.Format(time.RFC3339),
		// },
	}
	artifactType := fmt.Sprintf("application/vnd.%s+%s", strings.ToLower(strings.ReplaceAll(options.kustomization.APIVersion, "/", ".")), strings.ToLower(options.kustomization.Kind))
	manifestDescriptor, err := oras.PackManifest(ctx, local, oras.PackManifestVersion1_1, artifactType, manifestOptions)
	if err != nil {
		return err
	}

	tag := descriptor.Digest.Encoded()
	if err = local.Tag(ctx, manifestDescriptor, tag); err != nil {
		return err
	}

	credentialStore, err := credentials.NewStoreFromDocker(credentials.StoreOptions{DetectDefaultNativeStore: true})
	if err != nil {
		return err
	}

	log.Printf("Authorization configured: %t", credentialStore.IsAuthConfigured())
	val := os.Getenv("DOCKER_CONFIG")

	log.Printf("docker config: %s", val)

	log.Println("after cred store")

	for _, remoteReference := range options.targets {
		repo, err := remote.NewRepository(remoteReference.Name())
		if err != nil {
			return err
		}

		log.Println("after repo creation")

		repo.Client = &auth.Client{
			Client:     retry.DefaultClient,
			Cache:      auth.NewCache(),
			Credential: credentials.Credential(credentialStore),
		}

		_, err = oras.Copy(ctx, local, tag, repo, remoteReference.Tag(), oras.DefaultCopyOptions)
		if err != nil {
			return err
		}
	}

	return nil
}
