package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var modules = [...]string{
	"kyaml", "api", "kstatus", "cmd/config",
	"cmd/resource", "cmd/kubectl", "pluginator", "kustomize",
}
var verbose bool   // Enable verbose or not
var tempDir string // Temporary directory path for git worktree
var pwd string     // Current working directory

// === Log helper functions ===

func logDebug(format string, v ...interface{}) {
	if verbose {
		log.Printf("DEBUG "+format, v...)
	}
}

func logInfo(format string, v ...interface{}) {
	log.Printf("INFO "+format, v...)
}

func logFatal(format string, v ...interface{}) {
	log.Fatalf("FATAL "+format, v...)
}

// === Command line commands ===

var rootCmd = &cobra.Command{
	Use:   "releasing",
	Short: "This go program is used to improve the modules releasing process in Kustomize repository.",
}

var listSubCmd = &cobra.Command{
	Use:   "list",
	Short: "List current version of all covered modules",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		pwd, err = os.Getwd()
		if err != nil {
			logFatal(err.Error())
		}
		logDebug("Working directory: %s", pwd)
		remote := "upstream"
		// Check remotes
		checkRemoteExistence(pwd, remote)
		// Fetch latest tags from remote
		fetchTags(pwd, remote)
		res := []string{} // Store result strings
		for _, mod := range modules {
			res = append(res, fmt.Sprintf("%s/%s", mod, getModuleCurrentVersion(mod)))
		}
		for _, l := range res {
			fmt.Println(l)
		}
	},
}

var release = &cobra.Command{
	Use:   "release",
	Short: "Release a new version of specified module",
	PreRun: func(cmd *cobra.Command, args []string) {
		logDebug("Preparing Git environemnt")
		prepareGit()
	},
	Run: func(cmd *cobra.Command, args []string) {
		logInfo("Done")
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		logDebug("Cleaning Git environment")
		cleanGit()
	},
}

var subCmds = [...]*cobra.Command{
	listSubCmd,
	release,
}

// === Main function ===

func main() {
	for _, cmd := range subCmds {
		rootCmd.AddCommand(cmd)
	}
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	if err := rootCmd.Execute(); err != nil {
		logFatal(err.Error())
	}
}

func getModuleCurrentVersion(modName string) string {
	mod := newModule(modName, pwd)
	mod.UpdateCurrentVersion()
	v := mod.version.String()
	logDebug("module %s version.toString => %s", mod.name, v)
	return v
}

func checkRemoteExistence(path string, remote string) {
	logDebug("Checking remote %s in %s", remote, path)
	cmd := exec.Command("git", "remote")
	cmd.Dir = path
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		logFatal(err.Error())
	}
	logDebug("Remotes:\n%s", out.String())

	regString := fmt.Sprintf("(?m)^%s$", remote)
	reg := regexp.MustCompile(regString)
	if !reg.MatchString(out.String()) {
		logFatal("Cannot find remote named %s", remote)
	}
	logDebug("Remote %s exists", remote)
}

func fetchTags(path string, remote string) {
	logDebug("Fetching latest tags")
	cmd := exec.Command("git", "fetch", "-t", remote)
	cmd.Dir = path
	err := cmd.Run()
	if err != nil {
		logFatal(err.Error())
	}
	logDebug("Finished fetching")
}

// === module version struct and functions definition ===

type moduleVersion struct {
	major int
	minor int
	patch int
}

func (v moduleVersion) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.major, v.minor, v.patch)
}

func (v *moduleVersion) Set(major int, minor int, patch int) {
	v.major = major
	v.minor = minor
	v.patch = patch
}

func (v *moduleVersion) FromString(vs string) {
	versions := strings.Split(vs, ".")
	major, err := strconv.Atoi(versions[0])
	if err != nil {
		logFatal(err.Error())
	}
	minor, err := strconv.Atoi(versions[1])
	if err != nil {
		logFatal(err.Error())
	}
	patch, err := strconv.Atoi(versions[2])
	if err != nil {
		logFatal(err.Error())
	}
	v.Set(major, minor, patch)
}

// === module struct and functions definition ===

type module struct {
	name    string
	path    string
	version moduleVersion
}

func newModule(modName string, path string) module {
	mod := module{
		name: modName,
		path: path,
	}
	logDebug("Created module struct for %s", modName)
	return mod
}

func (m *module) UpdateCurrentVersion() {
	logDebug("Getting latest tag for %s", m.name)
	cmd := exec.Command("git", "tag", "-l")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Dir = m.path
	err := cmd.Run()
	if err != nil {
		logFatal(err.Error())
	}

	// Search for module tag
	regString := fmt.Sprintf("(?m)^%s/v(\\d+\\.){2}\\d+$", m.name)
	reg := regexp.MustCompile(regString)
	tagsString := reg.FindAllString(out.String(), -1)
	logDebug("Tags for module %s:\n%s", m.name, tagsString)
	var versions []moduleVersion
	for _, tag := range tagsString {
		tag = tag[len(m.name)+2:]
		v := moduleVersion{}
		v.FromString(tag)

		versions = append(versions, v)
	}
	// Sort to find latest tag
	sort.Slice(versions, func(i, j int) bool {
		if versions[i].major == versions[j].major && versions[i].minor == versions[j].minor {
			return versions[i].patch > versions[j].patch
		} else if versions[i].major == versions[j].major {
			return versions[i].minor > versions[j].minor
		} else {
			return versions[i].major > versions[j].major
		}
	})

	m.version = versions[0]
}

// === Git environment functions ===

func prepareGit() {
	var err error
	tempDir, err = ioutil.TempDir("", "kustomize-releases")
	if err != nil {
		logFatal(err.Error())
	}
	logDebug("Created git temp dir: " + tempDir)
}

func cleanGit() {
	logDebug("Deleting git temp dir: " + tempDir)
	err := os.RemoveAll(tempDir)
	if err != nil {
		logFatal(err.Error())
	}
	logDebug("Deleting done")
}
