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
	"errors"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/purell"
)

const (
	fragmentRune     = `#`
	emptyPointer     = ``
	pointerSeparator = `/`
	invalidStart     = `JSON pointer must be empty or start with a "` + pointerSeparator
)

// NewJSONSchemaRef creates a new reference for the given string
func NewJSONSchemaRef(jsonReferenceString string) (JSONSchemaRef, error) {
	var r JSONSchemaRef
	err := r.parseJSONSchemaRef(jsonReferenceString)
	return r, err
}

// MustCreateJSONSchemaRef parses the ref string and panics when it's invalid.
// Use the NewJSONSchemaRef method for a version that returns an error
func MustCreateJSONSchemaRef(ref string) JSONSchemaRef {
	r, err := NewJSONSchemaRef(ref)
	if err != nil {
		panic(err)
	}
	return r
}

// GetURL gets the URL for this reference
func (r *JSONSchemaRef) GetURL() *url.URL {
	return r.ReferenceURL
}

// GetPointer gets the json pointer for this reference
func (r *JSONSchemaRef) GetPointer() *JSONSchemaPointer {
	return &r.ReferencePointer
}

// String returns the best version of the url for this reference
func (r *JSONSchemaRef) String() string {
	if r.ReferenceURL != nil {
		return r.ReferenceURL.String()
	}
	if r.HasFragmentOnly {
		return fragmentRune + r.ReferencePointer.String()
	}
	return r.ReferencePointer.String()
}

// IsRoot returns true if this reference is a root document
func (r *JSONSchemaRef) IsRoot() bool {
	return r.ReferenceURL != nil &&
		!r.IsCanonical() &&
		!r.HasURLPathOnly &&
		r.ReferenceURL.Fragment == ""
}

// IsCanonical returns true when this pointer starts with http(s):// or file://
func (r *JSONSchemaRef) IsCanonical() bool {
	return (r.HasFileScheme && r.HasFullFilePath) || (!r.HasFileScheme && r.HasFullURL)
}

// "Constructor", parses the given string JSON reference
func (r *JSONSchemaRef) parseJSONSchemaRef(jsonReferenceString string) error {
	parsed, err := url.Parse(jsonReferenceString)
	if err != nil {
		return err
	}

	r.ReferenceURL, _ = url.Parse(purell.NormalizeURL(parsed, purell.FlagsSafe|purell.FlagRemoveDuplicateSlashes))
	refURL := r.ReferenceURL

	if refURL.Scheme != "" && refURL.Host != "" {
		r.HasFullURL = true
	} else {
		if refURL.Path != "" {
			r.HasURLPathOnly = true
		} else if refURL.RawQuery == "" && refURL.Fragment != "" {
			r.HasFragmentOnly = true
		}
	}
	r.HasFileScheme = refURL.Scheme == "file"
	r.HasFullFilePath = strings.HasPrefix(refURL.Path, "/")

	// invalid json-pointer error means url has no json-pointer fragment. simply ignore error
	r.ReferencePointer, _ = NewJSONSchemaPointer(refURL.Fragment)

	return nil
}

// JSONSchemaPointer to string representation function
func (p *JSONSchemaPointer) String() string {
	if len(p.ReferenceTokens) == 0 {
		return emptyPointer
	}
	pointerString := pointerSeparator + strings.Join(p.ReferenceTokens, pointerSeparator)
	return pointerString
}

// NewJSONSchemaPointer creates a new json pointer for the given string
func NewJSONSchemaPointer(jsonPointerString string) (JSONSchemaPointer, error) {
	var p JSONSchemaPointer
	err := p.parseJSONSchemaPointer(jsonPointerString)
	return p, err
}

// "Constructor", parses the given string JSON pointer
func (p *JSONSchemaPointer) parseJSONSchemaPointer(jsonPointerString string) error {
	var err error
	if jsonPointerString != emptyPointer {
		if !strings.HasPrefix(jsonPointerString, pointerSeparator) {
			err = errors.New(invalidStart)
		} else {
			referenceTokens := strings.Split(jsonPointerString, pointerSeparator)
			for _, referenceToken := range referenceTokens[1:] {
				p.ReferenceTokens = append(p.ReferenceTokens, referenceToken)
			}
		}
	}
	return err
}

func Float64Ptr(f float64) *float64 {
	return &f
}
func Int64Ptr(f int64) *int64 {
	return &f
}
