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
	allErrs = append(allErrs, ValidateCustomResourceDefinitionSchema(spec.Validation.JSONSchema, fldPath.Child("Validation.JSONSchema"))...)

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

// ValidateCustomResourceDefinitionSchema statically validates
func ValidateCustomResourceDefinitionSchema(JSONSchema *apiextensions.JSONSchemaProps, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if JSONSchema == nil {
		return allErrs
	}

	if JSONSchema.UniqueItems == true {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("UniqueItems"), "uniqueItems cannot be set to true since the runtime complexity becomes quadratic"))
	}

	if JSONSchema.AdditionalProperties != nil {
		if JSONSchema.AdditionalProperties.Allows == false {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("AdditionalProperties"), "additionalProperties cannot be set to false"))
		}
		allErrs = append(allErrs, ValidateCustomResourceDefinitionSchema(JSONSchema.AdditionalProperties.Schema, fldPath.Child("AdditionalProperties"))...)
	}

	if JSONSchema.Ref != "" {
		openapiRef, err := spec.NewRef(JSONSchema.Ref)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("Ref"), JSONSchema.Ref, err.Error()))
		}
		if !openapiRef.IsValidURI() {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("Ref"), JSONSchema.Ref, "Ref does not point to a valid URI."))
		}
	}

	if JSONSchema.AdditionalItems != nil {
		allErrs = append(allErrs, ValidateCustomResourceDefinitionSchema(JSONSchema.AdditionalItems.Schema, fldPath.Child("AdditionalItems"))...)
	}

	allErrs = append(allErrs, ValidateCustomResourceDefinitionSchema(JSONSchema.Not, fldPath.Child("Not"))...)

	if len(JSONSchema.AllOf) != 0 {
		for _, jsonSchema := range JSONSchema.AllOf {
			allErrs = append(allErrs, ValidateCustomResourceDefinitionSchema(&jsonSchema, fldPath.Child("AllOf"))...)
		}
	}

	if len(JSONSchema.OneOf) != 0 {
		for _, jsonSchema := range JSONSchema.OneOf {
			allErrs = append(allErrs, ValidateCustomResourceDefinitionSchema(&jsonSchema, fldPath.Child("OneOf"))...)
		}
	}

	if len(JSONSchema.AnyOf) != 0 {
		for _, jsonSchema := range JSONSchema.AnyOf {
			allErrs = append(allErrs, ValidateCustomResourceDefinitionSchema(&jsonSchema, fldPath.Child("AnyOf"))...)
		}
	}

	if len(JSONSchema.Properties) != 0 {
		for property, jsonSchema := range JSONSchema.Properties {
			allErrs = append(allErrs, ValidateCustomResourceDefinitionSchema(&jsonSchema, fldPath.Child("Properties."+property))...)
		}
	}

	if len(JSONSchema.PatternProperties) != 0 {
		for property, jsonSchema := range JSONSchema.PatternProperties {
			allErrs = append(allErrs, ValidateCustomResourceDefinitionSchema(&jsonSchema, fldPath.Child("PatternProperties."+property))...)
		}
	}

	if len(JSONSchema.Definitions) != 0 {
		for definition, jsonSchema := range JSONSchema.Definitions {
			allErrs = append(allErrs, ValidateCustomResourceDefinitionSchema(&jsonSchema, fldPath.Child("Definitions."+definition))...)
		}
	}

	if JSONSchema.Items != nil {
		allErrs = append(allErrs, ValidateCustomResourceDefinitionSchema(JSONSchema.Items.Schema, fldPath.Child("Items"))...)
		if len(JSONSchema.Items.JSONSchemas) != 0 {
			for _, jsonSchema := range JSONSchema.Items.JSONSchemas {
				allErrs = append(allErrs, ValidateCustomResourceDefinitionSchema(&jsonSchema, fldPath.Child("Items"))...)
			}
		}
	}

	if JSONSchema.Dependencies != nil {
		for dependency, jsonSchemaPropsOrStringArray := range JSONSchema.Dependencies {
			allErrs = append(allErrs, ValidateCustomResourceDefinitionSchema(jsonSchemaPropsOrStringArray.Schema, fldPath.Child("Dependencies.."+dependency))...)
		}
	}

	return allErrs
}
