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
	"errors"
	"fmt"
	"strings"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
)

// CustomResourceStorage includes dummy storage for CustomResources, and their Status and Scale subresources.
type CustomResourceStorage struct {
	CustomResource *REST
	Status         *StatusREST
	Scale          *ScaleREST
}

func NewStorage(resource schema.GroupResource, listKind schema.GroupVersionKind, strategy customResourceStrategy, optsGetter generic.RESTOptionsGetter, scale *apiextensions.CustomResourceSubResourceScale) CustomResourceStorage {
	customResourceREST, customResourceStatusREST := NewREST(resource, listKind, strategy, optsGetter)
	customResourceRegistry := NewRegistry(customResourceREST)

	s := CustomResourceStorage{
		CustomResource: customResourceREST,
		Status:         customResourceStatusREST,
	}

	if scale != nil {
		s.Scale = &ScaleREST{
			registry:           customResourceRegistry,
			specReplicasPath:   scale.SpecReplicasPath,
			statusReplicasPath: scale.StatusReplicasPath,
			labelSelectorPath:  scale.LabelSelectorPath,
		}
	}

	return s
}

// REST implements a RESTStorage for API services against etcd
type REST struct {
	*genericregistry.Store
}

// NewREST returns a RESTStorage object that will work against API services.
func NewREST(resource schema.GroupResource, listKind schema.GroupVersionKind, strategy customResourceStrategy, optsGetter generic.RESTOptionsGetter) (*REST, *StatusREST) {
	store := &genericregistry.Store{
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
	statusStore.UpdateStrategy = NewStatusStrategy(strategy)
	return &REST{store}, &StatusREST{store: &statusStore}
}

// StatusREST implements the REST endpoint for changing the status of a CustomResource
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
func (r *StatusREST) Update(ctx genericapirequest.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc) (runtime.Object, bool, error) {
	return r.store.Update(ctx, name, objInfo, createValidation, updateValidation)
}

type ScaleREST struct {
	registry           Registry
	specReplicasPath   string
	statusReplicasPath string
	labelSelectorPath  string
}

// ScaleREST implements Patcher
var _ = rest.Patcher(&ScaleREST{})
var _ = rest.GroupVersionKindProvider(&ScaleREST{})

func (r *ScaleREST) GroupVersionKind(containingGV schema.GroupVersion) schema.GroupVersionKind {
	return autoscalingv1.SchemeGroupVersion.WithKind("Scale")
}

// New creates a new Scale object
func (r *ScaleREST) New() runtime.Object {
	return &autoscalingv1.Scale{}
}

func (r *ScaleREST) Get(ctx genericapirequest.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	cr, err := r.registry.GetCustomResource(ctx, name, options)
	if err != nil {
		return nil, err
	}

	scaleObject, err := scaleFromCustomResource(cr, r.specReplicasPath, r.statusReplicasPath, r.labelSelectorPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v", err))
	}
	return scaleObject, err
}

func (r *ScaleREST) Update(ctx genericapirequest.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc) (runtime.Object, bool, error) {
	cr, err := r.registry.GetCustomResource(ctx, name, &metav1.GetOptions{})
	if err != nil {
		return nil, false, err
	}

	oldScale, err := scaleFromCustomResource(cr, r.specReplicasPath, r.statusReplicasPath, r.labelSelectorPath)
	if err != nil {
		return nil, false, err
	}

	obj, err := objInfo.UpdatedObject(ctx, oldScale)
	if err != nil {
		return nil, false, err
	}
	if obj == nil {
		return nil, false, apierrors.NewBadRequest(fmt.Sprintf("nil update passed to Scale"))
	}

	scale, ok := obj.(*autoscalingv1.Scale)
	if !ok {
		return nil, false, apierrors.NewBadRequest(fmt.Sprintf("wrong object passed to Scale update: %v", obj))
	}

	specReplicasPath := strings.TrimPrefix(r.specReplicasPath, ".") // ignore leading period
	if ok = unstructured.SetNestedField(cr.Object, int64(scale.Spec.Replicas), strings.Split(specReplicasPath, ".")...); !ok {
		return nil, false, fmt.Errorf("spec replicas value cannot be set because one of the nesting levels is not a map[string]interface{}")
	}
	cr.SetResourceVersion(scale.ResourceVersion)

	cr, err = r.registry.UpdateCustomResource(ctx, cr, createValidation, updateValidation)
	if err != nil {
		return nil, false, err
	}

	newScale, err := scaleFromCustomResource(cr, r.specReplicasPath, r.statusReplicasPath, r.labelSelectorPath)
	if err != nil {
		return nil, false, apierrors.NewBadRequest(fmt.Sprintf("%v", err))
	}
	return newScale, false, err
}

// scaleFromCustomResource returns a scale subresource for a customresource.
func scaleFromCustomResource(cr *unstructured.Unstructured, specReplicasPath, statusReplicasPath, labelSelectorPath string) (*autoscalingv1.Scale, error) {
	specReplicasPath = strings.TrimPrefix(specReplicasPath, ".") // ignore leading period
	specReplicas, ok := unstructured.NestedInt64(cr.UnstructuredContent(), strings.Split(specReplicasPath, ".")...)
	if !ok {
		return nil, fmt.Errorf("invalid spec replicas value: %v", specReplicas)
	}
	if specReplicas < 0 {
		return nil, fmt.Errorf("spec replicas value should be a non-negative integer")
	}

	statusReplicasPath = strings.TrimPrefix(statusReplicasPath, ".") // ignore leading period
	statusReplicas, ok := unstructured.NestedInt64(cr.UnstructuredContent(), strings.Split(statusReplicasPath, ".")...)
	if !ok {
		statusReplicas = 0
	}
	if statusReplicas < 0 {
		return nil, fmt.Errorf("status replicas value should be a non-negative integer")
	}

	labelSelectorPath = strings.TrimPrefix(labelSelectorPath, ".") // ignore leading period
	labelSelector, _ := unstructured.NestedString(cr.UnstructuredContent(), strings.Split(labelSelectorPath, ".")...)

	scale := &autoscalingv1.Scale{
		ObjectMeta: metav1.ObjectMeta{
			Name:              cr.GetName(),
			Namespace:         cr.GetNamespace(),
			UID:               cr.GetUID(),
			ResourceVersion:   cr.GetResourceVersion(),
			CreationTimestamp: cr.GetCreationTimestamp(),
		},
		Spec: autoscalingv1.ScaleSpec{
			Replicas: int32(specReplicas),
		},
		Status: autoscalingv1.ScaleStatus{
			Replicas: int32(statusReplicas),
			Selector: labelSelector,
		},
	}

	return scale, nil
}
