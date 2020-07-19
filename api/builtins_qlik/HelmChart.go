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
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/kustomize/api/builtins_qlik/utils"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/yaml"
)

const (
	defaultRepoIndexFileStaleAfterSeconds = 60
)

var helmRunMutex sync.Mutex

type HelmChartPlugin struct {
	ChartName                      string                 `json:"chartName,omitempty" yaml:"chartName,omitempty"`
	ChartHome                      string                 `json:"chartHome,omitempty" yaml:"chartHome,omitempty"`
	TmpChartHome                   string                 `json:"tmpChartHome,omitempty" yaml:"tmpChartHome,omitempty"`
	ChartVersion                   string                 `json:"chartVersion,omitempty" yaml:"chartVersion,omitempty"`
	ChartRepo                      string                 `json:"chartRepo,omitempty" yaml:"chartRepo,omitempty"`
	ChartRepoName                  string                 `json:"chartRepoName,omitempty" yaml:"chartRepoName,omitempty"`
	ValuesFrom                     string                 `json:"valuesFrom,omitempty" yaml:"valuesFrom,omitempty"`
	Values                         map[string]interface{} `json:"values,omitempty" yaml:"values,omitempty"`
	HelmHome                       string                 `json:"helmHome,omitempty" yaml:"helmHome,omitempty"`
	ReleaseName                    string                 `json:"releaseName,omitempty" yaml:"releaseName,omitempty"`
	ReleaseNamespace               string                 `json:"releaseNamespace,omitempty" yaml:"releaseNamespace,omitempty"`
	ExtraArgs                      string                 `json:"extraArgs,omitempty" yaml:"extraArgs,omitempty"`
	SubChart                       string                 `json:"subChart,omitempty" yaml:"subChart,omitempty"`
	NewChartVersion                string                 `json:"newChartVersion,omitempty" yaml:"newChartVersion,omitempty"`
	RepoIndexFileStaleAfterSeconds int                    `json:"repoIndexFileStaleAfterSeconds,omitempty" yaml:"repoIndexFileStaleAfterSeconds,omitempty"`
	LockRetryDelayMinMilliSeconds  int                    `json:"lockRetryDelayMinMilliSeconds,omitempty" yaml:"lockRetryDelayMinMilliSeconds,omitempty"`
	LockRetryDelayMaxMilliSeconds  int                    `json:"lockRetryDelayMaxMilliSeconds,omitempty" yaml:"lockRetryDelayMaxMilliSeconds,omitempty"`
	LockTimeoutSeconds             int                    `json:"lockTimeoutSeconds,omitempty" yaml:"lockTimeoutSeconds,omitempty"`
	rf                             *resmap.Factory
	logger                         *log.Logger
	hash                           string
	hashFolder                     string
	cacheEnabled                   bool
}

func (p *HelmChartPlugin) Config(h *resmap.PluginHelpers, c []byte) (err error) {
	p.rf = h.ResmapFactory()

	p.cacheEnabled, _ = strconv.ParseBool(os.Getenv("KUST_HELM_CACHE"))

	if err = yaml.Unmarshal(c, p); err != nil {
		p.logger.Printf("error unmarshalling yaml config: %v, error: %v\n", string(c), err)
		return err
	}

	if p.HelmHome == "" {
		directory := filepath.Join(os.TempDir(), "dotHelm")
		p.HelmHome = directory
	}

	chartVersion := "latest"
	if p.ChartVersion != "" {
		chartVersion = p.ChartVersion
	}

	generatedChartHome := false
	if p.ChartHome == "" && p.TmpChartHome != "" {
		p.ChartHome = filepath.Join(os.TempDir(), p.TmpChartHome, fmt.Sprintf("%v-%v", p.ChartName, chartVersion))
		generatedChartHome = true
	}
	if p.ChartHome == "" {
		p.ChartHome = filepath.Join(os.TempDir(), fmt.Sprintf("%v-%v", p.ChartName, chartVersion))
		generatedChartHome = true
	}
	if generatedChartHome {
		if err = os.MkdirAll(p.ChartHome, os.ModePerm); err != nil {
			p.logger.Printf("error creating chartHome directory: %v, error: %v\n", p.ChartHome, err)
			return err
		}
	}

	if p.ChartRepo == "" {
		p.logger.Println("No chartRepo set in the config. If fetch is needed, it will fail.")
	}

	if p.ReleaseName == "" {
		p.ReleaseName = "release-name"
	}

	if p.ReleaseNamespace == "" {
		p.ReleaseNamespace = "default"
	}

	if p.cacheEnabled {
		p.hashFolder = filepath.Join(p.ChartHome, ".plugincache")
		if err = os.MkdirAll(p.hashFolder, os.ModePerm); err != nil {
			p.logger.Printf("error creating hashfolder: %v, error: %v\n", p.hashFolder, err)
			return err
		}
		chartHash := sha256.New()
		chartHash.Write(c)
		p.hash = hex.EncodeToString(chartHash.Sum(nil))
	}
	if p.RepoIndexFileStaleAfterSeconds == 0 {
		p.RepoIndexFileStaleAfterSeconds = defaultRepoIndexFileStaleAfterSeconds
	}
	if p.LockRetryDelayMinMilliSeconds == 0 {
		p.LockRetryDelayMinMilliSeconds = utils.DefaultLockRetryDelayMinMilliSeconds
	}
	if p.LockRetryDelayMaxMilliSeconds == 0 {
		p.LockRetryDelayMaxMilliSeconds = utils.DefaultLockRetryDelayMaxMilliSeconds
	}
	if p.LockTimeoutSeconds == 0 {
		p.LockTimeoutSeconds = utils.DefaultLockTimeoutSeconds
	}

	return nil
}

func (p *HelmChartPlugin) Generate() (resmap.ResMap, error) {

	var templatedYaml []byte
	var hashFilePath = ""

	if p.cacheEnabled {
		hashFilePath = filepath.Join(p.hashFolder, p.hash)
		lockFilePath := filepath.Join(p.hashFolder, fmt.Sprintf("%v.flock", p.hash))

		if unlockFn, err := utils.LockPath(lockFilePath, p.LockTimeoutSeconds, p.LockRetryDelayMinMilliSeconds, p.LockRetryDelayMaxMilliSeconds, p.logger); err != nil {
			p.logger.Printf("error locking %v, error: %v\n", hashFilePath, err)
			return nil, err
		} else {
			defer unlockFn()
		}
	}
	if _, err := os.Stat(hashFilePath); err != nil {
		if os.IsNotExist(err) {
			if templatedYaml, err = p.executeHelmTemplate(); err != nil {
				p.logger.Printf("error executing helm template, error: %v\n", err)
				return nil, err
			} else if p.cacheEnabled {
				if err = ioutil.WriteFile(hashFilePath, templatedYaml, 0644); err != nil {
					p.logger.Printf("error writing kustomization yaml to file: %v, error: %v\n", hashFilePath, err)
				}
			}
		} else {
			return nil, err
		}
	} else {
		p.logger.Println("reading previously generated yaml from plugin cache...")
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

	if err := p.helmFetchIfRequired(settings); err != nil {
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

	c, err := p.loadChartWithDependencies(settings, chartPath)
	if err != nil {
		p.logger.Printf("error building dependencies, err: %v\n", err)
		return nil, err
	}

	resources, err := p.helmTemplate(settings, c, p.ReleaseName, p.Values)
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

func (p *HelmChartPlugin) helmFetchIfRequired(settings *cli.EnvSettings) error {
	lockFilePath := filepath.Join(p.ChartHome, fmt.Sprintf("%v.flock", p.ChartName))
	if unlockFn, err := utils.LockPath(lockFilePath, p.LockTimeoutSeconds, p.LockRetryDelayMinMilliSeconds, p.LockRetryDelayMaxMilliSeconds, p.logger); err != nil {
		p.logger.Printf("error locking chart directory: %v in chartHome: %v, error: %v\n", p.ChartName, p.ChartHome, err)
		return err
	} else {
		defer unlockFn()
	}

	chartDir := filepath.Join(p.ChartHome, p.ChartName)
	if exists, err := p.directoryExists(chartDir); err != nil {
		p.logger.Printf("error checking if chart was already fetched to path: %v, err: %v\n", chartDir, err)
		return err
	} else if !exists {
		if repoName, err := p.helmConfigForChart(settings, p.ChartRepoName); err != nil {
			return err
		} else if err := p.helmFetch(settings, fmt.Sprintf("%v/%v", repoName, p.ChartName), p.ChartVersion, p.ChartHome); err != nil {
			p.logger.Printf("error fetching chart, err: %v\n", err)
			return err
		}
	} else {
		p.logger.Printf("no need to fetch, chart is already at path: %v\n", chartDir)
	}
	return nil
}

func (p *HelmChartPlugin) helmConfigForChart(settings *cli.EnvSettings, repoName string) (string, error) {
	helmConfigHomeAndCacheDir := filepath.Dir(settings.RepositoryConfig)
	lockFilePath := filepath.Join(helmConfigHomeAndCacheDir, "helm-repo-config.flock")
	if unlockFn, err := utils.LockPath(lockFilePath, p.LockTimeoutSeconds, p.LockRetryDelayMinMilliSeconds, p.LockRetryDelayMaxMilliSeconds, p.logger); err != nil {
		p.logger.Printf("error locking helm config home and cache directory: %v, error: %v\n", helmConfigHomeAndCacheDir, err)
		return "", err
	} else {
		defer unlockFn()
	}

	if repoName == "" {
		repoFileEntries, err := getRepoFileEntries(settings)
		if err != nil {
			return "", err
		} else if repoEntry := repoEntriesContainHttpUrl(repoFileEntries, p.ChartRepo); repoEntry != nil {
			repoName = repoEntry.Name
		} else {
			repoName = getAutoRepoName("auto")
		}
	}

	if err := p.helmRepoAdd(settings, &repo.Entry{
		Name:     repoName,
		URL:      p.ChartRepo,
		Username: os.Getenv(strings.ToUpper(fmt.Sprintf("%v_helm_repo_username", repoName))),
		Password: os.Getenv(strings.ToUpper(fmt.Sprintf("%v_helm_repo_password", repoName)))}); err != nil {
		p.logger.Printf("error adding repo: %v, err: %v\n", p.ChartRepo, err)
		return "", err
	}
	if err := p.helmReposUpdate(settings); err != nil {
		p.logger.Printf("error updating helm repos, err: %v\n", err)
		return "", err
	}
	return repoName, nil
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
		if entry := repoEntriesContainHttpUrl([]*repo.Entry{existingEntry}, repoEntry.URL); entry != nil {
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

	now := time.Now()
	for _, cfg = range repoFile.Repositories {
		r, err := repo.NewChartRepository(cfg, getter.All(settings))
		if err != nil {
			return err
		}

		//only re-download index file if it's older than some threshold:
		indexFilePath := filepath.Join(r.CachePath, helmpath.CacheIndexFile(r.Config.Name))
		if fileInfo, err := os.Stat(indexFilePath); err != nil {
			return err
		} else if timerSinceUpdate := now.Sub(fileInfo.ModTime()); timerSinceUpdate > time.Duration(p.RepoIndexFileStaleAfterSeconds)*time.Second {
			p.logger.Printf("Will be updating repository: %v\n", r.Config.Name)
			repos = append(repos, r)
		} else {
			p.logger.Printf("Will NOT be updating repository: %v, because it was updated: %v ago\n", r.Config.Name, timerSinceUpdate)
		}
	}

	if len(repos) > 0 {
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
	}

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

func (p *HelmChartPlugin) helmTemplate(settings *cli.EnvSettings, c *chart.Chart, releaseName string, vals map[string]interface{}) ([]byte, error) {
	validInstallableChart, err := isChartInstallable(c)
	if !validInstallableChart {
		return nil, err
	}

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), func(format string, v ...interface{}) {}); err != nil {
		return nil, err
	}

	client := action.NewInstall(actionConfig)
	client.DryRun = true
	client.ReleaseName = releaseName
	client.Replace = true // Skip the name check
	client.ClientOnly = true
	client.IncludeCRDs = true
	
	client.Namespace = settings.Namespace()

	helmRunMutex.Lock()
	rel, err := client.Run(c, vals)
	helmRunMutex.Unlock()
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

func (p *HelmChartPlugin) loadChartWithDependencies(settings *cli.EnvSettings, chartPath string) (*chart.Chart, error) {
	lockFilePath := fmt.Sprintf("%v.flock", chartPath)
	if unlockFn, err := utils.LockPath(lockFilePath, p.LockTimeoutSeconds, p.LockRetryDelayMinMilliSeconds, p.LockRetryDelayMaxMilliSeconds, p.logger); err != nil {
		p.logger.Printf("error locking chart directory: %v, error: %v\n", chartPath, err)
		return nil, err
	} else {
		defer unlockFn()
	}

	c, err := loader.Load(chartPath)
	if err != nil {
		return nil, err
	}

	if len(p.NewChartVersion) > 0 {
		c.Metadata.Version = p.NewChartVersion
	}

	buildingDependencies := false
	if req := c.Metadata.Dependencies; req != nil {
		if ok, err := p.checkDependencies(c, req); err != nil {
			p.logger.Printf("dependency check returned failed: %v\n", err)
			return nil, err
		} else if !ok {
			buildingDependencies = true

			if err := p.helmConfigForDependencies(settings, c); err != nil {
				return nil, err
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
			if err := man.Build(); err != nil {
				return nil, err
			}
		}
	}

	//re-load the chart if we had to build the dependencies:
	if buildingDependencies {
		c, err = loader.Load(chartPath)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (p *HelmChartPlugin) checkDependencies(ch *chart.Chart, reqs []*chart.Dependency) (bool, error) {
	if err := action.CheckDependencies(ch, reqs); err != nil {
		p.logger.Printf("helm dependency presence checker returned a non-fatal error: %v\n", err)
		return false, nil
	} else {
		dependenciesOnDisk := make(map[string]*chart.Chart)
		for _, dependency := range ch.Dependencies() {
			dependenciesOnDisk[dependency.Name()] = dependency
		}
		for _, r := range reqs {
			dependencyOnDisk := dependenciesOnDisk[r.Name]
			if dependencyConstraint, err := semver.NewConstraint(r.Version); err != nil {
				return false, fmt.Errorf("dependency: %v for chart: %v has an invalid version/constraint format: %w", r.Name, ch.Name(), err)
			} else if versionOfDependencyOnDisk, err := semver.NewVersion(dependencyOnDisk.Metadata.Version); err != nil {
				return false, fmt.Errorf("dependency on disk: %v for chart: %v has an invalid version format: %w", r.Name, ch.Name(), err)
			} else if !dependencyConstraint.Check(versionOfDependencyOnDisk) {
				return false, nil
			}
		}
		return true, nil
	}
}

func (p *HelmChartPlugin) helmConfigForDependencies(settings *cli.EnvSettings, c *chart.Chart) error {
	helmConfigHomeAndCacheDir := filepath.Dir(settings.RepositoryConfig)
	lockFilePath := filepath.Join(helmConfigHomeAndCacheDir, "helm-repo-config.flock")
	if unlockFn, err := utils.LockPath(lockFilePath, p.LockTimeoutSeconds, p.LockRetryDelayMinMilliSeconds, p.LockRetryDelayMaxMilliSeconds, p.logger); err != nil {
		p.logger.Printf("error locking helm config home and cache directory: %v, error: %v\n", helmConfigHomeAndCacheDir, err)
		return err
	} else {
		defer unlockFn()
	}

	if dependencyRepoEntries, err := resolveDependencyRepos(settings, c); err != nil {
		p.logger.Printf("error resolving dependency repos: %v\n", err)
		return err
	} else {
		for _, repoEntry := range dependencyRepoEntries {
			if err := p.helmRepoAdd(settings, repoEntry); err != nil {
				p.logger.Printf("error adding dependency repo: %v to the repo file: %v\n", repoEntry.Name, err)
				return err
			}
			if err := p.helmReposUpdate(settings); err != nil {
				p.logger.Printf("error updating helm repos while processing dependencies, err: %v\n", err)
				return err
			}
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
		if repoEntriesContainHttpUrl(newAliasedRepoEntries, dep.Repository) != nil {
			break
		} else if repoEntriesContainHttpUrl(repoFileEntries, dep.Repository) != nil {
			break
		}
		newAliasedRepoEntries = append(newAliasedRepoEntries, &repo.Entry{
			Name: getAutoRepoName("auto"),
			URL:  dep.Repository,
		})
	}
	return newAliasedRepoEntries, nil
}

func getAutoRepoName(prefix string) string {
	return fmt.Sprintf("%v-%v", prefix, uuid.New().String())
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
			Username: os.Getenv(strings.ToUpper(fmt.Sprintf("%v_helm_repo_username", repoEntryName))),
			Password: os.Getenv(strings.ToUpper(fmt.Sprintf("%v_helm_repo_password", repoEntryName))),
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

func repoEntriesContainHttpUrl(repoEntries []*repo.Entry, testUrl string) *repo.Entry {
	for _, entry := range repoEntries {
		if entry.URL == testUrl {
			return entry
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
					return entry
				}
			}
		}
	}
	return nil
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
