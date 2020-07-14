package utils

import (
	"fmt"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func GetHighestSemverGitTagForHead(dir string, defaultSemver string) (string, error) {
	type tagVersionT struct {
		tag     string
		version *semver.Version
	}

	if defaultSemver == "" {
		defaultSemver = "v0.0.0"
	}

	allSemverTags := make([]*tagVersionT, 0)
	headSemverTags := make(map[string]bool)
	var headRef *plumbing.Reference

	if r, err := git.PlainOpenWithOptions(dir, &git.PlainOpenOptions{DetectDotGit: true}); err != nil {
		return "", err
	} else if headRef, err = r.Head(); err != nil {
		return "", err
	} else if refIter, err := r.Storer.IterReferences(); err != nil {
		return "", err
	} else if err := refIter.ForEach(func(t *plumbing.Reference) error {
		if t.Name().IsTag() {
			tag := t.Name().Short()
			if version, err := semver.NewVersion(tag); err == nil {
				tagVersion := &tagVersionT{
					tag:     tag,
					version: version,
				}
				allSemverTags = append(allSemverTags, tagVersion)
				if t.Hash() == headRef.Hash() {
					headSemverTags[tag] = true
				}
			}
		}
		return nil
	}); err != nil {
		return "", err
	}

	latestSemverTag := ""
	if len(allSemverTags) == 0 {
		latestSemverTag = fmt.Sprintf("%s-%s", defaultSemver, headRef.Hash().String()[0:7])
	} else {
		sort.SliceStable(allSemverTags, func(i, j int) bool {
			return allSemverTags[j].version.LessThan(allSemverTags[i].version)
		})
		latestSemverTag = allSemverTags[0].tag
		if _, ok := headSemverTags[latestSemverTag]; !ok {
			latestSemverTag = fmt.Sprintf("%s-%s", latestSemverTag, headRef.Hash().String()[0:7])
		}
	}
	return latestSemverTag, nil
}
