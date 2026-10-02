// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	githubAPIBaseURL       = "https://api.github.com"
	githubRawBaseURL       = "https://raw.githubusercontent.com"
	kubernetesRepository   = "kubernetes/kubernetes"
	kubernetesOpenAPIPath  = "api/openapi-spec/swagger.json"
	openAPICacheEnv        = "KUSTOMIZE_OPENAPI_CACHE_DIR"
	sourceRequestTimeout   = 2 * time.Minute
	maximumGitTagPeelDepth = 8
	maximumGitHubErrorSize = 4 << 10
	maximumGitHubJSONSize  = 1 << 20
)

type httpDoer interface {
	Do(request *http.Request) (*http.Response, error)
}

type githubSourceProvider struct {
	httpClient httpDoer
	apiBaseURL string
	rawBaseURL string
	cacheDir   string
	token      string
	offline    bool
}

type githubGitObject struct {
	Type string `json:"type"`
	SHA  string `json:"sha"`
}

type githubRefResponse struct {
	Object githubGitObject `json:"object"`
}

type githubTagResponse struct {
	Object githubGitObject `json:"object"`
}

func newGitHubSourceProvider(cacheDir string, offline bool) *githubSourceProvider {
	if cacheDir == "" {
		cacheDir = os.Getenv(openAPICacheEnv)
	}
	if cacheDir == "" {
		if userCacheDir, err := os.UserCacheDir(); err == nil {
			cacheDir = filepath.Join(userCacheDir, "kustomize", "openapi")
		}
	}
	return &githubSourceProvider{
		httpClient: &http.Client{Timeout: sourceRequestTimeout},
		apiBaseURL: githubAPIBaseURL,
		rawBaseURL: githubRawBaseURL,
		cacheDir:   cacheDir,
		token:      os.Getenv("GITHUB_TOKEN"),
		offline:    offline,
	}
}

func (p *githubSourceProvider) resolve(ctx context.Context, version string) (string, error) {
	if p.offline {
		return "", fmt.Errorf("Kubernetes %s is not present in the bundle lock and cannot be resolved in offline mode", version)
	}
	refURL := p.apiBaseURL + "/repos/" + kubernetesRepository + "/git/ref/tags/" + url.PathEscape(version)
	var ref githubRefResponse
	if err := p.getGitHubJSON(ctx, refURL, &ref); err != nil {
		return "", err
	}
	object := ref.Object
	seenTags := make(map[string]struct{})
	for depth := 0; depth < maximumGitTagPeelDepth; depth++ {
		if err := validateGitHubObject(object); err != nil {
			return "", fmt.Errorf("resolve tag %s: %w", version, err)
		}
		switch object.Type {
		case "commit":
			return object.SHA, nil
		case "tag":
			if _, found := seenTags[object.SHA]; found {
				return "", fmt.Errorf("resolve tag %s: annotated tag cycle at %s", version, object.SHA)
			}
			seenTags[object.SHA] = struct{}{}
			var tag githubTagResponse
			tagURL := p.apiBaseURL + "/repos/" + kubernetesRepository + "/git/tags/" + object.SHA
			if err := p.getGitHubJSON(ctx, tagURL, &tag); err != nil {
				return "", err
			}
			object = tag.Object
		default:
			return "", fmt.Errorf("resolve tag %s: points to unsupported Git object type %q", version, object.Type)
		}
	}
	return "", fmt.Errorf("resolve tag %s: exceeds annotated tag peel depth %d", version, maximumGitTagPeelDepth)
}

func validateGitHubObject(object githubGitObject) error {
	if object.Type == "" {
		return errors.New("Git object type is empty")
	}
	if err := validateHexDigest(object.SHA, 40); err != nil {
		return fmt.Errorf("invalid Git object SHA: %w", err)
	}
	return nil
}

func (p *githubSourceProvider) getGitHubJSON(ctx context.Context, target string, result interface{}) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return fmt.Errorf("create GitHub API request: %w", err)
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("X-Github-Api-Version", "2022-11-28")
	request.Header.Set("User-Agent", "kustomize-openapi-bundle")
	if p.token != "" {
		request.Header.Set("Authorization", "Bearer "+p.token)
	}
	response, err := p.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("request %s: %w", target, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		message, _ := io.ReadAll(io.LimitReader(response.Body, maximumGitHubErrorSize))
		return fmt.Errorf("request %s: GitHub returned %s: %s",
			target, response.Status, strings.TrimSpace(string(message)))
	}
	input, err := readLimitedInput(response.Body, maximumGitHubJSONSize, "GitHub API response", target)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(input))
	if err := decoder.Decode(result); err != nil {
		return fmt.Errorf("decode GitHub API response from %s: %w", target, err)
	}
	var trailing interface{}
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return fmt.Errorf("decode GitHub API response from %s: expected one JSON value", target)
	}
	return nil
}

func (p *githubSourceProvider) fetch(
	ctx context.Context, version string, commit string, expectedSHA256 string,
) ([]byte, error) {
	if err := validateHexDigest(commit, 40); err != nil {
		return nil, fmt.Errorf("invalid Git commit %q: %w", commit, err)
	}
	if expectedSHA256 != "" {
		if err := validateHexDigest(expectedSHA256, 64); err != nil {
			return nil, fmt.Errorf("invalid locked SHA-256: %w", err)
		}
		input, found, err := p.readCache(expectedSHA256)
		if err != nil {
			return nil, err
		}
		if found {
			return input, nil
		}
	}
	if p.offline {
		return nil, fmt.Errorf("Kubernetes %s OpenAPI with SHA-256 %s is not present in the cache",
			version, expectedSHA256)
	}

	target := p.rawBaseURL + "/" + kubernetesRepository + "/" + commit + "/" + kubernetesOpenAPIPath
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, fmt.Errorf("create OpenAPI request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Accept-Encoding", "identity")
	request.Header.Set("User-Agent", "kustomize-openapi-bundle")
	response, err := p.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", target, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		message, _ := io.ReadAll(io.LimitReader(response.Body, maximumGitHubErrorSize))
		return nil, fmt.Errorf("request %s: GitHub returned %s: %s",
			target, response.Status, strings.TrimSpace(string(message)))
	}
	if response.Header.Get("Content-Encoding") != "" {
		return nil, fmt.Errorf("request %s: unexpected Content-Encoding %q", target, response.Header.Get("Content-Encoding"))
	}
	if response.ContentLength > maxInputSize {
		return nil, fmt.Errorf("OpenAPI source %s exceeds %d bytes", target, maxInputSize)
	}
	input, err := readLimitedInput(response.Body, maxInputSize, "OpenAPI source", target)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(input)
	digestString := hex.EncodeToString(digest[:])
	if expectedSHA256 != "" && digestString != expectedSHA256 {
		return nil, fmt.Errorf("downloaded SHA-256 is %s, want locked value %s", digestString, expectedSHA256)
	}
	if _, err := decodeSwagger(input); err != nil {
		return nil, fmt.Errorf("validate downloaded OpenAPI: %w", err)
	}
	if err := p.writeCache(digestString, input); err != nil {
		return nil, err
	}
	return input, nil
}

func (p *githubSourceProvider) cachePath(sha256Digest string) string {
	if p.cacheDir == "" {
		return ""
	}
	return filepath.Join(p.cacheDir, "v1", sha256Digest+".json.gz")
}

func (p *githubSourceProvider) readCache(expectedSHA256 string) ([]byte, bool, error) {
	path := p.cachePath(expectedSHA256)
	if path == "" {
		return nil, false, nil
	}
	input, err := readInput(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, false, nil
		}
		if removeErr := os.Remove(path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			joined := errors.Join(err, fmt.Errorf("remove corrupt cache entry %q: %w", path, removeErr))
			return nil, false, fmt.Errorf("discard invalid OpenAPI cache entry: %w", joined)
		}
		return nil, false, nil
	}
	digest := sha256.Sum256(input)
	if hex.EncodeToString(digest[:]) != expectedSHA256 {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, false, fmt.Errorf("remove cache entry with invalid digest %q: %w", path, err)
		}
		return nil, false, nil
	}
	return input, true, nil
}

func (p *githubSourceProvider) writeCache(sha256Digest string, input []byte) error {
	path := p.cachePath(sha256Digest)
	if path == "" {
		return nil
	}
	if err := writeGzip(path, input); err != nil {
		return fmt.Errorf("write OpenAPI cache entry %q: %w", path, err)
	}
	return nil
}
