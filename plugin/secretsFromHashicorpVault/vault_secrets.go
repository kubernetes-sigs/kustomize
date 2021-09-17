package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"sigs.k8s.io/kustomize/api/kv"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"

	vault "github.com/hashicorp/vault/api"
)

// const variables used for environment variable assignment
// NOTE: Without these variables set, this plugin WILL NOT WORK.
// Read more here: https://github.com/benmorehouse/kustomize/plugin/secretsFromHashicorpVault/README.md
const (
	// VaultAddressEnv configures the vault cluster that we are targeting
	// for secret management and access
	VaultAddressEnv = "VAULT_ADDR"
	// VaultUsernameEnv configures the vault username
	VaultUsernameEnv = "VAULT_USERNAME"
	// VaultPasswordEnv configures the vault passwork
	VaultPasswordEnv = "VAULT_PASSWORD"
)

// Plugin manifests vault secrets into Kubernetes configs, which is
// driven by the main kustomize cli
type vaultPlugin struct {
	h                *resmap.PluginHelpers
	types.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Secrets          []Secret `json:"secrets,omitempty" yaml:"secrets,omitempty"`

	secretClient *vault.Client
}

// Secret will configure and hold a secret from vault
type Secret struct {
	Name    string `json:"name,omitempty" yaml:"name,omitempty"`
	Path    string `json:"path,omitempty" yaml:"path,omitempty"`
	Version int    `json:"version,omitempty" yaml:"version,omitempty"`

	value []byte
}

// KustomizePlugin will ensure compilation and linting standards are met
var KustomizePlugin vaultPlugin

// Config will load in configuration variables that the plugin uses to get secrets, and also
// initialize the connection with the vault server
func (p *vaultPlugin) Config(h *resmap.PluginHelpers, c []byte) error {
	if err := p.connectToVault(); err != nil {
		return fmt.Errorf("error_connecting_to_vault: %s", err)
	}

	p.h = h
	return yaml.Unmarshal(c, p)
}

/*

// GeneratorArgs contains arguments common to ConfigMap and Secret generators.
type GeneratorArgs struct {
	// Namespace for the configmap, optional
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`

	// Name - actually the partial name - of the generated resource.
	// The full name ends up being something like
	// NamePrefix + this.Name + hash(content of generated resource).
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Behavior of generated resource, must be one of:
	//   'create': create a new one
	//   'replace': replace the existing one
	//   'merge': merge with the existing one
	Behavior string `json:"behavior,omitempty" yaml:"behavior,omitempty"`

	// KvPairSources for the generator.
	KvPairSources `json:",inline,omitempty" yaml:",inline,omitempty"`

	// Local overrides to global generatorOptions field.
	Options *GeneratorOptions `json:"options,omitempty" yaml:"options,omitempty"`
}

// KvPairSources defines places to obtain key value pairs.
type KvPairSources struct {
	// LiteralSources is a list of literal
	// pair sources. Each literal source should
	// be a key and literal value, e.g. `key=value`
	LiteralSources []string `json:"literals,omitempty" yaml:"literals,omitempty"`

	// FileSources is a list of file "sources" to
	// use in creating a list of key, value pairs.
	// A source takes the form:  [{key}=]{path}
	// If the "key=" part is missing, the key is the
	// path's basename. If they "key=" part is present,
	// it becomes the key (replacing the basename).
	// In either case, the value is the file contents.
	// Specifying a directory will iterate each named
	// file in the directory whose basename is a
	// valid configmap key.
	FileSources []string `json:"files,omitempty" yaml:"files,omitempty"`

	// EnvSources is a list of file paths.
	// The contents of each file should be one
	// key=value pair per line, e.g. a Docker
	// or npm ".env" file or a ".ini" file
	// (wikipedia.org/wiki/INI_file)
	EnvSources []string `json:"envs,omitempty" yaml:"envs,omitempty"`

	// Older, singular form of EnvSources.
	// On edits (e.g. `kustomize fix`) this is merged into the plural form
	// for consistency with LiteralSources and FileSources.
	EnvSource string `json:"env,omitempty" yaml:"env,omitempty"`


*/

// Generate configMaps that will push the secret pulled from vault into a Volume
func (p *vaultPlugin) Generate() (resmap.ResMap, error) {
	args := types.ConfigMapArgs{}
	args.Namespace = "default"
	args.Name = "VaultGeneratedSecrets"
	args.Behavior = "create"

	for _, secret := range p.Secrets {
		// first we need to read the secret from Vault
		data, err := p.secretClient.Logical().Read(secret.Path)
		if err != nil {
			return nil, fmt.Errorf("error_loading_secret: %s; secret_path: %s", err, secret.Path)
		}

		rawValue, err := json.Marshal(data.Data)
		if err != nil {
			return nil, fmt.Errorf("error_marshalling_secret: %s; secret_path: %s", err, secret.Path)
		}
		args.LiteralSources = append(args.LiteralSources, fmt.Sprintf("%s=%s", secret.Name, string(rawValue)))
	}

	return p.h.ResmapFactory().FromConfigMapArgs(
		kv.NewLoader(p.h.Loader(), p.h.Validator()), args)
}

func (p *vaultPlugin) connectToVault() error {
	vaultAddr := os.Getenv(VaultAddressEnv)
	if vaultAddr == "" {
		return errors.New("vault_address_environment_variable_not_set: VAULT_ADDR")
	}
	vaultUsername := os.Getenv(VaultUsernameEnv)
	if vaultUsername == "" {
		return errors.New("vault_username_environment_variable_not_set: VAULT_USERNAME")
	}
	vaultPassword := os.Getenv(VaultPasswordEnv)
	if vaultPassword == "" {
		return errors.New("vault_password_environment_variable_not_set: VAULT_PASSWORD")
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	client, err := vault.NewClient(&vault.Config{Address: vaultAddr, HttpClient: httpClient})
	if err != nil {
		return fmt.Errorf("error_creating_vault_client: %s", err)
	}

	options := map[string]interface{}{
		"password": vaultPassword,
	}
	path := fmt.Sprintf("auth/userpass/login/%s", vaultUsername)

	// PUT call to get a vault token
	secret, err := client.Logical().Write(path, options)
	if err != nil {
		return fmt.Errorf("error_authenticating_with_vault_cluster: %s", err)
	}
	client.SetToken(secret.Auth.ClientToken)

	p.secretClient = client
	return nil
}
