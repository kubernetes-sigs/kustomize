package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var modules = [...]string{
	"kyaml", "api", "cmd/config",
	"cmd/resource", "cmd/kubectl", "pluginator", "kustomize",
}

// Enable verbose or not
var verbose bool

// Disable dry run
var noDryRun bool

// Enable module tests
var doTest bool

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

func logFatalE(e error) {
	log.Fatalf("FATAL %s", e.Error())
}

// === Command line commands ===

var rootCmd = &cobra.Command{
	Use: "releasing",
	Short: `This go program is used to improve the modules releasing process in Kustomize repository.
Note: You may need to run fixgomod.sh in the module to make the module ready to release.`,
}

func listCmdImpl() error {
	gr, err := newGitRunner(false)
	if err != nil {
		return err
	}
	logDebug("Working directory: %s", gr.originalGitPath)
	remote := "upstream"

	err = gr.CheckRemoteExistence(remote)
	if err != nil {
		return err
	}
	err = gr.FetchTags(remote)
	if err != nil {
		return err
	}
	tags, err := gr.GetTags()
	if err != nil {
		return err
	}

	res := []string{} // Store result strings
	for _, modName := range modules {
		mod := module{
			name: modName,
		}
		err = mod.UpdateVersion(tags)
		if err != nil {
			return err
		}
		res = append(res, fmt.Sprintf("%s/%s", mod.name, mod.version.String()))
	}
	err = gr.Close()
	if err != nil {
		return err
	}
	for _, l := range res {
		fmt.Println(l)
	}
	return nil
}

var listSubCmd = &cobra.Command{
	Use:   "list",
	Short: "List current version of all covered modules",
	Run: func(cmd *cobra.Command, args []string) {
		err := listCmdImpl()
		if err != nil {
			logFatalE(err)
		}
	},
}

func checkReleaseArgs(args []string) error {
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
}

func releaseCmdImpl(args []string) error {
	modName := args[0]
	versionType := args[1]
	gr, err := newGitRunner(true)
	if err != nil {
		return err
	}
	logInfo("Creating tag for module %s", modName)
	logDebug("Working directory: %s", gr.originalGitPath)
	remote := "upstream"

	err = gr.CheckRemoteExistence(remote)
	if err != nil {
		return err
	}
	err = gr.FetchTags(remote)
	if err != nil {
		return err
	}
	tags, err := gr.GetTags()
	if err != nil {
		return err
	}

	gitPath, err := gr.OriginalGitPath()
	if err != nil {
		return err
	}
	mod := module{
		name: modName,
		path: gitPath,
	}
	err = mod.UpdateVersion(tags)
	if err != nil {
		return err
	}

	oldVersion := mod.version.String()
	err = mod.version.Bump(versionType)
	if err != nil {
		return err
	}
	newVersion := mod.version.String()
	logInfo("Bumping version: %s => %s", oldVersion, newVersion)

	// Create branch
	branch := fmt.Sprintf("release-%s-v%d.%d", mod.name, mod.version.major, mod.version.minor)
	err = gr.NewBranch(branch)
	if err != nil {
		return err
	}

	err = gr.AddWorktree(branch)
	if err != nil {
		return err
	}

	err = gr.Merge("upstream/master")
	if err != nil {
		return err
	}
	// Update module path
	worktreePath, err := gr.WorktreePath()
	if err != nil {
		return err
	}
	mod.path = worktreePath

	logInfo(
		"Releasing summary:\nDir:\t%s\nModule:\t%s %s\nBranch:\t%s\nTag:\t%s",
		worktreePath,
		mod.name,
		mod.version.String(),
		branch,
		mod.Tag(),
	)

	// Check is there replace statement in go.mod
	err = mod.CheckModReplace()
	if err != nil {
		return err
	}

	// Run module tests
	output, err := mod.RunTest()
	if err != nil {
		logWarn(output)
	} else if !noDryRun {
		logInfo("Skipping push module %s. Run with --no-dry-run to push the release.", mod.name)
	} else {
		err = gr.PushRelease(branch, mod)
		if err != nil {
			return err
		}
	}
	// Clean
	err = gr.Close()
	if err != nil {
		return err
	}
	err = gr.DeleteBranch(branch)
	if err != nil {
		return err
	}
	logInfo("Releasing for module %s completes", mod.name)
	return nil
}

var release = &cobra.Command{
	Use:   "release [module name] [version type]",
	Short: "Release a new version of specified module",
	Args: func(cmd *cobra.Command, args []string) error {
		return checkReleaseArgs(args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := releaseCmdImpl(args)
		if err != nil {
			logFatalE(err)
		}
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
	release.Flags().BoolVarP(&doTest, "do-test", "", false, "run module tests before releasing")

	if err := rootCmd.Execute(); err != nil {
		logFatal(err.Error())
	}
}
