/*
Copyright 2022 Rancher Labs, Inc.

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

package v3

import (
	"context"
	"time"

	"github.com/rancher/lasso/pkg/client"
	"github.com/rancher/lasso/pkg/controller"
	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/condition"
	"github.com/rancher/wrangler/pkg/generic"
	"github.com/rancher/wrangler/pkg/kv"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

type ClusterScanHandler func(string, *v3.ClusterScan) (*v3.ClusterScan, error)

type ClusterScanController interface {
	generic.ControllerMeta
	ClusterScanClient

	OnChange(ctx context.Context, name string, sync ClusterScanHandler)
	OnRemove(ctx context.Context, name string, sync ClusterScanHandler)
	Enqueue(namespace, name string)
	EnqueueAfter(namespace, name string, duration time.Duration)

	Cache() ClusterScanCache
}

type ClusterScanClient interface {
	Create(*v3.ClusterScan) (*v3.ClusterScan, error)
	Update(*v3.ClusterScan) (*v3.ClusterScan, error)
	UpdateStatus(*v3.ClusterScan) (*v3.ClusterScan, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	Get(namespace, name string, options metav1.GetOptions) (*v3.ClusterScan, error)
	List(namespace string, opts metav1.ListOptions) (*v3.ClusterScanList, error)
	Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error)
	Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v3.ClusterScan, err error)
}

type ClusterScanCache interface {
	Get(namespace, name string) (*v3.ClusterScan, error)
	List(namespace string, selector labels.Selector) ([]*v3.ClusterScan, error)

	AddIndexer(indexName string, indexer ClusterScanIndexer)
	GetByIndex(indexName, key string) ([]*v3.ClusterScan, error)
}

type ClusterScanIndexer func(obj *v3.ClusterScan) ([]string, error)

type clusterScanController struct {
	controller    controller.SharedController
	client        *client.Client
	gvk           schema.GroupVersionKind
	groupResource schema.GroupResource
}

func NewClusterScanController(gvk schema.GroupVersionKind, resource string, namespaced bool, controller controller.SharedControllerFactory) ClusterScanController {
	c := controller.ForResourceKind(gvk.GroupVersion().WithResource(resource), gvk.Kind, namespaced)
	return &clusterScanController{
		controller: c,
		client:     c.Client(),
		gvk:        gvk,
		groupResource: schema.GroupResource{
			Group:    gvk.Group,
			Resource: resource,
		},
	}
}

func FromClusterScanHandlerToHandler(sync ClusterScanHandler) generic.Handler {
	return func(key string, obj runtime.Object) (ret runtime.Object, err error) {
		var v *v3.ClusterScan
		if obj == nil {
			v, err = sync(key, nil)
		} else {
			v, err = sync(key, obj.(*v3.ClusterScan))
		}
		if v == nil {
			return nil, err
		}
		return v, err
	}
}

func (c *clusterScanController) Updater() generic.Updater {
	return func(obj runtime.Object) (runtime.Object, error) {
		newObj, err := c.Update(obj.(*v3.ClusterScan))
		if newObj == nil {
			return nil, err
		}
		return newObj, err
	}
}

func UpdateClusterScanDeepCopyOnChange(client ClusterScanClient, obj *v3.ClusterScan, handler func(obj *v3.ClusterScan) (*v3.ClusterScan, error)) (*v3.ClusterScan, error) {
	if obj == nil {
		return obj, nil
	}

	copyObj := obj.DeepCopy()
	newObj, err := handler(copyObj)
	if newObj != nil {
		copyObj = newObj
	}
	if obj.ResourceVersion == copyObj.ResourceVersion && !equality.Semantic.DeepEqual(obj, copyObj) {
		return client.Update(copyObj)
	}

	return copyObj, err
}

func (c *clusterScanController) AddGenericHandler(ctx context.Context, name string, handler generic.Handler) {
	c.controller.RegisterHandler(ctx, name, controller.SharedControllerHandlerFunc(handler))
}

func (c *clusterScanController) AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler) {
	c.AddGenericHandler(ctx, name, generic.NewRemoveHandler(name, c.Updater(), handler))
}

func (c *clusterScanController) OnChange(ctx context.Context, name string, sync ClusterScanHandler) {
	c.AddGenericHandler(ctx, name, FromClusterScanHandlerToHandler(sync))
}

func (c *clusterScanController) OnRemove(ctx context.Context, name string, sync ClusterScanHandler) {
	c.AddGenericHandler(ctx, name, generic.NewRemoveHandler(name, c.Updater(), FromClusterScanHandlerToHandler(sync)))
}

func (c *clusterScanController) Enqueue(namespace, name string) {
	c.controller.Enqueue(namespace, name)
}

func (c *clusterScanController) EnqueueAfter(namespace, name string, duration time.Duration) {
	c.controller.EnqueueAfter(namespace, name, duration)
}

func (c *clusterScanController) Informer() cache.SharedIndexInformer {
	return c.controller.Informer()
}

func (c *clusterScanController) GroupVersionKind() schema.GroupVersionKind {
	return c.gvk
}

func (c *clusterScanController) Cache() ClusterScanCache {
	return &clusterScanCache{
		indexer:  c.Informer().GetIndexer(),
		resource: c.groupResource,
	}
}

func (c *clusterScanController) Create(obj *v3.ClusterScan) (*v3.ClusterScan, error) {
	result := &v3.ClusterScan{}
	return result, c.client.Create(context.TODO(), obj.Namespace, obj, result, metav1.CreateOptions{})
}

func (c *clusterScanController) Update(obj *v3.ClusterScan) (*v3.ClusterScan, error) {
	result := &v3.ClusterScan{}
	return result, c.client.Update(context.TODO(), obj.Namespace, obj, result, metav1.UpdateOptions{})
}

func (c *clusterScanController) UpdateStatus(obj *v3.ClusterScan) (*v3.ClusterScan, error) {
	result := &v3.ClusterScan{}
	return result, c.client.UpdateStatus(context.TODO(), obj.Namespace, obj, result, metav1.UpdateOptions{})
}

func (c *clusterScanController) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	if options == nil {
		options = &metav1.DeleteOptions{}
	}
	return c.client.Delete(context.TODO(), namespace, name, *options)
}

func (c *clusterScanController) Get(namespace, name string, options metav1.GetOptions) (*v3.ClusterScan, error) {
	result := &v3.ClusterScan{}
	return result, c.client.Get(context.TODO(), namespace, name, result, options)
}

func (c *clusterScanController) List(namespace string, opts metav1.ListOptions) (*v3.ClusterScanList, error) {
	result := &v3.ClusterScanList{}
	return result, c.client.List(context.TODO(), namespace, result, opts)
}

func (c *clusterScanController) Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	return c.client.Watch(context.TODO(), namespace, opts)
}

func (c *clusterScanController) Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (*v3.ClusterScan, error) {
	result := &v3.ClusterScan{}
	return result, c.client.Patch(context.TODO(), namespace, name, pt, data, result, metav1.PatchOptions{}, subresources...)
}

type clusterScanCache struct {
	indexer  cache.Indexer
	resource schema.GroupResource
}

func (c *clusterScanCache) Get(namespace, name string) (*v3.ClusterScan, error) {
	obj, exists, err := c.indexer.GetByKey(namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(c.resource, name)
	}
	return obj.(*v3.ClusterScan), nil
}

func (c *clusterScanCache) List(namespace string, selector labels.Selector) (ret []*v3.ClusterScan, err error) {

	err = cache.ListAllByNamespace(c.indexer, namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v3.ClusterScan))
	})

	return ret, err
}

func (c *clusterScanCache) AddIndexer(indexName string, indexer ClusterScanIndexer) {
	utilruntime.Must(c.indexer.AddIndexers(map[string]cache.IndexFunc{
		indexName: func(obj interface{}) (strings []string, e error) {
			return indexer(obj.(*v3.ClusterScan))
		},
	}))
}

func (c *clusterScanCache) GetByIndex(indexName, key string) (result []*v3.ClusterScan, err error) {
	objs, err := c.indexer.ByIndex(indexName, key)
	if err != nil {
		return nil, err
	}
	result = make([]*v3.ClusterScan, 0, len(objs))
	for _, obj := range objs {
		result = append(result, obj.(*v3.ClusterScan))
	}
	return result, nil
}

type ClusterScanStatusHandler func(obj *v3.ClusterScan, status v3.ClusterScanStatus) (v3.ClusterScanStatus, error)

type ClusterScanGeneratingHandler func(obj *v3.ClusterScan, status v3.ClusterScanStatus) ([]runtime.Object, v3.ClusterScanStatus, error)

func RegisterClusterScanStatusHandler(ctx context.Context, controller ClusterScanController, condition condition.Cond, name string, handler ClusterScanStatusHandler) {
	statusHandler := &clusterScanStatusHandler{
		client:    controller,
		condition: condition,
		handler:   handler,
	}
	controller.AddGenericHandler(ctx, name, FromClusterScanHandlerToHandler(statusHandler.sync))
}

func RegisterClusterScanGeneratingHandler(ctx context.Context, controller ClusterScanController, apply apply.Apply,
	condition condition.Cond, name string, handler ClusterScanGeneratingHandler, opts *generic.GeneratingHandlerOptions) {
	statusHandler := &clusterScanGeneratingHandler{
		ClusterScanGeneratingHandler: handler,
		apply:                        apply,
		name:                         name,
		gvk:                          controller.GroupVersionKind(),
	}
	if opts != nil {
		statusHandler.opts = *opts
	}
	controller.OnChange(ctx, name, statusHandler.Remove)
	RegisterClusterScanStatusHandler(ctx, controller, condition, name, statusHandler.Handle)
}

type clusterScanStatusHandler struct {
	client    ClusterScanClient
	condition condition.Cond
	handler   ClusterScanStatusHandler
}

func (a *clusterScanStatusHandler) sync(key string, obj *v3.ClusterScan) (*v3.ClusterScan, error) {
	if obj == nil {
		return obj, nil
	}

	origStatus := obj.Status.DeepCopy()
	obj = obj.DeepCopy()
	newStatus, err := a.handler(obj, obj.Status)
	if err != nil {
		// Revert to old status on error
		newStatus = *origStatus.DeepCopy()
	}

	if a.condition != "" {
		if errors.IsConflict(err) {
			a.condition.SetError(&newStatus, "", nil)
		} else {
			a.condition.SetError(&newStatus, "", err)
		}
	}
	if !equality.Semantic.DeepEqual(origStatus, &newStatus) {
		if a.condition != "" {
			// Since status has changed, update the lastUpdatedTime
			a.condition.LastUpdated(&newStatus, time.Now().UTC().Format(time.RFC3339))
		}

		var newErr error
		obj.Status = newStatus
		newObj, newErr := a.client.UpdateStatus(obj)
		if err == nil {
			err = newErr
		}
		if newErr == nil {
			obj = newObj
		}
	}
	return obj, err
}

type clusterScanGeneratingHandler struct {
	ClusterScanGeneratingHandler
	apply apply.Apply
	opts  generic.GeneratingHandlerOptions
	gvk   schema.GroupVersionKind
	name  string
}

func (a *clusterScanGeneratingHandler) Remove(key string, obj *v3.ClusterScan) (*v3.ClusterScan, error) {
	if obj != nil {
		return obj, nil
	}

	obj = &v3.ClusterScan{}
	obj.Namespace, obj.Name = kv.RSplit(key, "/")
	obj.SetGroupVersionKind(a.gvk)

	return nil, generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects()
}

func (a *clusterScanGeneratingHandler) Handle(obj *v3.ClusterScan, status v3.ClusterScanStatus) (v3.ClusterScanStatus, error) {
	if !obj.DeletionTimestamp.IsZero() {
		return status, nil
	}

	objs, newStatus, err := a.ClusterScanGeneratingHandler(obj, status)
	if err != nil {
		return newStatus, err
	}

	return newStatus, generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects(objs...)
}
