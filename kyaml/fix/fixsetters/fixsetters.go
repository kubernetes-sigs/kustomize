// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fixsetters

import (
	"os"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/setters2"
	"sigs.k8s.io/kustomize/kyaml/setters2/settersutil"
)

// SetterFixer fixes setters in the input package
type SetterFixer struct {
	// PkgPath is path to the resource package
	PkgPath string

	// OpenAPIPath is path to the openAPI file in the package
	OpenAPIPath string

	// DryRun only displays the actions without performing
	DryRun bool
}

// SetterFixerV1Result holds the results of V1 setters fix
type SetterFixerV1Result struct {
	// NeedFix indicates if the resource in pkgPath are on V1 version of setters
	// and need to be fixed
	NeedFix bool

	// CreatedSetters are setters created as part of this fix
	CreatedSetters []string

	// CreatedSubst are substitutions created as part of this fix
	CreatedSubst []string

	// FailedSetters are setters failed to create from current V1 setters
	FailedSetters map[string]error

	// FailedSubst are substitutions failed to be created from V1 partial setters
	FailedSubst map[string]error
}

// FixSettersV1 reads the package and upgrades v1 version of setters
// to latest
func (f *SetterFixer) FixV1Setters() (SetterFixerV1Result, error) {
	sfr := SetterFixerV1Result{
		FailedSetters: make(map[string]error),
		FailedSubst:   make(map[string]error),
	}
	// KrmFile need not exist for dryRun
	if !f.DryRun {
		_, err := os.Stat(f.OpenAPIPath)
		if err != nil {
			return sfr, err
		}
	}

	var err error
	// lookup for all setters and partial setters in v1 format
	// delete the v1 format comments after lookup
	l := UpgradeV1Setters{}
	if f.DryRun {
		err = applyReadFilter(&l, f.PkgPath)
	} else {
		err = applyWriteFilter(&l, f.PkgPath)
	}
	if err != nil {
		return sfr, err
	}
	if len(l.SetterCounts) > 0 {
		sfr.NeedFix = true
	} else {
		return sfr, nil
	}

	// for each v1 setter create the equivalent in v2,
	for _, setter := range l.SetterCounts {
		sd := setters2.SetterDefinition{
			Name:        setter.Name,
			Value:       setter.Value,
			Description: setter.Description,
			SetBy:       setter.SetBy,
			Type:        setter.Type,
		}
		var err error
		if !f.DryRun {
			err = sd.AddToFile(f.OpenAPIPath)
		}

		if err != nil {
			sfr.FailedSetters[setter.Name] = err
		} else {
			sfr.CreatedSetters = append(sfr.CreatedSetters, setter.Name)
		}
	}

	// for each group of partial setters, create equivalent substitution
	for _, subst := range l.Substitutions {
		sc := settersutil.SubstitutionCreator{
			Name:          subst.Name,
			FieldValue:    subst.FieldVale,
			Pattern:       subst.Pattern,
			ResourcesPath: f.PkgPath,
			OpenAPIPath:   f.OpenAPIPath,
		}
		var err error
		if !f.DryRun {
			err = applyWriteFilter(&sc, f.PkgPath)
		}
		if err != nil {
			sfr.FailedSubst[subst.Name] = err
		} else {
			sfr.CreatedSubst = append(sfr.CreatedSubst, subst.Name)
		}
	}

	return sfr, nil
}

func applyWriteFilter(f kio.Filter, pkgPath string) error {
	rw := &kio.LocalPackageReadWriter{
		PackagePath: pkgPath,
	}
	return kio.Pipeline{
		Inputs:  []kio.Reader{rw},
		Filters: []kio.Filter{f},
		Outputs: []kio.Writer{rw},
	}.Execute()
}

func applyReadFilter(f kio.Filter, pkgPath string) error {
	rw := &kio.LocalPackageReader{
		PackagePath: pkgPath,
	}
	return kio.Pipeline{
		Inputs:  []kio.Reader{rw},
		Filters: []kio.Filter{f},
	}.Execute()
}
