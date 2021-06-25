module sigs.k8s.io/kustomize/plugin/someteam.example.com/v1/starlarkmixer

go 1.16

require sigs.k8s.io/kustomize/api v0.8.9

replace sigs.k8s.io/kustomize/api => ../../../../api

replace sigs.k8s.io/kustomize/kyaml => ../../../../kyaml

replace gopkg.in/yaml.v3 => github.com/natasha41575/yaml v0.0.0-20210625180258-d648d2e6eae5
