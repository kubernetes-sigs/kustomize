package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var modules = [...]string{
	"kyaml", "api", "kstatus", "cmd/config",
	"cmd/resource", "cmd/kubectl", "pluginator", "kustomize",
}
var verbose bool

// === Log helper functions ===
func logDebug(s string) {
	if verbose {
		log.Println("DEBUG " + s)
	}
}

func logInfo(s string) {
	log.Println("INFO " + s)
}

func logFatal(s string) {
	log.Fatalln("FATAL " + s)
}

var rootCmd = &cobra.Command{
	Use:   "releasing",
	Short: "This go program is used to improve the modules releasing process in Kustomize repository.",
}

var listSubCmd = &cobra.Command{
	Use:   "list",
	Short: "List current version of all covered modules",
	Run: func(cmd *cobra.Command, args []string) {
		res := []string{}
		for _, mod := range modules {
			res = append(res, fmt.Sprintf("%s/%s", mod, getModuleCurrentVersion(mod)))
		}
		for _, l := range res {
			fmt.Println(l)
		}
	},
}

var subCmds = [...]*cobra.Command{
	listSubCmd,
}

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
	mod := newModule(modName)
	mod.updateCurrentVersion()
	return mod.currentVersion
}

type module struct {
	name           string
	path           string
	currentVersion string
}

func newModule(modName string) module {
	mod := module{
		name: modName,
	}
	logDebug(fmt.Sprintf("Created module struct for %s", modName))
	return mod
}

func (m *module) updateCurrentVersion() {
	m.currentVersion = "v1.0.0"
}
