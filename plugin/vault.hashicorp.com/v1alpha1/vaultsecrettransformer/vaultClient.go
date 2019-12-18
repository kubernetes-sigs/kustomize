package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/vault/api"
)

type vaultClient struct {
	vclient            *api.Client
	logical            *api.Logical
}

type secretReceiver interface {
	getData(path string) (map[string]interface{}, error)
}

func initVaultClient() (secretReceiver, error) {
	httpClient := new(http.Client)
	httpClient.Timeout = 5 * time.Second

	var vaultURL string
	if os.Getenv("VAULT_ADDR") != "" {
		vaultURL = os.Getenv("VAULT_ADDR")
	}

	vclient, err := api.NewClient(&api.Config{Address: vaultURL, HttpClient: httpClient})
	if err != nil {
		return nil, fmt.Errorf("ERROR: Unable to create Vault vaultClient: %s", err)

	}

	logical := vclient.Logical()

	vaultClient := &vaultClient{
		vclient: vclient,
		logical: logical,
	}

	var vaultToken string
	if os.Getenv("VAULT_TOKEN") != "" {
		vaultToken = os.Getenv("VAULT_TOKEN")
	}
	vaultClient.vclient.SetToken(vaultToken)

	return vaultClient, err
}

func (c *vaultClient) getData(path string) (map[string]interface{}, error) {
	secret, err := c.logical.Read(path)
	if secret == nil {
		return nil, fmt.Errorf("secret not found: %s", path)
	}
	if err != nil {
		return nil, fmt.Errorf("read secret error: %s, %s", path, err)
	}

	switch data := secret.Data["data"].(type) {
	case map[string]interface{}:
		return data, nil
	}

	return nil, fmt.Errorf("secret doesn't have data: %s", path)
}
