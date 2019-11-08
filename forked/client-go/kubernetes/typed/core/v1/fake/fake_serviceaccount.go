/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	corev1 "sigs.k8s.io/kustomize/forked/api/core/v1"
	v1 "sigs.k8s.io/kustomize/forked/apimachinery/pkg/apis/meta/v1"
	labels "sigs.k8s.io/kustomize/forked/apimachinery/pkg/labels"
	schema "sigs.k8s.io/kustomize/forked/apimachinery/pkg/runtime/schema"
	types "sigs.k8s.io/kustomize/forked/apimachinery/pkg/types"
	watch "sigs.k8s.io/kustomize/forked/apimachinery/pkg/watch"
	testing "sigs.k8s.io/kustomize/forked/client-go/testing"
)

// FakeServiceAccounts implements ServiceAccountInterface
type FakeServiceAccounts struct {
	Fake *FakeCoreV1
	ns   string
}

var serviceaccountsResource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "serviceaccounts"}

var serviceaccountsKind = schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ServiceAccount"}

// Get takes name of the serviceAccount, and returns the corresponding serviceAccount object, and an error if there is any.
func (c *FakeServiceAccounts) Get(name string, options v1.GetOptions) (result *corev1.ServiceAccount, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(serviceaccountsResource, c.ns, name), &corev1.ServiceAccount{})

	if obj == nil {
		return nil, err
	}
	return obj.(*corev1.ServiceAccount), err
}

// List takes label and field selectors, and returns the list of ServiceAccounts that match those selectors.
func (c *FakeServiceAccounts) List(opts v1.ListOptions) (result *corev1.ServiceAccountList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(serviceaccountsResource, serviceaccountsKind, c.ns, opts), &corev1.ServiceAccountList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &corev1.ServiceAccountList{ListMeta: obj.(*corev1.ServiceAccountList).ListMeta}
	for _, item := range obj.(*corev1.ServiceAccountList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested serviceAccounts.
func (c *FakeServiceAccounts) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(serviceaccountsResource, c.ns, opts))

}

// Create takes the representation of a serviceAccount and creates it.  Returns the server's representation of the serviceAccount, and an error, if there is any.
func (c *FakeServiceAccounts) Create(serviceAccount *corev1.ServiceAccount) (result *corev1.ServiceAccount, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(serviceaccountsResource, c.ns, serviceAccount), &corev1.ServiceAccount{})

	if obj == nil {
		return nil, err
	}
	return obj.(*corev1.ServiceAccount), err
}

// Update takes the representation of a serviceAccount and updates it. Returns the server's representation of the serviceAccount, and an error, if there is any.
func (c *FakeServiceAccounts) Update(serviceAccount *corev1.ServiceAccount) (result *corev1.ServiceAccount, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(serviceaccountsResource, c.ns, serviceAccount), &corev1.ServiceAccount{})

	if obj == nil {
		return nil, err
	}
	return obj.(*corev1.ServiceAccount), err
}

// Delete takes name of the serviceAccount and deletes it. Returns an error if one occurs.
func (c *FakeServiceAccounts) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(serviceaccountsResource, c.ns, name), &corev1.ServiceAccount{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeServiceAccounts) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(serviceaccountsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &corev1.ServiceAccountList{})
	return err
}

// Patch applies the patch and returns the patched serviceAccount.
func (c *FakeServiceAccounts) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *corev1.ServiceAccount, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(serviceaccountsResource, c.ns, name, pt, data, subresources...), &corev1.ServiceAccount{})

	if obj == nil {
		return nil, err
	}
	return obj.(*corev1.ServiceAccount), err
}
