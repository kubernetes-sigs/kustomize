package main

import (
	"bytes"
	"text/template"

	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
)

type plugin struct {
	ServiceName string
	Port        string
}

var Generator plugin

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

func (p *plugin) Config(k ifc.Kunstructured) error {
	var err error
	p.ServiceName, err = k.GetFieldValue("service")
	if err != nil {
		return err
	}
	p.Port, err = k.GetFieldValue("port")
	if err != nil {
		return err
	}
	return nil
}

func (p *plugin) Generate() (resmap.ResMap, error) {
	var buf bytes.Buffer

	temp := template.Must(template.New("manifest").Parse(manifest))
	err := temp.Execute(&buf, p)
	if err != nil {
		return nil, err
	}
	rf := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()))
	return rf.NewResMapFromBytes(buf.Bytes())
}
