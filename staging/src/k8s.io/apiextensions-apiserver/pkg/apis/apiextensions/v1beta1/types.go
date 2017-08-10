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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/jsonpath"
)

// CustomResourceDefinitionSpec describes how a user wants their resource to appear
type CustomResourceDefinitionSpec struct {
	// Group is the group this resource belongs in
	Group string `json:"group" protobuf:"bytes,1,opt,name=group"`
	// Version is the version this resource belongs in
	Version string `json:"version" protobuf:"bytes,2,opt,name=version"`
	// Names are the names used to describe this custom resource
	Names CustomResourceDefinitionNames `json:"names" protobuf:"bytes,3,opt,name=names"`

	// Scope indicates whether this resource is cluster or namespace scoped.  Default is namespaced
	Scope ResourceScope `json:"scope" protobuf:"bytes,4,opt,name=scope,casttype=ResourceScope"`
	// Selector is a label query over CustomResources that should match the Replicas count.
	Selector *metav1.LabelSelector `json:"selector,omitempty"`
	// SubResources describes the sub-resources for CustomResources
	SubResources CustomResourceSubResourceType `json:"subResources,omitempty"`
}

// CustomResourceDefinitionNames indicates the names to serve this CustomResourceDefinition
type CustomResourceDefinitionNames struct {
	// Plural is the plural name of the resource to serve.  It must match the name of the CustomResourceDefinition-registration
	// too: plural.group and it must be all lowercase.
	Plural string `json:"plural" protobuf:"bytes,1,opt,name=plural"`
	// Singular is the singular name of the resource.  It must be all lowercase  Defaults to lowercased <kind>
	Singular string `json:"singular,omitempty" protobuf:"bytes,2,opt,name=singular"`
	// ShortNames are short names for the resource.  It must be all lowercase.
	ShortNames []string `json:"shortNames,omitempty" protobuf:"bytes,3,opt,name=shortNames"`
	// Kind is the serialized kind of the resource.  It is normally CamelCase and singular.
	Kind string `json:"kind" protobuf:"bytes,4,opt,name=kind"`
	// ListKind is the serialized kind of the list for this resource.  Defaults to <kind>List.
	ListKind string `json:"listKind,omitempty" protobuf:"bytes,5,opt,name=listKind"`
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
	Type CustomResourceDefinitionConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=CustomResourceDefinitionConditionType"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=ConditionStatus"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,3,opt,name=lastTransitionTime"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,5,opt,name=message"`
}

// CustomResourceDefinitionStatus indicates the state of the CustomResourceDefinition
type CustomResourceDefinitionStatus struct {
	// Conditions indicate state for particular aspects of a CustomResourceDefinition
	Conditions []CustomResourceDefinitionCondition `json:"conditions" protobuf:"bytes,1,opt,name=conditions"`

	// AcceptedNames are the names that are actually being used to serve discovery
	// They may be different than the names in spec.
	AcceptedNames CustomResourceDefinitionNames `json:"acceptedNames" protobuf:"bytes,2,opt,name=acceptedNames"`
}

// CustomResourceCleanupFinalizer is the name of the finalizer which will delete instances of
// a CustomResourceDefinition
const CustomResourceCleanupFinalizer = "customresourcecleanup.apiextensions.k8s.io"

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CustomResourceDefinition represents a resource that should be exposed on the API server.  Its name MUST be in the format
// <.spec.name>.<.spec.group>.
type CustomResourceDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec describes how the user wants the resources to appear
	Spec CustomResourceDefinitionSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	// Status indicates the actual state of the CustomResourceDefinition
	Status CustomResourceDefinitionStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CustomResourceDefinitionList is a list of CustomResourceDefinition objects.
type CustomResourceDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items individual CustomResourceDefinitions
	Items []CustomResourceDefinition `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// CustomResourceSubResourceType defines the status and scale sub-resources for CRs
type CustomResourceSubResourceType struct {
	// Status denotes the status subresource for CustomResources
	Status *CustomResourceSubResourceStatus `json:"status,omitempty"`
	// Scale denotes the scale subresource for CustomResources
	Scale *CustomResourceSubResourceStatus `json:"scale,omitempty"`
}

// CustomResourceSubResourceStatus serves the HTTP path under <CR Name>/status
type CustomResourceSubResourceStatus struct {
	// Transfer the whole object over the wire, but only allow mutation of the
	// given JSON path. JSON path into a CustomResource i.e. “.status”, without
	// any array notation.
	Path jsonpath.JSONPath `json:"path,omitempty"`
}

type CustomResourceSubResourceScale struct {
	// required, e.g. “.spec.replicas”
	SpecReplicasPath jsonpath.JSONPath `json:"specReplicasPath,omitempty"`
	// optional, e.g. “.status.replicas”
	StatusReplicasPath jsonpath.JSONPath `json:"statusReplicasPath,omitempty"`
	// optional, e.g. “.spec.labelSelector”
	SelectorPath jsonpath.JSONPath `json:"selectorPath,omitempty"`
}

// The following is the payload to send over the wire for /scale. It happens
// to be defined here in apiextensions.k8s.io because we don’t have a global
// Scale type in meta/v1. Ref: https://github.com/kubernetes/kubernetes/issues/49504

// Scale represents a scaling request for a resource.
type Scale struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata; More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// defines the behavior of the scale. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status.
	// +optional
	Spec ScaleSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// current status of the scale. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status. Read-only.
	// +optional
	Status ScaleStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// ScaleSpec describes the attributes of a scale subresource
type ScaleSpec struct {
	// desired number of instances for the scaled object.
	// +optional
	Replicas int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`
}

// ScaleStatus represents the current status of a scale subresource.
type ScaleStatus struct {
	// actual number of observed instances of the scaled object.
	Replicas int32 `json:"replicas" protobuf:"varint,1,opt,name=replicas"`

	// label query over pods that should match the replicas count.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
	// +optional
	Selector *metav1.LabelSelector
}
