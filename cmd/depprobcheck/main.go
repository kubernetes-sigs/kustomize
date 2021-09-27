package main

import (
	"fmt"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kube-openapi/pkg/validation/spec"
	//"github.com/go-openapi/swag"
)

func main() {
	//hey := yaml.TypeMeta{}
	//fmt.Printf("%v\n", hey)

	//	fmt.Printf("%v\n", swag.BooleanProperty())

	genericclioptions.NewConfigFlags(true)
	fmt.Printf("%v\n", spec.Schema{})
}
