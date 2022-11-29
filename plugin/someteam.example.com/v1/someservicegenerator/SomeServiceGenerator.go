// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"text/template"

	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

// A simple generator example.  Makes one service.
type plugin struct {
	rf               *resmap.Factory
	types.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Port             string `json:"port,omitempty" yaml:"port,omitempty"`
}

var KustomizePlugin plugin //nolint:gochecknoglobals

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

func (p *plugin) Config(h *resmap.PluginHelpers, config []byte) error {
	p.rf = h.ResmapFactory()
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
