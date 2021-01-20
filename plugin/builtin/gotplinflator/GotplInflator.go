// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"

	//"sigs.k8s.io/kustomize/v3/pkg/ifc"
	//"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/yaml"

	yamlv2 "gopkg.in/yaml.v2"
)

var gotplFilePattern = "*.t*pl"
var manifestFilePattern = "*.y*ml"
var renderedManifestFilePattern = "*.rendered.y*ml"
var SprigCustomFuncs = map[string]interface{}{
	"handleEnvVars": func(rawEnvs interface{}) map[string]string {
		envs := map[string]string{}
		if str, ok := rawEnvs.(string); ok {
			err := json.Unmarshal([]byte(str), &envs)
			if err != nil {
				log.Fatal("failed to unmarshal Envs,", err)
			}
		}
		return envs
	},
	//
	// Shameless copy from:
	// https://github.com/helm/helm/blob/master/pkg/engine/engine.go#L107
	//
	// Some more Helm template functions:
	// https://github.com/helm/helm/blob/master/pkg/engine/funcs.go
	//
	"toYaml": func(v interface{}) string {
		data, err := yaml.Marshal(v)
		if err != nil {
			// Swallow errors inside of a template.
			return ""
		}
		return strings.TrimSuffix(string(data), "\n")
	},
}

// GotplInflatorPlugin is a plugin to generate resources
// from a remote or local go templates.
type GotplInflatorPlugin struct {
	//type remoteResource struct {
	h                *resmap.PluginHelpers
	types.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	types.GotplInflatorArgs

	rf *resmap.Factory

	TempDir string

	GotplInflatorRoot string
}

// Config uses the input plugin configurations `config` to setup the generator
//func (p *plugin) Config(
func (p *GotplInflatorPlugin) Config(h *resmap.PluginHelpers, config []byte) error {

	p.h = h
	err := yaml.Unmarshal(config, p)
	if err != nil {
		return err
	}
	return nil
}

// GotplRender process templates
func (p *GotplInflatorPlugin) GotplRender(t string) error {

	// read template
	tContent, err := ioutil.ReadFile(t)
	if err != nil {
		return fmt.Errorf("Read template failed: %v", err)
	}

	// init
	fMap := sprig.TxtFuncMap()
	for k, v := range SprigCustomFuncs {
		fMap[k] = v
	}
	//tOpt := strings.Split(rs.TemplateOpts, ",")
	tpl := template.Must(
		//.Option(tOpt)
		//.ParseGlob("*.html")
		template.New(t).Funcs(fMap).Parse(string(tContent)),
	)

	//render
	var rb bytes.Buffer
	err = tpl.Execute(&rb, p.Values)
	if err != nil {
		return err
	}

	// write
	tBasename := strings.TrimSuffix(t, filepath.Ext(t))
	tBasename = strings.TrimSuffix(t, filepath.Ext(tBasename)) // removes .yaml
	err = ioutil.WriteFile(tBasename+".rendered.yaml", rb.Bytes(), 0640)
	if err != nil {
		//log.Fatal("Write template failed:", err)
		return fmt.Errorf("Write template failed: %v", err)
	}
	return nil
}

// Generate fetch, render and return manifests from remote sources
func (p *GotplInflatorPlugin) Generate() (resmap.ResMap, error) {

	//DEBUG
	//for _, e := range os.Environ() {
	//    pair := strings.SplitN(e, "=", 2)
	//	fmt.Printf("#DEBUG %s='%s'\n", pair[0], pair[1])
	//}

	// FIXME - hardcoded /envs/ will go away and will be replaced by config option
	var pluginConfigRoot = os.Getenv("KUSTOMIZE_PLUGIN_CONFIG_ROOT")
	if os.Getenv("KUSTOMIZE_GOTPLINFLATOR_ROOT") == "" {
		var envsPath = strings.SplitAfter(pluginConfigRoot, "/envs/")
		if len(envsPath) > 1 {
			os.Setenv("KUSTOMIZE_GOTPLINFLATOR_ROOT", filepath.Join(envsPath[0], "../repos"))
			os.Setenv("ENV", strings.Split(envsPath[1], "/")[0])
		}
	}

	// where to fetch, render, otherwise tempdir
	p.GotplInflatorRoot = os.Getenv("KUSTOMIZE_GOTPLINFLATOR_ROOT")

	// tempdir
	err := p.getTempDir()
	if err != nil {
		return nil, fmt.Errorf("Failed to create temp dir: %s", p.TempDir)
	}

	// normalize context values used for template rendering
	nv := make(map[string]interface{})
	FlattenMap("", p.Values, nv)
	p.Values = nv
	//DEBUG
	//for k, v := range p.Values {
	//	fmt.Printf("%s:%s\n", k, v)
	//}

	// fetch dependencies
	err = p.fetchDependencies()
	if err != nil {
		return nil, fmt.Errorf("Error getting remote source: %v", err)
	}

	// render to files
	err = p.RenderDependencies()
	if err != nil {
		return nil, fmt.Errorf("Template rendering failed: %v", err)
	}

	// prepare output buffer
	var output bytes.Buffer
	output.WriteString("\n---\n")

	// collect, filter, parse manifests
	err = p.ReadManifests(output)
	if err != nil {
		return nil, fmt.Errorf("Read manifest failed: %v", err)
	}

	// cleanup
	var cleanupOpt = os.Getenv("KUSTOMIZE_GOTPLINFLATOR_CLEANUP")
	if p.GotplInflatorRoot != "" && cleanupOpt == "ALWAYS" {
		err = p.CleanWorkdir()
		if err != nil {
			return nil, fmt.Errorf("Cleanup failed: %v", err)
		}
	}

	return p.h.ResmapFactory().NewResMapFromBytes(output.Bytes())
}

// getTempDir prepare working directory
func (p *GotplInflatorPlugin) getTempDir() error {
	if p.GotplInflatorRoot != "" {
		//p.TempDir = filepath.Join(p.GotplInflatorRoot, os.Getenv("ENV"))
		p.TempDir = filepath.Join(p.GotplInflatorRoot)
		if _, err := os.Stat(p.TempDir); os.IsNotExist(err) {
			err := os.MkdirAll(p.TempDir, 0770)
			if err != nil {
				return err
			}
		}
	} else {
		var _tempDir filesys.ConfirmedDir
		_tempDir, err := filesys.NewTmpConfirmedDir()
		p.TempDir = _tempDir.String()
		if err != nil {
			return err
		}
	}
	// DEBUG
	// fmt.Println("# TempDir:", p.TempDir)
	return nil
}

// RenderDependencies render gotpl manifests
func (p *GotplInflatorPlugin) RenderDependencies() error {
	for _, rs := range p.Dependencies {

		fmt.Println("# Rendering:", rs.Name)

		// TODO, render manifests to output buffer directly. So it does not require

		// find templates
		if rs.TemplatePattern != "" {
			gotplFilePattern = rs.TemplatePattern
		}
		templates, err := WalkMatch(rs.Dir, gotplFilePattern)
		if os.IsNotExist(err) {
			//fmt.Println("# - no templates found")
			continue
		} else if err != nil {
			return err
		}

		// actual render
		for _, t := range templates {
			fmt.Printf("# - %s\n", strings.SplitAfter(t, p.TempDir)[1])
			err := p.GotplRender(t)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// fetchDependencies calls go-getter to fetch remote sources
func (p *GotplInflatorPlugin) fetchDependencies() error {
	for _, rs := range p.Dependencies {

		// TODO, to function
		// identify fetched repo with branch/commit/etc..
		var reporefspec = strings.SplitAfter(rs.Repo, "ref=")
		var reporef string
		if len(reporefspec) > 1 {
			reporef = strings.Split(reporefspec[1], "?")[0]
		}
		if reporef == "" {
			//original idea
			//reporef = fmt.sprintf("%d", hash(rs.repo))
			//
			// workaround, go-getter fails to fetch repo with subpath without it
			reporef = "master"
			rs.Repo = rs.Repo + "?ref=master"
		}
		var reporeferal = fmt.Sprintf("%s-%s", rs.Name, reporef)
		var repotempdir = filepath.Join(p.TempDir, reporeferal)
		rs.Dir = repotempdir
		if rs.Path != "" {
			// if subpath defined
			rs.Dir = filepath.Join(p.TempDir, reporeferal, rs.Path)
		}

		//handle credentials
		//https://github.com/hashicorp/go-getter#git-git
		//gettercreds := getrepocreds(rs.repocreds)

		// skip fetch if present and not forced
		_, err := os.Stat(rs.Dir)
		if err == nil && rs.Pull != "always" && os.Getenv("KUSTOMIZE_GOTPLINFLATOR_PULL") != "always" {
			//fmt.PrintLn("#- skipped, already exist")
			continue
		}

		// cleanup/create temp dir
		if os.IsNotExist(err) {
			_ = os.MkdirAll(repotempdir, 0770)
		} else {
			err := os.RemoveAll(repotempdir)
			if err != nil {
				return err
			}
		}

		//fetch
		//DEBUG
		//fmt.println("# go-getter", rs.repo, repotempdir)
		cmd := exec.Command("go-getter", rs.Repo, repotempdir)
		//cmd.stdout = os.stdout
		//cmd.stderr = os.stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("go-getter failed to clone repo %s", err)
		}

		//fetch2 - to be used within go native plugin
		//
		//issue: does not properly handle gitlab, no credentials to fetch
		//
		//pwd, err := os.getwd()
		//if err != nil {
		//	log.fatalf("error getting wd: %s", err)
		//}
		//opts := []getter.clientoption{}
		//client := &getter.client{
		//	ctx:  context.todo(),
		//	src:  rs.repo + gettercreds,
		//	dst:  repotempdir,
		//	pwd:  pwd,
		//	mode: getter.clientmodeany,
		//	detectors: []getter.detector{
		//		new(getter.githubdetector),
		//		new(getter.gitlabdetector),
		//		new(getter.gitdetector),
		//		new(getter.s3detector),
		//		new(getter.gcsdetector),
		//		new(getter.filedetector),
		//		new(getter.bitbucketdetector),
		//	},
		//	options: opts,
		//}
		//return client.get()
	}
	return nil
}

// ReadManifests locate & filter rendered manifests and print them in bytes.Buffer
func (p *GotplInflatorPlugin) ReadManifests(output bytes.Buffer) error {
	for _, rs := range p.Dependencies {
		manifests, err := WalkMatch(rs.Dir, renderedManifestFilePattern)
		if os.IsNotExist(err) {
			//fmt.Println(" - no manifests found")
			continue
		} else if err != nil {
			return err
		}
		for _, m := range manifests {
			mContent, err := ioutil.ReadFile(m)
			if err != nil {
				return err
			}
			// TODO - to function
			// test/parse rendered manifest
			mk := make(map[interface{}]interface{})
			err = yamlv2.Unmarshal([]byte(mContent), &mk)
			if err != nil {
				return err
			}
			// Kustomize lacks resource removal and multiple namespace manifests from dependencies cause `already registered id: ~G_v1_Namespace|~X|sre\`
			// https://kubectl.docs.kubernetes.io/faq/kustomize/eschewedfeatures/#removal-directives
			k := mk["kind"]
			if k != nil {
				kLcs := strings.ToLower(k.(string))
				if k == "namespace" || stringInSlice("!"+kLcs, rs.Kinds) {
					continue
				}
				if len(rs.Kinds) == 0 || stringInSlice(kLcs, rs.Kinds) {
					output.Write([]byte(mContent))
					output.WriteString("\n---\n")
				}
			}
		}
	}
	return nil
}

// CleanWorkdir cleanup temporary files plugin uses for build
func (p *GotplInflatorPlugin) CleanWorkdir() error {
	if os.Getenv("KUSTOMIZE_DEBUG") == "" {
		err := os.RemoveAll(p.TempDir)
		if err != nil {
			return err
		}
	} else {
		// DEBUG
		fmt.Println("# TempDir:", p.TempDir)
	}
	return nil
}

//
// PLUGIN UTILS

// stringInSlice boolean function
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// FlattenMap flatten context values to snake_case
// How about https://godoc.org/github.com/jeremywohl/flatten
func FlattenMap(prefix string, src map[string]interface{}, dest map[string]interface{}) {
	if len(prefix) > 0 {
		prefix += "_"
	}
	for k, v := range src {
		switch child := v.(type) {
		case map[string]interface{}:
			FlattenMap(prefix+k, child, dest)
		default:
			dest[prefix+k] = v
		}
	}
}

//WalkMatch returns list of files to render/process
func WalkMatch(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
			// TODO recursive
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

//getRepoCreds returns go-getter URI based on required credential profile (see: plugin configuration)
func getRepoCreds(repoCreds string) string {
	// not required if exec is used for go-getter => os env
	// for S3 you may want better way to get tokens, keys etc..
	// FIXME, builtin plugin, load repoCreds from plugin config?
	var cr = ""
	if repoCreds != "" {
		for _, e := range strings.Split(repoCreds, ",") {
			pair := strings.SplitN(e, "=", 2)
			if pair[0] == "sshkey" {
				key, err := ioutil.ReadFile(pair[1])
				if err != nil {
					log.Fatal(err)
				}
				keyb64 := base64.StdEncoding.EncodeToString([]byte(strings.TrimSpace(string(key))))
				cr = fmt.Sprintf("%s?sshkey=%s", cr, string(keyb64))
			}
		}
	}
	return cr
}

// hash generate fowler–noll–vo hash from string
func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
