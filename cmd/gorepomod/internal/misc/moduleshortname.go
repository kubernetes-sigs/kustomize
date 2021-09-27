package misc

import (
	"path/filepath"
	"strings"
)

// ModuleShortName is the in-repo path to the directory holding the module
// (holding the go.mod file).  It's the unique in-repo name of the module.
// It's the name used to tag the repo at a particular module version.
// E.g. "" (empty), "kyaml", "cmd/config", "plugin/example/whatever".
type ModuleShortName string

// Never used in a tag.
const ModuleAtTop = ModuleShortName("{top}")
const ModuleUnknown = ModuleShortName("{unknown}")

func (m ModuleShortName) Depth() int {
	if m == ModuleAtTop || m == ModuleUnknown {
		return 0
	}
	return strings.Count(string(m), string(filepath.Separator)) + 1
}
