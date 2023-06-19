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

package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	helpers "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	generate "github.com/squaremo/comprehension-controller/api/v1alpha1"
)

type generatorFunc func(ev *Evaluator, ar map[string]interface{}) ([]interface{}, error)

func compileGenerator(e *env, expr *generate.Generator) (generatorFunc, error) {
	switch {
	case expr.List != nil:
		return compileList(e, expr)
	case expr.Query != nil:
		return compileQuery(e, expr)
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
			return func(ev *Evaluator, ar map[string]interface{}) ([]interface{}, error) {
				for i := range evals {
					if err := evals[i](ar); err != nil {
						return nil, err
					}
				}
				return deepcopy(items).([]interface{}), nil
			}, nil
		}
	}

	return func(_ *Evaluator, _ map[string]interface{}) ([]interface{}, error) {
		return items, nil
	}, nil
}

func compileQuery(e *env, expr *generate.Generator) (generatorFunc, error) {
	ce, err := e.celEnv()
	if err != nil {
		return nil, err
	}

	query := *expr.Query
	// copy MatchLabels, otherwise we might overwrite the original
	if query.MatchLabels != nil {
		copy := map[string]string{}
		for k, v := range query.MatchLabels {
			copy[k] = v
		}
		query.MatchLabels = copy
	}

	replaceStrPointer := func(p *string) replaceFunc {
		return func(v interface{}) {
			if s, ok := v.(string); ok {
				*p = s
				return
			}
			panic("tried to replace a string with a non-string value")
		}
	}

	replaceStrMap := func(m map[string]string, k string) replaceFunc {
		return func(v interface{}) {
			if s, ok := v.(string); ok {
				m[k] = s
			}
			panic("tried to replace a string in a map with a non-string value")
		}
	}

	var evals []evaluationFunc

	apiVersionEvals, err := compileAny(ce, query.APIVersion, replaceStrPointer(&query.APIVersion))
	if err != nil {
		return nil, err
	}
	kindEvals, err := compileAny(ce, query.Kind, replaceStrPointer(&query.Kind))
	if err != nil {
		return nil, err
	}
	nameEvals, err := compileAny(ce, query.Name, replaceStrPointer(&query.Name))
	if err != nil {
		return nil, err
	}
	evals = append(evals, apiVersionEvals...)
	evals = append(evals, kindEvals...)
	evals = append(evals, nameEvals...)

	// Having an expression in a value can mutate that entry, but it
	// can't create or delete entries; so, it's OK to always mutate
	// the map value.
	for k, v := range query.MatchLabels {
		eval, err := compileString(ce, v, replaceStrMap(query.MatchLabels, k))
		if err != nil {
			return nil, err
		}
		if eval != nil {
			evals = append(evals, eval)
		}
	}

	if len(evals) == 0 {
		// nothing to evaluate; just evaluate the query and use the results.
		var (
			objects []interface{}
			err     error
			once    sync.Once
		)

		return func(ev *Evaluator, _ map[string]interface{}) ([]interface{}, error) {
			once.Do(func() {
				objects, err = ev.generateObjectQuery(expr.Query)
			})
			return objects, err
		}, nil
	}

	return func(ev *Evaluator, ar map[string]interface{}) ([]interface{}, error) {
		for i := range evals {
			if err := evals[i](ar); err != nil {
				return nil, err
			}
		}
		return ev.generateObjectQuery(&query)
	}, nil
}

func (ev *Evaluator) generateObjectQuery(gen *generate.ObjectQuery) ([]interface{}, error) {
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
