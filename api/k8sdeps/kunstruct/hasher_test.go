package kunstruct

import (
	"testing"
)

func TestHasher(t *testing.T) {
	input := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: foo
data:
  one: ""
binaryData:
  two: ""
`
	expect := "698h7c7t9m"

	factory := NewKunstructuredFactoryImpl()
	k, err := factory.SliceFromBytes([]byte(input))
	if err != nil {
		t.Fatal(err)
	}

	hasher := NewKustHash()
	result, err := hasher.Hash(k[0])
	if err != nil {
		t.Fatal(err)
	}
	if result != expect {
		t.Fatalf("expect %s but got %s", expect, result)
	}
}
