package doc

import (
	"path"
	"time"

	"sigs.k8s.io/kustomize/api/internal/git"
)

type Document struct {
	RepositoryURL string     `json:"repositoryUrl,omitempty"`
	FilePath      string     `json:"filePath,omitempty"`
	DefaultBranch string     `json:"defaultBranch,omitempty"`
	DocumentData  string     `json:"document,omitempty"`
	CreationTime  *time.Time `json:"creationTime,omitempty"`
	IsSame        bool       `json:"-"`
}

// Implements the CrawlerDocument interface.
func (doc *Document) GetDocument() *Document {
	return doc
}

// Implements the CrawlerDocument interface.
func (doc *Document) WasCached() bool {
	return doc.IsSame
}

func (doc *Document) FromRelativePath(newFile string) (Document, error) {
	repoSpec, err := git.NewRepoSpecFromUrl(newFile)
	if err == nil {
		return Document{
			RepositoryURL: repoSpec.Host + path.Clean(repoSpec.OrgRepo),
			FilePath:      path.Clean(repoSpec.Path),
			DefaultBranch: repoSpec.Ref,
		}, nil
	}
	// else document is probably relative path.

	ret := Document{
		RepositoryURL: doc.RepositoryURL,
		DefaultBranch: doc.DefaultBranch,
	}
	ogDir, _ := path.Split(doc.FilePath)

	cleaned := path.Clean(newFile)
	if !path.IsAbs(cleaned) {
		cleaned = path.Clean(ogDir + "/" + cleaned)
	}

	ret.FilePath = cleaned
	return ret, nil
}

func (doc *Document) ID() string {
	return doc.RepositoryURL + "/" +
		doc.DefaultBranch + "/" + doc.FilePath
}
