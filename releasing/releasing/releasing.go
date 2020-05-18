package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
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
var noDryRun bool  // Disable dry run
var noTest bool    // Disable module tests
var tempDir string // Temporary directory path for git worktree

// === Log helper functions ===

func logDebug(format string, v ...interface{}) {
	if verbose {
		log.Printf("DEBUG "+format, v...)
	}
}

func logInfo(format string, v ...interface{}) {
	log.Printf("INFO "+format, v...)
}

func logWarn(format string, v ...interface{}) {
	log.Printf("WARN "+format, v...)
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
		pwd, err := os.Getwd()
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
			res = append(res, fmt.Sprintf("%s/%s", mod, getModuleCurrentVersion(mod, pwd)))
		}
		for _, l := range res {
			fmt.Println(l)
		}
	},
}

var release = &cobra.Command{
	Use:   "release [module name] [version type]",
	Short: "Release a new version of specified module",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New("2 arguments are required")
		}
		found := false
		for _, mod := range modules {
			if mod == args[0] {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%s is not a valid module. Valid modules are %s", args[0], modules)
		}
		types := []string{"major", "minor", "patch"}
		found = false
		for _, t := range types {
			if t == args[1] {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%s is not a valid version type. Valid types are %s", args[1], types)
		}
		return nil
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		logDebug("Preparing Git environemnt")
		prepareGit()
	},
	Run: func(cmd *cobra.Command, args []string) {
		modName := args[0]
		versionType := args[1]
		logInfo("Creating tag for module %s", modName)
		pwd, err := os.Getwd()
		if err != nil {
			logFatal(err.Error())
		}
		logDebug("Working directory: %s", pwd)
		remote := "upstream"
		// Check remotes
		checkRemoteExistence(pwd, remote)
		// Fetch latest tags from remote
		fetchTags(pwd, remote)

		mod := module{
			name: modName,
			path: pwd,
		}
		mod.UpdateCurrentVersion()

		oldVersion := mod.version.String()
		mod.version.Bump(versionType)
		newVersion := mod.version.String()
		logInfo("Bumping version: %s => %s", oldVersion, newVersion)

		// Create branch
		branch := fmt.Sprintf("release-%s-v%d.%d", mod.name, mod.version.major, mod.version.minor)
		newBranch(pwd, branch)

		addWorktree(pwd, tempDir, branch)

		merge(tempDir, "upstream/master")
		// Update module path
		mod.path = tempDir

		logInfo(
			"Releasing summary:\nDir:\t%s\nModule:\t%s %s\nBranch:\t%s\nTag:\t%s",
			tempDir,
			mod.name,
			mod.version.String(),
			branch,
			mod.Tag(),
		)

		// Run module tests
		output, err := mod.RunTest()
		if err != nil {
			logWarn(output)
		} else if !noDryRun {
			logInfo("Skipping push module %s. Run with --no-dry-run to push the release.", mod.name)
		} else {
			pushRelease(tempDir, branch, mod)
		}
		// Clean
		cleanGit()
		pruneWorktree(pwd)
		deleteBranch(pwd, branch)
		logInfo("Module %s completes", mod.name)
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
	release.Flags().BoolVarP(&noDryRun, "no-dry-run", "", false, "disable dry-run")
	release.Flags().BoolVarP(&noTest, "no-test", "", false, "don't run module tests")

	if err := rootCmd.Execute(); err != nil {
		logFatal(err.Error())
	}
}

func getModuleCurrentVersion(modName, path string) string {
	mod := module{
		name: modName,
		path: path,
	}
	mod.UpdateCurrentVersion()
	v := mod.version.String()
	logDebug("module %s version.toString => %s", mod.name, v)
	return v
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

func (v *moduleVersion) Set(major, minor, patch int) {
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

func (v *moduleVersion) Bump(t string) {
	if t == "major" {
		v.major++
		v.minor = 0
		v.patch = 0
	} else if t == "minor" {
		v.minor++
		v.patch = 0
	} else if t == "patch" {
		v.patch++
	} else {
		logFatal("Invalid version type: %s", t)
	}
}

// === module struct and functions definition ===

type module struct {
	name    string
	path    string
	version moduleVersion
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

func (m *module) Tag() string {
	return m.name + "/" + m.version.String()
}

func (m *module) RunTest() (string, error) {
	if noTest {
		logInfo("Tests disabled.")
		return "", nil
	}
	testPath := path.Join(m.path, m.name)
	logInfo("Running tests in %s...", testPath)
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = testPath
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return string(stdoutStderr), err
	}
	logInfo("Tests are successfully finished")
	return "", nil
}

// === Git environment functions ===

func prepareGit() {
	var err error
	// Create temporary directory
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

func checkRemoteExistence(path, remote string) {
	logDebug("Checking remote %s in %s", remote, path)
	cmd := exec.Command("git", "remote")
	cmd.Dir = path
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		logFatal(string(stdoutStderr))
	}
	logDebug("Remotes:\n%s", string(stdoutStderr))

	regString := fmt.Sprintf("(?m)^%s$", remote)
	reg := regexp.MustCompile(regString)
	if !reg.MatchString(string(stdoutStderr)) {
		logFatal("Cannot find remote named %s", remote)
	}
	logDebug("Remote %s exists", remote)
}

func fetchTags(path, remote string) {
	logDebug("Fetching latest tags")
	cmd := exec.Command("git", "fetch", "-t", remote)
	cmd.Dir = path
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		logFatal(string(stdoutStderr))
	}
	logDebug("Finished fetching")
}

func checkBranchExistence(path, name string) bool {
	logDebug("Checking branch %s existence", name)
	cmd := exec.Command("git", "branch", "-a")
	cmd.Dir = path
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		logFatal(string(stdoutStderr))
	}
	return strings.Contains(string(stdoutStderr), name)
}

func newBranch(path, name string) {
	logInfo("Creating new branch %s", name)
	upstreamBranch := "upstream/" + name
	cmd := exec.Command("git", "branch", name, upstreamBranch)
	if !checkBranchExistence(path, upstreamBranch) {
		logInfo("Remote branch %s doesn't exist", upstreamBranch)
		cmd = exec.Command("git", "branch", name)
	}
	cmd.Dir = path
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		logFatal(string(stdoutStderr))
	}
	logInfo("Finished creating branch")
}

func deleteBranch(path, name string) {
	logDebug("Deleting branch %s", name)
	cmd := exec.Command("git", "branch", "-D", name)
	cmd.Dir = path
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		logFatal(string(stdoutStderr))
	}
	logDebug("Finished deleting branch")
}

func addWorktree(path, tempDir, branch string) {
	logInfo("Adding worktree %s for branch %s", tempDir, branch)
	cmd := exec.Command("git", "worktree", "add", tempDir, branch)
	cmd.Dir = path
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		logFatal(string(stdoutStderr))
	}
	logInfo("Finished adding worktree")
}

func pruneWorktree(path string) {
	logDebug("Pruning worktree for repo %s", path)
	cmd := exec.Command("git", "worktree", "prune")
	cmd.Dir = path
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		logFatal(string(stdoutStderr))
	}
	logDebug("Finished pruning worktree")
}

func merge(path, branch string) {
	logInfo("Merging %s", branch)
	logDebug("Working dir: %s", path)
	cmd := exec.Command("git", "merge", branch)
	cmd.Dir = path
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		logFatal(string(stdoutStderr))
	}
	logInfo("Finished merging")
}

func pushRelease(path, branch string, mod module) {
	logInfo("Pushing branch %s", branch)
	cmd := exec.Command("git", "push", "upstream", branch)
	cmd.Dir = path
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		logFatal(string(stdoutStderr))
	}

	logInfo("Creating tag %s", mod.Tag())
	cmd = exec.Command(
		"git", "tag",
		"-a", mod.Tag(),
		"-m", fmt.Sprintf("Release %s on branch %s", mod.Tag(), branch),
	)
	cmd.Dir = path
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		logFatal(string(stdoutStderr))
	}

	logInfo("Pushing tag %s", mod.Tag())
	cmd = exec.Command("git", "push", "upstream", mod.Tag())
	cmd.Dir = path
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		logFatal(string(stdoutStderr))
	}
}
