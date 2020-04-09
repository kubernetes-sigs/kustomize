module sigs.k8s.io/kustomize/plugin/someteam.example.com/v1/gogetter

go 1.13

require sigs.k8s.io/kustomize/api v0.3.1

replace sigs.k8s.io/kustomize/api v0.3.1 => ../../../../api
