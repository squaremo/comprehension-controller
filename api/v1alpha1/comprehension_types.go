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
// expr := forExpr
//       | generateExpr
//
// templateExpr := "output": template
//
// forExpr := "for": var
//            "in": generator
//            "do": expr
//
// var := dnslabel
//
// generator := "list": object*
//        // | others TBD
//
// template := /* { TypeMeta... } */
//

type Expr struct {
	*ForExpr      `json:",omitempty"`
	*TemplateExpr `json:",omitempty"`
}

type ForExpr struct {
	For string    `json:"for"`
	In  Generator `json:"in"`
	Do  Expr      `json:"do"`
}

type TemplateExpr struct {
	Template *apiextensions.JSON `json:"template,omitempty"`
}

type Generator struct {
	List []string `json:"list,omitempty"` // stand-in for now
}

// ComprehensionSpec defines the desired state of Comprehension
type ComprehensionSpec struct {
	ForExpr `json:""`
}

// ComprehensionStatus defines the observed state of Comprehension
type ComprehensionStatus struct {
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
