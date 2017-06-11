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
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
)

func addConversionFuncs(scheme *runtime.Scheme) error {
	// Add non-generated conversion functions
	err := scheme.AddConversionFuncs(
		Convert_apiextensions_JSON_To_v1beta1_JSON,
		Convert_v1beta1_JSON_To_apiextensions_JSON,
	)
	if err != nil {
		return err
	}
	return nil
}

func Convert_apiextensions_JSON_To_v1beta1_JSON(in *apiextensions.JSON, out *JSON, s conversion.Scope) error {
	if in != nil {
		raw, err := json.Marshal(*in)
		if err != nil {
			return err
		}
		out.Raw = raw
	} else {
		out = nil
	}
	return nil
}

func Convert_v1beta1_JSON_To_apiextensions_JSON(in *JSON, out *apiextensions.JSON, s conversion.Scope) error {
	if in != nil {
		var i interface{}
		if err := json.Unmarshal(in.Raw, &i); err != nil {
			return err
		}
		*out = i
	} else {
		out = nil
	}
	return nil
}
