/*
Copyright 2023 Rancher Labs, Inc.

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

// Code generated by main. DO NOT EDIT.

package fake

import (
	"context"

	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeRoleTemplates implements RoleTemplateInterface
type FakeRoleTemplates struct {
	Fake *FakeManagementV3
}

var roletemplatesResource = schema.GroupVersionResource{Group: "management.cattle.io", Version: "v3", Resource: "roletemplates"}

var roletemplatesKind = schema.GroupVersionKind{Group: "management.cattle.io", Version: "v3", Kind: "RoleTemplate"}

// Get takes name of the roleTemplate, and returns the corresponding roleTemplate object, and an error if there is any.
func (c *FakeRoleTemplates) Get(ctx context.Context, name string, options v1.GetOptions) (result *v3.RoleTemplate, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(roletemplatesResource, name), &v3.RoleTemplate{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v3.RoleTemplate), err
}

// List takes label and field selectors, and returns the list of RoleTemplates that match those selectors.
func (c *FakeRoleTemplates) List(ctx context.Context, opts v1.ListOptions) (result *v3.RoleTemplateList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(roletemplatesResource, roletemplatesKind, opts), &v3.RoleTemplateList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v3.RoleTemplateList{ListMeta: obj.(*v3.RoleTemplateList).ListMeta}
	for _, item := range obj.(*v3.RoleTemplateList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested roleTemplates.
func (c *FakeRoleTemplates) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(roletemplatesResource, opts))
}

// Create takes the representation of a roleTemplate and creates it.  Returns the server's representation of the roleTemplate, and an error, if there is any.
func (c *FakeRoleTemplates) Create(ctx context.Context, roleTemplate *v3.RoleTemplate, opts v1.CreateOptions) (result *v3.RoleTemplate, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(roletemplatesResource, roleTemplate), &v3.RoleTemplate{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v3.RoleTemplate), err
}

// Update takes the representation of a roleTemplate and updates it. Returns the server's representation of the roleTemplate, and an error, if there is any.
func (c *FakeRoleTemplates) Update(ctx context.Context, roleTemplate *v3.RoleTemplate, opts v1.UpdateOptions) (result *v3.RoleTemplate, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(roletemplatesResource, roleTemplate), &v3.RoleTemplate{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v3.RoleTemplate), err
}

// Delete takes name of the roleTemplate and deletes it. Returns an error if one occurs.
func (c *FakeRoleTemplates) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(roletemplatesResource, name, opts), &v3.RoleTemplate{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeRoleTemplates) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(roletemplatesResource, listOpts)

	_, err := c.Fake.Invokes(action, &v3.RoleTemplateList{})
	return err
}

// Patch applies the patch and returns the patched roleTemplate.
func (c *FakeRoleTemplates) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v3.RoleTemplate, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(roletemplatesResource, name, pt, data, subresources...), &v3.RoleTemplate{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v3.RoleTemplate), err
}
