package oci

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	gcrv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// @TODO: Remove once the discussion settles
// ref: https://github.com/kubernetes/kube-openapi/pull/402 + https://github.com/google/gnostic/issues/397
// Pkg github.com/fluxcd/pkg/oci creates a conflict with kube-openapi gnostic version
// Pulling the necessary functions here for the time being

const (
	// UserAgent string used for OCI calls.
	UserAgent = "flux/v2"

	// SourceAnnotation is the OpenContainers annotation for specifying
	// the upstream source of an OCI artifact.
	SourceAnnotation = "org.opencontainers.image.source"

	// RevisionAnnotation is the OpenContainers annotation for specifying
	// the upstream source revision of an OCI artifact.
	RevisionAnnotation = "org.opencontainers.image.revision"

	// CreatedAnnotation is the OpenContainers annotation for specifying
	// the date and time on which the OCI artifact was built (RFC 3339).
	CreatedAnnotation = "org.opencontainers.image.created"

	// OCIRepositoryPrefix is the prefix used for OCIRepository URLs.
	OCIRepositoryPrefix = "oci://"

	// DefaultMaxUntarSize defines the default (100MB) max amount of bytes that Untar will process. (100 times 2, 20 times)
	DefaultMaxUntarSize = 100 << 20

	// UnlimitedUntarSize defines the value which disables untar size checks for maxUntarSize.
	UnlimitedUntarSize = -1

	// bufferSize defines the size of the buffer used when copying the tar file entries.
	bufferSize = 32 * 1024

	// set the directory permissions
	uRWXgRXoRX = 0755
)

// Client holds the options for accessing remote OCI registries.
type Client struct {
	options []crane.Option
}

// Metadata holds the upstream information about on artifact's source.
// https://github.com/opencontainers/image-spec/blob/main/annotations.md
type Metadata struct {
	Created     string            `json:"created"`
	Source      string            `json:"source_url"`      //nolint:tagliatelle
	Revision    string            `json:"source_revision"` //nolint:tagliatelle
	Digest      string            `json:"digest"`
	URL         string            `json:"url"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type tarOpts struct {
	// maxUntarSize represents the limit size (bytes) for archives being decompressed by Untar.
	// When max is a negative value the size checks are disabled.
	maxUntarSize int
}

// TarOption represents options to be applied to Tar.
type TarOption func(*tarOpts)

// NewClient returns an OCI client configured with the given crane options.
func NewClient(opts []crane.Option) *Client {
	options := []crane.Option{
		crane.WithUserAgent(UserAgent),
	}
	options = append(options, opts...)

	return &Client{options: options}
}

// DefaultOptions returns an array containing crane.WithPlatform
// to set the platform to flux.
func DefaultOptions() []crane.Option {
	return []crane.Option{
		crane.WithPlatform(&gcrv1.Platform{
			Architecture: "flux",
			OS:           "flux",
			OSVersion:    "v2",
		}),
	}
}

// GetOptions returns the list of crane.Option used by this Client.
func (c *Client) GetOptions() []crane.Option {
	return c.options
}

// optionsWithContext returns the crane options for the given context.
func (c *Client) optionsWithContext(ctx context.Context) []crane.Option {
	options := []crane.Option{
		crane.WithContext(ctx),
	}
	return append(options, c.options...)
}

// WithRetryBackOff returns a function for setting the given backoff on crane.Option.
func WithRetryBackOff(backoff remote.Backoff) crane.Option {
	return func(options *crane.Options) {
		options.Remote = append(options.Remote, remote.WithRetryBackoff(backoff))
	}
}

// ParseArtifactURL validates the OCI URL and returns the address of the artifact.
func ParseArtifactURL(ociURL string) (string, error) {
	if !strings.HasPrefix(ociURL, OCIRepositoryPrefix) {
		return "", fmt.Errorf("URL must be in format 'oci://<domain>/<org>/<repo>'")
	}

	url := strings.TrimPrefix(ociURL, OCIRepositoryPrefix)
	if _, err := name.ParseReference(url); err != nil {
		return "", fmt.Errorf("'%s' invalid URL: %w", ociURL, err)
	}

	return url, nil
}

// Pull downloads an artifact from an OCI repository and extracts the content to the given directory.
func (c *Client) Pull(ctx context.Context, url, outDir string) (*Metadata, error) {
	ref, err := name.ParseReference(url)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	img, err := crane.Pull(url, c.optionsWithContext(ctx)...)
	if err != nil {
		return nil, fmt.Errorf("[Pull] error: %w", err)
	}

	digest, err := img.Digest()
	if err != nil {
		return nil, fmt.Errorf("parsing digest failed: %w", err)
	}

	manifest, err := img.Manifest()
	if err != nil {
		return nil, fmt.Errorf("parsing manifest failed: %w", err)
	}

	meta, err := MetadataFromAnnotations(manifest.Annotations)
	if err != nil {
		return nil, fmt.Errorf("[Pull] MetadataFromAnnotations error: %w", err)
	}
	meta.Digest = ref.Context().Digest(digest.String()).String()

	layers, err := img.Layers()
	if err != nil {
		return nil, fmt.Errorf("failed to list layers: %w", err)
	}

	if len(layers) < 1 {
		return nil, fmt.Errorf("no layers found in artifact")
	}

	blob, err := layers[0].Compressed()
	if err != nil {
		return nil, fmt.Errorf("extracting first layer failed: %w", err)
	}

	if err = Untar(blob, outDir, WithMaxUntarSize(-1)); err != nil {
		return nil, fmt.Errorf("failed to untar first layer: %w", err)
	}

	return meta, nil
}

// MetadataFromAnnotations parses the OpenContainers annotations and returns a Metadata object.
func MetadataFromAnnotations(annotations map[string]string) (*Metadata, error) {
	created, ok := annotations[CreatedAnnotation]
	if !ok {
		return nil, fmt.Errorf("'%s' annotation not found", CreatedAnnotation)
	}

	source, ok := annotations[SourceAnnotation]
	if !ok {
		return nil, fmt.Errorf("'%s' annotation not found", SourceAnnotation)
	}

	revision, ok := annotations[RevisionAnnotation]
	if !ok {
		return nil, fmt.Errorf("'%s' annotation not found", RevisionAnnotation)
	}

	m := Metadata{
		Created:     created,
		Source:      source,
		Revision:    revision,
		Annotations: annotations,
	}

	return &m, nil
}

// WithMaxUntarSize sets the limit size for archives being decompressed by Untar.
// When max is equal or less than 0 disables size checks.
func WithMaxUntarSize(max int) TarOption {
	return func(t *tarOpts) {
		t.maxUntarSize = max
	}
}

func (t *tarOpts) applyOpts(tarOpts ...TarOption) {
	for _, clientOpt := range tarOpts {
		clientOpt(t)
	}
}

// Untar reads the gzip-compressed tar file from r and writes it into dir.
//
// If dir is a relative path, it cannot ascend from the current working dir.
// If dir exists, it must be a directory.
//
//nolint:gocyclo
func Untar(r io.Reader, dir string, inOpts ...TarOption) (err error) {
	opts := tarOpts{
		maxUntarSize: DefaultMaxUntarSize,
	}
	opts.applyOpts(inOpts...)

	dir = filepath.Clean(dir)
	if !filepath.IsAbs(dir) {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("[Untar] Getwd error: %w", err)
		}

		dir, err = securejoin.SecureJoin(cwd, dir)
		if err != nil {
			return fmt.Errorf("[Untar] SecureJoin error: %w", err)
		}
	}

	fi, err := os.Lstat(dir)
	// Dir does not need to exist, as it can later be created.
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot lstat '%s': %w", dir, err)
	}

	if err == nil && !fi.IsDir() {
		return fmt.Errorf("dir '%s' must be a directory", dir)
	}

	madeDir := map[string]bool{}
	zr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("requires gzip-compressed body: %w", err)
	}
	tr := tar.NewReader(zr)
	processedBytes := 0
	t0 := time.Now()

	// For improved concurrency, this could be optimised by sourcing
	// the buffer from a sync.Pool.
	buf := make([]byte, bufferSize)
	for {
		f, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("tar error: %w", err)
		}
		processedBytes += int(f.Size)
		if opts.maxUntarSize > UnlimitedUntarSize &&
			processedBytes > opts.maxUntarSize {
			return fmt.Errorf("tar %q is bigger than max archive size of %d bytes", f.Name, opts.maxUntarSize)
		}
		if !validRelPath(f.Name) {
			return fmt.Errorf("tar contained invalid name error %q", f.Name)
		}
		rel := filepath.FromSlash(f.Name)
		abs := filepath.Join(dir, rel)

		fi := f.FileInfo()
		mode := fi.Mode()

		switch {
		case mode.IsRegular():
			// Make the directory. This is redundant because it should
			// already be made by a directory entry in the tar
			// beforehand. Thus, don't check for errors; the next
			// write will fail with the same error.
			dir := filepath.Dir(abs)
			if !madeDir[dir] {
				if err := os.MkdirAll(filepath.Dir(abs), uRWXgRXoRX); err != nil {
					return fmt.Errorf("[Untar] MkdirAll error: %w", err)
				}
				madeDir[dir] = true
			}
			if runtime.GOOS == "darwin" && mode&0111 != 0 {
				// The darwin kernel caches binary signatures
				// and SIGKILLs binaries with mismatched
				// signatures. Overwriting a binary with
				// O_TRUNC does not clear the cache, rendering
				// the new copy unusable. Removing the original
				// file first does clear the cache. See #54132.
				err := os.Remove(abs)
				if err != nil && !errors.Is(err, fs.ErrNotExist) {
					return fmt.Errorf("[Untar] Remove error: %w", err)
				}
			}
			wf, err := os.OpenFile(abs, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode.Perm())
			if err != nil {
				return fmt.Errorf("[Untar] OpenFile error: %w", err)
			}

			n, err := copyBuffer(wf, tr, buf)
			if err != nil && !errors.Is(err, io.EOF) {
				return fmt.Errorf("error copying buffer: %w", err)
			}

			if closeErr := wf.Close(); closeErr != nil && err == nil {
				err = closeErr
			}
			if err != nil {
				return fmt.Errorf("error writing to %s: %w", abs, err)
			}
			if n != f.Size {
				return fmt.Errorf("only wrote %d bytes to %s; expected %d", n, abs, f.Size)
			}
			modTime := f.ModTime
			if modTime.After(t0) {
				// Ensures that files extracted are not newer then the
				// current system time.
				modTime = t0
			}
			if !modTime.IsZero() {
				if err = os.Chtimes(abs, modTime, modTime); err != nil {
					return fmt.Errorf("error changing file time %s: %w", abs, err)
				}
			}
		case mode.IsDir():
			if err := os.MkdirAll(abs, uRWXgRXoRX); err != nil {
				return fmt.Errorf("[Untar] MkdirAll error: %w", err)
			}
			madeDir[abs] = true
		default:
			return fmt.Errorf("tar file entry %s contained unsupported file type %v", f.Name, mode)
		}
	}
	return nil
}

// Uses a variant of io.CopyBuffer which ensures that a buffer is being used.
// The upstream version prioritises the use of interfaces WriterTo and ReadFrom
// which in this case causes the entirety of the tar file entry to be loaded
// into memory.
//
// Original source:
// https://github.com/golang/go/blob/6f445a9db55f65e55c5be29d3c506ecf3be37915/src/io/io.go#L405
func copyBuffer(dst io.Writer, src io.Reader, buf []byte) (written int64, err error) {
	if buf == nil {
		return 0, fmt.Errorf("[copyBuffer] error: buffer is nil")
	}
	for {
		nr, er := src.Read(buf)
		if nr > 0 { //nolint:nestif
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = fmt.Errorf("errInvalidWrite")
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}

func validRelPath(p string) bool {
	if p == "" || strings.Contains(p, `\`) || strings.HasPrefix(p, "/") || strings.Contains(p, "../") {
		return false
	}
	return true
}
