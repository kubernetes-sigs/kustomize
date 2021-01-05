package edit

import (
	"fmt"
	"os/exec"
	"strings"

	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/misc"
	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/semver"
)

// Editor runs `go mod` commands on an instance of Module.
// If doIt is false, the command is printed, but not run.
type Editor struct {
	module misc.LaModule
	doIt   bool
}

func New(m misc.LaModule, doIt bool) *Editor {
	return &Editor{
		doIt:   doIt,
		module: m,
	}
}

func (e *Editor) run(args ...string) error {
	c := exec.Command(
		"go",
		append([]string{"mod"}, args...)...)
	c.Dir = string(e.module.ShortName())
	if e.doIt {
		out, err := c.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s out=%q", err.Error(), out)
		}
	} else {
		fmt.Printf("in %-60s; %s\n", c.Dir, c.String())
	}
	return nil
}

func upstairs(depth int) string {
	var b strings.Builder
	for i := 0; i < depth; i++ {
		b.WriteString("../")
	}
	return b.String()
}

func (e *Editor) Tidy() error {
	return e.run("tidy")
}

func (e *Editor) Pin(target misc.LaModule, oldV, newV semver.SemVer) error {
	err := e.run(
		"edit",
		"-dropreplace="+target.ImportPath(),
		"-dropreplace="+target.ImportPath()+"@"+oldV.String(),
		"-require="+target.ImportPath()+"@"+newV.String(),
	)
	if err != nil {
		return err
	}
	return e.run("tidy")
}

func (e *Editor) UnPin(target misc.LaModule, oldV semver.SemVer) error {
	var r strings.Builder
	r.WriteString(target.ImportPath())
	// Don't specify the old version.
	// r.WriteString("@")
	// r.WriteString(oldV.String())
	r.WriteString("=")
	r.WriteString(upstairs(e.module.ShortName().Depth()))
	r.WriteString(string(target.ShortName()))
	err := e.run(
		"edit",
		"-replace="+r.String(),
	)
	if err != nil {
		return err
	}
	return e.run("tidy")
}
