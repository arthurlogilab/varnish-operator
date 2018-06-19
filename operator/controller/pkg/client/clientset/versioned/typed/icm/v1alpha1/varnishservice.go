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

package v1alpha1

import (
	v1alpha1 "icm-varnish-k8s-operator/operator/controller/pkg/apis/icm/v1alpha1"
	scheme "icm-varnish-k8s-operator/operator/controller/pkg/client/clientset/versioned/scheme"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// VarnishServicesGetter has a method to return a VarnishServiceInterface.
// A group's client should implement this interface.
type VarnishServicesGetter interface {
	VarnishServices(namespace string) VarnishServiceInterface
}

// VarnishServiceInterface has methods to work with VarnishService resources.
type VarnishServiceInterface interface {
	Create(*v1alpha1.VarnishService) (*v1alpha1.VarnishService, error)
	Update(*v1alpha1.VarnishService) (*v1alpha1.VarnishService, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.VarnishService, error)
	List(opts v1.ListOptions) (*v1alpha1.VarnishServiceList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.VarnishService, err error)
	VarnishServiceExpansion
}

// varnishServices implements VarnishServiceInterface
type varnishServices struct {
	client rest.Interface
	ns     string
}

// newVarnishServices returns a VarnishServices
func newVarnishServices(c *IcmV1alpha1Client, namespace string) *varnishServices {
	return &varnishServices{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the varnishService, and returns the corresponding varnishService object, and an error if there is any.
func (c *varnishServices) Get(name string, options v1.GetOptions) (result *v1alpha1.VarnishService, err error) {
	result = &v1alpha1.VarnishService{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("varnishservices").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of VarnishServices that match those selectors.
func (c *varnishServices) List(opts v1.ListOptions) (result *v1alpha1.VarnishServiceList, err error) {
	result = &v1alpha1.VarnishServiceList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("varnishservices").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested varnishServices.
func (c *varnishServices) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("varnishservices").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a varnishService and creates it.  Returns the server's representation of the varnishService, and an error, if there is any.
func (c *varnishServices) Create(varnishService *v1alpha1.VarnishService) (result *v1alpha1.VarnishService, err error) {
	result = &v1alpha1.VarnishService{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("varnishservices").
		Body(varnishService).
		Do().
		Into(result)
	return
}

// Update takes the representation of a varnishService and updates it. Returns the server's representation of the varnishService, and an error, if there is any.
func (c *varnishServices) Update(varnishService *v1alpha1.VarnishService) (result *v1alpha1.VarnishService, err error) {
	result = &v1alpha1.VarnishService{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("varnishservices").
		Name(varnishService.Name).
		Body(varnishService).
		Do().
		Into(result)
	return
}

// Delete takes name of the varnishService and deletes it. Returns an error if one occurs.
func (c *varnishServices) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("varnishservices").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *varnishServices) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("varnishservices").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched varnishService.
func (c *varnishServices) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.VarnishService, err error) {
	result = &v1alpha1.VarnishService{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("varnishservices").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}