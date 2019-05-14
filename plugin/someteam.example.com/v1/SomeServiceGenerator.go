// +build plugin

/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"text/template"

	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/yaml"
)

// A simple generator example.  Makes one service.
type plugin struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	Port string `json:"port,omitempty" yaml:"port,omitempty"`
}

var KustomizePlugin plugin

const tmpl = `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: dev
  name: {{.Name}}
spec:
  ports:
  - port: {{.Port}}
  selector:
    app: dev
`

func (p *plugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, config []byte) error {
	return yaml.Unmarshal(config, p)
}

func (p *plugin) Generate() (resmap.ResMap, error) {
	var buf bytes.Buffer
	temp := template.Must(template.New("tmpl").Parse(tmpl))
	err := temp.Execute(&buf, p)
	if err != nil {
		return nil, err
	}
	rf := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()))
	return rf.NewResMapFromBytes(buf.Bytes())
}
