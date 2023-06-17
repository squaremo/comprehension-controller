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
	"encoding/json"

	"sigs.k8s.io/controller-runtime/pkg/client"

	generate "github.com/squaremo/comprehension-controller/api/v1alpha1"
)

type evaluator struct {
	client.Client
}

type env struct {
	name string
	next *env
}

type generated struct {
	name   string
	values []interface{}
}

func (ev *evaluator) evalTop(expr *generate.ComprehensionSpec) ([]interface{}, error) {
	// At present, each generator is independent; so it's sufficient
	// to run each, collect the generated values, then run each
	// combination through the template.
	generatedValues := make([]generated, len(expr.For))
	var e *env
	for i := range expr.For {
		// TODO: detect duplicate var names
		values, err := ev.generateItems(e, &expr.For[i].In)
		if err != nil {
			return nil, err
		}
		// If any of the generated lists is empty, the product is empty.
		if len(values) == 0 {
			return nil, nil
		}
		name := expr.For[i].Var
		e = &env{name: name, next: e}
		generatedValues[i] = generated{name: name, values: values}
	}

	var template interface{}
	if err := json.Unmarshal(expr.Yield.Template.Raw, &template); err != nil {
		return nil, err
	}

	t, err := compileTemplate(e, template)
	if err != nil {
		return nil, err
	}
	return instantiateTemplate(t, map[string]interface{}{}, generatedValues)
}

func instantiateTemplate(t *template, ar map[string]interface{}, rest []generated) ([]interface{}, error) {
	if len(rest) == 0 {
		return nil, nil
	}

	g := rest[0]

	if len(rest) == 1 {
		out := make([]interface{}, len(g.values))
		for i := range g.values {
			ar[g.name] = g.values[i]
			val, err := t.evaluate(ar)
			if err != nil {
				return nil, err
			}
			out[i] = val
		}
		return out, nil
	}

	var out []interface{}
	for i := range g.values {
		ar[g.name] = g.values[i]
		more, err := instantiateTemplate(t, ar, rest[1:])
		if err != nil {
			return nil, err
		}
		out = append(out, more...)
	}
	return out, nil
}
