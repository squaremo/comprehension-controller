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

// This is borrowed from Flux's Kustomize controller, where it is used
// to similar ends: keep track of what was created, so that anything
// superfluous can be removed.

// Inventory enumerates the objects created by a Comprehension.
type Inventory struct {
	Entries []ObjectRef `json:"entries,omitempty"`
}

// ObjectRef keeps flattened reference to a Kubernetes object, with a
// name (namespace and name), and an API version and kind (GroupKind
// and Version). The fields are intended to be readable.
type ObjectRef struct {
	NamespacedName string `json:"namespacedName"`
	GroupVersion   string `json:"groupVersion"`
	Kind           string `json:"kind"`
}
