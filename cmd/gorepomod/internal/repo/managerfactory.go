package repo

import (
	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/misc"
	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/mod"
)

// ManagerFactory is a collection of clean data needed to build
// clean, fully wired up instances of Manager.
type ManagerFactory struct {
	dg               *DotGitData
	modules          []*protoModule
	remoteName       misc.TrackedRepo
	versionMapLocal  misc.VersionMap
	versionMapRemote misc.VersionMap
}

func (mf *ManagerFactory) NewRepoManager() *Manager {
	result := &Manager{
		dg:         mf.dg,
		remoteName: mf.remoteName,
	}
	var modules misc.LesModules
	for _, pm := range mf.modules {
		shortName := pm.ShortName(mf.dg.RepoPath())
		modules = append(
			modules,
			mod.New(
				result, shortName, pm.mf,
				mf.versionMapLocal.Latest(shortName),
				mf.versionMapRemote.Latest(shortName)))
	}
	result.modules = modules
	return result
}
