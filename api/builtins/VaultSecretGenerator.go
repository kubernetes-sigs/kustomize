package builtins

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

const (
	// KustomizeVaultTimeStampFormat sets the format for timestamps of last known vault connection
	KustomizeVaultTimeStampFormat = "Mon Jan _2 15:04:05 2006"
	// KustomizeVaultDotFile is the dotfile mounted in the home dir of the machine, which leaves when
	// the last time you connected to vault, etc
	KustomizeVaultDotFile = ".kustomize-vault.yml"
)

// vaultClientConnectionRetries holds how many times kustomize tries to connect to vault
const vaultClientConnectionRetries = 5

// secretClient is the client that actually holds the connection to your vault instance
// It is a global variable to prevent multiple connections
var secretClient *vault.Client

// VaultSecretGeneratorPlugin manifests vault secrets into Kubernetes configs, which is
// driven by the main kustomize cli
type VaultSecretGeneratorPlugin struct {
	h                *resmap.PluginHelpers
	types.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	types.VaultSecretArgs
}

// vaultSecretGeneratorPluginConfig is used to configure whenever the vault secret gen plugin
// is connecting with Vault itself. If your environment connects to vault a bunch of times, Vault will
// refuse you. We should limit the amount of times you reconnect, and do so with a dotfile in your root
// called $HOME/.kustomize-vault.yml
type vaultSecretGeneratorPluginConfig struct {
	Token     string `yaml:"token"`
	Timestamp string `yaml:"timestamp"`

	path string `yaml:"-"`
}

// Config will load in configuration variables that the plugin uses to get secrets, and also
// initialize the connection with the vault server
func (p *VaultSecretGeneratorPlugin) Config(h *resmap.PluginHelpers, config []byte) (err error) {
	if secretClient == nil {
		if err := p.connectToVault(); err != nil {
			return fmt.Errorf("unable_to_connect_to_vault: %s", err)
		}
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
	data, err := secretClient.Logical().Read(p.Path)
	if err != nil {
		return nil, fmt.Errorf("error_loading_secret: %s; secret_path: %s", err, p.Path)
	}
	if data == nil {
		return nil, fmt.Errorf("secret_does_not_exist; secret_path: %s", p.Path)
	}
	if _, exists := data.Data[p.SecretKey]; !exists {
		return nil, fmt.Errorf("secret_does_not_exist; secret_path: %s", p.Path)
	}

	value, ok := data.Data[p.SecretKey].(string)
	if !ok {
		return nil, fmt.Errorf("secret_type_conversion_fail; secret_path: %s", p.Path)
	}

	args.LiteralSources = append(args.LiteralSources, fmt.Sprintf("%s=%s", p.VaultSecretArgs.Name, value))

	return p.h.ResmapFactory().FromVaultSecretArgs(
		kv.NewLoader(p.h.Loader(), p.h.Validator()), args)
}

// persistConnectionIfPresent will persist the connection to vault in the environment if it's present
// and returns true if so. returns false, if no connection present, err upon unexpected environment issues
func (p *VaultSecretGeneratorPlugin) persistConnectionIfPresent() (bool, error) {
	vaultAddr, exists := os.LookupEnv(VaultAddressEnv)
	if !exists {
		return false, fmt.Errorf("vault_addr_dne")
	}

	conf, err := newVaultSecretGeneratorPluginConfig()
	if err != nil {
		return false, fmt.Errorf("vault_secret_config: %s", err)
	}
	if err := conf.read(); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("vault_secret_config_read: %s", err)
	}

	timeStamp, err := time.Parse(KustomizeVaultTimeStampFormat, conf.Timestamp)
	if err != nil {
		return false, fmt.Errorf("vault_time_stamp_not_parseable: %s", err)
	}
	if time.Now().UTC().Add(time.Hour * -1).After(timeStamp) {
		return false, nil
	}

	secretClient, err = vault.NewClient(&vault.Config{
		Address:    vaultAddr,
		HttpClient: &http.Client{Timeout: 10 * time.Second},
	})
	if err != nil {
		return false, fmt.Errorf("error_creating_vault_client: %s", err)
	}

	secretClient.SetToken(conf.Token)
	return true, nil
}

func (p *VaultSecretGeneratorPlugin) connectToVault() (err error) {
	connected, err := p.persistConnectionIfPresent()
	if err != nil {
		return fmt.Errorf("persisted_vault_conn_fail: %s", err)
	}
	if connected {
		return nil
	}

	for try := 0; try < vaultClientConnectionRetries; try++ {
		err = p.connect()
		if err != nil {
			log.Printf("failed_vault_connection: %s  ...retrying...", err)
			time.Sleep(time.Second * 2)
			continue
		}
		conf, err := newVaultSecretGeneratorPluginConfig()
		if err != nil {
			return fmt.Errorf("vault_gen_conf_err: %s", err)
		}

		conf.Token = secretClient.Token()
		conf.Timestamp = time.Now().UTC().Format(KustomizeVaultTimeStampFormat)
		if err := conf.write(); err != nil {
			return fmt.Errorf("conf_write: %s", err)
		}
		return nil
	}
	return fmt.Errorf("failed_vault_connection_after_%d_retries: %s", vaultClientConnectionRetries, err)
}

// connect runs the actual process to connect to hashicorp vault
func (p *VaultSecretGeneratorPlugin) connect() error {
	vaultAddr, exists := os.LookupEnv(VaultAddressEnv)
	if !exists {
		return errors.New("vault_address_environment_variable_not_set: VAULT_ADDR")
	}
	vaultUsername, exists := os.LookupEnv(VaultUsernameEnv)
	if !exists {
		return errors.New("vault_username_environment_variable_not_set: VAULT_USERNAME")
	}
	vaultPassword, exists := os.LookupEnv(VaultPasswordEnv)
	if !exists {
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

	secretClient = client
	return nil
}

func (p *vaultSecretGeneratorPluginConfig) write() error {
	raw, err := yaml.Marshal(p)
	if err != nil {
		return fmt.Errorf("yaml_marshal: %s", err)
	}
	if err := ioutil.WriteFile(p.path, raw, 0744); err != nil {
		return fmt.Errorf("vault_yaml_write: %s", err)
	}
	return nil
}

func (p *vaultSecretGeneratorPluginConfig) read() error {
	raw, err := ioutil.ReadFile(p.path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("conf_read_file: %s", err)
	}
	if errors.Is(err, os.ErrNotExist) {
		return os.ErrNotExist
	}

	if err := yaml.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("file_unmarshal_err: %s", err)
	}
	return nil
}

// NewVaultSecretGeneratorPlugin creates a generator for vault secrets
func NewVaultSecretGeneratorPlugin() resmap.GeneratorPlugin {
	return &VaultSecretGeneratorPlugin{}
}

func newVaultSecretGeneratorPluginConfig() (*vaultSecretGeneratorPluginConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home_dir: %s", err)
	}

	path := filepath.Join(homeDir, KustomizeVaultDotFile)
	return &vaultSecretGeneratorPluginConfig{path: path}, nil
}
