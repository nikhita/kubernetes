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

package apiextensions

import (
	"net/url"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CustomResourceDefinitionSpec describes how a user wants their resource to appear
type CustomResourceDefinitionSpec struct {
	// Group is the group this resource belongs in
	Group string
	// Version is the version this resource belongs in
	Version string
	// Names are the names used to describe this custom resource
	Names CustomResourceDefinitionNames
	// Scope indicates whether this resource is cluster or namespace scoped.  Default is namespaced
	Scope ResourceScope
	// Validation describes the validation methods for CustomResources
	Validation *CustomResourceValidation `json:"validation,omitempty"`
}

// CustomResourceDefinitionNames indicates the names to serve this CustomResourceDefinition
type CustomResourceDefinitionNames struct {
	// Plural is the plural name of the resource to serve.  It must match the name of the CustomResourceDefinition-registration
	// too: plural.group and it must be all lowercase.
	Plural string
	// Singular is the singular name of the resource.  It must be all lowercase  Defaults to lowercased <kind>
	Singular string
	// ShortNames are short names for the resource.  It must be all lowercase.
	ShortNames []string
	// Kind is the serialized kind of the resource.  It is normally CamelCase and singular.
	Kind string
	// ListKind is the serialized kind of the list for this resource.  Defaults to <kind>List.
	ListKind string
}

// ResourceScope is an enum defining the different scopes availabe to a custom resource
type ResourceScope string

const (
	ClusterScoped   ResourceScope = "Cluster"
	NamespaceScoped ResourceScope = "Namespaced"
)

type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in the condition.
// "ConditionFalse" means a resource is not in the condition. "ConditionUnknown" means kubernetes
// can't decide if a resource is in the condition or not. In the future, we could add other
// intermediate conditions, e.g. ConditionDegraded.
const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

// CustomResourceDefinitionConditionType is a valid value for CustomResourceDefinitionCondition.Type
type CustomResourceDefinitionConditionType string

const (
	// Established means that the resource has become active. A resource is established when all names are
	// accepted without a conflict for the first time. A resource stays established until deleted, even during
	// a later NamesAccepted due to changed names. Note that not all names can be changed.
	Established CustomResourceDefinitionConditionType = "Established"
	// NamesAccepted means the names chosen for this CustomResourceDefinition do not conflict with others in
	// the group and are therefore accepted.
	NamesAccepted CustomResourceDefinitionConditionType = "NamesAccepted"
	// Terminating means that the CustomResourceDefinition has been deleted and is cleaning up.
	Terminating CustomResourceDefinitionConditionType = "Terminating"
)

// CustomResourceDefinitionCondition contains details for the current condition of this pod.
type CustomResourceDefinitionCondition struct {
	// Type is the type of the condition.
	Type CustomResourceDefinitionConditionType
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status ConditionStatus
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time
	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string
	// Human-readable message indicating details about last transition.
	// +optional
	Message string
}

// CustomResourceDefinitionStatus indicates the state of the CustomResourceDefinition
type CustomResourceDefinitionStatus struct {
	// Conditions indicate state for particular aspects of a CustomResourceDefinition
	Conditions []CustomResourceDefinitionCondition

	// AcceptedNames are the names that are actually being used to serve discovery
	// They may be different than the names in spec.
	AcceptedNames CustomResourceDefinitionNames
}

// CustomResourceCleanupFinalizer is the name of the finalizer which will delete instances of
// a CustomResourceDefinition
const CustomResourceCleanupFinalizer = "customresourcecleanup.apiextensions.k8s.io"

// +genclient=true
// +nonNamespaced=true

// CustomResourceDefinition represents a resource that should be exposed on the API server.  Its name MUST be in the format
// <.spec.name>.<.spec.group>.
type CustomResourceDefinition struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	// Spec describes how the user wants the resources to appear
	Spec CustomResourceDefinitionSpec
	// Status indicates the actual state of the CustomResourceDefinition
	Status CustomResourceDefinitionStatus
}

// CustomResourceDefinitionList is a list of CustomResourceDefinition objects.
type CustomResourceDefinitionList struct {
	metav1.TypeMeta
	metav1.ListMeta

	// Items individual CustomResourceDefinitions
	Items []CustomResourceDefinition
}

// CustomResourceValidation is a list of validation methods for CustomResources
type CustomResourceValidation struct {
	// JSONSchema is the JSON Schema to be validated against.
	// Can add other validation methods later if needed.
	JSONSchema *JSONSchemaProps `json:"jsonSchema,omitempty"`
}

// JSONSchemaProps is a JSON-Schema following Specification Draft 4 (http://json-schema.org/).
type JSONSchemaProps struct {
	ID               string        `json:"id,omitempty"`
	Schema           JSONSchemaURL `json:"-,omitempty"`
	Ref              JSONSchemaRef `json:"-,omitempty"`
	Description      string        `json:"description,omitempty"`
	Type             StringOrArray `json:"type,omitempty"`
	Format           string        `json:"format,omitempty"`
	Title            string        `json:"title,omitempty"`
	Default          interface{}   `json:"default,omitempty"`
	Maximum          *float64      `json:"maximum,omitempty"`
	ExclusiveMaximum bool          `json:"exclusiveMaximum,omitempty"`
	Minimum          *float64      `json:"minimum,omitempty"`
	ExclusiveMinimum bool          `json:"exclusiveMinimum,omitempty"`
	MaxLength        *int64        `json:"maxLength,omitempty"`
	MinLength        *int64        `json:"minLength,omitempty"`
	Pattern          string        `json:"pattern,omitempty"`
	MaxItems         *int64        `json:"maxItems,omitempty"`
	MinItems         *int64        `json:"minItems,omitempty"`
	// disable uniqueItems for now because it can cause the validation runtime
	// complexity to become quadratic.
	UniqueItems          bool                       `json:"uniqueItems,omitempty"`
	MultipleOf           *float64                   `json:"multipleOf,omitempty"`
	Enum                 []interface{}              `json:"enum,omitempty"`
	MaxProperties        *int64                     `json:"maxProperties,omitempty"`
	MinProperties        *int64                     `json:"minProperties,omitempty"`
	Required             []string                   `json:"required,omitempty"`
	Items                *JSONSchemaPropsOrArray    `json:"items,omitempty"`
	AllOf                []JSONSchemaProps          `json:"allOf,omitempty"`
	OneOf                []JSONSchemaProps          `json:"oneOf,omitempty"`
	AnyOf                []JSONSchemaProps          `json:"anyOf,omitempty"`
	Not                  *JSONSchemaProps           `json:"not,omitempty"`
	Properties           map[string]JSONSchemaProps `json:"properties,omitempty"`
	AdditionalProperties *JSONSchemaPropsOrBool     `json:"additionalProperties,omitempty"`
	PatternProperties    map[string]JSONSchemaProps `json:"patternProperties,omitempty"`
	Dependencies         JSONSchemaDependencies     `json:"dependencies,omitempty"`
	AdditionalItems      *JSONSchemaPropsOrBool     `json:"additionalItems,omitempty"`
	Definitions          JSONSchemaDefinitions      `json:"definitions,omitempty"`
}

// JSONSchemaRef represents a JSON reference that is potentially resolved.
// It is marshaled into a string using a custom JSON marshaller.
type JSONSchemaRef struct {
	ReferenceURL     *url.URL
	ReferencePointer JSONSchemaPointer
	HasFullURL       bool
	HasURLPathOnly   bool
	HasFragmentOnly  bool
	HasFileScheme    bool
	HasFullFilePath  bool
}

// JSONSchemaPointer is the JSON pointer representation.
type JSONSchemaPointer struct {
	ReferenceTokens []string
}

// JSONSchemaURL represents a schema url. Defaults to JSON Schema Specification Draft 4.
type JSONSchemaURL string

const (
	// JSONSchemaDraft4URL is the url for JSON Schema Specification Draft 4.
	JSONSchemaDraft4URL JSONSchemaURL = "http://json-schema.org/draft-04/schema#"
)

// StringOrArray represents a value that can either be a string or an array of strings.
// Mainly here for serialization purposes.
type StringOrArray []string

// JSONSchemaPropsOrArray represents a value that can either be a JSONSchemaProps
// or an array of JSONSchemaProps. Mainly here for serialization purposes.
type JSONSchemaPropsOrArray struct {
	Schema      *JSONSchemaProps
	JSONSchemas []JSONSchemaProps
}

// JSONSchemaPropsOrBool represents JSONSchemaProps or a boolean value.
// Defaults to true for the boolean property.
type JSONSchemaPropsOrBool struct {
	Allows bool
	Schema *JSONSchemaProps
}

// JSONSchemaDependencies represent a dependencies property.
type JSONSchemaDependencies map[string]JSONSchemaPropsOrStringArray

// JSONSchemaPropsOrStringArray represents a JSONSchemaProps or a string array.
type JSONSchemaPropsOrStringArray struct {
	Schema   *JSONSchemaProps
	Property []string
}

// JSONSchemaDefinitions contains the models explicitly defined in this spec.
type JSONSchemaDefinitions map[string]JSONSchemaProps
