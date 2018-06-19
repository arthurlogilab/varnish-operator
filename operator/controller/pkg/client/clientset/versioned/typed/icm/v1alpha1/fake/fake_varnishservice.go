/*
Copyright 2018 The Kubernetes Authors.

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
	v1alpha1 "icm-varnish-k8s-operator/operator/controller/pkg/apis/icm/v1alpha1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeVarnishServices implements VarnishServiceInterface
type FakeVarnishServices struct {
	Fake *FakeIcmV1alpha1
	ns   string
}

var varnishservicesResource = schema.GroupVersionResource{Group: "icm.ibm.com", Version: "v1alpha1", Resource: "varnishservices"}

var varnishservicesKind = schema.GroupVersionKind{Group: "icm.ibm.com", Version: "v1alpha1", Kind: "VarnishService"}

// Get takes name of the varnishService, and returns the corresponding varnishService object, and an error if there is any.
func (c *FakeVarnishServices) Get(name string, options v1.GetOptions) (result *v1alpha1.VarnishService, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(varnishservicesResource, c.ns, name), &v1alpha1.VarnishService{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VarnishService), err
}

// List takes label and field selectors, and returns the list of VarnishServices that match those selectors.
func (c *FakeVarnishServices) List(opts v1.ListOptions) (result *v1alpha1.VarnishServiceList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(varnishservicesResource, varnishservicesKind, c.ns, opts), &v1alpha1.VarnishServiceList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.VarnishServiceList{}
	for _, item := range obj.(*v1alpha1.VarnishServiceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested varnishServices.
func (c *FakeVarnishServices) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(varnishservicesResource, c.ns, opts))

}

// Create takes the representation of a varnishService and creates it.  Returns the server's representation of the varnishService, and an error, if there is any.
func (c *FakeVarnishServices) Create(varnishService *v1alpha1.VarnishService) (result *v1alpha1.VarnishService, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(varnishservicesResource, c.ns, varnishService), &v1alpha1.VarnishService{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VarnishService), err
}

// Update takes the representation of a varnishService and updates it. Returns the server's representation of the varnishService, and an error, if there is any.
func (c *FakeVarnishServices) Update(varnishService *v1alpha1.VarnishService) (result *v1alpha1.VarnishService, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(varnishservicesResource, c.ns, varnishService), &v1alpha1.VarnishService{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VarnishService), err
}

// Delete takes name of the varnishService and deletes it. Returns an error if one occurs.
func (c *FakeVarnishServices) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(varnishservicesResource, c.ns, name), &v1alpha1.VarnishService{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeVarnishServices) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(varnishservicesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.VarnishServiceList{})
	return err
}

// Patch applies the patch and returns the patched varnishService.
func (c *FakeVarnishServices) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.VarnishService, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(varnishservicesResource, c.ns, name, data, subresources...), &v1alpha1.VarnishService{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VarnishService), err
}
