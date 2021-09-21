package types

// VaultSecretArgs parses parameters needed to generate vault secret volume mounts
type VaultSecretArgs struct {
	GeneratorArgs `json:",inline,omitempty" yaml:",inline,omitempty"`
	Path          string `json:"path,omitempty" yaml:"path,omitempty"`
	SecretKey     string `json:"secretKey,omitempty" yaml:"secretKey,omitempty"`
}
