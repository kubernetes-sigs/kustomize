package wait

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// keyFromResourceIdentifier creates a resourceKey from a ResourceIdentifier.
func keyFromResourceIdentifier(i ResourceIdentifier) resourceKey {
	return resourceKey{
		apiVersion: i.GetAPIVersion(),
		kind:       i.GetKind(),
		name:       i.GetName(),
		namespace:  i.GetNamespace(),
	}
}

// keyFromObject creates a resourceKey from an Object.
func keyFromObject(obj runtime.Object) resourceKey {
	gvk := obj.GetObjectKind().GroupVersionKind()
	r := obj.(metav1.Object)
	return resourceKey{
		apiVersion: gvk.GroupVersion().String(),
		kind:       gvk.Kind,
		name:       r.GetName(),
		namespace:  r.GetNamespace(),
	}
}
