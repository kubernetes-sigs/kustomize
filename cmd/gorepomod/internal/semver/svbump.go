package semver

type SvBump int

const (
	Patch SvBump = iota
	Minor
	Major
)

func (b SvBump) String() string {
	return map[SvBump]string{
		Patch: "Patch",
		Minor: "Minor",
		Major: "Major",
	}[b]
}
