// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"log"

	"github.com/distribution/reference"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type PushOptions struct {
	fSys          filesys.FileSystem
	kustomization *types.Kustomization
	kustFileName  string
	targets       []reference.NamedTagged
	annotations   map[string]string
}

func PushToOciRegistries(options *PushOptions) error {
	// ctx := context.Background()

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

	// var dir string = filepath.Dir(mf.GetPath())

	// fs, err := file.New("")
	// if err != nil {
	// 	return err
	// }
	// defer fs.Close()

	// fileDescriptor, err := fs.Add(ctx, ".", "", dir)
	// if err != nil {
	// 	return err
	// }

	// d := memory.New()

	// mediaTypes := make(map[string]string)

	// if err := options.fSys.Walk(string(options.root), func(path string, info os.FileInfo, err error) error {
	// 	if err != nil {
	// 		return err
	// 	}

	// 	if info.IsDir() {
	// 		return nil
	// 	}

	// 	var b string = filepath.Base(path)
	// 	for _, kfilename := range konfig.RecognizedKustomizationFileNames() {
	// 		if b == kfilename {
	// 			mediaTypes[path] = "application/vnd.kustomize.config.k8s.io.v1beta1+yaml"
	// 			return nil
	// 		}
	// 	}

	// 	mediaTypes[path] = "application/vnd.kustomize.unknown.v1beta1"

	// 	return nil
	// }); err != nil {
	// 	return err
	// }

	// ms := memory.New()

	// st, err := oci.New("")
	// vs, err := file.New("")

	// descriptors := make([]v1.Descriptor, len(mediaTypes))

	// for path, mediaType := range mediaTypes {
	// 	bt, err := options.fSys.ReadFile(path)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	descriptor := v1.Descriptor{
	// 		MediaType: mediaType,
	// 		Digest:    digest.FromBytes(bt),
	// 		Size:      int64(len(bt)),
	// 	}

	// 	descriptors = append(descriptors, descriptor)

	// 	ms.Push(ctx, descriptor, bytes.NewReader(bt))

	// }

	// oras.PackManifest(ctx, d, oras.PackManifestVersion1_1, "sdfsdf", oras.PackManifestOptions{})

	// f, err := fSys.Open("")
	// defer f.Close()

	// d.Push(ctx, fileDescriptor, f)

	// opts := oras.PackManifestOptions{
	// 	Layers: []v1.Descriptor{
	// 		fileDescriptor,
	// 	},
	// 	ManifestAnnotations: map[string]string{
	// 		"org.opencontainers.image.created": o.createdAt.Format(time.RFC3339),
	// 	},
	// }

	// artifactType := fmt.Sprintf("application/vnd.%s+%s", strings.ToLower(strings.ReplaceAll(kustomization.APIVersion, "/", ".")), strings.ToLower(kustomization.Kind))
	// manifestDescriptor, err := oras.PackManifest(ctx, fs, oras.PackManifestVersion1_1, artifactType, opts)
	// if err != nil {
	// 	return err
	// }

	// tag := "latest"
	// if err = fs.Tag(ctx, manifestDescriptor, tag); err != nil {
	// 	return err
	// }

	// dst, err := oci.New(o.registry)
	// if err != nil {
	// 	return err
	// }

	// // }
	// // reg, err := remote.NewRegistry(o.registry)
	// // reg.PlainHTTP = true

	// // dst, err := reg.Repository(ctx, "destination")
	// // if err != nil {
	// // 	panic(err) // Handle error
	// // }

	// desc, err := oras.Copy(ctx, fs, tag, dst, tag, oras.DefaultCopyOptions)
	// if err != nil {
	// 	return err
	// }

	// log.Printf(`SUCCESS: published %s:%s@%s\n`, o.registry, tag, desc.Digest)

	// return nil

	// src, err := oci.New(repoSpec.RepoPath)
	// if err != nil {
	// 	return err
	// }
	// dir, err := filesys.NewTmpConfirmedDir()
	// if err != nil {
	// 	return err
	// }

	// repoSpec.Dir = dir

	// fs, err := file.New(dir.String())
	// if err != nil {
	// 	return err
	// }
	// defer fs.Close()

	// reference := "latest"
	// if repoSpec.Tag != "" {
	// 	reference = repoSpec.Tag
	// } else if repoSpec.Digest != "" {
	// 	reference = repoSpec.Digest
	// }

	// desc, err := oras.Copy(ctx, src, reference, fs, "", oras.DefaultCopyOptions)
	// if err != nil {
	// 	return err
	// } else if repoSpec.Digest != "" && repoSpec.Digest != desc.Digest.String() {
	// 	return errors.Errorf("expected digest %s, but pulled artifact with digest %s", repoSpec.Digest, desc.Digest)
	// }

	return nil
}
