/*
Copyright 2017 The Kubernetes Authors.

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

package customresource

import (
	"fmt"
	"strconv"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
)

// CustomResourceStorage includes dummy storage for CustomResources and for Scale subresource.
type CustomResourceStorage struct {
	CustomResource *REST
	Status         *StatusREST
	Scale          *ScaleREST
}

func NewStorage(resource schema.GroupResource, listKind schema.GroupVersionKind, copier runtime.ObjectCopier, strategy CustomResourceDefinitionStorageStrategy, optsGetter generic.RESTOptionsGetter) CustomResourceStorage {
	customResourceREST, customResourceStatusREST := NewREST(resource, listKind, copier, strategy, optsGetter)
	customResourceRegistry := NewRegistry(customResourceREST)

	return CustomResourceStorage{
		CustomResource: customResourceREST,
		Status:         customResourceStatusREST,
		Scale:          &ScaleREST{registry: customResourceRegistry},
	}
}

// REST implements a RESTStorage for API services against etcd
type REST struct {
	*genericregistry.Store
}

// NewREST returns a RESTStorage object that will work against API services.
func NewREST(resource schema.GroupResource, listKind schema.GroupVersionKind, copier runtime.ObjectCopier, strategy CustomResourceDefinitionStorageStrategy, optsGetter generic.RESTOptionsGetter) (*REST, *StatusREST) {
	store := &genericregistry.Store{
		Copier:  copier,
		NewFunc: func() runtime.Object { return &unstructured.Unstructured{} },
		NewListFunc: func() runtime.Object {
			// lists are never stored, only manufactured, so stomp in the right kind
			ret := &unstructured.UnstructuredList{}
			ret.SetGroupVersionKind(listKind)
			return ret
		},
		PredicateFunc:            strategy.MatchCustomResourceDefinitionStorage,
		DefaultQualifiedResource: resource,

		CreateStrategy: strategy,
		UpdateStrategy: strategy,
		DeleteStrategy: strategy,
	}
	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: strategy.GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		panic(err) // TODO: Propagate error up
	}

	statusStore := *store
	statusStore.UpdateStrategy = customResourceDefinitionStorageStatusStrategy{strategy}

	return &REST{store}, &StatusREST{store: &statusStore}
}

// StatusREST implements the REST endpoint for changing the status of a CustomResource.
type StatusREST struct {
	store *genericregistry.Store
}

func (r *StatusREST) New() runtime.Object {
	return &unstructured.Unstructured{}
}

// Get retrieves the object from the storage. It is required to support Patch.
func (r *StatusREST) Get(ctx genericapirequest.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return r.store.Get(ctx, name, options)
}

// Update alters the status subset of an object.
func (r *StatusREST) Update(ctx genericapirequest.Context, name string, objInfo rest.UpdatedObjectInfo) (runtime.Object, bool, error) {
	return r.store.Update(ctx, name, objInfo)
}

type ScaleREST struct {
	registry Registry
}

// ScaleREST implements Patcher
var _ = rest.Patcher(&ScaleREST{})

// New creates a new Scale object
func (r *ScaleREST) New() runtime.Object {
	return &apiextensions.Scale{}
}

func (r *ScaleREST) Get(ctx genericapirequest.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	customResource, err := r.registry.GetCustomResource(ctx, name, options)
	if err != nil {
		return nil, err
	}
	scale, err := scaleFromCustomResource(customResource)
	if err != nil {
		return nil, errors.NewBadRequest(fmt.Sprintf("%v", err))
	}
	return scale, nil
}

func (r *ScaleREST) Update(ctx genericapirequest.Context, name string, objInfo rest.UpdatedObjectInfo) (runtime.Object, bool, error) {
	customResource, err := r.registry.GetCustomResource(ctx, name, &metav1.GetOptions{})
	if err != nil {
		return nil, false, err
	}

	oldScale, err := scaleFromCustomResource(customResource)
	if err != nil {
		return nil, false, err
	}

	obj, err := objInfo.UpdatedObject(ctx, oldScale)
	if err != nil {
		return nil, false, err
	}
	if obj == nil {
		return nil, false, errors.NewBadRequest(fmt.Sprintf("nil update passed to Scale"))
	}
	scale, ok := obj.(*apiextensions.Scale)
	if !ok {
		return nil, false, errors.NewBadRequest(fmt.Sprintf("wrong object passed to Scale update: %v", obj))
	}

	// TODO: Add validation for the apiextensions.Scale object.
	/*if errs := extvalidation.ValidateScale(scale); len(errs) > 0 {
		return nil, false, errors.NewInvalid(extensions.Kind("Scale"), scale.Name, errs)
	}*/

	customResourceContent := customResource.UnstructuredContent()
	customResourceSpecReplicasString, err := apiextensions.ParseJSONPath(customResourceContent, "specReplicas", "{.spec.replicas}")
	if err != nil {
		return nil, false, errors.NewBadRequest(fmt.Sprintf("cannot parse custom resource: %v", err))
	}

	customResourceSpecReplicas, err := strconv.ParseInt(customResourceSpecReplicasString, 10, 32)
	if err != nil {
		return nil, false, err
	}

	int32(customResourceSpecReplicas) = scale.Spec.Replicas
	customResource.SetResourceVersion(scale.ResourceVersion)
	customResource, err = r.registry.UpdateCustomrsource(ctx, customResource)
	if err != nil {
		return nil, false, err
	}

	newScale, err := scaleFromCustomResource(customResource)
	if err != nil {
		return nil, false, errors.NewBadRequest(fmt.Sprintf("%v", err))
	}
	return newScale, false, err
}

// scaleFromCustomResource returns a scale subresource for a custom resource.
func scaleFromCustomResource(customResource *unstructured.Unstructured) (*apiextensions.Scale, error) {
	customResourceContent := customResource.UnstructuredContent()
	customResourceSpecReplicas, err := apiextensions.ParseJSONPath(customResourceContent, "specReplicas", "{.spec.replicas}")
	if err != nil {
		return nil, err
	}
	customResourceStatusReplicas, err := apiextensions.ParseJSONPath(customResourceContent, "statusReplicas", "{.status.replicas}")
	if err != nil {
		return nil, err
	}
	customResourceStatusSelector, err := apiextensions.ParseJSONPath(customResourceContent, "statusSelector", "{.status.selector}")
	if err != nil {
		return nil, err
	}

	return &apiextensions.Scale{
		ObjectMeta: metav1.ObjectMeta{
			Name:              customResource.GetName(),
			Namespace:         customResource.GetNamespace(),
			UID:               customResource.GetUID(),
			ResourceVersion:   customResource.GetResourceVersion(),
			CreationTimestamp: customResource.GetCreationTimestamp(),
		},
		Spec: apiextensions.ScaleSpec{
			Replicas: customResourceSpecReplicas,
		},
		Status: apiextensions.ScaleStatus{
			Replicas: customResourceStatusReplicas,
			Selector: customResourceStatusSelector,
		},
	}, nil
}
