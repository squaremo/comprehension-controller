/*
Copyright 2023 Michael Bridgen.

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

package v1alpha1

import (
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Grammar:
//
// top := templateExpr forExpr+

// templateExpr := "template": template
//
// forExpr := "var": var
//            "in": generator
//            "when": CELexpr
//
// var := DNSLABEL
//
// generator := "list" object*
//           | "query" apiVersion kind name|matchLabels
//        // | others TBD
//
// template := k8sTemplate+ /* { TypeMeta... } */

type ForExpr struct {
	Var  string    `json:"var"`
	In   Generator `json:"in"`
	When string    `json:"when,omitempty"`
}

type TemplateExpr struct {
	Template *apiextensions.JSON `json:"template,omitempty"`
}

type Generator struct {
	List    *apiextensions.JSON `json:"list,omitempty"`
	Query   *ObjectQuery        `json:"query,omitempty"`
	Request *HttpRequest        `json:"request,omitempty"`
}

type ObjectQuery struct {
	APIVersion  string            `json:"apiVersion"`
	Kind        string            `json:"kind"`
	Name        string            `json:"name,omitempty"`
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

type HttpRequest struct {
	URL     string   `json:"url"`
	Headers []string `json:"headers,omitempty"`
}

// ComprehensionSpec defines the desired state of Comprehension
type ComprehensionSpec struct {
	Yield TemplateExpr `json:"yield"`
	For   []ForExpr    `json:"for"`
}

// ComprehensionStatus defines the observed state of Comprehension
type ComprehensionStatus struct {
	Inventory *Inventory `json:"inventory,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Comprehension is the Schema for the comprehensions API
type Comprehension struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComprehensionSpec   `json:"spec,omitempty"`
	Status ComprehensionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ComprehensionList contains a list of Comprehension
type ComprehensionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Comprehension `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Comprehension{}, &ComprehensionList{})
}
