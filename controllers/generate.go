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

package controllers

import (
	"context"
	"fmt"

	helpers "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	generate "github.com/squaremo/comprehension-controller/api/v1alpha1"
)

func (ev *evaluator) generateItems(e *env, gen *generate.Generator) ([]interface{}, error) {
	switch {
	case gen.List != nil:
		return anyify(gen.List), nil
	case gen.Query != nil:
		return ev.generateObjectQuery(e, gen.Query)
	default:
		return nil, fmt.Errorf("unknown generator %#v", gen)
	}
}

func (ev *evaluator) generateObjectQuery(e *env, gen *generate.ObjectQuery) ([]interface{}, error) {
	switch {
	case gen.MatchLabels == nil && gen.Name != "":
		var obj unstructured.Unstructured
		obj.SetAPIVersion(gen.APIVersion)
		obj.SetKind(gen.Kind)
		if err := ev.Get(context.TODO(), types.NamespacedName{
			Name: gen.Name,
		}, &obj); err != nil {
			return nil, fmt.Errorf("unable to fetch named object: %w", err)
		}
		return []interface{}{obj.Object}, nil

	case gen.Name == "" && gen.MatchLabels != nil:
		var objs unstructured.UnstructuredList
		objs.SetAPIVersion(gen.APIVersion)
		// unstructuredClient lets you give the item kind rather than the list kind
		objs.SetKind(gen.Kind)
		selector, err := helpers.LabelSelectorAsSelector(&helpers.LabelSelector{MatchLabels: gen.MatchLabels})
		if err != nil {
			return nil, err
		}
		if err := ev.List(context.TODO(), &objs, &client.ListOptions{LabelSelector: selector}); err != nil {
			return nil, fmt.Errorf("unable to fetch selected objects: %w", err)
		}
		out := make([]interface{}, len(objs.Items))
		for i := range objs.Items {
			out[i] = interface{}(objs.Items[i].Object)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("objects query generator must specify one of .name or .matchLabels")
	}
}

func anyify(ss []string) []interface{} {
	out := make([]interface{}, len(ss))
	for i := range ss {
		out[i] = interface{}(ss[i])
	}
	return out
}
