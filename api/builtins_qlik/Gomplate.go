package builtins_qlik

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"sigs.k8s.io/kustomize/api/builtins_qlik/utils"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/yaml"
)

type GomplatePlugin struct {
	DataSource map[string]interface{} `json:"dataSource,omitempty" yaml:"dataSource,omitempty"`
	Pwd        string
	ldr        ifc.Loader
	rf         *resmap.Factory
	logger     *log.Logger
}

func (p *GomplatePlugin) Config(h *resmap.PluginHelpers, c []byte) (err error) {
	p.ldr = h.Loader()
	p.rf = h.ResmapFactory()
	p.Pwd = h.Loader().Root()
	return yaml.Unmarshal(c, p)
}

func (p *GomplatePlugin) Transform(m resmap.ResMap) error {
	var env []string
	var vaultAddressPath, vaultTokenPath string
	var vaultAddress, vaultToken string
	if p.DataSource["vault"] != nil {
		vaultAddressPath = fmt.Sprintf("%s", p.DataSource["vault"].(map[string]interface{})["addressPath"])
		vaultTokenPath = fmt.Sprintf("%s", p.DataSource["vault"].(map[string]interface{})["tokenPath"])

		if _, err := os.Stat(vaultAddressPath); os.IsNotExist(err) {
			readBytes, err := ioutil.ReadFile(vaultAddressPath)
			if err != nil {
				p.logger.Printf("error reading vault address file: %v, error: %v\n", vaultAddressPath, err)
				return err
			}
			vaultAddress = fmt.Sprintf("VAULT_ADDR=%s", string(readBytes))
			env = append(env, vaultAddress)
		} else if err != nil {
			p.logger.Printf("error executing stat on vault address file: %v, error: %v\n", vaultAddressPath, err)
		}

		if _, err := os.Stat(vaultTokenPath); os.IsNotExist(err) {
			readBytes, err := ioutil.ReadFile(vaultTokenPath)
			if err != nil {
				p.logger.Printf("error reading vault token file: %v, error: %v\n", vaultTokenPath, err)
				return err
			}
			vaultToken = fmt.Sprintf("VAULT_TOKEN=%s", string(readBytes))
			env = append(env, vaultToken)
		} else if err != nil {
			p.logger.Printf("error executing stat on vault token file: %v, error: %v\n", vaultTokenPath, err)
		}
	}

	var ejsonKey string
	if p.DataSource["ejson"] != nil {
		if _, isSet := p.DataSource["ejson"].(map[string]interface{})["privateKeyPath"]; isSet {
			if ejsonPrivateKeyPath, ok := p.DataSource["ejson"].(map[string]interface{})["privateKeyPath"].(string); !ok {
				err := errors.New("privateKeyPath must be a string")
				p.logger.Printf("error: %v\n", err)
				return err
			} else if _, err := os.Stat(ejsonPrivateKeyPath); err != nil {
				p.logger.Printf("error executing stat on ejson private key file: %v, error: %v\n", ejsonPrivateKeyPath, err)
			} else {
				readBytes, err := ioutil.ReadFile(ejsonPrivateKeyPath)
				if err != nil {
					p.logger.Printf("error reading ejson private key file: %v, error: %v\n", ejsonPrivateKeyPath, err)
					return err
				}
				ejsonKey = fmt.Sprintf("EJSON_KEY=%s", string(readBytes))
				env = append(env, ejsonKey)
			}
		}
	}
	if os.Getenv("EJSON_KEY") != "" && ejsonKey == "" {
		ejsonKey = fmt.Sprintf("EJSON_KEY=%s", os.Getenv("EJSON_KEY"))
		env = append(env, ejsonKey)
	}

	var dataSource string
	if p.DataSource["ejson"] != nil {
		dataSource = fmt.Sprintf("%s", p.DataSource["ejson"].(map[string]interface{})["filePath"])
	} else if vaultAddress != "" && vaultToken != "" {
		dataSource = fmt.Sprintf("%s", p.DataSource["vault"].(map[string]interface{})["secretPath"])
	} else if p.DataSource["file"] != nil {
		dataSource = fmt.Sprintf("%s", p.DataSource["file"].(map[string]interface{})["path"])
	} else {
		dataSource = fmt.Sprintf("%s", p.DataSource["file"].(map[string]interface{})["path"])
	}

	for _, r := range m.Resources() {
		yamlByte, err := r.AsYAML()
		if err != nil {
			p.logger.Printf("error getting resource as yaml: %v, error: %v\n", r.GetName(), err)
			return err
		}

		if output, err := utils.RunGomplate(dataSource, p.Pwd, env, string(yamlByte), p.logger); err != nil {
			p.logger.Printf("error executing runGomplate() on dataSource: %v, in directory: %v, error: %v\n", dataSource, p.Pwd, err)
		} else {
			res, err := p.rf.RF().FromBytes(output)
			if err != nil {
				p.logger.Printf("error unmarshalling resource from bytes: %v\n", err)
				return err
			}
			r.SetMap(res.Map())
		}
	}
	return nil
}

func NewGomplatePlugin() resmap.TransformerPlugin {
	return &GomplatePlugin{logger: utils.GetLogger("GomplatePlugin")}
}
