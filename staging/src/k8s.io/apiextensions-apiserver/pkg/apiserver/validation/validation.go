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

package validation

import (
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
)

func ValidateCustomResource(customresource interface{}, schema *spec.Schema) error {
	if err := spec.ExpandSchema(schema, nil, nil); err != nil {
		return err
	}

	validator := validate.NewSchemaValidator(schema, nil, "spec", strfmt.Default)
	result := validator.Validate(customresource)
	if result.AsError() != nil {
		return result.AsError()
	}
	return nil
}

func ConvertToOpenAPITypes(in *apiextensions.CustomResourceDefinition, out *spec.Schema) error {
	if err := convertJSONSchemaProps(in.Spec.Validation.JSONSchema, out); err != nil {
		return err
	}
	return nil
}

func convertJSONSchemaProps(in *apiextensions.JSONSchemaProps, out *spec.Schema) error {
	if in != nil {
		out.ID = in.ID
		out.Schema = spec.SchemaURL(in.Schema)
		out.Description = in.Description
		out.Type = spec.StringOrArray(in.Type)
		out.Format = in.Format
		out.Title = in.Title
		out.Default = in.Default
		out.ExclusiveMaximum = in.ExclusiveMaximum
		out.Minimum = in.Minimum
		out.ExclusiveMinimum = in.ExclusiveMinimum
		out.MaxLength = in.MaxLength
		out.MinLength = in.MinLength
		out.Pattern = in.Pattern
		out.MaxItems = in.MaxItems
		out.MinItems = in.MinItems
		// disable uniqueItems because it can cause the validation runtime
		// complexity to become quadratic.
		out.UniqueItems = false
		out.MultipleOf = in.MultipleOf
		out.Enum = in.Enum
		out.MaxProperties = in.MaxProperties
		out.MinProperties = in.MinProperties
		out.Required = in.Required
		if err := convertJSONSchemaRef(&in.Ref, &out.Ref); err != nil {
			return err
		}
		if err := convertJSONSchemaPropsOrArray(in.Items, out.Items); err != nil {
			return err
		}
		if err := convertSliceOfJSONSchemaProps(&in.AllOf, &out.AllOf); err != nil {
			return err
		}
		if err := convertSliceOfJSONSchemaProps(&in.OneOf, &out.OneOf); err != nil {
			return err
		}
		if err := convertSliceOfJSONSchemaProps(&in.AnyOf, &out.AnyOf); err != nil {
			return err
		}
		if err := convertJSONSchemaProps(in.Not, out.Not); err != nil {
			return err
		}
		var err error
		out.Properties, err = convertMapOfJSONSchemaProps(in.Properties)
		if err != nil {
			return err
		}
		if err := convertJSONSchemaPropsorBool(in.AdditionalProperties, out.AdditionalProperties); err != nil {
			return err
		}
		out.PatternProperties, err = convertMapOfJSONSchemaProps(in.PatternProperties)
		if err != nil {
			return err
		}
		if err := convertJSONSchemaDependencies(in.Dependencies, out.Dependencies); err != nil {
			return err
		}
		if err := convertJSONSchemaPropsorBool(in.AdditionalItems, out.AdditionalItems); err != nil {
			return err
		}
		out.Definitions, err = convertMapOfJSONSchemaProps(in.Definitions)
		if err != nil {
			return err
		}
	}
	return nil
}

// TODO: Convert the ReferenceURL and ReferencePointer fields
func convertJSONSchemaRef(in *apiextensions.JSONSchemaRef, out *spec.Ref) error {
	if in != nil {
		out.HasFullURL = in.HasFullURL
		out.HasURLPathOnly = in.HasURLPathOnly
		out.HasFragmentOnly = in.HasFragmentOnly
		out.HasFileScheme = in.HasFileScheme
		out.HasFullFilePath = in.HasFullFilePath
	}
	return nil
}

func convertSliceOfJSONSchemaProps(in *[]apiextensions.JSONSchemaProps, out *[]spec.Schema) error {
	if in != nil {
		for _, jsonSchemaProps := range *in {
			schema := spec.Schema{}
			if err := convertJSONSchemaProps(&jsonSchemaProps, &schema); err != nil {
				return err
			}
			*out = append(*out, schema)
		}
	}
	return nil
}

func convertMapOfJSONSchemaProps(in map[string]apiextensions.JSONSchemaProps) (map[string]spec.Schema, error) {
	out := make(map[string]spec.Schema)
	if len(in) != 0 {
		for k, jsonSchemaProps := range in {
			schema := spec.Schema{}
			if err := convertJSONSchemaProps(&jsonSchemaProps, &schema); err != nil {
				return nil, err
			}
			out[k] = schema
		}
	}
	return out, nil
}

func convertJSONSchemaPropsOrArray(in *apiextensions.JSONSchemaPropsOrArray, out *spec.SchemaOrArray) error {
	if in != nil {
		if err := convertJSONSchemaProps(in.Schema, out.Schema); err != nil {
			return err
		}
		if err := convertSliceOfJSONSchemaProps(&in.JSONSchemas, &out.Schemas); err != nil {
			return err
		}
	}
	return nil
}

func convertJSONSchemaPropsorBool(in *apiextensions.JSONSchemaPropsOrBool, out *spec.SchemaOrBool) error {
	if in != nil {
		// always allow additionalProperties
		out.Allows = true
		if err := convertJSONSchemaProps(in.Schema, out.Schema); err != nil {
			return err
		}
	}
	return nil
}

func convertJSONSchemaDependencies(in apiextensions.JSONSchemaDependencies, out spec.Dependencies) error {
	if in != nil {
		for k, v := range in {
			schemaOrArray := spec.SchemaOrStringArray{}
			if err := convertJSONSchemaPropsOrStringArray(&v, &schemaOrArray); err != nil {
				return err
			}
			out[k] = schemaOrArray
		}
	}
	return nil
}

func convertJSONSchemaPropsOrStringArray(in *apiextensions.JSONSchemaPropsOrStringArray, out *spec.SchemaOrStringArray) error {
	if in != nil {
		out.Property = in.Property
		if err := convertJSONSchemaProps(in.Schema, out.Schema); err != nil {
			return err
		}
	}
	return nil
}
