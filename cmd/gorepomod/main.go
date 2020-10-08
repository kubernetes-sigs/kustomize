package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/arguments"
	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/misc"
	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/repo"
)

//go:generate go run internal/gen/main.go

func loadRepoManager(args *arguments.Args) (*repo.Manager, error) {
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	dg, err := repo.NewDotGitDataFromPath(path)
	if err != nil {
		return nil, err
	}
	pr, err := dg.NewRepoFactory(args.Exclusions())
	if err != nil {
		return nil, err
	}
	return pr.NewRepoManager(), nil
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

	var targetModule misc.LaModule = nil
	if args.ModuleName() != misc.ModuleUnknown {
		targetModule = mgr.FindModule(args.ModuleName())
		if targetModule == nil {
			return fmt.Errorf(
				"cannot find module %q in repo %s",
				args.ModuleName(), mgr.RepoPath())
		}
	}

	switch args.GetCommand() {
	case arguments.List:
		return mgr.List()
	case arguments.Tidy:
		return mgr.Tidy(args.DoIt())
	case arguments.Pin:
		v := args.Version()
		if v.IsZero() {
			v = targetModule.VersionLocal()
		}
		return mgr.Pin(args.DoIt(), targetModule, v)
	case arguments.UnPin:
		return mgr.UnPin(args.DoIt(), targetModule)
	case arguments.Release:
		return mgr.Release(targetModule, args.Bump(), args.DoIt())
	case arguments.UnRelease:
		return mgr.UnRelease(targetModule, args.DoIt())
	case arguments.Debug:
		return mgr.Debug(targetModule, args.DoIt())
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
