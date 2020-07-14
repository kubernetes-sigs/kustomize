package builtins_qlik

import (
	"fmt"
	"log"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"sigs.k8s.io/kustomize/api/builtins"
	"sigs.k8s.io/kustomize/api/builtins_qlik/utils"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

type GitImageTagPluginImage struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}

type GitImageTagPlugin struct {
	Images []GitImageTagPluginImage `json:"images,omitempty" yaml:"images,omitempty"`
	pwd    string
	logger *log.Logger
}

func (p *GitImageTagPlugin) Config(h *resmap.PluginHelpers, c []byte) (err error) {
	p.Images = make([]GitImageTagPluginImage, 0)
	p.pwd = h.Loader().Root()
	return yaml.Unmarshal(c, p)
}

func (p *GitImageTagPlugin) Transform(m resmap.ResMap) error {
	if headGitTag, err := getHighestSemverGitTagForHead(p.pwd, p.logger); err != nil {
		return err
	} else {
		for _, image := range p.Images {
			if err := transformForImage(&m, &image, headGitTag); err != nil {
				return err
			}
		}
	}
	return nil
}

func transformForImage(m *resmap.ResMap, gitImageTagPluginImage *GitImageTagPluginImage, newTag string) error {
	if newTag != "" {
		imageTagTransformerPlugin := builtins.ImageTagTransformerPlugin{
			ImageTag: types.Image{
				Name:   gitImageTagPluginImage.Name,
				NewTag: newTag,
			},
		}
		if err := imageTagTransformerPlugin.Transform(*m); err != nil {
			return err
		}
	}
	return nil
}

func getHighestSemverGitTagForHead(dir string, logger *log.Logger) (string, error) {
	type tagVersionT struct {
		tag     string
		version *semver.Version
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
			if version, err := semver.NewVersion(tag); err != nil {
				logger.Printf("could not parse tag: %v as semver version, error: %v\n", tag, err)
			} else {
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
		latestSemverTag = fmt.Sprintf("v0.0.0-%s", headRef.Hash().String()[0:7])
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

func NewGitImageTagPlugin() resmap.TransformerPlugin {
	return &GitImageTagPlugin{logger: utils.GetLogger("GitImageTagPlugin")}
}
