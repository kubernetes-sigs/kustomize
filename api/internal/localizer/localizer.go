// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package localizer

import (
	"log"
	"path/filepath"

	"sigs.k8s.io/kustomize/api/ifc"
	pLdr "sigs.k8s.io/kustomize/api/internal/plugins/loader"
	"sigs.k8s.io/kustomize/api/internal/target"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/yaml"
)

// Localizer encapsulates all state needed to localize the root at ldr.
type Localizer struct {
	fSys filesys.FileSystem

	// kusttarget fields
	validator ifc.Validator
	rFactory  *resmap.Factory
	pLdr      *pLdr.Loader

	// underlying type is Loader
	ldr ifc.Loader

	// destination directory in newDir that mirrors ldr's current root.
	dst filesys.ConfirmedDir
}

// NewLocalizer is the factory method for Localizer
func NewLocalizer(ldr *Loader, validator ifc.Validator, rFactory *resmap.Factory, pLdr *pLdr.Loader) (*Localizer, error) {
	toDst, err := filepath.Rel(ldr.args.Scope.String(), ldr.Root())
	if err != nil {
		log.Fatalf("cannot find path from directory %q to %q inside directory: %s", ldr.args.Scope.String(),
			ldr.Root(), err.Error())
	}
	dst := ldr.args.NewDir.Join(toDst)
	if err = ldr.fSys.MkdirAll(dst); err != nil {
		return nil, errors.WrapPrefixf(err, "unable to create directory in localize destination")
	}
	return &Localizer{
		fSys:      ldr.fSys,
		validator: validator,
		rFactory:  rFactory,
		pLdr:      pLdr,
		ldr:       ldr,
		dst:       filesys.ConfirmedDir(dst),
	}, nil
}

// Localize localizes the root that lc is at
func (lc *Localizer) Localize() error {
	kt := target.NewKustTarget(lc.ldr, lc.validator, lc.rFactory, lc.pLdr)
	err := kt.Load()
	if err != nil {
		return errors.Wrap(err)
	}

	kust := lc.processKust(kt)

	content, err := yaml.Marshal(kust)
	if err != nil {
		return errors.WrapPrefixf(err, "unable to serialize localized kustomization file")
	}
	if err = lc.fSys.WriteFile(lc.dst.Join(konfig.DefaultKustomizationFileName()), content); err != nil {
		return errors.WrapPrefixf(err, "unable to write localized kustomization file")
	}
	return nil
}

// TODO(annasong): implement
// processKust returns a copy of the kustomization at kt with paths localized.
func (lc *Localizer) processKust(kt *target.KustTarget) *types.Kustomization {
	kust := kt.Kustomization()
	return &kust
}
