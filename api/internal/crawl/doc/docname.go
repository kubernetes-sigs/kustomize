package doc

import (
	"crypto/sha256"
	"fmt"
	"path"
	"strings"
	"time"

	"sigs.k8s.io/kustomize/api/internal/git"
)

type Document struct {
	RepositoryURL string `json:"repositoryUrl,omitempty"`
	// User makes it easy to aggregate data in the user level instead
	// of the repository level
	User          string     `json:"user,omitempty"`
	FilePath      string     `json:"filePath,omitempty"`
	DefaultBranch string     `json:"defaultBranch,omitempty"`
	DocumentData  string     `json:"document,omitempty"`
	CreationTime  *time.Time `json:"creationTime,omitempty"`
	IsSame        bool       `json:"-"`
	// FileType can be one of the following:
	// "generator", "transformer", "resource", "".
	FileType string `json:"fileType,omitempty"`
}

// Implements the CrawlerDocument interface.
func (doc *Document) GetDocument() *Document {
	return doc
}

func (doc *Document) Copy() *Document {
	return &Document{
		RepositoryURL: doc.RepositoryURL,
		User:          doc.User,
		FilePath:      doc.FilePath,
		DefaultBranch: doc.DefaultBranch,
		DocumentData:  doc.DocumentData,
		CreationTime:  doc.CreationTime,
		IsSame:        doc.IsSame,
		FileType:      doc.FileType,
	}
}

func (doc *Document) Path() string {
	return fmt.Sprintf("repoURL: %s filePath: %s branch: %s",
		doc.RepositoryURL, doc.FilePath, doc.DefaultBranch)
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
			User:          UserName(repoSpec.Host + path.Clean(repoSpec.OrgRepo)),
		}, nil
	}
	// else document is probably relative path.

	ret := Document{
		RepositoryURL: doc.RepositoryURL,
		DefaultBranch: doc.DefaultBranch,
		User:          UserName(doc.RepositoryURL),
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
	sum := sha256.Sum256([]byte(strings.Join(
		[]string{
			doc.RepositoryURL,
			doc.DefaultBranch,
			doc.FilePath,
		},
		"---|---")))
	return fmt.Sprintf("%x", sum)
}

func (doc *Document) RepositoryFullName() string {
	url := TrimUrl(doc.RepositoryURL)
	sections := strings.Split(url, "/")
	l := len(sections)
	if l < 2 {
		return url
	}
	return path.Join(sections[l-2], sections[l-1])
}

// TrimUrl removes all the trailing slashes and the "git@github.com:" prefix (if exists).
func TrimUrl(s string) string {
	url := strings.TrimRight(s, "/")

	gitPrefix := "git@github.com:"
	if strings.HasPrefix(url, gitPrefix) {
		url = url[len(gitPrefix):]
	}
	return url
}

func UserName(repositoryURL string) string {
	url := TrimUrl(repositoryURL)
	sections := strings.Split(url, "/")
	l := len(sections)
	if l < 2 {
		return url
	}
	return sections[l-2]
}
