module sigs.k8s.io/kustomize/plugin/someteam.example.com/v1/chartinflator

go 1.13

require sigs.k8s.io/kustomize/api v0.1.1

replace sigs.k8s.io/kustomize/api v0.1.1 => ../../../../api
