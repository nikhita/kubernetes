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

package v1beta1

import (
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func addConversionFuncs(scheme *runtime.Scheme) error {
	// Add non-generated conversion funtions
	err := scheme.AddConversionFuncs(
		Convert_v1beta1_CustomResourceDefinitionSpec_To_apiextensions_CustomResourceDefinitionSpec,
		// Convert_apiextensions_CustomResourceDefinitionSpec_To_v1beta1_CustomResourceDefinitionSpec,
	)
	if err != nil {
		return err
	}
}

func Convert_v1beta1_CustomResourceDefinitionSpec_To_apiextensions_CustomResourceDefinitionSpec(in *CustomResourceDefinitionSpec, out *apiextensions.CustomResourceDefinitionSpec) error {
	out.Group = in.Group
	out.Version = in.Version
	if err := Convert_apiextensions_CustomResourceDefinitionNames_To_v1beta1_CustomResourceDefinitionNames(&in.Names, &out.Names, s); err != nil {
		return err
	}
	out.Scope = ResourceScope(in.Scope)
	if err := Convert_apiextensions_CustomResourceValidation_To_v1beta1_CustomResourceValidation(&in.Validation, &out.Validation, s); err != nil {
		return err
	}
	if in.Selector != nil {
		metav1.Convert_map_to_unversioned_LabelSelector(&in.Selector, out.Selector, s)
	}
	if err := Convert_apiextensions_CustomResourceSubResourceType_To_v1beta1_CustomResourceSubResourceType(&in.SubResources, &out.SubResources, s); err != nil {
		return err
	}
	return nil
}
