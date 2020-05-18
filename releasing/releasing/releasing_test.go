package main

import (
	"os"
	"regexp"
	"testing"
)

func TestGetModuleCurrentVersion(t *testing.T) {
	var err error
	pwd, err = os.Getwd()
	if err != nil {
		t.Errorf(err.Error())
	}
	remote := "upstream"
	// Check remotes
	checkRemoteExistence(pwd, remote)
	// Fetch latest tags from remote
	fetchTags(pwd, remote)
	for _, mod := range modules {
		v := getModuleCurrentVersion(mod)
		valid, err := regexp.MatchString("^v(\\d+\\.){2}\\d+$", v)
		if err != nil {
			t.Errorf(err.Error())
		}
		if !valid {
			t.Errorf("Returned version %s is not valid", v)
		}
	}
}
