module sigs.k8s.io/kustomize/plugin/someteam.example.com/v1/starlarkmixer

go 1.16

require sigs.k8s.io/kustomize/api v0.0.0

replace sigs.k8s.io/kustomize/kyaml => ../../../../kyaml

replace sigs.k8s.io/kustomize/api => ../../../../api

replace gopkg.in/yaml.v3 => gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
