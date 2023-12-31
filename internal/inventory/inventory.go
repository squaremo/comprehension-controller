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

package inventory

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	generatev1 "github.com/squaremo/comprehension-controller/api/v1alpha1"
)

func Add(inv *generatev1.Inventory, obj client.Object) {
	nsn := client.ObjectKeyFromObject(obj)
	gvk := obj.GetObjectKind().GroupVersionKind()
	ref := generatev1.ObjectRef{
		NamespacedName: nsn.String(),
		GroupVersion:   gvk.GroupVersion().String(),
		Kind:           gvk.Kind,
	}
	inv.Entries = append(inv.Entries, ref)
}
