// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"text/template"

	"sigs.k8s.io/kustomize/v3/pkg/ifc"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/types"
	"sigs.k8s.io/yaml"
)

// A simple generator example.  Makes one service.
type plugin struct {
	rf               *resmap.Factory
	types.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Port             string `json:"port,omitempty" yaml:"port,omitempty"`
}

//nolint: golint
//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

const tmpl = `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: dev
  name: {{.Name}}
  namespace: {{.Namespace}}
spec:
  ports:
  - port: {{.Port}}
  selector:
    app: dev
`

func (p *plugin) Config(
	_ ifc.Loader, rf *resmap.Factory, config []byte) error {
	p.rf = rf
	return yaml.Unmarshal(config, p)
}

func (p *plugin) Generate() (resmap.ResMap, error) {
	var buf bytes.Buffer
	temp := template.Must(template.New("tmpl").Parse(tmpl))
	err := temp.Execute(&buf, p)
	if err != nil {
		return nil, err
	}
	return p.rf.NewResMapFromBytes(buf.Bytes())
}
