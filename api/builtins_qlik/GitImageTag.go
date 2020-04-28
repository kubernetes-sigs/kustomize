package builtins_qlik

import (
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
	Name    string `json:"name,omitempty" yaml:"name,omitempty"`
	Default string `json:"default,omitempty" yaml:"default,omitempty"`
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
	if newTag == "" {
		newTag = gitImageTagPluginImage.Default
	}
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
	if tagsForHead, err := getGitTagsForHead(dir); err != nil {
		return "", err
	} else {
		return getHighestSemverTag(tagsForHead, logger)
	}
}

func getGitTagsForHead(dir string) ([]string, error) {
	tagsForRef := make([]string, 0)
	if r, err := git.PlainOpenWithOptions(dir, &git.PlainOpenOptions{DetectDotGit: true}); err != nil {
		return nil, err
	} else if ref, err := r.Head(); err != nil {
		return nil, err
	} else if refIter, err := r.Storer.IterReferences(); err != nil {
		return nil, err
	} else if err := refIter.ForEach(func(t *plumbing.Reference) error {
		if t.Name().IsTag() && t.Hash() == ref.Hash() {
			tagsForRef = append(tagsForRef, t.Name().Short())
		}
		return nil
	}); err != nil {
		return nil, err
	} else {
		return tagsForRef, nil
	}
}

func getHighestSemverTag(tags []string, logger *log.Logger) (string, error) {
	type tagVersionT struct {
		tag     string
		version *semver.Version
	}
	var tagVersions []*tagVersionT
	for _, tag := range tags {
		if version, err := semver.NewVersion(tag); err != nil {
			logger.Printf("error parsing tag: %v as semver version, error: %v\n", tag, err)
		} else {
			tagVersions = append(tagVersions, &tagVersionT{
				tag:     tag,
				version: version,
			})
		}
	}
	sort.SliceStable(tagVersions, func(i, j int) bool {
		return tagVersions[j].version.LessThan(tagVersions[i].version)
	})
	if len(tagVersions) == 0 {
		return "", nil
	}
	return tagVersions[0].tag, nil
}

func NewGitImageTagPlugin() resmap.TransformerPlugin {
	return &GitImageTagPlugin{logger: utils.GetLogger("GitImageTagPlugin")}
}
