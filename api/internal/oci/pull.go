package oci

import (
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
const notPulled = filesys.ConfirmedDir("/notPulled")

type OciSpec struct {
	raw      string
	provider SourceOCIProvider

	// Host, e.g. ghcr.io, gcr.io, docker.io, etc.
	Registry string

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

const (
	ociPrefix      = "oci://"
	pathSeparator  = "/" // do not use filepath.Separator, as this is a URL
	defaultTimeout = 5 * time.Minute
)

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
	fmt.Printf("\nWorking on: %s", n)
	ociSpec := &OciSpec{
		raw:     n,
		Dir:     notPulled,
		Timeout: defaultTimeout,
	}
	ociSpec.provider.Set("generic")

	// parse repo URL
	ociTag, err := name.NewTag(strings.Replace(n, ociPrefix, "", 1))
	fmt.Printf("\nMy tag is: %v - err: %v", ociTag, err)
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
	dir, err := filesys.NewTmpConfirmedDir()
	if err != nil {
		return err
	}
	ociSpec.Dir = dir
	ociURL, err := fluxClient.ParseArtifactURL(ociSpec.raw)
	fmt.Printf("\n\nCloning: %s - err? %v\n", ociURL, err)
	if err != nil {
		return err
	}
	return nil
}

/*
	timeout := 5 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ociClient := fluxClient.NewLocalClient()

	if localizeFlags.provider.String() == v1beta2.GenericOCIProvider && localizeFlags.creds != "" {
		log.Println("logging in to registry with credentials")
		if err := ociClient.LoginWithCredentials(localizeFlags.creds); err != nil {
			return fmt.Errorf("could not login with credentials: %w", err)
		}
	}

	if localizeFlags.provider.String() != v1beta2.GenericOCIProvider {
		log.Println("logging in to registry with provider credentials")
		ociProvider, err := localizeFlags.provider.ToOCIProvider()
		if err != nil {
			return fmt.Errorf("provider not supported: %w", err)
		}

		if err := ociClient.LoginWithProvider(ctx, ociURL, ociProvider); err != nil {
			return fmt.Errorf("error during login with provider: %w", err)
		}
	}

	log.Printf("pulling artifact from %s", ociURL)

	meta, err := ociClient.Pull(ctx, ociURL, output)
	if err != nil {
		return err
	}

	log.Printf("source %s", meta.Source)
	log.Printf("revision %s", meta.Revision)
	log.Printf("digest %s", meta.Digest)
	log.Printf("artifact content extracted to %s", output)
	return nil
}
*/
