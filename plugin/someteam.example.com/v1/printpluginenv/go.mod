module sigs.k8s.io/kustomize/plugin/someteam.example.com/v1/printpluginenv

go 1.16

require sigs.k8s.io/kustomize/api v0.8.5

replace sigs.k8s.io/kustomize/kyaml => ../../../../kyaml

replace sigs.k8s.io/kustomize/api => ../../../../api
