module sigs.k8s.io/kustomize/cmd/gorepomod

go 1.22.7

exclude (
	// These dependencies were updated to "master" in spf13/viper, but have no
	// code-changes since the last release. While a fix was merged in viper,
	// it's not released yet, and it may take time before other (indirect)
	// dependencies also downgraded.
	//
	// Exclude these versions so that go modules picks their latest release
	// before that. For additional details, see; https://github.com/kubernetes-sigs/kustomize/pull/5830
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2
)

require (
	github.com/spf13/viper v1.17.0
	github.com/stretchr/testify v1.9.0
	golang.org/x/mod v0.20.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/frankban/quicktest v1.14.6 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml/v2 v2.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/sagikazarmark/locafero v0.3.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/exp v0.0.0-20240103183307-be819d1f06fc // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
