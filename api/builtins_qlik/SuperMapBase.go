package builtins_qlik

import (
	"encoding/base64"
	"fmt"
	"log"

	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/internal/accumulator"
	"sigs.k8s.io/kustomize/api/internal/plugins/builtinconfig"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/transform"
	"sigs.k8s.io/kustomize/api/types"
)

type IDecorator interface {
	GetLogger() *log.Logger
	GetName() string
	GetNamespace() string
	SetNamespace(namespace string)
	GetType() string
	GetConfigData() map[string]string
	ShouldBase64EncodeConfigData() bool
	GetDisableNameSuffixHash() bool
	Generate() (resmap.ResMap, error)
}

type SuperMapPluginBase struct {
	AssumeTargetWillExist bool   `json:"assumeTargetWillExist,omitempty" yaml:"assumeTargetWillExist,omitempty"`
	Prefix                string `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	Rf                    *resmap.Factory
	Hasher                ifc.KunstructuredHasher
	Decorator             IDecorator
	Configurations        []string `json:"configurations,omitempty" yaml:"configurations,omitempty"`
	tConfig               *builtinconfig.TransformerConfig
}

func NewBase(rf *resmap.Factory, decorator IDecorator) SuperMapPluginBase {
	return SuperMapPluginBase{
		AssumeTargetWillExist: true,
		Prefix:                "",
		Rf:                    rf,
		Decorator:             decorator,
		Hasher:                rf.RF().Hasher(),
		Configurations:        make([]string, 0),
		tConfig:               nil,
	}
}

func (b *SuperMapPluginBase) SetupTransformerConfig(ldr ifc.Loader) error {
	b.tConfig = &builtinconfig.TransformerConfig{}
	tCustomConfig, err := builtinconfig.MakeTransformerConfig(ldr, b.Configurations)
	if err != nil {
		b.Decorator.GetLogger().Printf("error making transformer config, error: %v\n", err)
		return err
	}
	b.tConfig, err = b.tConfig.Merge(tCustomConfig)
	if err != nil {
		b.Decorator.GetLogger().Printf("error merging transformer config, error: %v\n", err)
		return err
	}
	return nil
}

func (b *SuperMapPluginBase) Transform(m resmap.ResMap) error {
	resource := b.find(b.Decorator.GetName(), b.Decorator.GetType(), m)
	if resource != nil {
		return b.executeBasicTransform(resource, m)
	} else if b.AssumeTargetWillExist && !b.Decorator.GetDisableNameSuffixHash() {
		return b.executeAssumeWillExistTransform(m)
	} else {
		b.Decorator.GetLogger().Printf("NOT executing anything because resource: %v is NOT in the input stream and AssumeTargetWillExist: %v, disableNameSuffixHash: %v\n", b.Decorator.GetName(), b.AssumeTargetWillExist, b.Decorator.GetDisableNameSuffixHash())
	}
	return nil
}

func (b *SuperMapPluginBase) executeAssumeWillExistTransform(m resmap.ResMap) error {
	b.Decorator.GetLogger().Printf("executeAssumeWillExistTransform() for imaginary resource: %v\n", b.Decorator.GetName())

	if b.Decorator.GetNamespace() == "" {
		if anyExistingResource := m.GetByIndex(0); anyExistingResource != nil && anyExistingResource.GetNamespace() != "" {
			b.Decorator.SetNamespace(anyExistingResource.GetNamespace())
		}
	}

	generateResourceMap, err := b.Decorator.Generate()
	if err != nil {
		b.Decorator.GetLogger().Printf("error generating temp resource: %v, error: %v\n", b.Decorator.GetName(), err)
		return err
	}
	tempResource := b.find(b.Decorator.GetName(), b.Decorator.GetType(), generateResourceMap)
	if tempResource == nil {
		err := fmt.Errorf("error locating generated temp resource: %v", b.Decorator.GetName())
		b.Decorator.GetLogger().Printf("%v\n", err)
		return err
	}

	err = m.Append(tempResource)
	if err != nil {
		b.Decorator.GetLogger().Printf("error appending temp resource: %v to the resource map, error: %v\n", b.Decorator.GetName(), err)
		return err
	}

	resourceName := b.Decorator.GetName()
	if len(b.Prefix) > 0 {
		resourceName = fmt.Sprintf("%s%s", b.Prefix, resourceName)
	}
	tempResource.SetName(resourceName)

	nameWithHash, err := b.generateNameWithHash(tempResource)
	if err != nil {
		b.Decorator.GetLogger().Printf("error hashing resource: %v, error: %v\n", resourceName, err)
		return err
	}
	tempResource.SetName(nameWithHash)

	err = b.executeNameReferencesTransformer(m)
	if err != nil {
		b.Decorator.GetLogger().Printf("error executing nameReferenceTransformer.Transform(): %v\n", err)
		return err
	}
	err = m.Remove(tempResource.CurId())
	if err != nil {
		b.Decorator.GetLogger().Printf("error removing temp resource: %v from the resource map, error: %v\n", b.Decorator.GetName(), err)
		return err
	}
	return nil
}

func (b *SuperMapPluginBase) executeBasicTransform(resource *resource.Resource, m resmap.ResMap) error {
	b.Decorator.GetLogger().Printf("executeBasicTransform() for resource: %v...\n", resource)

	if err := b.appendData(resource, b.Decorator.GetConfigData(), false); err != nil {
		b.Decorator.GetLogger().Printf("error appending data to resource: %v, error: %v\n", b.Decorator.GetName(), err)
		return err
	}

	if !b.Decorator.GetDisableNameSuffixHash() {
		if err := m.Remove(resource.CurId()); err != nil {
			b.Decorator.GetLogger().Printf("error removing original resource on name change: %v\n", err)
			return err
		}
		newResource := b.Rf.RF().FromMapAndOption(resource.Map(), &types.GeneratorArgs{Behavior: "replace"})
		if err := m.Append(newResource); err != nil {
			b.Decorator.GetLogger().Printf("error re-adding resource on name change: %v\n", err)
			return err
		}
		b.Decorator.GetLogger().Printf("resource should have hashing enabled: %v\n", newResource)
	}
	return nil
}

func (b *SuperMapPluginBase) executeNameReferencesTransformer(m resmap.ResMap) error {
	ac := accumulator.MakeEmptyAccumulator()
	if err := ac.AppendAll(m); err != nil {
		return err
	} else if err := ac.MergeConfig(b.tConfig); err != nil {
		return err
	} else if err := ac.FixBackReferences(); err != nil {
		return err
	}
	return nil
}

func (b *SuperMapPluginBase) find(name string, resourceType string, m resmap.ResMap) *resource.Resource {
	for _, res := range m.Resources() {
		if res.GetKind() == resourceType && res.GetName() == b.Decorator.GetName() {
			return res
		}
	}
	return nil
}

func (b *SuperMapPluginBase) generateNameWithHash(res *resource.Resource) (string, error) {
	hash, err := b.Hasher.Hash(res)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s", res.GetName(), hash), nil
}

func (b *SuperMapPluginBase) appendData(res *resource.Resource, data map[string]string, straightCopy bool) error {
	for k, v := range data {
		pathToField := []string{"data", k}
		err := transform.MutateField(
			res.Map(),
			pathToField,
			true,
			func(interface{}) (interface{}, error) {
				var val string
				if !straightCopy && b.Decorator.ShouldBase64EncodeConfigData() {
					val = base64.StdEncoding.EncodeToString([]byte(v))
				} else {
					val = v
				}
				return val, nil
			})
		if err != nil {
			b.Decorator.GetLogger().Printf("error executing MutateField for resource: %v, pathToField: %v, error: %v\n", b.Decorator.GetName(), pathToField, err)
			return err
		}
	}
	return nil
}
