// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// openapi-bundle compiles Kubernetes OpenAPI v2 documents into the compact,
// deterministic bundle embedded by kyaml.
package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	openapi_v2 "github.com/google/gnostic-models/openapiv2"
	"google.golang.org/protobuf/proto"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/kyaml/openapi/internal/builtinopenapi"
)

const (
	gvkExtension                           = "x-kubernetes-group-version-kind"
	patchVersionExceptionLegacyBaseline    = "legacy-baseline"
	patchVersionExceptionOpenAPICorrection = "openapi-correction"
	legacyBuiltinOpenAPIAlias              = "v1.21.2"
	upstreamKubernetesPullRequestURLPrefix = "https://github.com/kubernetes/kubernetes/pull/"
)

const maxInputSize = 64 << 20

type options struct {
	manifest          string
	input             string
	lock              string
	output            string
	scopeOutput       string
	versionOutput     string
	provenanceOutput  string
	legacyProtoOutput string
	kubernetesVersion string
	cacheDir          string
	offline           bool
}

type versionManifest struct {
	Versions               []string                         `json:"versions"`
	PatchVersionExceptions map[string]patchVersionException `json:"patchVersionExceptions,omitempty"`
}

type patchVersionException struct {
	Reason              string `json:"reason"`
	UpstreamPullRequest string `json:"upstreamPullRequest,omitempty"`
}

type sourceDocument struct {
	kubernetesVersion string
	gitCommit         string
	input             []byte
	expectedSHA256    string
	curatedProvenance bool
}

type sourceProvider interface {
	resolve(ctx context.Context, version string) (string, error)
	fetch(ctx context.Context, version string, commit string, expectedSHA256 string) ([]byte, error)
}

type sourceProvenance struct {
	kubernetesVersion string
	gitCommit         string
	sha256            string
	apis              []builtinopenapi.Resource
	curated           bool
}

func main() {
	var opts options
	flag.StringVar(&opts.manifest, "manifest", "", "path to an oldest-first Kubernetes OpenAPI version manifest")
	flag.StringVar(&opts.input, "input", "", "path to a Kubernetes OpenAPI v2 JSON or protobuf document (optionally gzip-compressed)")
	flag.StringVar(&opts.lock, "lock", "", "existing bundle whose source metadata locks downloads (defaults to -output)")
	flag.StringVar(&opts.output, "output", "", "path to the generated .json.gz bundle")
	flag.StringVar(&opts.scopeOutput, "scope-output", "", "optional path to the generated Go resource-scope index")
	flag.StringVar(&opts.versionOutput, "version-output", "", "optional path to generated Go built-in version metadata")
	flag.StringVar(&opts.provenanceOutput, "provenance-output", "", "optional path to generated source provenance test data")
	flag.StringVar(&opts.legacyProtoOutput, "legacy-proto-output", "", "optional path to a deterministic gzip archive of the input protobuf")
	flag.StringVar(&opts.kubernetesVersion, "kubernetes-version", "", "Kubernetes version represented by the input")
	flag.StringVar(&opts.cacheDir, "cache-dir", "", "optional directory for downloaded Kubernetes OpenAPI documents")
	flag.BoolVar(&opts.offline, "offline", false, "use only source documents already present in the download cache")
	flag.Parse()

	if err := run(opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(opts options) error {
	return runWithSourceProvider(opts, newGitHubSourceProvider(opts.cacheDir, opts.offline))
}

func runWithSourceProvider(opts options, provider sourceProvider) error {
	if opts.output == "" {
		return errors.New("-output is required")
	}
	if opts.manifest != "" {
		return runManifest(opts, provider)
	}
	return runSingleSource(opts)
}

func runManifest(opts options, provider sourceProvider) error {
	if opts.input != "" || opts.kubernetesVersion != "" || opts.legacyProtoOutput != "" {
		return errors.New("-manifest cannot be combined with -input, -kubernetes-version, or -legacy-proto-output")
	}
	lockPath := opts.lock
	if lockPath == "" {
		lockPath = opts.output
	}
	manifest, bundle, provenance, err := compileManifest(opts.manifest, lockPath, provider)
	if err != nil {
		return fmt.Errorf("compile version manifest: %w", err)
	}
	if opts.provenanceOutput != "" {
		if err := validateSourceProvenance(provenance); err != nil {
			return fmt.Errorf("validate source provenance: %w", err)
		}
	}
	if err := writeBundle(opts.output, bundle); err != nil {
		return fmt.Errorf("write bundle: %w", err)
	}
	if opts.scopeOutput != "" {
		if err := writeScopeFile(opts.scopeOutput, bundle); err != nil {
			return fmt.Errorf("write resource-scope index: %w", err)
		}
	}
	if opts.versionOutput != "" {
		if err := writeVersionFile(opts.versionOutput, manifest); err != nil {
			return fmt.Errorf("write built-in version metadata: %w", err)
		}
	}
	if opts.provenanceOutput != "" {
		if err := writeProvenanceFile(opts.provenanceOutput, provenance); err != nil {
			return fmt.Errorf("write source provenance: %w", err)
		}
	}
	return nil
}

func runSingleSource(opts options) error {
	if opts.lock != "" || opts.versionOutput != "" || opts.provenanceOutput != "" || opts.cacheDir != "" || opts.offline {
		return errors.New("-lock, -version-output, -provenance-output, -cache-dir, and -offline require -manifest")
	}
	if opts.input == "" || opts.kubernetesVersion == "" {
		return errors.New("either -manifest or both -input and -kubernetes-version are required")
	}

	input, err := readInput(opts.input)
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}
	if opts.legacyProtoOutput != "" && isJSONOpenAPI(input) {
		return errors.New("-legacy-proto-output requires a protobuf -input")
	}
	bundle, err := compile(input, opts.kubernetesVersion)
	if err != nil {
		return fmt.Errorf("compile bundle: %w", err)
	}
	if err := writeBundle(opts.output, bundle); err != nil {
		return fmt.Errorf("write bundle: %w", err)
	}
	if opts.scopeOutput != "" {
		if err := writeScopeFile(opts.scopeOutput, bundle); err != nil {
			return fmt.Errorf("write resource-scope index: %w", err)
		}
	}
	if opts.legacyProtoOutput != "" {
		if err := writeGzip(opts.legacyProtoOutput, input); err != nil {
			return fmt.Errorf("write legacy protobuf archive: %w", err)
		}
	}
	return nil
}

func readInput(path string) ([]byte, error) {
	return readInputWithLimit(path, maxInputSize)
}

func readInputWithLimit(path string, limit int64) (result []byte, resultErr error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("close input %q: %w", path, err))
		}
	}()

	reader := bufio.NewReader(file)
	magic, err := reader.Peek(2)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("inspect input %q: %w", path, err)
	}
	if len(magic) < 2 || magic[0] != 0x1f || magic[1] != 0x8b {
		return readLimitedInput(reader, limit, "input", path)
	}

	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("open gzip input %q: %w", path, err)
	}
	uncompressed, readErr := readLimitedInput(gzipReader, limit, "decompressed input", path)
	closeErr := gzipReader.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close gzip input %q: %w", path, closeErr)
	}
	return uncompressed, nil
}

func readLimitedInput(reader io.Reader, limit int64, description, path string) ([]byte, error) {
	contents, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, fmt.Errorf("read %s %q: %w", description, path, err)
	}
	if int64(len(contents)) > limit {
		return nil, fmt.Errorf("%s %q exceeds %d bytes", description, path, limit)
	}
	return contents, nil
}

func compile(input []byte, kubernetesVersion string) (*builtinopenapi.Bundle, error) {
	return compileSources([]sourceDocument{{
		kubernetesVersion: kubernetesVersion,
		input:             input,
	}}, builtinopenapi.Coverage{Floor: kubernetesVersion, Ceiling: kubernetesVersion})
}

func compileManifest(
	path string, existingBundlePath string, provider sourceProvider,
) (*versionManifest, *builtinopenapi.Bundle, []sourceProvenance, error) {
	manifestBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read manifest %q: %w", path, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(manifestBytes))
	decoder.DisallowUnknownFields()
	var manifest versionManifest
	if err := decoder.Decode(&manifest); err != nil {
		return nil, nil, nil, fmt.Errorf("decode manifest %q: %w", path, err)
	}
	var trailing interface{}
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return nil, nil, nil, fmt.Errorf("decode manifest %q: expected one JSON value", path)
	}
	if err := validateManifest(&manifest); err != nil {
		return nil, nil, nil, fmt.Errorf("validate manifest %q: %w", path, err)
	}
	if provider == nil {
		return nil, nil, nil, errors.New("a source provider is required to compile a manifest")
	}

	lockedSourceList, err := readLockedSources(existingBundlePath)
	if err != nil {
		return nil, nil, nil, err
	}
	if err := validateManifestAgainstLockedSources(&manifest, lockedSourceList); err != nil {
		return nil, nil, nil, fmt.Errorf("validate version list against bundle lock: %w", err)
	}
	lockedSources := make(map[string]builtinopenapi.Source, len(lockedSourceList))
	for _, source := range lockedSourceList {
		lockedSources[source.KubernetesVersion] = source
	}

	documents := make([]sourceDocument, 0, len(manifest.Versions))
	resolvedVersions := make(map[string]struct{})
	ctx := context.Background()
	for i := len(manifest.Versions) - 1; i >= 0; i-- {
		version := manifest.Versions[i]
		locked := lockedSources[version]
		commit := locked.GitCommit
		if commit == "" {
			commit, err = provider.resolve(ctx, version)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("resolve Kubernetes %s tag: %w", version, err)
			}
			resolvedVersions[version] = struct{}{}
		}
		input, err := provider.fetch(ctx, version, commit, locked.SHA256)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("fetch Kubernetes %s OpenAPI: %w", version, err)
		}
		documents = append(documents, sourceDocument{
			kubernetesVersion: version,
			gitCommit:         commit,
			input:             input,
			expectedSHA256:    locked.SHA256,
			curatedProvenance: i == 0 && len(curatedProvenanceAPIs(version)) != 0,
		})
	}
	bundle, provenance, err := compileSourcesWithProvenance(documents, builtinopenapi.Coverage{
		Floor:   manifest.Versions[0],
		Ceiling: manifest.Versions[len(manifest.Versions)-1],
	})
	if err != nil {
		return nil, nil, nil, err
	}
	for _, source := range bundle.Sources {
		if _, wasResolved := resolvedVersions[source.KubernetesVersion]; wasResolved {
			fmt.Printf("resolved Kubernetes %s to %s (SHA-256 %s)\n",
				source.KubernetesVersion, source.GitCommit, source.SHA256)
		}
	}
	return &manifest, bundle, provenance, nil
}

func validateManifestAgainstLockedSources(manifest *versionManifest, lockedSources []builtinopenapi.Source) error {
	if len(manifest.Versions) < len(lockedSources) {
		return errors.New("version list must not remove covered Kubernetes minors from the bundle lock")
	}
	for i := range lockedSources {
		lockedVersion := lockedSources[len(lockedSources)-1-i].KubernetesVersion
		selectedVersion := manifest.Versions[i]
		if selectedVersion == lockedVersion {
			continue
		}
		locked, err := parseKubernetesVersion(lockedVersion)
		if err != nil {
			return fmt.Errorf("bundle lock contains invalid Kubernetes version %q", lockedVersion)
		}
		selected, err := parseKubernetesVersion(selectedVersion)
		if err != nil {
			return err
		}
		if locked[0] != selected[0] || locked[1] != selected[1] {
			return fmt.Errorf("version list must retain locked Kubernetes minor v%d.%d at position %d, got %s",
				locked[0], locked[1], i, selectedVersion)
		}
		exception, found := manifest.PatchVersionExceptions[selectedVersion]
		if !found || exception.Reason != patchVersionExceptionOpenAPICorrection {
			return fmt.Errorf("cannot replace locked source %s with %s without an openapi-correction exception",
				lockedVersion, selectedVersion)
		}
	}
	return nil
}

func validateManifest(manifest *versionManifest) error {
	if len(manifest.Versions) == 0 {
		return errors.New("at least one Kubernetes version is required")
	}
	var previous [3]int
	seenVersions := make(map[string]struct{}, len(manifest.Versions))
	for i, versionString := range manifest.Versions {
		version, err := parseKubernetesVersion(versionString)
		if err != nil {
			return fmt.Errorf("version %d: %w", i, err)
		}
		if i > 0 && compareVersions(previous, version) >= 0 {
			return fmt.Errorf("versions must be ordered oldest-first: %s is not newer than %s",
				versionString, manifest.Versions[i-1])
		}
		if i > 0 && (previous[0] != version[0] || version[1]-previous[1] != 1) {
			return fmt.Errorf("versions must contain exactly one snapshot for every Kubernetes minor between %s and %s",
				manifest.Versions[i-1], versionString)
		}
		previous = version
		seenVersions[versionString] = struct{}{}
		exception, hasException := manifest.PatchVersionExceptions[versionString]
		if err := validatePatchVersionException(version, exception, hasException, i == 0); err != nil {
			return fmt.Errorf("version %s: %w", versionString, err)
		}
	}
	for version := range manifest.PatchVersionExceptions {
		if _, found := seenVersions[version]; !found {
			return fmt.Errorf("patchVersionExceptions contains version %s that is not selected", version)
		}
	}
	return nil
}

func validatePatchVersionException(
	version [3]int, exception patchVersionException, hasException bool, oldest bool,
) error {
	if version[2] == 0 {
		if hasException {
			return errors.New("patchVersionException is only valid for a non-zero patch version")
		}
		return nil
	}
	if !hasException {
		return errors.New("non-zero patch version requires patchVersionException")
	}
	switch exception.Reason {
	case patchVersionExceptionLegacyBaseline:
		if !oldest {
			return errors.New("legacy-baseline is only valid for the oldest source")
		}
		if exception.UpstreamPullRequest != "" {
			return errors.New("legacy-baseline must not specify upstreamPullRequest")
		}
	case patchVersionExceptionOpenAPICorrection:
		if !strings.HasPrefix(exception.UpstreamPullRequest, upstreamKubernetesPullRequestURLPrefix) {
			return fmt.Errorf("openapi-correction requires upstreamPullRequest in the form %s<number>",
				upstreamKubernetesPullRequestURLPrefix)
		}
		pullRequest := strings.TrimPrefix(exception.UpstreamPullRequest, upstreamKubernetesPullRequestURLPrefix)
		value, err := strconv.Atoi(pullRequest)
		if err != nil || value <= 0 || strconv.Itoa(value) != pullRequest {
			return fmt.Errorf("openapi-correction requires upstreamPullRequest in the form %s<number>",
				upstreamKubernetesPullRequestURLPrefix)
		}
	default:
		return fmt.Errorf("unsupported patchVersionException reason %q", exception.Reason)
	}
	return nil
}

func parseKubernetesVersion(version string) ([3]int, error) {
	var result [3]int
	parts := strings.Split(strings.TrimPrefix(version, "v"), ".")
	if !strings.HasPrefix(version, "v") || len(parts) != len(result) {
		return result, fmt.Errorf("invalid Kubernetes version %q", version)
	}
	for i, part := range parts {
		value, err := strconv.Atoi(part)
		if err != nil || value < 0 {
			return result, fmt.Errorf("invalid Kubernetes version %q", version)
		}
		result[i] = value
	}
	return result, nil
}

func compareVersions(left, right [3]int) int {
	for i := range left {
		if left[i] < right[i] {
			return -1
		}
		if left[i] > right[i] {
			return 1
		}
	}
	return 0
}

func validateHexDigest(value string, length int) error {
	if len(value) != length {
		return fmt.Errorf("must contain %d hexadecimal characters", length)
	}
	if _, err := hex.DecodeString(value); err != nil {
		return errors.New("must contain only hexadecimal characters")
	}
	if value != strings.ToLower(value) {
		return errors.New("must use lowercase hexadecimal characters")
	}
	return nil
}

func compileSources(sources []sourceDocument, coverage builtinopenapi.Coverage) (*builtinopenapi.Bundle, error) {
	bundle, _, err := compileSourcesWithProvenance(sources, coverage)
	return bundle, err
}

func compileSourcesWithProvenance(
	sources []sourceDocument, coverage builtinopenapi.Coverage,
) (*builtinopenapi.Bundle, []sourceProvenance, error) {
	if len(sources) == 0 {
		return nil, nil, errors.New("no OpenAPI sources provided")
	}
	definitions := spec.Definitions{}
	resources := map[string]builtinopenapi.Resource{}
	bundleSources := make([]builtinopenapi.Source, 0, len(sources))
	sourceInventories := make([]sourceProvenance, 0, len(sources))

	for _, source := range sources {
		digest := sha256.Sum256(source.input)
		digestString := hex.EncodeToString(digest[:])
		if source.expectedSHA256 != "" && source.expectedSHA256 != digestString {
			return nil, nil, fmt.Errorf("Kubernetes %s source SHA-256 is %s, want %s",
				source.kubernetesVersion, digestString, source.expectedSHA256)
		}
		swagger, err := decodeSwagger(source.input)
		if err != nil {
			return nil, nil, fmt.Errorf("decode Kubernetes %s source: %w", source.kubernetesVersion, err)
		}
		if err := validateDefinitionReferences(swagger.Definitions); err != nil {
			return nil, nil, fmt.Errorf("validate Kubernetes %s definitions: %w", source.kubernetesVersion, err)
		}
		sourceResources, err := collectResources(swagger)
		if err != nil {
			return nil, nil, fmt.Errorf("collect Kubernetes %s resources: %w", source.kubernetesVersion, err)
		}
		sourceInventories = append(sourceInventories, sourceProvenance{
			kubernetesVersion: source.kubernetesVersion,
			gitCommit:         source.gitCommit,
			sha256:            digestString,
			apis:              sourceResources,
			curated:           source.curatedProvenance,
		})
		for name, definition := range swagger.Definitions {
			if _, found := definitions[name]; !found {
				definitions[name] = definition
			}
		}
		for _, resource := range sourceResources {
			key := resourceKey(resource.APIVersion, resource.Kind)
			existing, found := resources[key]
			if !found {
				resources[key] = resource
				continue
			}
			if existing.Definition == "" {
				existing.Definition = resource.Definition
			} else if resource.Definition != "" && existing.Definition != resource.Definition {
				return nil, nil, fmt.Errorf("GVK %s/%s maps to definitions %q and %q",
					resource.APIVersion, resource.Kind, existing.Definition, resource.Definition)
			}
			if existing.Scope == builtinopenapi.ScopeUnknown {
				existing.Scope = resource.Scope
			} else if resource.Scope != builtinopenapi.ScopeUnknown && existing.Scope != resource.Scope {
				return nil, nil, fmt.Errorf("GVK %s/%s has scopes %q and %q",
					resource.APIVersion, resource.Kind, existing.Scope, resource.Scope)
			}
			resources[key] = existing
		}
		bundleSources = append(bundleSources, builtinopenapi.Source{
			KubernetesVersion: source.kubernetesVersion,
			GitCommit:         source.gitCommit,
			SHA256:            digestString,
		})
	}

	if err := validateDefinitionReferences(definitions); err != nil {
		return nil, nil, err
	}
	resourceList := make([]builtinopenapi.Resource, 0, len(resources))
	for _, resource := range resources {
		resourceList = append(resourceList, resource)
	}
	builtinopenapi.SortResources(resourceList)
	if err := normalizeDefinitionGVKs(definitions, resourceList); err != nil {
		return nil, nil, fmt.Errorf("normalize definition GVKs: %w", err)
	}
	bundle := &builtinopenapi.Bundle{
		FormatVersion:   builtinopenapi.FormatVersion,
		Coverage:        coverage,
		SelectionPolicy: builtinopenapi.SelectionPolicy,
		Sources:         bundleSources,
		Definitions:     definitions,
		Resources:       resourceList,
	}
	if err := bundle.Validate(); err != nil {
		return nil, nil, fmt.Errorf("validate compiled bundle: %w", err)
	}
	provenance, err := collectSourceProvenance(sourceInventories)
	if err != nil {
		return nil, nil, fmt.Errorf("collect source provenance: %w", err)
	}
	return bundle, provenance, nil
}

func collectSourceProvenance(inventories []sourceProvenance) ([]sourceProvenance, error) {
	const maximumExamples = 3
	seen := make(map[string]struct{})
	result := make([]sourceProvenance, 0, len(inventories))
	for i := len(inventories) - 1; i >= 0; i-- {
		inventory := inventories[i]
		examples := make([]builtinopenapi.Resource, 0, maximumExamples)
		curated := curatedProvenanceAPIs(inventory.kubernetesVersion)
		if inventory.curated {
			resources := make(map[string]builtinopenapi.Resource, len(inventory.apis))
			for _, resource := range inventory.apis {
				resources[resourceKey(resource.APIVersion, resource.Kind)] = resource
			}
			for _, wanted := range curated {
				resource, found := resources[resourceKey(wanted.APIVersion, wanted.Kind)]
				if !found || resource.Definition != wanted.Definition {
					return nil, fmt.Errorf("Kubernetes %s does not contain curated provenance API %s/%s",
						inventory.kubernetesVersion, wanted.APIVersion, wanted.Kind)
				}
				examples = append(examples, resource)
			}
		} else {
			for _, resource := range inventory.apis {
				key := resourceKey(resource.APIVersion, resource.Kind)
				if _, found := seen[key]; found || !isProvenanceCandidate(resource) {
					continue
				}
				if len(examples) < maximumExamples {
					examples = append(examples, resource)
				}
			}
		}
		for _, resource := range inventory.apis {
			seen[resourceKey(resource.APIVersion, resource.Kind)] = struct{}{}
		}
		result = append(result, sourceProvenance{
			kubernetesVersion: inventory.kubernetesVersion,
			gitCommit:         inventory.gitCommit,
			sha256:            inventory.sha256,
			apis:              examples,
			curated:           inventory.curated,
		})
	}
	return result, nil
}

func curatedProvenanceAPIs(version string) []builtinopenapi.Resource {
	parsed, err := parseKubernetesVersion(version)
	if err != nil || parsed[0] != 1 || parsed[1] != 21 {
		return nil
	}
	// There is no source older than the compatibility floor in the union to
	// compare automatically. These APIs are known Kubernetes v1.21 additions;
	// generation verifies that the selected v1.21 source still advertises them.
	return []builtinopenapi.Resource{
		{APIVersion: "batch/v1", Kind: "CronJob", Definition: "io.k8s.api.batch.v1.CronJob"},
		{APIVersion: "discovery.k8s.io/v1", Kind: "EndpointSlice", Definition: "io.k8s.api.discovery.v1.EndpointSlice"},
		{APIVersion: "policy/v1", Kind: "PodDisruptionBudget", Definition: "io.k8s.api.policy.v1.PodDisruptionBudget"},
	}
}

func isProvenanceCandidate(resource builtinopenapi.Resource) bool {
	// Meta API schemas such as DeleteOptions and WatchEvent are reused as
	// synthetic path responses under many API versions. They do not represent
	// a Kubernetes API introduced in that release and would make poor examples.
	return resource.Definition != "" &&
		!strings.HasPrefix(resource.Definition, "io.k8s.apimachinery.pkg.apis.meta.") &&
		!strings.HasSuffix(resource.Kind, "List")
}

// normalizeDefinitionGVKs makes the resource inventory authoritative for GVK
// discovery. Schema fields still come as a whole from the newest source that
// contains a definition; only the GVK extension is synthesized across sources.
func normalizeDefinitionGVKs(definitions spec.Definitions, resources []builtinopenapi.Resource) error {
	gvksByDefinition := make(map[string][]interface{})
	for _, resource := range resources {
		if resource.Definition == "" {
			continue
		}
		if _, found := definitions[resource.Definition]; !found {
			return fmt.Errorf("resource %s/%s references missing definition %q",
				resource.APIVersion, resource.Kind, resource.Definition)
		}
		group, version := splitAPIVersion(resource.APIVersion)
		gvksByDefinition[resource.Definition] = append(gvksByDefinition[resource.Definition], map[string]interface{}{
			"group":   group,
			"version": version,
			"kind":    resource.Kind,
		})
	}

	for name, definition := range definitions {
		extensions := make(spec.Extensions, len(definition.Extensions)+1)
		for key, value := range definition.Extensions {
			if key != gvkExtension {
				extensions[key] = value
			}
		}
		if gvks := gvksByDefinition[name]; len(gvks) != 0 {
			extensions[gvkExtension] = gvks
		}
		if len(extensions) == 0 {
			definition.Extensions = nil
		} else {
			definition.Extensions = extensions
		}
		definitions[name] = definition
	}
	return nil
}

func splitAPIVersion(apiVersion string) (group, version string) {
	group, version, found := strings.Cut(apiVersion, "/")
	if !found {
		return "", group
	}
	return group, version
}

func decodeSwagger(input []byte) (*spec.Swagger, error) {
	if len(bytes.TrimSpace(input)) == 0 {
		return nil, errors.New("OpenAPI source is empty")
	}
	if isJSONOpenAPI(input) {
		var swagger spec.Swagger
		decoder := json.NewDecoder(bytes.NewReader(input))
		if err := decoder.Decode(&swagger); err != nil {
			return nil, fmt.Errorf("unmarshal OpenAPI JSON: %w", err)
		}
		var trailing interface{}
		if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
			return nil, errors.New("OpenAPI JSON must contain exactly one value")
		}
		return &swagger, nil
	}

	document := &openapi_v2.Document{}
	if err := proto.Unmarshal(input, document); err != nil {
		return nil, fmt.Errorf("unmarshal OpenAPI protobuf: %w", err)
	}

	var swagger spec.Swagger
	ok, err := swagger.FromGnostic(document)
	if err != nil {
		return nil, fmt.Errorf("convert gnostic document: %w", err)
	}
	if !ok {
		return nil, errors.New("gnostic document cannot be converted without data loss")
	}
	return &swagger, nil
}

func isJSONOpenAPI(input []byte) bool {
	return bytes.HasPrefix(bytes.TrimSpace(input), []byte("{"))
}

func readLockedSources(path string) ([]builtinopenapi.Source, error) {
	input, err := readInput(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("read existing bundle lock %q: %w (restore the generated bundle before regenerating)",
				path, err)
		}
		return nil, fmt.Errorf("read existing bundle lock %q: %w", path, err)
	}
	var bundle builtinopenapi.Bundle
	if err := json.Unmarshal(input, &bundle); err != nil {
		return nil, fmt.Errorf("decode existing bundle lock %q: %w", path, err)
	}
	if err := bundle.Validate(); err != nil {
		return nil, fmt.Errorf("validate existing bundle lock %q: %w", path, err)
	}
	return bundle.Sources, nil
}

func writeVersionFile(path string, manifest *versionManifest) error {
	newest := manifest.Versions[len(manifest.Versions)-1]
	version, err := parseKubernetesVersion(newest)
	if err != nil {
		return err
	}
	defaultVersion := fmt.Sprintf("v%d.%d", version[0], version[1])

	aliases := make([]string, 0, len(manifest.Versions)+1)
	seenAliases := make(map[string]struct{}, len(manifest.Versions)+1)
	addAlias := func(alias string) {
		if _, found := seenAliases[alias]; found {
			return
		}
		seenAliases[alias] = struct{}{}
		aliases = append(aliases, alias)
	}
	addAlias("builtin")
	for _, exactVersion := range manifest.Versions {
		parsed, err := parseKubernetesVersion(exactVersion)
		if err != nil {
			return err
		}
		minorAlias := fmt.Sprintf("v%d.%d", parsed[0], parsed[1])
		addAlias(minorAlias)
		if minorAlias == "v1.21" {
			addAlias(legacyBuiltinOpenAPIAlias)
		}
	}

	var source bytes.Buffer
	source.WriteString(`// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Code generated by openapi-bundle. DO NOT EDIT.

package openapi

`)
	fmt.Fprintf(&source, `const (
	// DefaultOpenAPI identifies the built-in union schema. Its suffix is the
	// newest Kubernetes minor represented by the artifact; exact source
	// coverage is recorded in the bundle metadata.
	DefaultOpenAPI = %q

	// BuiltinSchemaInfo is the value printed by kustomize openapi info.
	BuiltinSchemaInfo = "{title:Kubernetes,version:" + DefaultOpenAPI + "}"
)

func hasBuiltinOpenAPIVersion(version string) bool {
	switch version {
`, defaultVersion)
	for _, alias := range aliases {
		fmt.Fprintf(&source, "\tcase %q:\n\t\treturn true\n", alias)
	}
	source.WriteString(`	default:
		return false
	}
}
`)
	formatted, err := format.Source(source.Bytes())
	if err != nil {
		return fmt.Errorf("format generated Go: %w", err)
	}
	return writeFile(path, formatted)
}

func writeProvenanceFile(path string, provenance []sourceProvenance) error {
	if err := validateSourceProvenance(provenance); err != nil {
		return err
	}
	var source bytes.Buffer
	source.WriteString(`// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Code generated by openapi-bundle. DO NOT EDIT.

package openapi

var generatedBuiltinSourceProvenance = []builtinSourceProvenance{ //nolint:gochecknoglobals
`)
	for _, entry := range provenance {
		fmt.Fprintf(&source, "\t{kubernetesVersion: %q, gitCommit: %q, sha256: %q, apis: []builtinProvenanceAPI{\n",
			entry.kubernetesVersion, entry.gitCommit, entry.sha256)
		for _, api := range entry.apis {
			fmt.Fprintf(&source, "\t\t{apiVersion: %q, kind: %q, definition: %q},\n",
				api.APIVersion, api.Kind, api.Definition)
		}
		source.WriteString("\t}},\n")
	}
	source.WriteString("}\n")
	formatted, err := format.Source(source.Bytes())
	if err != nil {
		return fmt.Errorf("format generated Go: %w", err)
	}
	return writeFile(path, formatted)
}

func validateSourceProvenance(provenance []sourceProvenance) error {
	for _, entry := range provenance {
		if err := validateHexDigest(entry.gitCommit, 40); err != nil {
			return fmt.Errorf("Kubernetes %s has invalid source commit: %w", entry.kubernetesVersion, err)
		}
		if err := validateHexDigest(entry.sha256, 64); err != nil {
			return fmt.Errorf("Kubernetes %s has invalid source SHA-256: %w", entry.kubernetesVersion, err)
		}
		if len(entry.apis) == 0 {
			return fmt.Errorf("Kubernetes %s has no API provenance examples", entry.kubernetesVersion)
		}
	}
	return nil
}

func writeScopeFile(path string, bundle *builtinopenapi.Bundle) error {
	var source bytes.Buffer
	source.WriteString(`// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Code generated by openapi-bundle. DO NOT EDIT.

package openapi

import "sigs.k8s.io/kustomize/kyaml/yaml"

// precomputedIsNamespaceScoped avoids loading the full built-in schema when
// only the scope of a known Kubernetes resource is needed.
var precomputedIsNamespaceScoped = map[yaml.TypeMeta]bool{ //nolint:gochecknoglobals
`)
	for _, resource := range bundle.Resources {
		if resource.Scope == builtinopenapi.ScopeUnknown {
			continue
		}
		fmt.Fprintf(&source, "\t{APIVersion: %q, Kind: %q}: %t,\n",
			resource.APIVersion, resource.Kind, resource.Scope == builtinopenapi.ScopeNamespaced)
	}
	source.WriteString("}\n")
	formatted, err := format.Source(source.Bytes())
	if err != nil {
		return fmt.Errorf("format generated Go: %w", err)
	}
	return writeFile(path, formatted)
}

func collectResources(swagger *spec.Swagger) ([]builtinopenapi.Resource, error) {
	resources := map[string]builtinopenapi.Resource{}
	for definitionName, definition := range swagger.Definitions {
		extension, found := definition.Extensions[gvkExtension]
		if !found {
			continue
		}
		entries, ok := extension.([]interface{})
		if !ok {
			return nil, fmt.Errorf("definition %q has a malformed %s extension", definitionName, gvkExtension)
		}
		for _, entry := range entries {
			apiVersion, kind, err := parseGVK(entry)
			if err != nil {
				return nil, fmt.Errorf("definition %q: %w", definitionName, err)
			}
			key := resourceKey(apiVersion, kind)
			resource := resources[key]
			if resource.Definition != "" && resource.Definition != definitionName {
				return nil, fmt.Errorf("GVK %s/%s is advertised by definitions %q and %q",
					apiVersion, kind, resource.Definition, definitionName)
			}
			resource.APIVersion = apiVersion
			resource.Kind = kind
			resource.Definition = definitionName
			resources[key] = resource
		}
	}

	if err := collectPathResources(swagger.Paths, resources); err != nil {
		return nil, err
	}

	result := make([]builtinopenapi.Resource, 0, len(resources))
	for _, resource := range resources {
		result = append(result, resource)
	}
	builtinopenapi.SortResources(result)
	return result, nil
}

func collectPathResources(paths *spec.Paths, resources map[string]builtinopenapi.Resource) error {
	if paths == nil {
		return nil
	}
	for path, pathInfo := range paths.Paths {
		if pathInfo.Get == nil {
			continue
		}
		extension, found := pathInfo.Get.Extensions[gvkExtension]
		if !found {
			continue
		}
		apiVersion, kind, err := parseGVK(extension)
		if err != nil {
			return fmt.Errorf("path %q: %w", path, err)
		}
		key := resourceKey(apiVersion, kind)
		resource := resources[key]
		resource.APIVersion = apiVersion
		resource.Kind = kind
		if strings.Contains(path, "namespaces/{namespace}") {
			resource.Scope = builtinopenapi.ScopeNamespaced
		} else if resource.Scope == builtinopenapi.ScopeUnknown {
			resource.Scope = builtinopenapi.ScopeCluster
		}
		resources[key] = resource
	}
	return nil
}

func parseGVK(value interface{}) (string, string, error) {
	entry, ok := value.(map[string]interface{})
	if !ok {
		return "", "", fmt.Errorf("malformed %s extension entry", gvkExtension)
	}
	version, versionOK := entry["version"].(string)
	kind, kindOK := entry["kind"].(string)
	if !versionOK || version == "" || !kindOK || kind == "" {
		return "", "", fmt.Errorf("incomplete %s extension entry", gvkExtension)
	}
	groupValue, hasGroup := entry["group"]
	group, groupOK := groupValue.(string)
	if hasGroup && !groupOK {
		return "", "", fmt.Errorf("malformed %s extension entry: group must be a string", gvkExtension)
	}
	if group != "" {
		return group + "/" + version, kind, nil
	}
	return version, kind, nil
}

func resourceKey(apiVersion, kind string) string {
	return apiVersion + "\x00" + kind
}

func validateDefinitionReferences(definitions spec.Definitions) error {
	b, err := json.Marshal(definitions)
	if err != nil {
		return fmt.Errorf("marshal definitions for reference validation: %w", err)
	}
	var value interface{}
	if err := json.Unmarshal(b, &value); err != nil {
		return fmt.Errorf("unmarshal definitions for reference validation: %w", err)
	}
	return walkReferences(value, definitions)
}

func walkReferences(value interface{}, definitions spec.Definitions) error {
	switch typed := value.(type) {
	case []interface{}:
		for _, item := range typed {
			if err := walkReferences(item, definitions); err != nil {
				return err
			}
		}
	case map[string]interface{}:
		for key, item := range typed {
			if key == "$ref" {
				if err := validateReference(item, definitions); err != nil {
					return err
				}
			}
			if err := walkReferences(item, definitions); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateReference(value interface{}, definitions spec.Definitions) error {
	ref, ok := value.(string)
	// A schema may itself describe an object with a property named "$ref".
	// Such a property has a schema object as its value and is not an OpenAPI
	// reference keyword.
	if !ok {
		return nil
	}
	const prefix = "#/definitions/"
	if !strings.HasPrefix(ref, prefix) {
		return fmt.Errorf("OpenAPI definition contains unsupported reference %q", ref)
	}
	name := strings.TrimPrefix(ref, prefix)
	name = strings.ReplaceAll(strings.ReplaceAll(name, "~1", "/"), "~0", "~")
	if _, found := definitions[name]; !found {
		return fmt.Errorf("OpenAPI definition references missing definition %q", name)
	}
	return nil
}

func writeBundle(path string, bundle *builtinopenapi.Bundle) error {
	jsonBytes, err := json.Marshal(bundle)
	if err != nil {
		return fmt.Errorf("marshal bundle: %w", err)
	}
	return writeGzip(path, jsonBytes)
}

func writeFile(path string, contents []byte) (resultErr error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create output directory %q: %w", dir, err)
	}
	tmp, err := os.CreateTemp(dir, ".openapi-bundle-*")
	if err != nil {
		return fmt.Errorf("create temporary output: %w", err)
	}
	tmpName := tmp.Name()
	tmpClosed := false
	defer func() {
		if !tmpClosed {
			if err := tmp.Close(); err != nil {
				resultErr = errors.Join(resultErr, fmt.Errorf("close temporary output: %w", err))
			}
		}
		if err := os.Remove(tmpName); err != nil && !errors.Is(err, os.ErrNotExist) {
			resultErr = errors.Join(resultErr, fmt.Errorf("remove temporary output: %w", err))
		}
	}()
	if _, err := tmp.Write(contents); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	if err := tmp.Chmod(0o644); err != nil {
		return fmt.Errorf("set output permissions: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temporary output: %w", err)
	}
	tmpClosed = true
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("replace output %q: %w", path, err)
	}
	return nil
}

func writeGzip(path string, contents []byte) (resultErr error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create output directory %q: %w", dir, err)
	}
	tmp, err := os.CreateTemp(dir, ".openapi-bundle-*")
	if err != nil {
		return fmt.Errorf("create temporary output: %w", err)
	}
	tmpName := tmp.Name()
	tmpClosed := false
	defer func() {
		if !tmpClosed {
			if err := tmp.Close(); err != nil {
				resultErr = errors.Join(resultErr, fmt.Errorf("close temporary output: %w", err))
			}
		}
		if err := os.Remove(tmpName); err != nil && !errors.Is(err, os.ErrNotExist) {
			resultErr = errors.Join(resultErr, fmt.Errorf("remove temporary output: %w", err))
		}
	}()

	writer, err := gzip.NewWriterLevel(tmp, gzip.BestCompression)
	if err != nil {
		return fmt.Errorf("create gzip writer: %w", err)
	}
	writerClosed := false
	defer func() {
		if !writerClosed {
			if err := writer.Close(); err != nil {
				resultErr = errors.Join(resultErr, fmt.Errorf("close gzip writer: %w", err))
			}
		}
	}()
	writer.Header.ModTime = time.Time{}
	writer.Header.Name = ""
	writer.Header.Comment = ""
	writer.Header.Extra = nil
	writer.Header.OS = 255

	if _, err := writer.Write(contents); err != nil {
		return fmt.Errorf("write compressed output: %w", err)
	}
	closeWriterErr := writer.Close()
	writerClosed = true
	if closeWriterErr != nil {
		return fmt.Errorf("close gzip writer: %w", closeWriterErr)
	}
	if err := tmp.Chmod(0o644); err != nil {
		return fmt.Errorf("set output permissions: %w", err)
	}
	closeTempErr := tmp.Close()
	tmpClosed = true
	if closeTempErr != nil {
		return fmt.Errorf("close temporary output: %w", closeTempErr)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("replace output %q: %w", path, err)
	}
	return nil
}
