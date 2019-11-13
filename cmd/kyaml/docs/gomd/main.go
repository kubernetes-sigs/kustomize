// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package main generates cobra.Command go variables containing documentation read from .md files.
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "requires source dir and dest dir as args\n")
		os.Exit(1)
	}
	source := os.Args[1]
	dest := os.Args[2]

	files, err := ioutil.ReadDir(source)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	var pairs []Pair
	for _, f := range files {
		if filepath.Ext(f.Name()) != ".md" {
			continue
		}
		b, err := ioutil.ReadFile(filepath.Join(source, f.Name()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		name := strings.ReplaceAll(f.Name(), filepath.Ext(f.Name()), "")
		name = strings.Title(name)
		name = strings.ReplaceAll(name, "-", "")

		value := string(b)
		value = strings.ReplaceAll(value, "`", "` + \"`\" + `")
		pairs = append(pairs, Pair{Name: name, Value: value})
	}

	out := `
// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package ` + filepath.Base(dest) + `

`
	for i := range pairs {
		out = out + fmt.Sprintf("var %s=`%s\n`\n", pairs[i].Name, pairs[i].Value)
	}

	err = ioutil.WriteFile(filepath.Join(dest, "docs.go"), []byte(out), 0600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

type Pair struct {
	Name  string
	Value string
}
