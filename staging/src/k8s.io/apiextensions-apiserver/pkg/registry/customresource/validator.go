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

	"github.com/go-openapi/validate"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"

	apiservervalidation "k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
)

type customResourceValidator struct {
	namespaceScoped       bool
	kind                  schema.GroupVersionKind
	schemaValidator       *validate.SchemaValidator
	statusSchemaValidator *validate.SchemaValidator
}

func (a customResourceValidator) Validate(ctx genericapirequest.Context, obj runtime.Object) field.ErrorList {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return field.ErrorList{field.Invalid(field.NewPath("metadata"), nil, err.Error())}
	}
	typeAccessor, err := meta.TypeAccessor(obj)
	if err != nil {
		return field.ErrorList{field.Invalid(field.NewPath("kind"), nil, err.Error())}
	}
	if typeAccessor.GetKind() != a.kind.Kind {
		return field.ErrorList{field.Invalid(field.NewPath("kind"), typeAccessor.GetKind(), fmt.Sprintf("must be %v", a.kind.Kind))}
	}
	if typeAccessor.GetAPIVersion() != a.kind.Group+"/"+a.kind.Version {
		return field.ErrorList{field.Invalid(field.NewPath("apiVersion"), typeAccessor.GetAPIVersion(), fmt.Sprintf("must be %v", a.kind.Group+"/"+a.kind.Version))}
	}

	customResourceObject, ok := obj.(*unstructured.Unstructured)
	// this will never happen.
	if !ok {
		return field.ErrorList{field.Invalid(field.NewPath(""), customResourceObject, fmt.Sprintf("has type %T. Must be a pointer to an Unstructured type", customResourceObject))}
	}
	customResource := customResourceObject.UnstructuredContent()

	if err = apiservervalidation.ValidateCustomResource(customResource, a.schemaValidator); err != nil {
		return field.ErrorList{field.Invalid(field.NewPath(""), customResource, err.Error())}
	}

	return validation.ValidateObjectMetaAccessor(accessor, a.namespaceScoped, validation.NameIsDNSSubdomain, field.NewPath("metadata"))
}

func (a customResourceValidator) ValidateUpdate(ctx genericapirequest.Context, obj, old runtime.Object) field.ErrorList {
	objAccessor, err := meta.Accessor(obj)
	if err != nil {
		return field.ErrorList{field.Invalid(field.NewPath("metadata"), nil, err.Error())}
	}
	oldAccessor, err := meta.Accessor(old)
	if err != nil {
		return field.ErrorList{field.Invalid(field.NewPath("metadata"), nil, err.Error())}
	}
	typeAccessor, err := meta.TypeAccessor(obj)
	if err != nil {
		return field.ErrorList{field.Invalid(field.NewPath("kind"), nil, err.Error())}
	}
	if typeAccessor.GetKind() != a.kind.Kind {
		return field.ErrorList{field.Invalid(field.NewPath("kind"), typeAccessor.GetKind(), fmt.Sprintf("must be %v", a.kind.Kind))}
	}
	if typeAccessor.GetAPIVersion() != a.kind.Group+"/"+a.kind.Version {
		return field.ErrorList{field.Invalid(field.NewPath("apiVersion"), typeAccessor.GetAPIVersion(), fmt.Sprintf("must be %v", a.kind.Group+"/"+a.kind.Version))}
	}

	customResourceObject, ok := obj.(*unstructured.Unstructured)
	// this will never happen.
	if !ok {
		return field.ErrorList{field.Invalid(field.NewPath(""), customResourceObject, fmt.Sprintf("has type %T. Must be a pointer to an Unstructured type", customResourceObject))}
	}
	customResource := customResourceObject.UnstructuredContent()

	if err = apiservervalidation.ValidateCustomResource(customResource, a.schemaValidator); err != nil {
		return field.ErrorList{field.Invalid(field.NewPath(""), customResource, err.Error())}
	}

	return validation.ValidateObjectMetaAccessorUpdate(objAccessor, oldAccessor, field.NewPath("metadata"))
}

func (a customResourceValidator) ValidateStatusUpdate(ctx genericapirequest.Context, obj, old runtime.Object) field.ErrorList {
	objAccessor, err := meta.Accessor(obj)
	if err != nil {
		return field.ErrorList{field.Invalid(field.NewPath("metadata"), nil, err.Error())}
	}
	oldAccessor, err := meta.Accessor(old)
	if err != nil {
		return field.ErrorList{field.Invalid(field.NewPath("metadata"), nil, err.Error())}
	}
	typeAccessor, err := meta.TypeAccessor(obj)
	if err != nil {
		return field.ErrorList{field.Invalid(field.NewPath("kind"), nil, err.Error())}
	}
	if typeAccessor.GetKind() != a.kind.Kind {
		return field.ErrorList{field.Invalid(field.NewPath("kind"), typeAccessor.GetKind(), fmt.Sprintf("must be %v", a.kind.Kind))}
	}
	if typeAccessor.GetAPIVersion() != a.kind.Group+"/"+a.kind.Version {
		return field.ErrorList{field.Invalid(field.NewPath("apiVersion"), typeAccessor.GetAPIVersion(), fmt.Sprintf("must be %v", a.kind.Group+"/"+a.kind.Version))}
	}

	customResourceObject, ok := obj.(*unstructured.Unstructured)
	// this will never happen.
	if !ok {
		return field.ErrorList{field.Invalid(field.NewPath(""), customResourceObject, fmt.Sprintf("has type %T. Must be a pointer to an Unstructured type", customResourceObject))}
	}
	customResource := customResourceObject.UnstructuredContent()

	// validate only the status
	customResourceStatus := customResource["status"]
	if err = apiservervalidation.ValidateCustomResource(customResourceStatus, a.statusSchemaValidator); err != nil {
		return field.ErrorList{field.Invalid(field.NewPath("status"), customResourceStatus, err.Error())}
	}

	return validation.ValidateObjectMetaAccessorUpdate(objAccessor, oldAccessor, field.NewPath("metadata"))
}
