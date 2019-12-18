package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

type plugin struct {
	rmf              *resmap.Factory
	ldr              ifc.Loader
	secretReceiver   secretReceiver
	types.TypeMeta   `json:",inline"`
	types.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
}

//noinspection GoUnusedGlobalVariable
// nolint: golint
var KustomizePlugin plugin

func (p *plugin) Config(h *resmap.PluginHelpers, c []byte) error {
	p.rmf = h.ResmapFactory()
	p.ldr = h.Loader()

	var err error
	if p.secretReceiver, err = initVaultClient(); err != nil {
		return fmt.Errorf("can't initialize Vault client: %s", err)
	}

	return yaml.Unmarshal(c, p)
}

func (p *plugin) Transform(m resmap.ResMap) error {

	targetId := resid.NewResId(VaultSecretGvk, "")
	for _, res := range m.Resources() {
		if gvkEquals(targetId, resmap.GetOriginalId(res)) {
			marshaled, err := res.MarshalJSON()
			if err != nil {
				log.Printf("WARN: Can't marshal an object: %s, %s: %s", res, res.GetName(), err)
				continue
			}
			ts := VaultSecret{}
			if err := json.Unmarshal(marshaled, &ts); err != nil {
				log.Printf("WARN: Can't unmarshal a VaultSecret object: %s, %s: %s", res, res.GetName(), err)
				continue
			}

			secretResource, err := p.buildSecretResource(ts.Name, ts.Namespace, ts.Spec.Path, ts.Spec.Type)
			if err != nil {
				log.Printf("WARN: No secret has been obtained for: %s, name: %s, tenant: %s: %s", ts, ts.Name, ts.Spec.Path, err)
				continue
			}
			if err := m.Append(secretResource); err != nil {
				log.Printf("WARN: Can't add a secret resource to a resource map: %s", err)
			}
		}
	}

	return nil
}

func (p *plugin) buildSecretResource(name string, namespace string, path string, secretType string) (*resource.Resource, error) {
	var err error

	data, err := p.secretReceiver.getData(path)
	if err != nil {
		return nil, fmt.Errorf("can't receive secret: %s", err)
	}

	kubeSecret := makeFreshSecret(name, namespace, secretType)

	dockerType := false
	if corev1.SecretType(secretType) == corev1.SecretTypeDockerConfigJson || secretType == "" {
		if username, ok := data["username"].(string); ok {
			if password, ok := data["password"].(string); ok {
				if server, ok := data["server"].(string); ok {
					var email string
					if _, ok = data["email"]; ok {
						email, _ = data["email"].(string)
					}
					dockerConfigJson, err := generateDockerConfigJson(username, password, email, server)
					if err != nil {
						return nil, errors.New(fmt.Sprintf("Can't generate docker config json: %s", err))
					}
					kubeSecret.Data[".dockerconfigjson"] = dockerConfigJson
					kubeSecret.Type = corev1.SecretTypeDockerConfigJson
					dockerType = true
				}
			}
		}
	}

	if dockerType == false {
		if secretType != "" {
			kubeSecret.Type = corev1.SecretType(secretType)
		} else {
			kubeSecret.Type = corev1.SecretTypeOpaque
		}
		for key, value := range data {
			if stringValue, ok := value.(string); ok {
				kubeSecret.Data[key] = []byte(stringValue)
			}
		}
	}

	unstr, err := buildUnstructured(kubeSecret)
	if err != nil {
		return nil, fmt.Errorf("WARN: Failed to build Unstructured: %s", err)
	}

	return buildResourceFromUnstructured(unstr, p.rmf.RF()), nil
}

func gvkEquals(id resid.ResId, o resid.ResId) bool {
	return id.Gvk.Equals(o.Gvk)
}

func makeFreshSecret(name string, namespace string, secretType string) *corev1.Secret {
	s := &corev1.Secret{}
	s.APIVersion = "v1"
	s.Kind = "Secret"
	s.Name = name
	s.Namespace = namespace
	s.Type = corev1.SecretType(secretType)
	if s.Type == "" {
		s.Type = corev1.SecretTypeOpaque
	}
	s.Data = map[string][]byte{}
	return s
}

func buildUnstructured(p interface{}) (*unstructured.Unstructured, error) {
	marshaled, err := json.Marshal(p)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: Json marshaling error", err))
	}
	unstr := unstructured.Unstructured{}
	if err = unstr.UnmarshalJSON(marshaled); err != nil {
		return nil, errors.New(fmt.Sprintf("%s: Json unmarshaling error", err))
	}
	unstr.SetCreationTimestamp(metav1.Time{})
	return &unstr, nil
}

func buildResourceFromUnstructured(u *unstructured.Unstructured, rf *resource.Factory) *resource.Resource {
	k := interface{}(&kunstruct.UnstructAdapter{Unstructured: *u}).(ifc.Kunstructured)
	return rf.FromKunstructured(k)
}

func generateDockerConfigJson(username, password, email, server string) ([]byte, error) {
	dockercfgAuth := DockerConfigEntry{
		Username: username,
		Password: password,
		Email:    email,
		Auth:     encodeDockerConfigFieldAuth(username, password),
	}

	dockerCfgJSON := DockerConfigJSON{
		Auths: map[string]DockerConfigEntry{server: dockercfgAuth},
	}

	return json.Marshal(dockerCfgJSON)
}

func encodeDockerConfigFieldAuth(username, password string) string {
	fieldValue := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(fieldValue))
}

type DockerConfigJSON struct {
	Auths DockerConfig `json:"auths"`
	// +optional
	HttpHeaders map[string]string `json:"HttpHeaders,omitempty"`
}

type DockerConfig map[string]DockerConfigEntry

type DockerConfigEntry struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
	Auth     string `json:"auth,omitempty"`
}
