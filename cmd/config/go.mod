module sigs.k8s.io/kustomize/cmd/config

go 1.13

require (
	github.com/go-errors/errors v1.0.1
	github.com/go-openapi/spec v0.19.5
	github.com/olekukonko/tablewriter v0.0.4
	github.com/posener/complete/v2 v2.0.1-alpha.12
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	k8s.io/apimachinery v0.17.0
	sigs.k8s.io/kustomize/kyaml v0.0.0 // Don't change this!
)

// Don't change this!
//
// This line is managed by the release script -- releasing/releasemodule.sh
// Pinning to a released version of kyaml will invalidate the e2e tests used to
// test kyaml changes as the e2e tests will run against the pinned version, not
// the HEAD.
//
// releasing/releasemodule.sh will remove this line and set the require version
// to the kyaml version specified in releasing/VERSIONS
replace sigs.k8s.io/kustomize/kyaml v0.0.0 => ../../kyaml
