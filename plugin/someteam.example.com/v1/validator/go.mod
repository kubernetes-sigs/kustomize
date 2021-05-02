module sigs.k8s.io/kustomize/plugin/someteam.example.com/v1/validator

go 1.16

require sigs.k8s.io/kustomize/api v0.8.9

replace sigs.k8s.io/kustomize/api => ../../../../api

replace sigs.k8s.io/kustomize/kyaml => ../../../../kyaml
