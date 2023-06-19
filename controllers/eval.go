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

	"github.com/google/cel-go/cel"
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
	when   cel.Program
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

		var when cel.Program
		if w := expr.For[i].When; w != "" {
			ce, err := e.celEnv()
			if err != nil {
				return nil, err
			}
			when, err = compileExpr(ce, w)
			if err != nil {
				return nil, err
			}
		}
		generatedValues[i] = generated{name: name, values: values, when: when}
	}

	var template interface{}
	if err := json.Unmarshal(expr.Yield.Template.Raw, &template); err != nil {
		return nil, err
	}

	t, err := compileTemplate(e, template)
	if err != nil {
		return nil, err
	}
	return instantiateTemplate(t, map[string]interface{}{}, generatedValues, nil)
}

func instantiateTemplate(t *template, ar map[string]interface{}, rest []generated, out []interface{}) ([]interface{}, error) {
	if len(rest) == 0 {
		val, err := t.evaluate(ar)
		if err != nil {
			return nil, err
		}
		return append(out, val), nil
	}

	g := rest[0]
	for i := range g.values {
		ar[g.name] = g.values[i]

		if g.when != nil {
			ref, _, err := g.when.Eval(ar)
			if err != nil {
				return nil, err
			}
			if !truthy(ref.Value()) {
				continue
			}
		}

		var err error
		out, err = instantiateTemplate(t, ar, rest[1:], out)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// truthy here is anything that isn't `false`.
func truthy(val interface{}) bool {
	if b, ok := val.(bool); ok {
		return b
	}
	return true
}
