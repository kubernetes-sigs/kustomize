package oci

import (
	"fmt"
	"strings"

	fluxOCI "github.com/fluxcd/pkg/oci"
	"github.com/fluxcd/source-controller/api/v1beta2"
)

var supportedSourceOCIProviders = []string{
	v1beta2.GenericOCIProvider,
	v1beta2.AmazonOCIProvider,
	v1beta2.AzureOCIProvider,
	v1beta2.GoogleOCIProvider,
}

var sourceOCIProvidersToOCIProvider = map[string]fluxOCI.Provider{
	v1beta2.GenericOCIProvider: fluxOCI.ProviderGeneric,
	v1beta2.AmazonOCIProvider:  fluxOCI.ProviderAWS,
	v1beta2.AzureOCIProvider:   fluxOCI.ProviderAzure,
	v1beta2.GoogleOCIProvider:  fluxOCI.ProviderGCP,
}

type SourceOCIProvider string

func (p *SourceOCIProvider) String() string {
	return string(*p)
}

func (p *SourceOCIProvider) Set(str string) error {
	if strings.TrimSpace(str) == "" {
		return fmt.Errorf("no source OCI provider given, please specify %s",
			p.Description())
	}
	if !containsItemString(supportedSourceOCIProviders, str) {
		return fmt.Errorf("source OCI provider '%s' is not supported, must be one of: %v",
			str, strings.Join(supportedSourceOCIProviders, ", "))
	}
	*p = SourceOCIProvider(str)
	return nil
}

func (p *SourceOCIProvider) Type() string {
	return "sourceOCIProvider"
}

func (p *SourceOCIProvider) Description() string {
	return fmt.Sprintf(
		"the OCI provider name, available options are: (%s)",
		strings.Join(supportedSourceOCIProviders, ", "),
	)
}

func (p *SourceOCIProvider) ToOCIProvider() (fluxOCI.Provider, error) {
	value, ok := sourceOCIProvidersToOCIProvider[p.String()]
	if !ok {
		return 0, fmt.Errorf("no mapping between source OCI provider %s and OCI provider", p.String())
	}

	return value, nil
}

func containsItemString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
