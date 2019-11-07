package types

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Compile time check that the forked version of the libraries are being used
var _ = []bool{
	corev1.APIKustomizeFork,
	metav1.ApiMachineryKustomizeFork,
	kubernetes.ClientGoKustomizeFork,
}
