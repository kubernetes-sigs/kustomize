package builtins

import (
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

// VaultSecretGeneratorPlugin manifests vault secrets into Kubernetes configs, which is
// driven by the main kustomize cli
type VaultSecretGeneratorPlugin struct {
	h                *resmap.PluginHelpers
	types.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	types.VaultSecretArgs

	secretClient *vault.Client
}

// Config will load in configuration variables that the plugin uses to get secrets, and also
// initialize the connection with the vault server
func (p *VaultSecretGeneratorPlugin) Config(h *resmap.PluginHelpers, config []byte) (err error) {
	if err := p.connectToVault(); err != nil {
		return fmt.Errorf("error_connecting_to_vault: %s", err)
	}
	p.VaultSecretArgs = types.VaultSecretArgs{}
	if err = yaml.Unmarshal(config, p); err != nil {
		return fmt.Errorf("error_unmarshal_kustomization_config: %s", err)
	}
	if p.VaultSecretArgs.Name == "" {
		p.VaultSecretArgs.Name = p.Name
	}
	if p.VaultSecretArgs.Namespace == "" {
		p.VaultSecretArgs.Namespace = p.Namespace
	}
	p.h = h
	return
}

// Generate configMaps that will push the secret pulled from vault into a Volume
func (p *VaultSecretGeneratorPlugin) Generate() (resmap.ResMap, error) {
	args := types.VaultSecretArgs{}
	args.Namespace = "default"
	args.Name = p.VaultSecretArgs.Name
	args.Behavior = "create"
	args.Path = p.Path

	// first we need to read the secret from Vault
	data, err := p.secretClient.Logical().Read(p.Path)
	if err != nil {
		return nil, fmt.Errorf("error_loading_secret: %s; secret_path: %s", err, p.Path)
	}

	value, ok := data.Data[p.SecretKey].(string)
	if !ok {
		return nil, fmt.Errorf("secret_type_conversion_fail; secret_path: %s", p.Path)
	}

	args.LiteralSources = append(args.LiteralSources, fmt.Sprintf("%s=%s", p.VaultSecretArgs.Name, value))

	return p.h.ResmapFactory().FromVaultSecretArgs(
		kv.NewLoader(p.h.Loader(), p.h.Validator()), args)
}

func (p *VaultSecretGeneratorPlugin) connectToVault() error {
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

// NewVaultSecretGeneratorPlugin creates a generator for vault secrets
func NewVaultSecretGeneratorPlugin() resmap.GeneratorPlugin {
	return &VaultSecretGeneratorPlugin{}
}
