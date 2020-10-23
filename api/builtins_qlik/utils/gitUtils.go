package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetGitDescribeForHead returns tag/version from a git repository based on HEAD
func GetGitDescribeForHead(dir string, defaultVersion string) (string, error) {
	if defaultVersion == "" {
		defaultVersion = "0.0.0"
	}

	out, err := exec.Command("git", "-C", dir, "describe", "--tags", "--abbrev=7", "--match", "v[0-9]*.[0-9]*.[0-9]*").Output()
	if err != nil {
		out, err = exec.Command("git", "-C", dir, "rev-parse", "--short=7", "HEAD").Output()
		if err != nil {
			return fmt.Sprintf("%s", strings.TrimPrefix(defaultVersion, "v")), nil
		}
		out = []byte(fmt.Sprintf("0.0.0-0-g%v", string(out)))
	}

	tag := strings.TrimSpace(string(out))
	version := strings.TrimPrefix(tag, "v")

	return string(version), nil
}
