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
	"fmt"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	genericvalidation "k8s.io/apimachinery/pkg/api/validation"
	validationutil "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

// ValidateCustomResourceDefinition statically validates
func ValidateCustomResourceDefinition(obj *apiextensions.CustomResourceDefinition) field.ErrorList {
	nameValidationFn := func(name string, prefix bool) []string {
		ret := genericvalidation.NameIsDNSSubdomain(name, prefix)
		requiredName := obj.Spec.Names.Plural + "." + obj.Spec.Group
		if name != requiredName {
			ret = append(ret, fmt.Sprintf(`must be spec.names.plural+"."+spec.group`))
		}
		return ret
	}

	allErrs := genericvalidation.ValidateObjectMeta(&obj.ObjectMeta, false, nameValidationFn, field.NewPath("metadata"))
	allErrs = append(allErrs, ValidateCustomResourceDefinitionSpec(&obj.Spec, field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidateCustomResourceDefinitionStatus(&obj.Status, field.NewPath("status"))...)
	return allErrs
}

// ValidateCustomResourceDefinitionUpdate statically validates
func ValidateCustomResourceDefinitionUpdate(obj, oldObj *apiextensions.CustomResourceDefinition) field.ErrorList {
	allErrs := genericvalidation.ValidateObjectMetaUpdate(&obj.ObjectMeta, &oldObj.ObjectMeta, field.NewPath("metadata"))
	allErrs = append(allErrs, ValidateCustomResourceDefinitionSpecUpdate(&obj.Spec, &oldObj.Spec, apiextensions.IsCRDConditionTrue(oldObj, apiextensions.Established), field.NewPath("spec"))...)
	allErrs = append(allErrs, ValidateCustomResourceDefinitionStatus(&obj.Status, field.NewPath("status"))...)
	return allErrs
}

// ValidateUpdateCustomResourceDefinitionStatus statically validates
func ValidateUpdateCustomResourceDefinitionStatus(obj, oldObj *apiextensions.CustomResourceDefinition) field.ErrorList {
	allErrs := genericvalidation.ValidateObjectMetaUpdate(&obj.ObjectMeta, &oldObj.ObjectMeta, field.NewPath("metadata"))
	allErrs = append(allErrs, ValidateCustomResourceDefinitionStatus(&obj.Status, field.NewPath("status"))...)
	return allErrs
}

// ValidateCustomResourceDefinitionSpec statically validates
func ValidateCustomResourceDefinitionSpec(spec *apiextensions.CustomResourceDefinitionSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(spec.Group) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("group"), ""))
	} else if errs := validationutil.IsDNS1123Subdomain(spec.Group); len(errs) > 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("group"), spec.Group, strings.Join(errs, ",")))
	} else if len(strings.Split(spec.Group, ".")) < 2 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("group"), spec.Group, "should be a domain with at least one dot"))
	}

	if len(spec.Version) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("version"), ""))
	} else if errs := validationutil.IsDNS1035Label(spec.Version); len(errs) > 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("version"), spec.Version, strings.Join(errs, ",")))
	}

	switch spec.Scope {
	case "":
		allErrs = append(allErrs, field.Required(fldPath.Child("scope"), ""))
	case apiextensions.ClusterScoped, apiextensions.NamespaceScoped:
	default:
		allErrs = append(allErrs, field.NotSupported(fldPath.Child("scope"), spec.Scope, []string{string(apiextensions.ClusterScoped), string(apiextensions.NamespaceScoped)}))
	}

	// in addition to the basic name restrictions, some names are required for spec, but not for status
	if len(spec.Names.Plural) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("names", "plural"), ""))
	}
	if len(spec.Names.Singular) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("names", "singular"), ""))
	}
	if len(spec.Names.Kind) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("names", "kind"), ""))
	}
	if len(spec.Names.ListKind) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("names", "listKind"), ""))
	}

	allErrs = append(allErrs, ValidateCustomResourceDefinitionNames(&spec.Names, fldPath.Child("names"))...)

	return allErrs
}

// ValidateCustomResourceDefinitionSpecUpdate statically validates
func ValidateCustomResourceDefinitionSpecUpdate(spec, oldSpec *apiextensions.CustomResourceDefinitionSpec, established bool, fldPath *field.Path) field.ErrorList {
	allErrs := ValidateCustomResourceDefinitionSpec(spec, fldPath)

	if established {
		// these effect the storage and cannot be changed therefore
		allErrs = append(allErrs, genericvalidation.ValidateImmutableField(spec.Version, oldSpec.Version, fldPath.Child("version"))...)
		allErrs = append(allErrs, genericvalidation.ValidateImmutableField(spec.Scope, oldSpec.Scope, fldPath.Child("scope"))...)
		allErrs = append(allErrs, genericvalidation.ValidateImmutableField(spec.Names.Kind, oldSpec.Names.Kind, fldPath.Child("names", "kind"))...)
	}

	// these affects the resource name, which is always immutable, so this can't be updated.
	allErrs = append(allErrs, genericvalidation.ValidateImmutableField(spec.Group, oldSpec.Group, fldPath.Child("group"))...)
	allErrs = append(allErrs, genericvalidation.ValidateImmutableField(spec.Names.Plural, oldSpec.Names.Plural, fldPath.Child("names", "plural"))...)

	return allErrs
}

// ValidateCustomResourceDefinitionStatus statically validates
func ValidateCustomResourceDefinitionStatus(status *apiextensions.CustomResourceDefinitionStatus, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, ValidateCustomResourceDefinitionNames(&status.AcceptedNames, fldPath.Child("acceptedNames"))...)
	return allErrs
}

// ValidateCustomResourceDefinitionNames statically validates
func ValidateCustomResourceDefinitionNames(names *apiextensions.CustomResourceDefinitionNames, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if errs := validationutil.IsDNS1035Label(names.Plural); len(names.Plural) > 0 && len(errs) > 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("plural"), names.Plural, strings.Join(errs, ",")))
	}
	if errs := validationutil.IsDNS1035Label(names.Singular); len(names.Singular) > 0 && len(errs) > 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("singular"), names.Singular, strings.Join(errs, ",")))
	}
	if errs := validationutil.IsDNS1035Label(strings.ToLower(names.Kind)); len(names.Kind) > 0 && len(errs) > 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("kind"), names.Kind, "may have mixed case, but should otherwise match: "+strings.Join(errs, ",")))
	}
	if errs := validationutil.IsDNS1035Label(strings.ToLower(names.ListKind)); len(names.ListKind) > 0 && len(errs) > 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("listKind"), names.ListKind, "may have mixed case, but should otherwise match: "+strings.Join(errs, ",")))
	}

	for i, shortName := range names.ShortNames {
		if errs := validationutil.IsDNS1035Label(shortName); len(errs) > 0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("shortNames").Index(i), shortName, strings.Join(errs, ",")))
		}

	}

	// kind and listKind may not be the same or parsing become ambiguous
	if len(names.Kind) > 0 && names.Kind == names.ListKind {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("listKind"), names.ListKind, "kind and listKind may not be the same"))
	}

	return allErrs
}

func ValidateCustomResource(customresource interface{}, crd apiextensionsv1beta1.CustomResourceDefinition) error {
	schema := spec.Schema{}
	if err := convertToOpenAPITypes(&crd, &schema); err != nil {
		return err
	}

	if err := spec.ExpandSchema(&schema, nil, nil); err != nil {
		return err
	}

	validator := validate.NewSchemaValidator(&schema, nil, "spec", strfmt.Default)
	result := validator.Validate(customresource)
	if result.AsError() != nil {
		return result.AsError()
	}
	return nil
}

func convertToOpenAPITypes(in *apiextensionsv1beta1.CustomResourceDefinition, out *spec.Schema) error {
	if in.Spec.Validation != nil {
		if err := convertJSONSchemaProps(in.Spec.Validation.JSONSchema, out); err != nil {
			return err
		}
	}
	return nil
}

func convertJSONSchemaProps(in *apiextensionsv1beta1.JSONSchemaProps, out *spec.Schema) error {
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
func convertJSONSchemaRef(in *apiextensionsv1beta1.JSONSchemaRef, out *spec.Ref) error {
	if in != nil {
		out.HasFullURL = in.HasFullURL
		out.HasURLPathOnly = in.HasURLPathOnly
		out.HasFragmentOnly = in.HasFragmentOnly
		out.HasFileScheme = in.HasFileScheme
		out.HasFullFilePath = in.HasFullFilePath
	}
	return nil
}

func convertSliceOfJSONSchemaProps(in *[]apiextensionsv1beta1.JSONSchemaProps, out *[]spec.Schema) error {
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

func convertMapOfJSONSchemaProps(in map[string]apiextensionsv1beta1.JSONSchemaProps) (map[string]spec.Schema, error) {
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

func convertJSONSchemaPropsOrArray(in *apiextensionsv1beta1.JSONSchemaPropsOrArray, out *spec.SchemaOrArray) error {
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

func convertJSONSchemaPropsorBool(in *apiextensionsv1beta1.JSONSchemaPropsOrBool, out *spec.SchemaOrBool) error {
	if in != nil {
		// always allow additionalProperties
		out.Allows = true
		if err := convertJSONSchemaProps(in.Schema, out.Schema); err != nil {
			return err
		}
	}
	return nil
}

func convertJSONSchemaDependencies(in apiextensionsv1beta1.JSONSchemaDependencies, out spec.Dependencies) error {
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

func convertJSONSchemaPropsOrStringArray(in *apiextensionsv1beta1.JSONSchemaPropsOrStringArray, out *spec.SchemaOrStringArray) error {
	if in != nil {
		out.Property = in.Property
		if err := convertJSONSchemaProps(in.Schema, out.Schema); err != nil {
			return err
		}
	}
	return nil
}
