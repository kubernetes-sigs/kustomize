package k8sdeps

import (
	"encoding/json"
	"fmt"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"strings"
)

var _ ifc.FunStruct = &KustFunStruct{}

type KustFunStruct struct {
	unstructured.Unstructured
}

func (fs *KustFunStruct) GetGvk() gvk.Gvk {
	return gvk.FromSchemaGvk(fs.GroupVersionKind())
}

func (fs *KustFunStruct) Copy() ifc.FunStruct {
	return &KustFunStruct{*fs.DeepCopy()}
}

// NewKustFunStruct returns a new instance of KustFunStruct.
func NewKustFunStructFromObject(obj runtime.Object) (*KustFunStruct, error) {
	// Convert obj to a byte stream, then convert that to JSON (Unstructured).
	marshaled, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	var u unstructured.Unstructured
	err = u.UnmarshalJSON(marshaled)
	// creationTimestamp always 'null', remove it
	u.SetCreationTimestamp(metav1.Time{})
	return &KustFunStruct{Unstructured: u}, err
}

// NewKustFunStruct returns a new instance of KustFunStruct.
func NewKustFunStructFromMap(m map[string]interface{}) *KustFunStruct {
	return NewKustFunStructFromUnstruct(unstructured.Unstructured{Object: m})
}

// NewKustFunStructFromUnstruct returns a new instance of KustFunStruct.
func NewKustFunStructFromUnstruct(u unstructured.Unstructured) *KustFunStruct {
	return &KustFunStruct{Unstructured: u}
}

// Map returns the unstructured content map.
func (fs *KustFunStruct) Map() map[string]interface{} {
	return fs.Object
}

// SetMap overrides the unstructured content map.
func (fs *KustFunStruct) SetMap(m map[string]interface{}) {
	fs.Object = m
}

// GetFieldValue returns value at the given fieldpath.
func (fs *KustFunStruct) GetFieldValue(fieldPath string) (string, error) {
	return getFieldValue(fs.UnstructuredContent(), strings.Split(fieldPath, "."))
}

func getFieldValue(m map[string]interface{}, pathToField []string) (string, error) {
	if len(pathToField) == 0 {
		return "", fmt.Errorf("field not found")
	}
	if len(pathToField) == 1 {
		if v, found := m[pathToField[0]]; found {
			if s, ok := v.(string); ok {
				return s, nil
			}
			return "", fmt.Errorf("value at fieldpath is not of string type")
		}
		return "", fmt.Errorf("field at given fieldpath does not exist")
	}
	v := m[pathToField[0]]
	switch typedV := v.(type) {
	case map[string]interface{}:
		return getFieldValue(typedV, pathToField[1:])
	default:
		return "", fmt.Errorf("%#v is not expected to be a primitive type", typedV)
	}
}

// NewFunStructSliceFromBytes unmarshalls bytes into a FunStruct slice.
func NewFunStructSliceFromBytes(
	in []byte, decoder ifc.Decoder) ([]ifc.FunStruct, error) {
	decoder.SetInput(in)
	var result []ifc.FunStruct
	var err error
	for err == nil || isEmptyYamlError(err) {
		var out unstructured.Unstructured
		err = decoder.Decode(&out)
		if err == nil {
			result = append(result, &KustFunStruct{Unstructured: out})
		}
	}
	if err != io.EOF {
		return nil, err
	}
	return result, nil
}

func isEmptyYamlError(err error) bool {
	return strings.Contains(err.Error(), "is missing in 'null'")
}
