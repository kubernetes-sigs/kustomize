module sigs.k8s.io/kustomize/cmd/resource

go 1.12

require (
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.4.0
	golang.org/x/net v0.0.0-20191004110552-13f9640d40b9 // indirect
	k8s.io/api v0.0.0-20190918155943-95b840bb6a1f
	k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
	k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/kustomize/kstatus v0.0.0-20191204200457-7c1b477ff62d
	sigs.k8s.io/kustomize/kyaml v0.0.0-20191202204815-0a19a5dbd9b8
)

replace (
	sigs.k8s.io/kustomize/kstatus v0.0.0-20191204200457-7c1b477ff62d => ../../kstatus
	sigs.k8s.io/kustomize/kyaml v0.0.0-20191202204815-0a19a5dbd9b8 => ../../kyaml
)
