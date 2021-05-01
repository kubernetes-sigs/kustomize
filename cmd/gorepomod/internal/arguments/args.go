package arguments

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/misc"
	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/semver"
	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/utils"
)

const (
	doItFlag     = "--doIt"
	cmdPin       = "pin"
	cmdUnPin     = "unpin"
	cmdTidy      = "tidy"
	cmdList      = "list"
	cmdRelease   = "release"
	cmdUnRelease = "unrelease"
	cmdDebug     = "debug"
)

var (
	commands = []string{
		cmdPin, cmdUnPin, cmdTidy, cmdList, cmdRelease, cmdUnRelease, cmdDebug}

	// TODO: make this a PATH-like flag
	// e.g.: --excludes ".git:.idea:site:docs"
	excSlice = []string{
		".git",
		".github",
		".idea",
		"docs",
		"examples",
		"hack",
		"plugin",
		"releasing",
		"site",
	}
	// TODO: make this a PATH-like flag
	allowedReplacements = []string{
		"gopkg.in/yaml.v3",
	}
)

type Command int

const (
	Tidy Command = iota
	UnPin
	Pin
	List
	Release
	UnRelease
	Debug
)

type Args struct {
	cmd               Command
	moduleName        misc.ModuleShortName
	conditionalModule misc.ModuleShortName
	version           semver.SemVer
	bump              semver.SvBump
	doIt              bool
}

func (a *Args) GetCommand() Command {
	return a.cmd
}

func (a *Args) AllowedReplacements() (result []string) {
	// Make sure the list has no repeats.
	for k := range utils.SliceToSet(allowedReplacements) {
		result = append(result, k)
	}
	return
}

func (a *Args) Bump() semver.SvBump {
	return a.bump
}

func (a *Args) Version() semver.SemVer {
	return a.version
}

func (a *Args) ModuleName() misc.ModuleShortName {
	return a.moduleName
}

func (a *Args) ConditionalModule() misc.ModuleShortName {
	return a.conditionalModule
}

func (a *Args) Exclusions() (result []string) {
	// Make sure the list has no repeats.
	for k := range utils.SliceToSet(excSlice) {
		result = append(result, k)
	}
	return
}

func (a *Args) DoIt() bool {
	return a.doIt
}

type myArgs struct {
	args []string
	doIt bool
}

func (a *myArgs) next() (result string) {
	if !a.more() {
		panic("no args left")
	}
	result = a.args[0]
	a.args = a.args[1:]
	return
}

func (a *myArgs) more() bool {
	return len(a.args) > 0
}

func newArgs() *myArgs {
	result := &myArgs{}
	for _, a := range os.Args[1:] {
		if a == doItFlag {
			result.doIt = true
		} else {
			result.args = append(result.args, a)
		}
	}
	return result
}

func Parse() (result *Args, err error) {
	result = &Args{}
	clArgs := newArgs()
	result.doIt = clArgs.doIt

	result.moduleName = misc.ModuleUnknown
	result.conditionalModule = misc.ModuleUnknown
	if !clArgs.more() {
		return nil, fmt.Errorf("command needs at least one arg")
	}
	command := clArgs.next()
	switch command {
	case cmdPin:
		if !clArgs.more() {
			return nil, fmt.Errorf("pin needs a moduleName to pin")
		}
		result.moduleName = misc.ModuleShortName(clArgs.next())
		if clArgs.more() {
			result.version, err = semver.Parse(clArgs.next())
			if err != nil {
				return nil, err
			}
		} else {
			result.version = semver.Zero()
		}
		result.cmd = Pin
	case cmdUnPin:
		if !clArgs.more() {
			return nil, fmt.Errorf("unpin needs a moduleName to unpin")
		}
		result.moduleName = misc.ModuleShortName(clArgs.next())
		if clArgs.more() {
			result.conditionalModule = misc.ModuleShortName(clArgs.next())
		}
		result.cmd = UnPin
	case cmdTidy:
		result.cmd = Tidy
	case cmdList:
		result.cmd = List
	case cmdRelease:
		if !clArgs.more() {
			return nil, fmt.Errorf("specify {module} to release")
		}
		result.moduleName = misc.ModuleShortName(clArgs.next())
		bump := "patch"
		if clArgs.more() {
			bump = clArgs.next()
		}
		switch bump {
		case "major":
			result.bump = semver.Major
		case "minor":
			result.bump = semver.Minor
		case "patch":
			result.bump = semver.Patch
		default:
			return nil, fmt.Errorf(
				"unknown bump %s; specify one of 'major', 'minor' or 'patch'", bump)
		}
		result.cmd = Release
	case cmdUnRelease:
		if !clArgs.more() {
			return nil, fmt.Errorf("specify {module} to unrelease")
		}
		result.moduleName = misc.ModuleShortName(clArgs.next())
		result.cmd = UnRelease
	case cmdDebug:
		if !clArgs.more() {
			return nil, fmt.Errorf("specify {module} to debug")
		}
		result.moduleName = misc.ModuleShortName(clArgs.next())
		result.cmd = Debug
	default:
		return nil, fmt.Errorf(
			"unknown command %q; must be one of %v", command, commands)
	}
	if clArgs.more() {
		return nil, fmt.Errorf("unknown extra args: %v", clArgs.args)
	}
	return
}
