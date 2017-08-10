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

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	genericvalidation "k8s.io/apimachinery/pkg/api/validation"
	validationutil "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
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
	allErrs = append(allErrs, ValidateCustomResourceDefinitionValidation(&spec.Validation, fldPath.Child("validation"))...)

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

// additionalSpecValidator applies validations for different OpenAPI specfication versions.
type additionalSpecValidator interface {
	validateCustomResourceDefinitionOpenAPISchema(spec *apiextensions.JSONSchemaProps, fldPath *field.Path) field.ErrorList
}

// ValidateCustomResourceDefinitionValidation statically validates
func ValidateCustomResourceDefinitionValidation(customResourceValidation *apiextensions.CustomResourceValidation, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if customResourceValidation.OpenAPIV2Schema != nil && customResourceValidation.OpenAPIV3Schema != nil {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child(""), "only one openAPI schema can be specified."))
	}

	if customResourceValidation.OpenAPIV2Schema != nil {
		openAPIV2Schema := &additionalSpecValidatorV2{}
		allErrs = append(allErrs, ValidateCustomResourceDefinitionOpenAPISchema(customResourceValidation.OpenAPIV2Schema, fldPath.Child("openAPIV2Schema"), openAPIV2Schema)...)
	}

	if customResourceValidation.OpenAPIV3Schema != nil {
		openAPIV3Schema := &additionalSpecValidatorV3{}
		allErrs = append(allErrs, ValidateCustomResourceDefinitionOpenAPISchema(customResourceValidation.OpenAPIV3Schema, fldPath.Child("openAPIV3Schema"), openAPIV3Schema)...)
	}

	return allErrs
}

// ValidateCustomResourceDefinitionOpenAPISchema statically validates
func ValidateCustomResourceDefinitionOpenAPISchema(schema *apiextensions.JSONSchemaProps, fldPath *field.Path, specVersion additionalSpecValidator) field.ErrorList {
	allErrs := field.ErrorList{}

	if schema == nil {
		return allErrs
	}

	allErrs = append(allErrs, specVersion.validateCustomResourceDefinitionOpenAPISchema(schema, fldPath)...)

	if schema.UniqueItems == true {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("uniqueItems"), "uniqueItems cannot be set to true since the runtime complexity becomes quadratic"))
	}

	if schema.AdditionalProperties != nil {
		if schema.AdditionalProperties.Allows == false {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("additionalProperties"), "additionalProperties cannot be set to false"))
		}
		allErrs = append(allErrs, ValidateCustomResourceDefinitionOpenAPISchema(schema.AdditionalProperties.Schema, fldPath.Child("AdditionalProperties"), specVersion)...)
	}

	if schema.Ref != nil {
		openapiRef, err := spec.NewRef(*schema.Ref)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("ref"), *schema.Ref, err.Error()))
		}

		if !openapiRef.IsValidURI() {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("ref"), *schema.Ref, "ref does not point to a valid URI."))
		}
	}

	allErrs = append(allErrs, ValidateCustomResourceDefinitionOpenAPISchema(schema.Not, fldPath.Child("not"), specVersion)...)

	if len(schema.AllOf) != 0 {
		for _, jsonSchema := range schema.AllOf {
			allErrs = append(allErrs, ValidateCustomResourceDefinitionOpenAPISchema(&jsonSchema, fldPath.Child("allOf"), specVersion)...)
		}
	}

	if len(schema.OneOf) != 0 {
		for _, jsonSchema := range schema.OneOf {
			allErrs = append(allErrs, ValidateCustomResourceDefinitionOpenAPISchema(&jsonSchema, fldPath.Child("oneOf"), specVersion)...)
		}
	}

	if len(schema.AnyOf) != 0 {
		for _, jsonSchema := range schema.AnyOf {
			allErrs = append(allErrs, ValidateCustomResourceDefinitionOpenAPISchema(&jsonSchema, fldPath.Child("anyOf"), specVersion)...)
		}
	}

	if len(schema.Properties) != 0 {
		for property, jsonSchema := range schema.Properties {
			allErrs = append(allErrs, ValidateCustomResourceDefinitionOpenAPISchema(&jsonSchema, fldPath.Child("properties."+property), specVersion)...)
		}
	}

	if schema.Items != nil {
		allErrs = append(allErrs, ValidateCustomResourceDefinitionOpenAPISchema(schema.Items.Schema, fldPath.Child("items"), specVersion)...)
		if len(schema.Items.JSONSchemas) != 0 {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("items"), "items must be a schema object and not an array"))
		}
	}

	return allErrs
}

type additionalSpecValidatorV2 struct{}

// validateCustomResourceDefinitionOpenAPISchema validates against OpenAPI Schema v2.
func (v *additionalSpecValidatorV2) validateCustomResourceDefinitionOpenAPISchema(schema *apiextensions.JSONSchemaProps, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if schema == nil {
		return allErrs
	}

	if len(schema.OneOf) != 0 {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("oneOf"), "oneOf is not supported in OpenAPI Spec v2."))
	}

	if len(schema.AnyOf) != 0 {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("anyOf"), "anyOf is not supported in OpenAPI Spec v2."))
	}

	if schema.Not != nil {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("not"), "not is not supported in OpenAPI Spec v2."))
	}

	// all restrictions for v3 also apply to v2
	openAPIV3Schema := &additionalSpecValidatorV3{}
	allErrs = append(allErrs, openAPIV3Schema.validateCustomResourceDefinitionOpenAPISchema(schema, fldPath)...)

	return allErrs
}

type additionalSpecValidatorV3 struct{}

// validateCustomResourceDefinitionOpenAPISchema validates against OpenAPI Schema v3.
func (v *additionalSpecValidatorV3) validateCustomResourceDefinitionOpenAPISchema(schema *apiextensions.JSONSchemaProps, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if schema == nil {
		return allErrs
	}

	if schema.Default != nil {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("default"), "default is not supported."))
	}

	if schema.ID != "" {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("id"), "id is not supported."))
	}

	if schema.AdditionalItems != nil {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("additionalItems"), "additionalItems is not supported."))
	}

	if len(schema.PatternProperties) != 0 {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("patternProperties"), "patternProperties is not supported."))
	}

	if len(schema.Definitions) != 0 {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("definitions"), "definitions is not supported."))
	}

	if schema.Dependencies != nil {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("dependencies"), "dependencies is not supported."))
	}

	if len(schema.Type) != 0 {
		if len(schema.Type) > 1 {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("type"), "multiple values via an array for type is not supported."))
		}
		if schema.Type[0] == "null" {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("type"), "type cannot be set to null."))
		}
	}

	if schema.Items != nil && len(schema.Items.JSONSchemas) != 0 {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("items"), "items must be a schema object and not an array"))
	}

	return allErrs
}
