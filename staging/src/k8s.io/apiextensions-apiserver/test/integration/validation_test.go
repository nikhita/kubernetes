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

package integration

import (
	"strings"
	"testing"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/test/integration/testserver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestForProperValidationErrors(t *testing.T) {
	stopCh, apiExtensionClient, clientPool, err := testserver.StartDefaultServer()
	if err != nil {
		t.Fatal(err)
	}
	defer close(stopCh)

	noxuDefinition := testserver.NewNoxuCustomResourceDefinition(apiextensionsv1beta1.NamespaceScoped)
	noxuVersionClient, err := testserver.CreateNewCustomResourceDefinition(noxuDefinition, apiExtensionClient, clientPool)
	if err != nil {
		t.Fatal(err)
	}

	ns := "not-the-default"
	noxuResourceClient := NewNamespacedCustomResourceClient(ns, noxuVersionClient, noxuDefinition)

	tests := []struct {
		name          string
		instanceFn    func() *unstructured.Unstructured
		expectedError string
	}{
		{
			name: "bad version",
			instanceFn: func() *unstructured.Unstructured {
				instance := testserver.NewNoxuInstance(ns, "foo")
				instance.Object["apiVersion"] = "mygroup.example.com/v2"
				return instance
			},
			expectedError: "the API version in the data (mygroup.example.com/v2) does not match the expected API version (mygroup.example.com/v1beta1)",
		},
		{
			name: "bad kind",
			instanceFn: func() *unstructured.Unstructured {
				instance := testserver.NewNoxuInstance(ns, "foo")
				instance.Object["kind"] = "SomethingElse"
				return instance
			},
			expectedError: `SomethingElse.mygroup.example.com "foo" is invalid: kind: Invalid value: "SomethingElse": must be WishIHadChosenNoxu`,
		},
	}

	for _, tc := range tests {
		_, err := noxuResourceClient.Create(tc.instanceFn())
		if err == nil {
			t.Errorf("%v: expected %v", tc.name, tc.expectedError)
			continue
		}
		// this only works when status errors contain the expect kind and version, so this effectively tests serializations too
		if !strings.Contains(err.Error(), tc.expectedError) {
			t.Errorf("%v: expected %v, got %v", tc.name, tc.expectedError, err)
			continue
		}
	}
}

func TestCustomResourceValidation(t *testing.T) {
	stopCh, apiExtensionClient, clientPool, err := testserver.StartDefaultServer()
	if err != nil {
		t.Fatal(err)
	}
	defer close(stopCh)

	noxuDefinition := testserver.NewNoxuValidationCRD(apiextensionsv1beta1.NamespaceScoped)
	noxuVersionClient, err := testserver.CreateNewCustomResourceDefinition(noxuDefinition, apiExtensionClient, clientPool)
	if err != nil {
		t.Fatal(err)
	}

	ns := "not-the-default"
	noxuResourceClient := NewNamespacedCustomResourceClient(ns, noxuVersionClient, noxuDefinition)
	_, err = instantiateCustomResource(t, testserver.NewNoxuValidationInstance(ns, "foo"), noxuResourceClient, noxuDefinition)
	if err != nil {
		t.Fatalf("unable to create noxu Instance:%v", err)
	}
}

func TestCustomResourceUpdateValidation(t *testing.T) {
	stopCh, apiExtensionClient, clientPool, err := testserver.StartDefaultServer()
	if err != nil {
		t.Fatal(err)
	}
	defer close(stopCh)

	noxuDefinition := testserver.NewNoxuValidationCRD(apiextensionsv1beta1.NamespaceScoped)
	noxuVersionClient, err := testserver.CreateNewCustomResourceDefinition(noxuDefinition, apiExtensionClient, clientPool)
	if err != nil {
		t.Fatal(err)
	}

	ns := "not-the-default"
	noxuResourceClient := NewNamespacedCustomResourceClient(ns, noxuVersionClient, noxuDefinition)
	_, err = instantiateCustomResource(t, testserver.NewNoxuValidationInstance(ns, "foo"), noxuResourceClient, noxuDefinition)
	if err != nil {
		t.Fatalf("unable to create noxu Instance:%v", err)
	}

	gottenNoxuInstance, err := noxuResourceClient.Get("foo", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	// invalidate the instance
	gottenNoxuInstance.Object["spec"] = map[string]interface{}{
		"gamma": "bar",
		"delta": "hello",
	}

	_, err = noxuResourceClient.Update(gottenNoxuInstance)
	if err == nil {
		t.Fatalf("unexpected non-error: spec.alpha and spec.beta are required while updating %v", gottenNoxuInstance)
	}
}

func TestCustomResourceValidationErrors(t *testing.T) {
	stopCh, apiExtensionClient, clientPool, err := testserver.StartDefaultServer()
	if err != nil {
		t.Fatal(err)
	}
	defer close(stopCh)

	noxuDefinition := testserver.NewNoxuValidationCRD(apiextensionsv1beta1.NamespaceScoped)
	noxuVersionClient, err := testserver.CreateNewCustomResourceDefinition(noxuDefinition, apiExtensionClient, clientPool)
	if err != nil {
		t.Fatal(err)
	}

	ns := "not-the-default"
	noxuResourceClient := NewNamespacedCustomResourceClient(ns, noxuVersionClient, noxuDefinition)

	tests := []struct {
		name          string
		instanceFn    func() *unstructured.Unstructured
		expectedError string
	}{
		{
			name: "bad alpha",
			instanceFn: func() *unstructured.Unstructured {
				instance := testserver.NewNoxuValidationInstance(ns, "foo")
				instance.Object["spec"] = map[string]interface{}{
					"alpha": "foo_123!",
					"beta":  10,
					"gamma": "bar",
					"delta": "hello",
				}
				return instance
			},
			expectedError: "spec.alpha in body should match '^[a-zA-Z0-9_]*$'",
		},
		{
			name: "bad beta",
			instanceFn: func() *unstructured.Unstructured {
				instance := testserver.NewNoxuValidationInstance(ns, "foo")
				instance.Object["spec"] = map[string]interface{}{
					"alpha": "foo_123",
					"beta":  5,
					"gamma": "bar",
					"delta": "hello",
				}
				return instance
			},
			expectedError: "spec.beta in body should be greater than or equal to 10",
		},
		{
			name: "bad gamma",
			instanceFn: func() *unstructured.Unstructured {
				instance := testserver.NewNoxuValidationInstance(ns, "foo")
				instance.Object["spec"] = map[string]interface{}{
					"alpha": "foo_123",
					"beta":  5,
					"gamma": "qux",
					"delta": "hello",
				}
				return instance
			},
			expectedError: "spec.gamma in body should be one of [foo bar baz]",
		},
		{
			name: "bad delta",
			instanceFn: func() *unstructured.Unstructured {
				instance := testserver.NewNoxuValidationInstance(ns, "foo")
				instance.Object["spec"] = map[string]interface{}{
					"alpha": "foo_123",
					"beta":  5,
					"gamma": "bar",
					"delta": "foobarbaz",
				}
				return instance
			},
			expectedError: "must validate at least one schema (anyOf)\nspec.delta in body should be at most 5 chars long",
		},
		{
			name: "absent alpha and beta",
			instanceFn: func() *unstructured.Unstructured {
				instance := testserver.NewNoxuValidationInstance(ns, "foo")
				instance.Object["spec"] = map[string]interface{}{
					"gamma": "bar",
					"delta": "hello",
				}
				return instance
			},
			expectedError: "spec.alpha in body is required\nspec.beta in body is required",
		},
	}

	for _, tc := range tests {
		_, err := noxuResourceClient.Create(tc.instanceFn())
		if err == nil {
			t.Errorf("%v: expected %v", tc.name, tc.expectedError)
			continue
		}
		// this only works when status errors contain the expect kind and version, so this effectively tests serializations too
		if !strings.Contains(err.Error(), tc.expectedError) {
			t.Errorf("%v: expected %v, got %v", tc.name, tc.expectedError, err)
			continue
		}
	}
}

// TODO: This should pass
func TestCRValidationOnCRDUpdate(t *testing.T) {
	stopCh, apiExtensionClient, clientPool, err := testserver.StartDefaultServer()
	if err != nil {
		t.Fatal(err)
	}
	defer close(stopCh)

	noxuDefinition := testserver.NewNoxuValidationCRD(apiextensionsv1beta1.NamespaceScoped)

	// set stricter schema
	beta := noxuDefinition.Spec.Validation.JSONSchema.Properties["beta"]
	beta.Minimum = apiextensionsv1beta1.Float64Ptr(12)

	noxuVersionClient, err := testserver.CreateNewCustomResourceDefinition(noxuDefinition, apiExtensionClient, clientPool)
	if err != nil {
		t.Fatal(err)
	}
	ns := "not-the-default"
	noxuResourceClient := NewNamespacedCustomResourceClient(ns, noxuVersionClient, noxuDefinition)

	// CR is rejected
	_, err = instantiateCustomResource(t, testserver.NewNoxuValidationInstance(ns, "foo"), noxuResourceClient, noxuDefinition)
	if err == nil {
		t.Fatalf("unexpected non-error: CR should be rejected")
	}

	gottenCRD, err := apiExtensionClient.ApiextensionsV1beta1().CustomResourceDefinitions().Get("noxus.mygroup.example.com", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	// update the CRD to a less stricter schema
	beta = gottenCRD.Spec.Validation.JSONSchema.Properties["beta"]
	beta.Minimum = apiextensionsv1beta1.Float64Ptr(10)

	updatedCRD, err := apiExtensionClient.ApiextensionsV1beta1().CustomResourceDefinitions().Update(gottenCRD)
	if err != nil {
		t.Fatal(err)
	}

	// CR is now accepted
	_, err = instantiateCustomResource(t, testserver.NewNoxuValidationInstance(ns, "foo"), noxuResourceClient, updatedCRD)
	if err != nil {
		t.Fatalf("unable to create noxu Instance:%v", err)
	}
}
