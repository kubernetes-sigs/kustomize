package krmfunction

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/rakyll/statik/fs"
	// load embedded func wrapper
	_ "sigs.k8s.io/kustomize/cmd/pluginator/v2/internal/krmfunction/funcwrapper"
)

// Converter is a converter to convert the
// plugin file to KRM function
type Converter struct {
	// Path to the output directory
	outputDir string
	// Path to the input file
	inputFile       string
	wrapperFileName string
	pluginFileName  string
	goModFileName   string
	dockerFileName  string
}

// NewConverter return a pointer to a new converter
func NewConverter(outputDir, inputFile string) *Converter {
	return &Converter{
		outputDir:       outputDir,
		inputFile:       inputFile,
		wrapperFileName: "main.go",
		pluginFileName:  "plugin.go",
		goModFileName:   "go.mod",
		dockerFileName:  "Dockerfile",
	}
}

// Convert converts the input file to a executable
// KRM function and writes to destination directory
func (c *Converter) Convert() error {
	// read and process executable wrapper
	wrapper, err := c.readEmbeddedFile(c.wrapperFileName)
	if err != nil {
		return err
	}
	wrapper = c.prepareWrapper(wrapper)

	if !strings.HasSuffix(c.inputFile, ".go") {
		return fmt.Errorf("input file %s is not a Go file", c.inputFile)
	}

	// read and process plugin code
	pluginCode, err := c.readDiskFile(c.inputFile)
	if err != nil {
		return err
	}
	_, c.pluginFileName = filepath.Split(c.inputFile)

	// go.mod file
	goMod, err := c.readEmbeddedFile(c.goModFileName + ".src")
	if err != nil {
		return err
	}

	// prepare destination directory
	err = c.mkDstDir()
	if err != nil {
		return err
	}

	// write
	return c.write(map[string]string{
		c.wrapperFileName: wrapper,
		c.pluginFileName:  pluginCode,
		c.goModFileName:   goMod,
		c.dockerFileName:  c.getDockerfile(),
	})
}

func (c *Converter) getDockerfile() string {
	return `FROM golang:1.13-stretch
ENV CGO_ENABLED=0
WORKDIR /go/src/
COPY . .
RUN go build -v -o /usr/local/bin/function ./
FROM alpine:latest
COPY --from=0 /usr/local/bin/function /usr/local/bin/function
CMD ["function"]
`
}

func (c *Converter) prepareWrapper(content string) string {
	b := bytes.NewBufferString(content)
	o := &bytes.Buffer{}
	scanner := bufio.NewScanner(b)
	for scanner.Scan() {
		line := scanner.Text()
		// Set the package name to main
		if strings.TrimSpace(line) == "package funcwrappersrc" {
			line = "package main"
		}
		// assign to plugin variable
		if strings.TrimSpace(line) == "var plugin resmap.Configurable" {
			line = line + `
	// KustomizePlugin is a global variable defined in every plugin
	plugin = &KustomizePlugin
`
		}
		o.WriteString(line + "\n")
	}
	return o.String()
}

// readEmbeddedFile read the file from embedded files with filename
// name. Return the file content if it's successful.
func (c *Converter) readEmbeddedFile(name string) (string, error) {
	statikFS, err := fs.New()
	if err != nil {
		return "", err
	}
	r, err := statikFS.Open("/" + name)
	if err != nil {
		return "", err
	}
	defer r.Close()
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}

	return string(contents), nil
}

func (c *Converter) readDiskFile(path string) (string, error) {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(f), nil
}

func (c *Converter) mkDstDir() error {
	p := c.outputDir
	f, err := os.Open(p)
	if err == nil || f != nil {
		return fmt.Errorf("directory %s has already existed", p)
	}

	return os.MkdirAll(p, 0755)
}

func (c *Converter) write(m map[string]string) error {
	for k, v := range m {
		p := filepath.Join(c.outputDir, k)
		err := ioutil.WriteFile(p, []byte(v), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}
