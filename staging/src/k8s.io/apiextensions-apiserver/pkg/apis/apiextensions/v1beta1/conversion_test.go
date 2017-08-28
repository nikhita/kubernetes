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
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
)

func TestJSONConversion(t *testing.T) {
	nullJSON := apiextensions.JSON(nil)
	stringJSON := apiextensions.JSON("foo")
	boolJSON := apiextensions.JSON(true)
	sliceJSON := apiextensions.JSON([]string{"foo", "bar", "baz"})

	testCases := map[string]struct {
		input    *apiextensions.JSON
		expected *JSON
	}{
		"omitempty": {
			expected: &JSON{},
		},
		"null": {
			input: &nullJSON,
			expected: &JSON{
				Raw: []byte(`null`),
			},
		},
		"string": {
			input: &stringJSON,
			expected: &JSON{
				Raw: []byte(`"foo"`),
			},
		},
		"bool": {
			input: &boolJSON,
			expected: &JSON{
				Raw: []byte(`true`),
			},
		},
		"slice": {
			input: &sliceJSON,
			expected: &JSON{
				Raw: []byte(`["foo","bar","baz"]`),
			},
		},
	}

	scheme := runtime.NewScheme()
	if err := AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	for k, tc := range testCases {
		external := &JSON{}
		if err := scheme.Convert(tc.input, external, nil); err != nil {
			t.Errorf("%s: unexpected error: %v", k, err)
		}

		if !reflect.DeepEqual(external, tc.expected) {
			t.Errorf("%s: expected\n\t%#v, got \n\t%#v", k, tc.expected, external)
		}
	}
}
