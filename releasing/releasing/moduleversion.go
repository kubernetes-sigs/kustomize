package main

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type moduleVersion struct {
	major int
	minor int
	patch int
}

func (v *moduleVersion) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.major, v.minor, v.patch)
}

func (v *moduleVersion) Bump(t string) error {
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
		return fmt.Errorf("invalid version type: %s", t)
	}
	return nil
}

func newModuleVersionFromString(vs string) (moduleVersion, error) {
	v := moduleVersion{}
	if len(vs) < 1 {
		return v, fmt.Errorf("invalid version string %s", vs)
	}
	if vs[0] == 'v' {
		vs = vs[1:]
	}
	versions := strings.Split(vs, ".")
	if len(versions) != 3 {
		return v, fmt.Errorf("invalid version string %s", vs)
	}
	major, err := strconv.Atoi(versions[0])
	if err != nil {
		return v, err
	}
	minor, err := strconv.Atoi(versions[1])
	if err != nil {
		return v, err
	}
	patch, err := strconv.Atoi(versions[2])
	if err != nil {
		return v, err
	}
	v = moduleVersion{
		major: major,
		minor: minor,
		patch: patch,
	}

	return v, nil
}

func newModuleVersionFromGitTags(tags, modName string) (moduleVersion, error) {
	// Search for module tag
	regString := fmt.Sprintf("(?m)^\\s*%s/v(\\d+\\.){2}\\d+\\s*$", modName)
	reg := regexp.MustCompile(regString)
	modTagsString := reg.FindAllString(tags, -1)
	logDebug("Tags for module %s:\n%s", modName, modTagsString)
	var versions []moduleVersion
	for _, tag := range modTagsString {
		tag = tag[len(modName)+2:]
		v, err := newModuleVersionFromString(tag)
		if err != nil {
			return moduleVersion{}, err
		}

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
	return versions[0], nil
}
