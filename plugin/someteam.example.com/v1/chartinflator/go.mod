module sigs.k8s.io/kustomize/plugin/someteam.example.com/v1/chartinflator

go 1.14

require sigs.k8s.io/kustomize/api v0.6.5

replace sigs.k8s.io/kustomize/api v0.6.5 => ../../../../api
replace sigs.k8s.io/kustomize/kyaml v0.9.4 => ../../../../kyaml
