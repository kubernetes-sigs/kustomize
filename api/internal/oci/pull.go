package oci

import (
	"context"
	"fmt"
	"strings"
	"time"

	fluxClient "github.com/fluxcd/pkg/oci/client"
	"github.com/google/go-containerregistry/pkg/name"
	"sigs.k8s.io/kustomize/api/filesys"
)

// Used as a temporary non-empty occupant of the cloneDir
// field, as something distinguishable from the empty string
// in various outputs (especially tests). Not using an
// actual directory name here, as that's a temporary directory
// with a unique name that isn't created until clone time.
const (
	notPulled      = filesys.ConfirmedDir("/notPulled")
	pathQuery      = "?path="
	pathSeparator  = "/" // do not use filepath.Separator, as this is a URL
	defaultTimeout = 5 * time.Minute
)

type OciSpec struct {
	raw   string
	image string

	// Host, e.g. ghcr.io, gcr.io, docker.io, etc.
	Registry string

	// Path to the kustomize.yaml file in the OCI archive (if non-root)
	// e.g. prod/cluster-only
	Path string

	// RepoPath name (Path to artifact org/name),
	// e.g. kubernetes-sigs/kustomize
	RepoPath string

	// Dir is where the artifact is unpacked.
	Dir filesys.ConfirmedDir

	// Relative path in the artifact, and in the cloneDir,
	// to a Kustomization.
	KustRootPath string

	// Oci manifest tag reference.
	Tag string

	// Timeout is the maximum duration allowed for pulling artifact.
	Timeout time.Duration
}

// CloneSpec returns a string suitable for "docker pull {spec}".
func (x *OciSpec) CloneSpec() string {
	return fmt.Sprintf("%s:%s", x.Registry+pathSeparator+x.RepoPath, x.Tag)
}

func (x *OciSpec) CloneDir() filesys.ConfirmedDir {
	return x.Dir
}

func (x *OciSpec) Raw() string {
	return x.raw
}

func (x *OciSpec) AbsPath() string {
	return x.Dir.Join(x.KustRootPath)
}

func (x *OciSpec) Cleaner(fSys filesys.FileSystem) func() error {
	return func() error { return fSys.RemoveAll(x.Dir.String()) }
}

// NewRepoSpecFromURL parses git-like urls.
// From strings like git@github.com:someOrg/someRepo.git or
// https://github.com/someOrg/someRepo?ref=someHash, extract
// the different parts of URL, set into a RepoSpec object and return RepoSpec object.
// It MUST return an error if the input is not a git-like URL, as this is used by some code paths
// to distinguish between local and remote paths.
//
// In particular, NewRepoSpecFromURL separates the URL used to clone the repo from the
// elements Kustomize uses for other purposes (e.g. query params that turn into args, and
// the path to the kustomization root within the repo).
func NewOCISpecFromURL(n string) (*OciSpec, error) {
	ociSpec := &OciSpec{
		raw:     n,
		Dir:     notPulled,
		Timeout: defaultTimeout,
	}

	splitUrl := strings.Split(n, pathQuery)
	if len(splitUrl) > 1 {
		ociSpec.Path = splitUrl[1]
	}

	// check if string starts with  "oci://"
	ociURL, err := fluxClient.ParseArtifactURL(splitUrl[0])
	if err != nil {
		return nil, err
	}
	ociSpec.image = ociURL

	// parse repo URL
	ociTag, err := name.NewTag(ociSpec.image)
	if err != nil {
		return nil, err
	}
	ociSpec.Registry = ociTag.RegistryStr()
	ociSpec.RepoPath = ociTag.RepositoryStr()
	ociSpec.Tag = ociTag.TagStr()

	return ociSpec, nil
}

// Puller is a function that can pull an OCI image
type Puller func(ociSpec *OciSpec) error

func PullArtifact(ociSpec *OciSpec) error {
	if ociSpec.Dir.String() == "" || ociSpec.Dir == notPulled {
		dir, err := filesys.NewTmpConfirmedDir()
		if err != nil {
			return err
		}
		ociSpec.Dir = dir
	}
	timeout := 5 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ociClient := fluxClient.NewLocalClient()
	_, err := ociClient.Pull(ctx, ociSpec.image, ociSpec.Dir.String())
	if err != nil {
		return err
	}
	return nil
}
