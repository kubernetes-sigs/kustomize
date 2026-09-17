// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type httpDoerFunc func(*http.Request) (*http.Response, error)

func (f httpDoerFunc) Do(request *http.Request) (*http.Response, error) {
	return f(request)
}

func httpResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode:    statusCode,
		Status:        fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		Header:        make(http.Header),
		Body:          io.NopCloser(strings.NewReader(body)),
		ContentLength: int64(len(body)),
	}
}

func TestGitHubSourceProviderResolvesAnnotatedTag(t *testing.T) {
	tagSHA := strings.Repeat("a", 40)
	commitSHA := strings.Repeat("b", 40)
	requests := 0
	provider := &githubSourceProvider{
		apiBaseURL: "https://api.example.test",
		rawBaseURL: "https://raw.example.test",
		token:      "token",
		httpClient: httpDoerFunc(func(request *http.Request) (*http.Response, error) {
			requests++
			require.Equal(t, "Bearer token", request.Header.Get("Authorization"))
			require.Equal(t, "application/vnd.github+json", request.Header.Get("Accept"))
			switch request.URL.Path {
			case "/repos/kubernetes/kubernetes/git/ref/tags/v1.37.0":
				return httpResponse(http.StatusOK, `{"object":{"type":"tag","sha":"`+tagSHA+`"}}`), nil
			case "/repos/kubernetes/kubernetes/git/tags/" + tagSHA:
				return httpResponse(http.StatusOK, `{"object":{"type":"commit","sha":"`+commitSHA+`"}}`), nil
			default:
				return nil, errors.New("unexpected request: " + request.URL.String())
			}
		}),
	}

	actual, err := provider.resolve(context.Background(), "v1.37.0")
	require.NoError(t, err)
	require.Equal(t, commitSHA, actual)
	require.Equal(t, 2, requests)
}

func TestGitHubSourceProviderFetchesAndCachesByDigest(t *testing.T) {
	commit := strings.Repeat("c", 40)
	openAPI := `{
  "swagger": "2.0",
  "info": {"title": "test", "version": "v1.37.0"},
  "paths": {},
  "definitions": {"example": {"type": "object"}}
}`
	digest := sha256.Sum256([]byte(openAPI))
	digestString := hex.EncodeToString(digest[:])
	requests := 0
	cacheDir := t.TempDir()
	provider := &githubSourceProvider{
		apiBaseURL: "https://api.example.test",
		rawBaseURL: "https://raw.example.test",
		cacheDir:   cacheDir,
		token:      "must-not-be-sent",
		httpClient: httpDoerFunc(func(request *http.Request) (*http.Response, error) {
			requests++
			require.Equal(t, "/kubernetes/kubernetes/"+commit+"/api/openapi-spec/swagger.json", request.URL.Path)
			require.Empty(t, request.Header.Get("Authorization"))
			require.Equal(t, "identity", request.Header.Get("Accept-Encoding"))
			return httpResponse(http.StatusOK, openAPI), nil
		}),
	}

	actual, err := provider.fetch(context.Background(), "v1.37.0", commit, digestString)
	require.NoError(t, err)
	require.Equal(t, []byte(openAPI), actual)
	require.Equal(t, 1, requests)
	require.FileExists(t, filepath.Join(cacheDir, "v1", digestString+".json.gz"))

	offline := &githubSourceProvider{
		cacheDir: cacheDir,
		offline:  true,
		httpClient: httpDoerFunc(func(request *http.Request) (*http.Response, error) {
			return nil, errors.New("network must not be used")
		}),
	}
	actual, err = offline.fetch(context.Background(), "v1.37.0", commit, digestString)
	require.NoError(t, err)
	require.Equal(t, []byte(openAPI), actual)
}

func TestGitHubSourceProviderDiscardsInvalidCacheEntry(t *testing.T) {
	commit := strings.Repeat("c", 40)
	openAPI := `{"swagger":"2.0","info":{"title":"test","version":"v1"},"paths":{},"definitions":{"x":{"type":"object"}}}`
	digest := sha256.Sum256([]byte(openAPI))
	digestString := hex.EncodeToString(digest[:])
	cacheDir := t.TempDir()
	cachePath := filepath.Join(cacheDir, "v1", digestString+".json.gz")
	require.NoError(t, os.MkdirAll(filepath.Dir(cachePath), 0o755))
	require.NoError(t, writeGzip(cachePath, []byte("wrong contents")))

	requests := 0
	provider := &githubSourceProvider{
		rawBaseURL: "https://raw.example.test",
		cacheDir:   cacheDir,
		httpClient: httpDoerFunc(func(request *http.Request) (*http.Response, error) {
			requests++
			return httpResponse(http.StatusOK, openAPI), nil
		}),
	}
	actual, err := provider.fetch(context.Background(), "v1.37.0", commit, digestString)
	require.NoError(t, err)
	require.Equal(t, []byte(openAPI), actual)
	require.Equal(t, 1, requests)
	cached, err := readInput(cachePath)
	require.NoError(t, err)
	require.Equal(t, []byte(openAPI), cached)
}

func TestGitHubSourceProviderOfflineCacheMissDoesNotUseNetwork(t *testing.T) {
	requests := 0
	provider := &githubSourceProvider{
		cacheDir: t.TempDir(),
		offline:  true,
		httpClient: httpDoerFunc(func(request *http.Request) (*http.Response, error) {
			requests++
			return nil, errors.New("network must not be used")
		}),
	}
	_, err := provider.resolve(context.Background(), "v1.37.0")
	require.ErrorContains(t, err, "offline mode")
	_, err = provider.fetch(context.Background(), "v1.37.0", strings.Repeat("c", 40), strings.Repeat("d", 64))
	require.ErrorContains(t, err, "not present in the cache")
	require.Zero(t, requests)
}

func TestGitHubSourceProviderRejectsUnexpectedResponses(t *testing.T) {
	commit := strings.Repeat("d", 40)
	validOpenAPI := `{"swagger":"2.0","info":{"title":"test","version":"v1"},"paths":{},"definitions":{"x":{"type":"object"}}}`
	malformedOpenAPI := "{"
	malformedDigest := sha256.Sum256([]byte(malformedOpenAPI))

	tests := []struct {
		name      string
		response  func() *http.Response
		expected  string
		wantedErr string
	}{
		{
			name: "HTTP status",
			response: func() *http.Response {
				return httpResponse(http.StatusNotFound, "missing")
			},
			wantedErr: "404",
		},
		{
			name: "content encoding",
			response: func() *http.Response {
				response := httpResponse(http.StatusOK, validOpenAPI)
				response.Header.Set("Content-Encoding", "gzip")
				return response
			},
			wantedErr: "Content-Encoding",
		},
		{
			name: "declared size",
			response: func() *http.Response {
				response := httpResponse(http.StatusOK, validOpenAPI)
				response.ContentLength = maxInputSize + 1
				return response
			},
			wantedErr: "exceeds",
		},
		{
			name: "hash mismatch",
			response: func() *http.Response {
				return httpResponse(http.StatusOK, validOpenAPI)
			},
			expected:  strings.Repeat("e", 64),
			wantedErr: "want locked value",
		},
		{
			name: "malformed OpenAPI",
			response: func() *http.Response {
				return httpResponse(http.StatusOK, malformedOpenAPI)
			},
			expected:  hex.EncodeToString(malformedDigest[:]),
			wantedErr: "validate downloaded OpenAPI",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			provider := &githubSourceProvider{
				rawBaseURL: "https://raw.example.test",
				httpClient: httpDoerFunc(func(request *http.Request) (*http.Response, error) {
					return tc.response(), nil
				}),
			}
			_, err := provider.fetch(context.Background(), "v1.37.0", commit, tc.expected)
			require.ErrorContains(t, err, tc.wantedErr)
		})
	}
}
