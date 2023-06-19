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
	"encoding/json"
	"fmt"

	helpers "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	generate "github.com/squaremo/comprehension-controller/api/v1alpha1"
)

type generatorFunc func(ev *evaluator, ar map[string]interface{}) ([]interface{}, error)

func compileGenerator(e *env, expr *generate.Generator) (generatorFunc, error) {
	switch {
	case expr.List != nil:
		return compileList(e, expr)
	case expr.Query != nil:
		return func(ev *evaluator, _ map[string]interface{}) ([]interface{}, error) {
			return ev.generateObjectQuery(e, expr.Query)
		}, nil

	default:
		return nil, fmt.Errorf("unknown generator %#v", expr)
	}
}

func compileList(e *env, expr *generate.Generator) (generatorFunc, error) {
	var items []interface{}
	if len(expr.List) > 0 {
		items = make([]interface{}, len(expr.List))
		for i := range expr.List {
			if err := json.Unmarshal(expr.List[i].Raw, &items[i]); err != nil {
				return nil, err
			}
		}

		ce, err := e.celEnv()
		if err != nil {
			return nil, err
		}
		evals, err := compileSlice(ce, items)
		if len(evals) > 0 {
			return func(ev *evaluator, ar map[string]interface{}) ([]interface{}, error) {
				for i := range evals {
					if err := evals[i](ar); err != nil {
						return nil, err
					}
				}
				return deepcopy(items).([]interface{}), nil
			}, nil
		}
	}

	return func(_ *evaluator, _ map[string]interface{}) ([]interface{}, error) {
		return items, nil
	}, nil
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
		if len(objs.Items) == 0 {
			return nil, nil
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
