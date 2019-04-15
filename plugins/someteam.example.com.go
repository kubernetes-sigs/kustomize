// +build plugin

package main

import (
	"bytes"
	"text/template"
	"time"

	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/transformers"
	"sigs.k8s.io/kustomize/pkg/transformers/config"
)

type plugin struct {
	cfg []ifc.Kunstructured
}

var KustomizePlugin plugin

func (p *plugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, name string, k []ifc.Kunstructured) error {
	p.cfg = k
	return nil
}

func (p *plugin) Transform(m resmap.ResMap) error {
	datePrefixer := ""
	stringPrefixer := ""
	for _, u := range p.cfg {
		switch u.GetKind() {
		case "DatePrefixer":
			datePrefixer = time.Now().Format("2006-01-02")+"-"
		case "StringPrefixer":
			p, err := u.GetFieldValue("prefix")
			if err != nil {
				return err
			}
			stringPrefixer = p
		}
	}
	prefix := datePrefixer + stringPrefixer
	tr, err := transformers.NewNamePrefixSuffixTransformer(
		prefix, "",
		config.MakeDefaultConfig().NamePrefix)
	if err != nil {
		return err
	}
	return tr.Transform(m)
}

func (p *plugin) Generate() (resmap.ResMap, error) {
	result := resmap.ResMap{}
	var rm resmap.ResMap
	var err error
	for _, u := range p.cfg {
		switch u.GetKind() {
		case "ServiceGenerator":
			args := serviceArgs{}
			args.ServiceName, err = u.GetFieldValue("service")
			if err != nil {
				return nil, err
			}
			args.Port, err = u.GetFieldValue("port")
			if err != nil {
				return nil, err
			}
			var buf bytes.Buffer
			temp := template.Must(template.New("manifest").Parse(manifest))
			err := temp.Execute(&buf, args)
			if err != nil {
				return nil, err
			}
			rf := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()))
			rm, err = rf.NewResMapFromBytes(buf.Bytes())
			if err != nil {
				return nil, err
			}
		}
		result, err = resmap.MergeWithErrorOnIdCollision(result, rm)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

var manifest = `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: dev
  name: {{.ServiceName}}
spec:
  ports:
  - port: {{.Port}}
  selector:
    app: dev
`

type serviceArgs struct {
	ServiceName string
	Port string
}