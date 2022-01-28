package framework_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// This function generates Graphana configuration in the form of ConfigMap. It
// accepts Revision and ID as input.

func Example_generator() {
	if err := command.AsMain(framework.ResourceListProcessorFunc(generate)); err != nil {
		os.Exit(1)
	}
}

// generate generates a ConfigMap.
func generate(rl *framework.ResourceList) error {
	if rl.FunctionConfig == nil {
		return framework.ErrMissingFnConfig{}
	}

	revision, foundRevision, err := rl.FunctionConfig.GetNestedString("data", "revision")
	if err != nil {
		return fmt.Errorf("failed to find field revision: %w", err)
	}
	id, foundId, err := rl.FunctionConfig.GetNestedString("data", "id")
	if err != nil {
		return fmt.Errorf("failed to find field id: %w", err)
	}
	if !foundRevision || !foundId {
		return nil
	}
	js, err := fetchDashboard(revision, id)
	if err != nil {
		return fmt.Errorf("fetch dashboard: %v", err)
	}

	// corev1.ConfigMap should be used here. But we can't use it here due to dependency restriction in the kustomize repo.
	cm := ConfigMap{
		ResourceMeta: yaml.ResourceMeta{
			TypeMeta: yaml.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			ObjectMeta: yaml.ObjectMeta{
				NameMeta: yaml.NameMeta{
					Name:      fmt.Sprintf("%v-gen", rl.FunctionConfig.GetName()),
					Namespace: rl.FunctionConfig.GetNamespace(),
				},
				Labels: map[string]string{
					"grafana_dashboard": "true",
				},
			},
		},
		Data: map[string]string{
			fmt.Sprintf("%v.json", rl.FunctionConfig.GetName()): fmt.Sprintf("%q", js),
		},
	}
	return rl.UpsertObjectToItems(cm, nil, false)
}

func fetchDashboard(revision, id string) (string, error) {
	url := fmt.Sprintf("https://grafana.com/api/dashboards/%s/revisions/%s/download", id, revision)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ConfigMap is a copy of corev1.ConfigMap.
type ConfigMap struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`
	Data              map[string]string `json:"data,omitempty" yaml:"data,omitempty"`
}
