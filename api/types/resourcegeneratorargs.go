package types

// ResourceGeneratorArgs configures the resource paths to accumulate.
type ResourceGeneratorArgs struct {
	// List of resource files to accumulate.
	Files             []string `json:"files,omitempty" yaml:"files,omitempty"`
	OriginAnnotations bool     `json:"originAnnotations" yaml:"originAnnotations,omitempty"`
}
