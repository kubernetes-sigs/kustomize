// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/parser"
)

//go:embed templates/*
var templateFS embed.FS

var annotationTemplate = `
metadata:
  annotations: 
    value: {{ .Value }}
`

func buildProcessor(value *string) framework.ResourceListProcessor {
	return framework.TemplateProcessor{
		ResourceTemplates: []framework.ResourceTemplate{{
			Templates: parser.TemplateFiles("templates").FromFS(templateFS),
		}},
		PatchTemplates: []framework.PatchTemplate{&framework.ResourcePatchTemplate{
			Templates: parser.TemplateStrings(annotationTemplate),
		}},
		// This will be populated from the --value flag if provided,
		// or the config file's `value` field if provided, with the latter taking precedence.
		TemplateData: &struct {
			Value *string `yaml:"value"`
		}{Value: value}}
}

func buildCmd() *cobra.Command {
	var value string
	cmd := command.Build(buildProcessor(&value), command.StandaloneEnabled, false)
	cmd.Flags().StringVar(&value, "value", "", "annotation value")
	return cmd
}

func main() {
	if err := buildCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
