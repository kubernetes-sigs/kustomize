package wait

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func resourceIdentifierFromObject(object KubernetesObject) ResourceIdentifier {
	return ResourceIdentifier{
		Name:      object.GetName(),
		Namespace: object.GetNamespace(),
		GroupKind: object.GroupVersionKind().GroupKind(),
	}
}

func resourceIdentifiersFromObjects(objects []KubernetesObject) []ResourceIdentifier {
	var resourceIdentifiers []ResourceIdentifier
	for _, object := range objects {
		resourceIdentifiers = append(resourceIdentifiers, resourceIdentifierFromObject(object))
	}
	return resourceIdentifiers
}

func resourceIdentifierFromRuntimeObject(object runtime.Object) ResourceIdentifier {
	gvk := object.GetObjectKind().GroupVersionKind()
	r := object.(metav1.Object)
	return ResourceIdentifier{
		GroupKind: gvk.GroupKind(),
		Name:      r.GetName(),
		Namespace: r.GetNamespace(),
	}
}
