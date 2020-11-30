// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// ResourceList reads the function input and writes the function output.
//
// Adheres to the spec: https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md
type ResourceList struct {
	// FunctionConfig is the ResourceList.functionConfig input value.  If FunctionConfig
	// is set to a value such as a struct or map[string]interface{} before ResourceList.Read()
	// is called, then the functionConfig will be parsed into that value.
	// If it is nil, the functionConfig will be set to a map[string]interface{}
	// before it is parsed.
	//
	// e.g. given the function input:
	//
	//    kind: ResourceList
	//    functionConfig:
	//      kind: Example
	//      spec:
	//        foo: var
	//
	// FunctionConfig will contain the Example unmarshalled into its value.
	FunctionConfig interface{}

	// Items is the ResourceList.items input and output value.  Items will be set by
	// ResourceList.Read() and written by ResourceList.Write().
	//
	// e.g. given the function input:
	//
	//    kind: ResourceList
	//    items:
	//    - kind: Deployment
	//      ...
	//    - kind: Service
	//      ...
	//
	// Items will be a slice containing the Deployment and Service resources
	Items []*yaml.RNode

	// Result is ResourceList.result output value.  Result will be written by
	// ResourceList.Write()
	Result *Result

	// DisableStandalone if set will not support standalone mode
	DisableStandalone bool

	// Args are the command args used for standalone mode
	Args []string

	// Flags are an optional set of flags to parse the ResourceList.functionConfig.data.
	// If non-nil, ResourceList.Read() will set the flag value for each flag name matching
	// a ResourceList.functionConfig.data map entry.
	//
	// e.g. given the function input:
	//
	//    kind: ResourceList
	//    functionConfig:
	//      data:
	//        foo: bar
	//        a: b
	//
	// The flags --a=b and --foo=bar will be set in Flags.
	Flags *pflag.FlagSet

	// Reader is used to read the function input (ResourceList).
	// Defaults to os.Stdin.
	Reader io.Reader

	// Writer is used to write the function output (ResourceList)
	// Defaults to os.Stdout.
	Writer io.Writer

	// rw reads function input and writes function output
	rw *kio.ByteReadWriter

	// NoPrintError if set will prevent the error from being printed
	NoPrintError bool

	Command *cobra.Command
}

// Read reads the ResourceList
func (r *ResourceList) Read() error {
	var in io.Reader = os.Stdin
	var out io.Writer = os.Stdout
	if r.Command != nil {
		in = r.Command.InOrStdin()
		out = r.Command.OutOrStdout()
	}

	// parse the inputs from the args
	var readStdinStandalone bool
	if len(r.Args) > 0 && !r.DisableStandalone {
		// write the files as input
		var buf bytes.Buffer
		for i := range r.Args {
			// the first argument is the resourceList.FunctionConfig and will be parsed
			// separately later, the rest of the arguments are the resourceList.items
			if i == 0 {
				continue
			}
			if r.Args[i] == "-" {
				// Read stdin separately
				readStdinStandalone = true
				continue
			}

			b, err := ioutil.ReadFile(r.Args[i])
			if err != nil {
				return errors.WrapPrefixf(err, "unable to read input file %s", r.Args[i])
			}
			buf.WriteString(string(b))
			buf.WriteString("\n---\n")
		}
		r.Reader = &buf
	}

	if r.Reader == nil {
		r.Reader = in
	}
	if r.Writer == nil {
		r.Writer = out
	}
	r.rw = &kio.ByteReadWriter{
		Reader:                r.Reader,
		Writer:                r.Writer,
		KeepReaderAnnotations: true,
	}

	// parse the resourceList.FunctionConfig from the first arg
	if len(r.Args) > 0 && !r.DisableStandalone {
		// Don't keep the reader annotations if we are in standalone mode
		r.rw.KeepReaderAnnotations = false
		// Don't wrap the resources in a resourceList -- we are in
		// standalone mode and writing to stdout to be applied
		r.rw.NoWrap = true

		b, err := ioutil.ReadFile(r.Args[0])
		if err != nil {
			return errors.WrapPrefixf(err, "unable to read configuration file %s", r.Args[0])
		}
		fc, err := yaml.Parse(string(b))
		if err != nil {
			return errors.WrapPrefixf(err, "unable to parse configuration file %s", r.Args[0])
		}
		// use this as the function config used to configure the function
		r.rw.FunctionConfig = fc
	}

	var err error
	r.Items, err = r.rw.Read()
	if err != nil {
		return errors.Wrap(err)
	}

	if readStdinStandalone {
		br := kio.ByteReader{Reader: in}
		items, err := br.Read()
		if err != nil {
			return errors.Wrap(err)
		}
		// stdin always comes first so files are patches
		r.Items = append(items, r.Items...)
	}

	// parse the functionConfig
	return func() error {
		if r.rw.FunctionConfig == nil {
			// no function config exists
			return nil
		}
		if r.FunctionConfig == nil {
			// set directly from r.rw
			r.FunctionConfig = r.rw.FunctionConfig
		} else {
			// unmarshal the functionConfig into the provided value
			err := yaml.Unmarshal([]byte(r.rw.FunctionConfig.MustString()), r.FunctionConfig)
			if err != nil {
				return errors.Wrap(err)
			}
		}

		// set the functionConfig values as flags so they are easy to access
		if r.Flags == nil || !r.Flags.HasFlags() {
			return nil
		}
		// flags are always set from the "data" field
		data, err := r.rw.FunctionConfig.Pipe(yaml.Lookup("data"))
		if err != nil || data == nil {
			return err
		}
		return data.VisitFields(func(node *yaml.MapNode) error {
			f := r.Flags.Lookup(node.Key.YNode().Value)
			if f == nil {
				return nil
			}
			return f.Value.Set(node.Value.YNode().Value)
		})
	}()
}

// Write writes the ResourceList
func (r *ResourceList) Write() error {
	// set the ResourceList.results for validating functions
	if r.Result != nil {
		if len(r.Result.Items) > 0 {
			b, err := yaml.Marshal(r.Result)
			if err != nil {
				return errors.Wrap(err)
			}
			y, err := yaml.Parse(string(b))
			if err != nil {
				return errors.Wrap(err)
			}
			r.rw.Results = y
		}
	}

	// write the results
	return r.rw.Write(r.Items)
}

// Command returns a cobra.Command to run a function.
//
// The cobra.Command will use the provided ResourceList to Read() the input,
// run the provided function, and then Write() the output.
//
// The returned cobra.Command will have a "gen" subcommand which can be used to generate
// a Dockerfile to build the function into a container image
//
//		go run main.go gen DIR/
func Command(resourceList *ResourceList, function Function) *cobra.Command {
	cmd := cobra.Command{}
	resourceList.Command = &cmd
	AddGenerateDockerfile(&cmd)
	var printStack bool
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		err := execute(resourceList, function, cmd, args)
		if err != nil && !resourceList.NoPrintError {
			fmt.Fprintf(cmd.ErrOrStderr(), "%v", err)
		}
		// print the stack if requested
		if s := errors.GetStack(err); printStack && s != "" {
			fmt.Fprintln(cmd.ErrOrStderr(), s)
		}
		return err
	}
	cmd.Flags().BoolVar(&printStack, "stack", false, "print the stack trace on failure")
	cmd.Args = cobra.MinimumNArgs(0)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	return &cmd
}

// TemplateCommand provides a cobra command to invoke a template
type TemplateCommand struct {
	// API is the function API provide to the template as input
	API interface{}

	// Template is a go template to render and is appended to Templates.
	Template *template.Template

	// Templates is a list of templates to render.
	Templates []*template.Template

	// TemplatesFn returns a list of templates
	TemplatesFn func(*ResourceList) ([]*template.Template, error)

	// PatchTemplates is a list of templates to render into Patches and apply.
	PatchTemplates []PatchTemplate

	// PatchTemplateFn returns a list of templates to render into Patches and apply.
	// PatchTemplateFn is called after the ResourceList has been parsed.
	PatchTemplatesFn func(*ResourceList) ([]PatchTemplate, error)

	// PatchContainerTemplates applies patches to matching container fields
	PatchContainerTemplates []ContainerPatchTemplate

	// PatchContainerTemplates returns a list of PatchContainerTemplates
	PatchContainerTemplatesFn func(*ResourceList) ([]ContainerPatchTemplate, error)

	// TemplateFiles list of templates to read from disk which are appended
	// to Templates.
	TemplatesFiles []string

	// MergeResources if set to true will apply input resources
	// as patches to the templates
	MergeResources bool

	// PreProcess is run on the ResourceList before the template is invoked
	PreProcess func(*ResourceList) error

	PreProcessFilters []kio.Filter

	// PostProcess is run on the ResourceList after the template is invoked
	PostProcess func(*ResourceList) error

	PostProcessFilters []kio.Filter
}

// ContainerPatchTemplate defines a patch to be applied to containers
type ContainerPatchTemplate struct {
	PatchTemplate

	ContainerNames []string
}

func (tc TemplateCommand) doTemplate(t *template.Template, rl *ResourceList) error {
	// invoke the template
	var b bytes.Buffer
	err := t.Execute(&b, tc.API)
	if err != nil {
		return errors.WrapPrefixf(err, "failed to render template %v", t.DefinedTemplates())
	}
	// split the resources so the error messaging is better
	for _, s := range strings.Split(b.String(), "\n---\n") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		nodes, err := (&kio.ByteReader{Reader: bytes.NewBufferString(s)}).Read()
		if err != nil {
			// create the debug string
			lines := strings.Split(s, "\n")
			for j := range lines {
				lines[j] = fmt.Sprintf("%03d %s", j+1, lines[j])
			}
			s = strings.Join(lines, "\n")
			return errors.WrapPrefixf(err, "failed to parse rendered template into a resource:\n%s\n", s)
		}

		if tc.MergeResources {
			// apply inputs as patches -- add the
			rl.Items = append(nodes, rl.Items...)
		} else {
			// add to the inputs list
			rl.Items = append(rl.Items, nodes...)
		}
	}
	return nil
}

// Defaulter is implemented by APIs to have Default invoked
type Defaulter interface {
	Default() error
}

func (tc *TemplateCommand) doPreProcess(rl *ResourceList) error {
	// do any preprocessing
	if tc.PreProcess != nil {
		if err := tc.PreProcess(rl); err != nil {
			return err
		}
	}

	// TODO: test this
	if tc.PreProcessFilters != nil {
		for i := range tc.PreProcessFilters {
			fltr := tc.PreProcessFilters[i]
			var err error
			rl.Items, err = fltr.Filter(rl.Items)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (tc *TemplateCommand) doMerge(rl *ResourceList) error {
	var err error
	if tc.MergeResources {
		rl.Items, err = filters.MergeFilter{}.Filter(rl.Items)
	}
	return err
}

func (tc *TemplateCommand) doPostProcess(rl *ResourceList) error {
	// finish up
	if tc.PostProcess != nil {
		if err := tc.PostProcess(rl); err != nil {
			return err
		}
	}
	// TODO: test this
	if tc.PostProcessFilters != nil {
		for i := range tc.PostProcessFilters {
			fltr := tc.PostProcessFilters[i]
			var err error
			rl.Items, err = fltr.Filter(rl.Items)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (tc *TemplateCommand) doTemplates(rl *ResourceList) error {
	if tc.Template != nil {
		tc.Templates = append(tc.Templates, tc.Template)
	}

	// TODO: test this
	if tc.TemplatesFn != nil {
		t, err := tc.TemplatesFn(rl)
		if err != nil {
			return err
		}
		tc.Templates = append(tc.Templates, t...)
	}

	for i := range tc.TemplatesFiles {
		tbytes, err := ioutil.ReadFile(tc.TemplatesFiles[i])
		if err != nil {
			return errors.WrapPrefixf(err, "unable to read template file")
		}
		t, err := template.New("files").Parse(string(tbytes))
		if err != nil {
			return errors.WrapPrefixf(err, "unable to parse template files %v", tc.TemplatesFiles)
		}
		tc.Templates = append(tc.Templates, t)
	}

	for i := range tc.Templates {
		if err := tc.doTemplate(tc.Templates[i], rl); err != nil {
			return err
		}
	}
	return nil
}

func (tc *TemplateCommand) doPatchTemplates(rl *ResourceList) error {
	if tc.PatchTemplatesFn != nil {
		pt, err := tc.PatchTemplatesFn(rl)
		if err != nil {
			return err
		}
		tc.PatchTemplates = append(tc.PatchTemplates, pt...)
	}
	for i := range tc.PatchTemplates {
		if err := tc.PatchTemplates[i].Apply(rl); err != nil {
			return err
		}
	}
	return nil
}

func (tc *TemplateCommand) doPatchContainerTemplates(rl *ResourceList) error {
	if tc.PatchContainerTemplatesFn != nil {
		ct, err := tc.PatchContainerTemplatesFn(rl)
		if err != nil {
			return err
		}
		tc.PatchContainerTemplates = append(tc.PatchContainerTemplates, ct...)
	}
	for i := range tc.PatchContainerTemplates {
		ct := tc.PatchContainerTemplates[i]
		matches, err := ct.Selector.GetMatches(rl)
		if err != nil {
			return err
		}
		err = PatchContainersWithTemplate(matches, ct.Template, rl.FunctionConfig, ct.ContainerNames...)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetCommand returns a new cobra command
func (tc TemplateCommand) GetCommand() *cobra.Command {
	rl := ResourceList{
		FunctionConfig: tc.API,
		NoPrintError:   true,
	}
	c := Command(&rl, func() error {
		if d, ok := rl.FunctionConfig.(Defaulter); ok {
			if err := d.Default(); err != nil {
				return err
			}
		}

		if err := tc.doPreProcess(&rl); err != nil {
			return err
		}
		if err := tc.doTemplates(&rl); err != nil {
			return err
		}
		if err := tc.doPatchTemplates(&rl); err != nil {
			return err
		}
		if err := tc.doPatchContainerTemplates(&rl); err != nil {
			return err
		}
		if err := tc.doMerge(&rl); err != nil {
			return err
		}
		if err := tc.doPostProcess(&rl); err != nil {
			return err
		}

		return nil
	})
	return c
}

// AddGenerateDockerfile adds a "gen" subcommand to create a Dockerfile for building
// the function as a container.
func AddGenerateDockerfile(cmd *cobra.Command) {
	gen := &cobra.Command{
		Use:  "gen",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ioutil.WriteFile(filepath.Join(args[0], "Dockerfile"), []byte(`FROM golang:1.13-stretch
ENV CGO_ENABLED=0
WORKDIR /go/src/
COPY . .
RUN go build -v -o /usr/local/bin/function ./

FROM alpine:latest
COPY --from=0 /usr/local/bin/function /usr/local/bin/function
CMD ["function"]
`), 0600)
		},
	}
	cmd.AddCommand(gen)
}

func execute(rl *ResourceList, function Function, cmd *cobra.Command, args []string) error {
	rl.Reader = cmd.InOrStdin()
	rl.Writer = cmd.OutOrStdout()
	rl.Flags = cmd.Flags()
	rl.Args = args

	if err := rl.Read(); err != nil {
		return err
	}

	retErr := function()

	if err := rl.Write(); err != nil {
		return err
	}

	return retErr
}

// Filters returns a function which returns the provided Filters
func Filters(fltrs ...kio.Filter) func(*ResourceList) []kio.Filter {
	return func(*ResourceList) []kio.Filter {
		return fltrs
	}
}
