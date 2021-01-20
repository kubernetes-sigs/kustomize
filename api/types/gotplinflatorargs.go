package types

// RemoteResource is generic specification for remote resources (git, s3, http...)
type RemoteResource struct {
	// local name for remote
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// go-getter compatible uri to remote
	Repo string `json:"repo" yaml:"repo"`
	// go-getter creds profile for private repos, s3, etc..
	RepoCreds string `json:"repoCreds" yaml:"repoCreds"`
	// PLACEHOLDER, subPath at repo
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
	// pull policy
	Pull string `json:"pull,omitempty" yaml:"pull,omitempty"`
	// template
	Template        string `json:"template,omitempty" yaml:"template,omitempty"`
	TemplatePattern string `json:"templatePattern,omitempty" yaml:"templatePattern,omitempty"`
	TemplateOpts    string `json:"templateOpts,omitempty" yaml:"templateOpts,omitempty"`
	// kinds
	Kinds []string `json:"kinds,omitempty" yaml:"kinds,omitempty"`

	// Dir is where the resource is cloned
	Dir string
}

// GotplInflatorArgs metadata to fetch and render remote templates
type GotplInflatorArgs struct {
	// local name for remote
	Name         string                 `json:"name,omitempty" yaml:"name,omitempty"`
	Dependencies []RemoteResource       `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	Values       map[string]interface{} `json:"values,omitempty" yaml:"values,omitempty"`
}
