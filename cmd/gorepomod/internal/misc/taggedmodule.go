package misc

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/semver"
)

// TaggedModule is a module known to be tagged with the given version.
type TaggedModule struct {
	M LaModule
	V semver.SemVer
}

func (p TaggedModule) String() string {
	if p.V.IsZero() {
		return string(p.M.ShortName())
	}
	return string(p.M.ShortName()) + "/" + p.V.String()
}

type TaggedModules []TaggedModule

func (s TaggedModules) String() string {
	// format := "%-"+strconv.Itoa(s.LenLongestString()+2)+"s"
	var b strings.Builder
	for i := range s {
		b.WriteString(fmt.Sprintf("%-19s", s[i]))
	}
	return b.String()
}

func (s TaggedModules) LenLongestString() (ans int) {
	for _, m := range s {
		l := len(m.String())
		if l > ans {
			ans = l
		}
	}
	return
}
