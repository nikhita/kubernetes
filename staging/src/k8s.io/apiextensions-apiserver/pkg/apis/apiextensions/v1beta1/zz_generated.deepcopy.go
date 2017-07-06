// +build !ignore_autogenerated

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

// This file was autogenerated by deepcopy-gen. Do not edit it manually!

package v1beta1

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
	url "net/url"
	reflect "reflect"
)

func init() {
	SchemeBuilder.Register(RegisterDeepCopies)
}

// RegisterDeepCopies adds deep-copy functions to the given scheme. Public
// to allow building arbitrary schemes.
func RegisterDeepCopies(scheme *runtime.Scheme) error {
	return scheme.AddGeneratedDeepCopyFuncs(
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1beta1_CustomResourceDefinition, InType: reflect.TypeOf(&CustomResourceDefinition{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1beta1_CustomResourceDefinitionCondition, InType: reflect.TypeOf(&CustomResourceDefinitionCondition{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1beta1_CustomResourceDefinitionList, InType: reflect.TypeOf(&CustomResourceDefinitionList{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1beta1_CustomResourceDefinitionNames, InType: reflect.TypeOf(&CustomResourceDefinitionNames{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1beta1_CustomResourceDefinitionSpec, InType: reflect.TypeOf(&CustomResourceDefinitionSpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1beta1_CustomResourceDefinitionStatus, InType: reflect.TypeOf(&CustomResourceDefinitionStatus{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1beta1_CustomResourceValidation, InType: reflect.TypeOf(&CustomResourceValidation{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1beta1_JSONSchemaPointer, InType: reflect.TypeOf(&JSONSchemaPointer{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1beta1_JSONSchemaProps, InType: reflect.TypeOf(&JSONSchemaProps{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1beta1_JSONSchemaPropsOrArray, InType: reflect.TypeOf(&JSONSchemaPropsOrArray{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1beta1_JSONSchemaPropsOrBool, InType: reflect.TypeOf(&JSONSchemaPropsOrBool{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1beta1_JSONSchemaPropsOrStringArray, InType: reflect.TypeOf(&JSONSchemaPropsOrStringArray{})},
		conversion.GeneratedDeepCopyFunc{Fn: DeepCopy_v1beta1_JSONSchemaRef, InType: reflect.TypeOf(&JSONSchemaRef{})},
	)
}

// DeepCopy_v1beta1_CustomResourceDefinition is an autogenerated deepcopy function.
func DeepCopy_v1beta1_CustomResourceDefinition(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*CustomResourceDefinition)
		out := out.(*CustomResourceDefinition)
		*out = *in
		if newVal, err := c.DeepCopy(&in.ObjectMeta); err != nil {
			return err
		} else {
			out.ObjectMeta = *newVal.(*v1.ObjectMeta)
		}
		if newVal, err := c.DeepCopy(&in.Spec); err != nil {
			return err
		} else {
			out.Spec = *newVal.(*CustomResourceDefinitionSpec)
		}
		if newVal, err := c.DeepCopy(&in.Status); err != nil {
			return err
		} else {
			out.Status = *newVal.(*CustomResourceDefinitionStatus)
		}
		return nil
	}
}

// DeepCopy_v1beta1_CustomResourceDefinitionCondition is an autogenerated deepcopy function.
func DeepCopy_v1beta1_CustomResourceDefinitionCondition(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*CustomResourceDefinitionCondition)
		out := out.(*CustomResourceDefinitionCondition)
		*out = *in
		out.LastTransitionTime = in.LastTransitionTime.DeepCopy()
		return nil
	}
}

// DeepCopy_v1beta1_CustomResourceDefinitionList is an autogenerated deepcopy function.
func DeepCopy_v1beta1_CustomResourceDefinitionList(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*CustomResourceDefinitionList)
		out := out.(*CustomResourceDefinitionList)
		*out = *in
		if in.Items != nil {
			in, out := &in.Items, &out.Items
			*out = make([]CustomResourceDefinition, len(*in))
			for i := range *in {
				if newVal, err := c.DeepCopy(&(*in)[i]); err != nil {
					return err
				} else {
					(*out)[i] = *newVal.(*CustomResourceDefinition)
				}
			}
		}
		return nil
	}
}

// DeepCopy_v1beta1_CustomResourceDefinitionNames is an autogenerated deepcopy function.
func DeepCopy_v1beta1_CustomResourceDefinitionNames(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*CustomResourceDefinitionNames)
		out := out.(*CustomResourceDefinitionNames)
		*out = *in
		if in.ShortNames != nil {
			in, out := &in.ShortNames, &out.ShortNames
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
		return nil
	}
}

// DeepCopy_v1beta1_CustomResourceDefinitionSpec is an autogenerated deepcopy function.
func DeepCopy_v1beta1_CustomResourceDefinitionSpec(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*CustomResourceDefinitionSpec)
		out := out.(*CustomResourceDefinitionSpec)
		*out = *in
		if newVal, err := c.DeepCopy(&in.Names); err != nil {
			return err
		} else {
			out.Names = *newVal.(*CustomResourceDefinitionNames)
		}
		if in.Validation != nil {
			in, out := &in.Validation, &out.Validation
			if newVal, err := c.DeepCopy(*in); err != nil {
				return err
			} else {
				*out = newVal.(*CustomResourceValidation)
			}
		}
		return nil
	}
}

// DeepCopy_v1beta1_CustomResourceDefinitionStatus is an autogenerated deepcopy function.
func DeepCopy_v1beta1_CustomResourceDefinitionStatus(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*CustomResourceDefinitionStatus)
		out := out.(*CustomResourceDefinitionStatus)
		*out = *in
		if in.Conditions != nil {
			in, out := &in.Conditions, &out.Conditions
			*out = make([]CustomResourceDefinitionCondition, len(*in))
			for i := range *in {
				if newVal, err := c.DeepCopy(&(*in)[i]); err != nil {
					return err
				} else {
					(*out)[i] = *newVal.(*CustomResourceDefinitionCondition)
				}
			}
		}
		if newVal, err := c.DeepCopy(&in.AcceptedNames); err != nil {
			return err
		} else {
			out.AcceptedNames = *newVal.(*CustomResourceDefinitionNames)
		}
		return nil
	}
}

// DeepCopy_v1beta1_CustomResourceValidation is an autogenerated deepcopy function.
func DeepCopy_v1beta1_CustomResourceValidation(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*CustomResourceValidation)
		out := out.(*CustomResourceValidation)
		*out = *in
		if in.JSONSchema != nil {
			in, out := &in.JSONSchema, &out.JSONSchema
			if newVal, err := c.DeepCopy(*in); err != nil {
				return err
			} else {
				*out = newVal.(*JSONSchemaProps)
			}
		}
		return nil
	}
}

// DeepCopy_v1beta1_JSONSchemaPointer is an autogenerated deepcopy function.
func DeepCopy_v1beta1_JSONSchemaPointer(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*JSONSchemaPointer)
		out := out.(*JSONSchemaPointer)
		*out = *in
		if in.ReferenceTokens != nil {
			in, out := &in.ReferenceTokens, &out.ReferenceTokens
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
		return nil
	}
}

// DeepCopy_v1beta1_JSONSchemaProps is an autogenerated deepcopy function.
func DeepCopy_v1beta1_JSONSchemaProps(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*JSONSchemaProps)
		out := out.(*JSONSchemaProps)
		*out = *in
		if newVal, err := c.DeepCopy(&in.Ref); err != nil {
			return err
		} else {
			out.Ref = *newVal.(*JSONSchemaRef)
		}
		if in.Type != nil {
			in, out := &in.Type, &out.Type
			*out = make(StringOrArray, len(*in))
			copy(*out, *in)
		}
		// in.Default is kind 'Interface'
		if in.Default != nil {
			if newVal, err := c.DeepCopy(&in.Default); err != nil {
				return err
			} else {
				out.Default = *newVal.(*interface{})
			}
		}
		if in.Maximum != nil {
			in, out := &in.Maximum, &out.Maximum
			*out = new(float64)
			**out = **in
		}
		if in.Minimum != nil {
			in, out := &in.Minimum, &out.Minimum
			*out = new(float64)
			**out = **in
		}
		if in.MaxLength != nil {
			in, out := &in.MaxLength, &out.MaxLength
			*out = new(int64)
			**out = **in
		}
		if in.MinLength != nil {
			in, out := &in.MinLength, &out.MinLength
			*out = new(int64)
			**out = **in
		}
		if in.MaxItems != nil {
			in, out := &in.MaxItems, &out.MaxItems
			*out = new(int64)
			**out = **in
		}
		if in.MinItems != nil {
			in, out := &in.MinItems, &out.MinItems
			*out = new(int64)
			**out = **in
		}
		if in.MultipleOf != nil {
			in, out := &in.MultipleOf, &out.MultipleOf
			*out = new(float64)
			**out = **in
		}
		if in.Enum != nil {
			in, out := &in.Enum, &out.Enum
			*out = make([]interface{}, len(*in))
			for i := range *in {
				if newVal, err := c.DeepCopy(&(*in)[i]); err != nil {
					return err
				} else {
					(*out)[i] = *newVal.(*interface{})
				}
			}
		}
		if in.MaxProperties != nil {
			in, out := &in.MaxProperties, &out.MaxProperties
			*out = new(int64)
			**out = **in
		}
		if in.MinProperties != nil {
			in, out := &in.MinProperties, &out.MinProperties
			*out = new(int64)
			**out = **in
		}
		if in.Required != nil {
			in, out := &in.Required, &out.Required
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
		if in.Items != nil {
			in, out := &in.Items, &out.Items
			if newVal, err := c.DeepCopy(*in); err != nil {
				return err
			} else {
				*out = newVal.(*JSONSchemaPropsOrArray)
			}
		}
		if in.AllOf != nil {
			in, out := &in.AllOf, &out.AllOf
			*out = make([]JSONSchemaProps, len(*in))
			for i := range *in {
				if newVal, err := c.DeepCopy(&(*in)[i]); err != nil {
					return err
				} else {
					(*out)[i] = *newVal.(*JSONSchemaProps)
				}
			}
		}
		if in.OneOf != nil {
			in, out := &in.OneOf, &out.OneOf
			*out = make([]JSONSchemaProps, len(*in))
			for i := range *in {
				if newVal, err := c.DeepCopy(&(*in)[i]); err != nil {
					return err
				} else {
					(*out)[i] = *newVal.(*JSONSchemaProps)
				}
			}
		}
		if in.AnyOf != nil {
			in, out := &in.AnyOf, &out.AnyOf
			*out = make([]JSONSchemaProps, len(*in))
			for i := range *in {
				if newVal, err := c.DeepCopy(&(*in)[i]); err != nil {
					return err
				} else {
					(*out)[i] = *newVal.(*JSONSchemaProps)
				}
			}
		}
		if in.Not != nil {
			in, out := &in.Not, &out.Not
			if newVal, err := c.DeepCopy(*in); err != nil {
				return err
			} else {
				*out = newVal.(*JSONSchemaProps)
			}
		}
		if in.Properties != nil {
			in, out := &in.Properties, &out.Properties
			*out = make(map[string]JSONSchemaProps)
			for key, val := range *in {
				if newVal, err := c.DeepCopy(&val); err != nil {
					return err
				} else {
					(*out)[key] = *newVal.(*JSONSchemaProps)
				}
			}
		}
		if in.AdditionalProperties != nil {
			in, out := &in.AdditionalProperties, &out.AdditionalProperties
			if newVal, err := c.DeepCopy(*in); err != nil {
				return err
			} else {
				*out = newVal.(*JSONSchemaPropsOrBool)
			}
		}
		if in.PatternProperties != nil {
			in, out := &in.PatternProperties, &out.PatternProperties
			*out = make(map[string]JSONSchemaProps)
			for key, val := range *in {
				if newVal, err := c.DeepCopy(&val); err != nil {
					return err
				} else {
					(*out)[key] = *newVal.(*JSONSchemaProps)
				}
			}
		}
		if in.Dependencies != nil {
			in, out := &in.Dependencies, &out.Dependencies
			*out = make(JSONSchemaDependencies)
			for key, val := range *in {
				if newVal, err := c.DeepCopy(&val); err != nil {
					return err
				} else {
					(*out)[key] = *newVal.(*JSONSchemaPropsOrStringArray)
				}
			}
		}
		if in.AdditionalItems != nil {
			in, out := &in.AdditionalItems, &out.AdditionalItems
			if newVal, err := c.DeepCopy(*in); err != nil {
				return err
			} else {
				*out = newVal.(*JSONSchemaPropsOrBool)
			}
		}
		if in.Definitions != nil {
			in, out := &in.Definitions, &out.Definitions
			*out = make(JSONSchemaDefinitions)
			for key, val := range *in {
				if newVal, err := c.DeepCopy(&val); err != nil {
					return err
				} else {
					(*out)[key] = *newVal.(*JSONSchemaProps)
				}
			}
		}
		return nil
	}
}

// DeepCopy_v1beta1_JSONSchemaPropsOrArray is an autogenerated deepcopy function.
func DeepCopy_v1beta1_JSONSchemaPropsOrArray(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*JSONSchemaPropsOrArray)
		out := out.(*JSONSchemaPropsOrArray)
		*out = *in
		if in.Schema != nil {
			in, out := &in.Schema, &out.Schema
			if newVal, err := c.DeepCopy(*in); err != nil {
				return err
			} else {
				*out = newVal.(*JSONSchemaProps)
			}
		}
		if in.JSONSchemas != nil {
			in, out := &in.JSONSchemas, &out.JSONSchemas
			*out = make([]JSONSchemaProps, len(*in))
			for i := range *in {
				if newVal, err := c.DeepCopy(&(*in)[i]); err != nil {
					return err
				} else {
					(*out)[i] = *newVal.(*JSONSchemaProps)
				}
			}
		}
		return nil
	}
}

// DeepCopy_v1beta1_JSONSchemaPropsOrBool is an autogenerated deepcopy function.
func DeepCopy_v1beta1_JSONSchemaPropsOrBool(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*JSONSchemaPropsOrBool)
		out := out.(*JSONSchemaPropsOrBool)
		*out = *in
		if in.Schema != nil {
			in, out := &in.Schema, &out.Schema
			if newVal, err := c.DeepCopy(*in); err != nil {
				return err
			} else {
				*out = newVal.(*JSONSchemaProps)
			}
		}
		return nil
	}
}

// DeepCopy_v1beta1_JSONSchemaPropsOrStringArray is an autogenerated deepcopy function.
func DeepCopy_v1beta1_JSONSchemaPropsOrStringArray(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*JSONSchemaPropsOrStringArray)
		out := out.(*JSONSchemaPropsOrStringArray)
		*out = *in
		if in.Schema != nil {
			in, out := &in.Schema, &out.Schema
			if newVal, err := c.DeepCopy(*in); err != nil {
				return err
			} else {
				*out = newVal.(*JSONSchemaProps)
			}
		}
		if in.Property != nil {
			in, out := &in.Property, &out.Property
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
		return nil
	}
}

// DeepCopy_v1beta1_JSONSchemaRef is an autogenerated deepcopy function.
func DeepCopy_v1beta1_JSONSchemaRef(in interface{}, out interface{}, c *conversion.Cloner) error {
	{
		in := in.(*JSONSchemaRef)
		out := out.(*JSONSchemaRef)
		*out = *in
		if in.ReferenceURL != nil {
			in, out := &in.ReferenceURL, &out.ReferenceURL
			if newVal, err := c.DeepCopy(*in); err != nil {
				return err
			} else {
				*out = newVal.(*url.URL)
			}
		}
		if newVal, err := c.DeepCopy(&in.ReferencePointer); err != nil {
			return err
		} else {
			out.ReferencePointer = *newVal.(*JSONSchemaPointer)
		}
		return nil
	}
}
