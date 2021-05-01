package misc

import (
	"fmt"

	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/semver"
)

// ModFunc is a function accepting a module, and returning an error.
type ModFunc func(LaModule) error

type LaRepository interface {
	// RepoPath is the import of the of repository,
	// e.g. github.com/kubernetes-sigs/kustomize
	// The directory {srcRoot}/{importPath} should contain a
	// dotGit directory.
	// This directory might be a Go module, or contain directories
	// that are Go modules, or both.
	RepoPath() string

	// AbsPath is the full local filesystem path.
	AbsPath() string

	// FindModule returns a module or nil.
	FindModule(ModuleShortName) LaModule
}

type LaModule interface {
	// ShortName is the module's name without the repo.
	ShortName() ModuleShortName

	// ImportPath is the relative path below the Go src root,
	// which is the same path as would be used to
	// import the module.
	ImportPath() string

	// AbsPath is the absolute path to the module's
	// go.mod file on the local file system.
	AbsPath() string

	// Latest version tagged locally.
	VersionLocal() semver.SemVer

	// Latest version tagged remotely.
	VersionRemote() semver.SemVer

	// Does this module depend on the argument, and
	// if so at what version?
	DependsOn(LaModule) (bool, semver.SemVer)

	// GetReplacements returns a list of replacements.
	GetReplacements() []string

	// GetDisallowedReplacements returns a list of disallowed replacements.
	GetDisallowedReplacements([]string) []string
}

// VersionMap holds the versions associated with modules.
type VersionMap map[ModuleShortName]semver.Versions

func (m VersionMap) Report() {
	for n, versions := range m {
		fmt.Println(n)
		for _, v := range versions {
			fmt.Print("  ")
			fmt.Println(v)
		}
	}
}

func (m VersionMap) Latest(
	n ModuleShortName) semver.SemVer {
	versions := m[n]
	if versions == nil {
		return semver.Zero()
	}
	return versions[0]
}

type LesModules []LaModule

func (s LesModules) LenLongestName() (ans int) {
	for _, m := range s {
		l := len(m.ShortName())
		if l > ans {
			ans = l
		}
	}
	return
}

func (s LesModules) Apply(f ModFunc) error {
	for _, m := range s {
		err := f(m)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s LesModules) Find(target ModuleShortName) LaModule {
	for _, m := range s {
		if m.ShortName() == target {
			return m
		}
	}
	return nil
}

func (s LesModules) GetAllThatDependOn(
	target LaModule) (result TaggedModules) {
	for _, m := range s {
		if yes, v := m.DependsOn(target); yes {
			result = append(result, TaggedModule{M: m, V: v})
		}
	}
	return
}

func (s LesModules) InternalDeps(
	target LaModule) (result TaggedModules) {
	for _, m := range s {
		if yes, v := target.DependsOn(m); yes {
			result = append(result, TaggedModule{M: m, V: v})
		}
	}
	return
}
