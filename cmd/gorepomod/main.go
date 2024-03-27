// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/arguments"
	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/git"
	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/misc"
	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/repo"
	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/semver"
)

//go:generate go run internal/gen/main.go

func loadRepoManager(args *arguments.Args) (*repo.Manager, error) {
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	dg, err := repo.NewDotGitDataFromPath(path, args.LocalFlag())
	if err != nil {
		return nil, err
	}
	pr, err := dg.NewRepoFactory(args.Exclusions(), args.LocalFlag())
	if err != nil {
		return nil, err
	}

	if args.LocalFlag() {
		return pr.NewRepoManagerWithLocalFlag(args.AllowedReplacements()), nil
	}

	return pr.NewRepoManager(args.AllowedReplacements()), nil
}

func findModule(
	name misc.ModuleShortName, mgr *repo.Manager) (m misc.LaModule, err error) {
	if name != misc.ModuleUnknown {
		m = mgr.FindModule(name)
		if m == nil {
			return nil, fmt.Errorf(
				"cannot find module %q in repo %s", name, mgr.RepoPath())
		}
	}
	return
}

func actualMain() error {
	args, err := arguments.Parse()
	if err != nil {
		return err
	}
	mgr, err := loadRepoManager(args)
	if err != nil {
		return err
	}
	targetModule, err := findModule(args.ModuleName(), mgr)
	if err != nil {
		return err
	}
	conditionalModule, err := findModule(args.ConditionalModule(), mgr)
	if err != nil {
		return err
	}

	switch args.GetCommand() {
	case arguments.List:
		return mgr.List()
	case arguments.Tidy:
		return mgr.Tidy(args.DoIt())
	case arguments.Pin:
		v := args.Version()
		// Update branch history
		gr := git.NewQuiet(mgr.AbsPath(), args.DoIt(), false)
		err = gr.FetchRemote(misc.TrackedRepo(gr.GetMainBranch()))
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		if v.IsZero() {
			// Always use latest tag while does not removing manual usage capability
			releaseTag := string(targetModule.ShortName())
			fmt.Printf("new version not specified, fall back to latest version according to release tag: %s/*\n", releaseTag)
			latest, err := gr.GetLatestTag(releaseTag)
			if err != nil {
				v = targetModule.VersionLocal()
				err = mgr.Pin(args.DoIt(), targetModule, v)
				if err != nil {
					return fmt.Errorf("error: %w", err)
				}
				return nil
			}
			fmt.Printf("setting release tag to %s ...\n", latest)
			v, err = semver.Parse(latest)
			if err != nil {
				v = targetModule.VersionLocal()
				err = mgr.Pin(args.DoIt(), targetModule, v)
				if err != nil {
					return fmt.Errorf("error: %w", err)
				}
				return nil
			}
		}

		// If we can't find revision extracted from latest version specified on release branch
		err = mgr.Pin(args.DoIt(), targetModule, v)
		if err != nil {
			v = targetModule.VersionLocal()
			err = mgr.Pin(args.DoIt(), targetModule, v)
			if err != nil {
				return fmt.Errorf("error: %w", err)
			}
			return nil
		}
		return nil
	case arguments.UnPin:
		err = mgr.UnPin(args.DoIt(), targetModule, conditionalModule)
		if err != nil {
			return fmt.Errorf("error: %w", err)
		}
		return nil
	case arguments.Release:
		err = mgr.Release(targetModule, args.Bump(), args.DoIt(), args.LocalFlag())
		if err != nil {
			return fmt.Errorf("error: %w", err)
		}
		return nil
	case arguments.UnRelease:
		err = mgr.UnRelease(targetModule, args.DoIt(), args.LocalFlag())
		if err != nil {
			return fmt.Errorf("error: %w", err)
		}
		return nil
	case arguments.Debug:
		err = mgr.Debug(targetModule, args.DoIt(), args.LocalFlag())
		if err != nil {
			return fmt.Errorf("error: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("cannot handle cmd %v", args.GetCommand())
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usageMsg)
		return
	}
	if err := actualMain(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
