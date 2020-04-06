package builtins_qlik

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"

	"helm.sh/helm/v3/pkg/downloader"
	"k8s.io/client-go/util/homedir"

	"github.com/pkg/errors"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"

	"sigs.k8s.io/kustomize/api/builtins_qlik/utils"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/yaml"
)

type HelmChartPlugin struct {
	ChartName        string                 `json:"chartName,omitempty" yaml:"chartName,omitempty"`
	ChartHome        string                 `json:"chartHome,omitempty" yaml:"chartHome,omitempty"`
	TmpChartHome     string                 `json:"tmpChartHome,omitempty" yaml:"tmpChartHome,omitempty"`
	ChartVersion     string                 `json:"chartVersion,omitempty" yaml:"chartVersion,omitempty"`
	ChartRepo        string                 `json:"chartRepo,omitempty" yaml:"chartRepo,omitempty"`
	ChartRepoName    string                 `json:"chartRepoName,omitempty" yaml:"chartRepoName,omitempty"`
	ValuesFrom       string                 `json:"valuesFrom,omitempty" yaml:"valuesFrom,omitempty"`
	Values           map[string]interface{} `json:"values,omitempty" yaml:"values,omitempty"`
	HelmHome         string                 `json:"helmHome,omitempty" yaml:"helmHome,omitempty"`
	ReleaseName      string                 `json:"releaseName,omitempty" yaml:"releaseName,omitempty"`
	ReleaseNamespace string                 `json:"releaseNamespace,omitempty" yaml:"releaseNamespace,omitempty"`
	ExtraArgs        string                 `json:"extraArgs,omitempty" yaml:"extraArgs,omitempty"`
	SubChart         string                 `json:"subChart,omitempty" yaml:"subChart,omitempty"`
	rf               *resmap.Factory
	logger           *log.Logger
	hash             string
	hashFolder       string
}

func (p *HelmChartPlugin) Config(h *resmap.PluginHelpers, c []byte) (err error) {
	p.rf = h.ResmapFactory()

	if err = yaml.Unmarshal(c, p); err != nil {
		p.logger.Printf("error unmarshalling yaml config: %v, error: %v\n", string(c), err)
		return err
	}

	if p.HelmHome == "" {
		directory := path.Join(os.TempDir(), "dotHelm")
		p.HelmHome = directory
	}

	chartVersion := "latest"
	if p.ChartVersion != "" {
		chartVersion = p.ChartVersion
	}

	if p.ChartHome == "" && p.TmpChartHome != "" {
		p.ChartHome = path.Join(os.TempDir(), p.TmpChartHome, fmt.Sprintf("%v-%v", p.ChartName, chartVersion))
	}

	if p.ChartHome == "" {
		directory := path.Join(os.TempDir(), fmt.Sprintf("%v-%v", p.ChartName, chartVersion))
		p.ChartHome = directory
	}

	if p.ChartRepo == "" {
		p.ChartRepo = "https://qlik.bintray.com/edge"
	}

	if p.ChartRepoName == "" {
		p.ChartRepoName = "qlik"
	}

	if p.ReleaseName == "" {
		p.ReleaseName = "release-name"
	}

	if p.ReleaseNamespace == "" {
		p.ReleaseNamespace = "default"
	}

	p.hashFolder = filepath.Join(p.ChartHome, ".plugincache")
	if err = os.MkdirAll(p.hashFolder, os.ModePerm); err != nil {
		p.logger.Printf("error creating hashfolder: %v, error: %v\n", p.hashFolder, err)
		return err
	}

	chartHash := sha256.New()
	chartHash.Write(c)
	p.hash = hex.EncodeToString(chartHash.Sum(nil))

	return nil
}

func (p *HelmChartPlugin) Generate() (resmap.ResMap, error) {

	var templatedYaml []byte

	hashFilePath := filepath.Join(p.hashFolder, p.hash)
	if _, err := os.Stat(hashFilePath); err != nil {
		if os.IsNotExist(err) {
			if templatedYaml, err = p.executeHelmTemplate(); err != nil {
				p.logger.Printf("error executing helm template, error: %v\n", err)
				return nil, err
			} else if err = ioutil.WriteFile(hashFilePath, templatedYaml, 0644); err != nil {
				p.logger.Printf("error writing kustomization yaml to file: %v, error: %v\n", hashFilePath, err)
			}
		} else {
			return nil, err
		}
	} else {
		templatedYaml, err = ioutil.ReadFile(hashFilePath)
		if err != nil {
			p.logger.Printf("error reading file: %v, error: %v\n", hashFilePath, err)
			return nil, err
		}
	}

	return p.rf.NewResMapFromBytes(templatedYaml)
}

func (p *HelmChartPlugin) executeHelmTemplate() ([]byte, error) {
	os.Setenv("HELM_NAMESPACE", p.ReleaseNamespace)
	os.Setenv("XDG_CONFIG_HOME", p.HelmHome)
	os.Setenv("XDG_CACHE_HOME", p.HelmHome)
	settings := cli.New()

	if err := p.helmFetchIfRequired(settings, p.ChartRepoName); err != nil {
		p.logger.Printf("error checking/fetching chart, err: %v\n", err)
		return nil, err
	}

	var chartPath string
	var chartName string

	if p.SubChart != "" {
		chartName = p.SubChart
		chartPath = filepath.Join(p.ChartHome, p.ChartName, "charts", p.SubChart)
	} else {
		chartName = p.ChartName
		chartPath = filepath.Join(p.ChartHome, p.ChartName)
	}

	if err := p.helmDependenciesBuild(settings, chartPath); err != nil {
		p.logger.Printf("error building dependencies, err: %v\n", err)
		return nil, err
	}

	resources, err := p.helmTemplate(settings, chartPath, p.ReleaseName, p.Values)
	if err != nil {
		p.logger.Printf("error executing helm template for chart: %v at path: %v, err: %v\n", chartName, chartPath, err)
		return nil, err
	}

	return resources, nil
}

func (p *HelmChartPlugin) directoryExists(path string) (exists bool, err error) {
	if info, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		exists = false
		err = nil
	} else if err != nil && !os.IsNotExist(err) {
		exists = false
	} else if err == nil && info.IsDir() {
		exists = true
	} else if err == nil && !info.IsDir() {
		exists = false
		err = fmt.Errorf("path: %v is occupied by a file instead of a directory\n", path)
	}
	return exists, err
}

func (p *HelmChartPlugin) helmFetchIfRequired(settings *cli.EnvSettings, repoName string) error {
	chartDir := filepath.Join(p.ChartHome, p.ChartName)
	if exists, err := p.directoryExists(chartDir); err != nil {
		p.logger.Printf("error checking if chart was already fetched to path: %v, err: %v\n", chartDir, err)
		return err
	} else if !exists {
		if err := p.helmRepoAdd(settings, &repo.Entry{Name: repoName, URL: p.ChartRepo}); err != nil {
			p.logger.Printf("error adding repo: %v, err: %v\n", p.ChartRepo, err)
			return err
		}
		if err := p.helmReposUpdate(settings); err != nil {
			p.logger.Printf("error updating helm repos, err: %v\n", err)
			return err
		}
		if err := p.helmFetch(settings, fmt.Sprintf("%v/%v", repoName, p.ChartName), p.ChartVersion, p.ChartHome); err != nil {
			p.logger.Printf("error fetching chart, err: %v\n", err)
			return err
		}
	} else {
		p.logger.Printf("nothing to do, chart is already at path: %v\n", chartDir)
	}
	return nil
}

func (p *HelmChartPlugin) helmRepoAdd(settings *cli.EnvSettings, repoEntry *repo.Entry) error {
	repoFilePath := settings.RepositoryConfig

	b, err := ioutil.ReadFile(repoFilePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var repoFile repo.File
	if err := yaml.Unmarshal(b, &repoFile); err != nil {
		return err
	}

	if existingEntry := repoFile.Get(repoEntry.Name); existingEntry != nil {
		if existingEntry.URL == repoEntry.URL {
			return nil
		} else {
			return fmt.Errorf("cannot add helm repo name: %v, URL: %v, "+
				"because an entry with the same name but different URL already exists", repoEntry.Name, repoEntry.URL)
		}
	}

	chartRepository, err := repo.NewChartRepository(repoEntry, getter.All(settings))
	if err != nil {
		return err
	}

	if _, err := chartRepository.DownloadIndexFile(); err != nil {
		return errors.Wrapf(err, "looks like %q is not a valid chart repository or cannot be reached", repoEntry.URL)
	}

	repoFile.Update(repoEntry)

	if err := repoFile.WriteFile(repoFilePath, 0644); err != nil {
		return err
	}
	p.logger.Printf("%q has been added to your repositories\n", repoEntry.Name)
	return nil
}

func (p *HelmChartPlugin) helmReposUpdate(settings *cli.EnvSettings) error {
	var (
		repoFilePath = settings.RepositoryConfig
		err          error
		repoFile     *repo.File
		repos        []*repo.ChartRepository
		cfg          *repo.Entry
		wg           sync.WaitGroup
	)

	repoFile, err = repo.LoadFile(repoFilePath)
	if os.IsNotExist(errors.Cause(err)) || len(repoFile.Repositories) == 0 {
		return errors.New("no repositories found. You must add one before updating")
	}

	for _, cfg = range repoFile.Repositories {
		r, err := repo.NewChartRepository(cfg, getter.All(settings))
		if err != nil {
			return err
		}
		repos = append(repos, r)
	}

	// fmt.Printf("Downloading helm chart index ...\n")
	for _, r := range repos {
		wg.Add(1)
		go func(re *repo.ChartRepository) {
			defer wg.Done()
			if _, err := re.DownloadIndexFile(); err != nil {
				p.logger.Printf("...Unable to get an update from the %q chart repository (%s):\n\t%s\n", re.Config.Name, re.Config.URL, err)
			} else {
				p.logger.Printf("...Successfully got an update from the %q chart repository\n", re.Config.Name)
			}
		}(r)
	}
	wg.Wait()
	return nil
}

func (p *HelmChartPlugin) helmFetch(settings *cli.EnvSettings, chartRef, version, chartUntarDirPath string) error {
	client := action.NewPull()
	client.Untar = true
	client.UntarDir = chartUntarDirPath
	client.Settings = settings
	client.Version = version
	_, err := client.Run(chartRef)
	return err
}

func (p *HelmChartPlugin) helmTemplate(settings *cli.EnvSettings, chartPath, releaseName string, vals map[string]interface{}) ([]byte, error) {
	validate := false
	var extraAPIs []string

	actionConfig := new(action.Configuration)

	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), func(format string, v ...interface{}) {}); err != nil {
		return nil, err
	}

	client := action.NewInstall(actionConfig)
	client.DryRun = true
	client.ReleaseName = releaseName
	client.Replace = true // Skip the name check
	client.ClientOnly = !validate
	client.APIVersions = chartutil.VersionSet(extraAPIs)

	rel, err := p.runInstall(settings, chartPath, client, vals)
	if err != nil {
		return nil, err
	}

	var manifests bytes.Buffer

	fmt.Fprintln(&manifests, strings.TrimSpace(rel.Manifest))
	if !client.DisableHooks {
		for _, m := range rel.Hooks {
			fmt.Fprintf(&manifests, "---\n# Source: %s\n%s\n", m.Path, m.Manifest)
		}
	}

	return manifests.Bytes(), nil
}

func (p *HelmChartPlugin) runInstall(settings *cli.EnvSettings, chartPath string, client *action.Install, vals map[string]interface{}) (*release.Release, error) {
	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return nil, err
	}

	validInstallableChart, err := isChartInstallable(chartRequested)
	if !validInstallableChart {
		return nil, err
	}

	if req := chartRequested.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(chartRequested, req); err != nil {
			return nil, err
		}
	}

	client.Namespace = settings.Namespace()
	return client.Run(chartRequested, vals)
}

func isChartInstallable(ch *chart.Chart) (bool, error) {
	switch ch.Metadata.Type {
	case "", "application":
		return true, nil
	}
	return false, fmt.Errorf("%s charts are not installable", ch.Metadata.Type)
}

// defaultKeyring returns the expanded path to the default keyring.
func defaultKeyring() string {
	if v, ok := os.LookupEnv("GNUPGHOME"); ok {
		return filepath.Join(v, "pubring.gpg")
	}
	return filepath.Join(homedir.HomeDir(), ".gnupg", "pubring.gpg")
}

func (p *HelmChartPlugin) helmDependenciesBuild(settings *cli.EnvSettings, chartPath string) error {
	c, err := loader.Load(chartPath)
	if err != nil {
		return err
	}

	if req := c.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(c, req); err != nil {

			if dependencyRepoEntries, err := resolveDependencyRepos(settings, c); err != nil {
				p.logger.Printf("error resolving dependency repos: %v\n", err)
				return err
			} else {
				for _, repoEntry := range dependencyRepoEntries {
					if err := p.helmRepoAdd(settings, repoEntry); err != nil {
						p.logger.Printf("error adding dependency repo: %v to the repo file: %v\n", repoEntry.Name, err)
						return err
					}
				}
			}

			man := &downloader.Manager{
				Out:              p.logger.Writer(),
				ChartPath:        chartPath,
				Keyring:          defaultKeyring(),
				Getters:          getter.All(settings),
				RepositoryConfig: settings.RepositoryConfig,
				RepositoryCache:  settings.RepositoryCache,
				Debug:            settings.Debug,
				SkipUpdate:       true,
			}
			return man.Build()
		}
	}

	return nil
}

func resolveDependencyRepos(settings *cli.EnvSettings, c *chart.Chart) ([]*repo.Entry, error) {
	newAliasedRepoEntries := make([]*repo.Entry, 0)
	pureUrlDependencies := make([]*chart.Dependency, 0)

	//only consider named and unnamed repo entries with HTTP URLs:
	for _, dep := range c.Metadata.Dependencies {
		isAliasedRepoDependency := false
		for _, aliasMarker := range []string{"@", "alias:"} {
			if strings.HasPrefix(dep.Repository, aliasMarker) {
				if repoEntry, err := getRepoEntryForAliasedDependency(aliasMarker, dep, c); err != nil {
					return nil, err
				} else {
					newAliasedRepoEntries = append(newAliasedRepoEntries, repoEntry)
					isAliasedRepoDependency = true
					break
				}
			}
		}
		if isAliasedRepoDependency {
			continue
		}

		//need to process pure URL dependencies after all aliased(named) dependencies,
		//in case there are overlaps:
		if u := getAbsoluteHttpUrlObject(dep.Repository); u != nil {
			pureUrlDependencies = append(pureUrlDependencies, dep)
		}
	}

	repoFileEntries, err := getRepoFileEntries(settings)
	if err != nil {
		return nil, err
	}

	for _, dep := range pureUrlDependencies {
		if repoEntriesContainHttpUrl(newAliasedRepoEntries, dep.Repository) {
			break
		} else if repoEntriesContainHttpUrl(repoFileEntries, dep.Repository) {
			break
		}
		newAliasedRepoEntries = append(newAliasedRepoEntries, &repo.Entry{
			Name: fmt.Sprintf("auto-%v", uuid.New().String()),
			URL:  dep.Repository,
		})
	}
	return newAliasedRepoEntries, nil
}

func getRepoFileEntries(settings *cli.EnvSettings) ([]*repo.Entry, error) {
	repoFile, err := repo.LoadFile(settings.RepositoryConfig)
	if err != nil {
		if os.IsNotExist(errors.Cause(err)) {
			return make([]*repo.Entry, 0), nil
		} else {
			return nil, err
		}
	} else {
		return repoFile.Repositories, nil
	}
}

func getRepoEntryForAliasedDependency(aliasMarker string, dep *chart.Dependency, c *chart.Chart) (*repo.Entry, error) {
	repoEntryName := strings.TrimPrefix(dep.Repository, aliasMarker)
	repoEntryUrl := getRepoUrlFromChartLock(c.Lock, dep.Name)
	if repoEntryUrl == "" {
		return nil, fmt.Errorf("cannot find URL for dependency: %v, alias: %v", dep.Name, repoEntryName)
	} else {
		return &repo.Entry{
			Name:     repoEntryName,
			URL:      repoEntryUrl,
			Username: os.Getenv(fmt.Sprintf("%v-helm-repo-username", repoEntryName)),
			Password: os.Getenv(fmt.Sprintf("%v-helm-repo-password", repoEntryName)),
		}, nil
	}
}

func getRepoUrlFromChartLock(chartLock *chart.Lock, name string) string {
	for _, lockDep := range chartLock.Dependencies {
		if lockDep.Name == name {
			return lockDep.Repository
		}
	}
	return ""
}

func repoEntriesContainHttpUrl(repoEntries []*repo.Entry, testUrl string) bool {
	for _, entry := range repoEntries {
		if entry.URL == testUrl {
			return true
		} else {
			if entryUrlObject := getAbsoluteHttpUrlObject(entry.URL); entryUrlObject == nil {
				break
			} else if testUrlObject := getAbsoluteHttpUrlObject(testUrl); testUrlObject == nil {
				break
			} else {
				for _, u := range []*url.URL{entryUrlObject, testUrlObject} {
					if u.Path == "" {
						u.Path = "/"
					}
					u.Path = filepath.Clean(u.Path)
				}
				if entryUrlObject.String() == testUrlObject.String() {
					return true
				}
			}
		}
	}
	return false
}

func getAbsoluteHttpUrlObject(test string) *url.URL {
	if u, err := url.Parse(test); err != nil {
		return nil
	} else if !isAbsoluteHttpUrlObject(u) {
		return nil
	} else {
		return u
	}
}

func isAbsoluteHttpUrlObject(u *url.URL) bool {
	if !u.IsAbs() {
		return false
	} else if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return true
}

func NewHelmChartPlugin() resmap.GeneratorPlugin {
	return &HelmChartPlugin{logger: utils.GetLogger("HelmChartPlugin")}
}
