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

type generatorFunc func(ev *evaluator, ar map[string]interface{}) ([]interface{}, error)

type generated struct {
	name   string
	values generatorFunc
	when   cel.Program
}

func makeGeneratorFunc(e *env, expr *generate.Generator) generatorFunc {
	return func(ev *evaluator, _ map[string]interface{}) ([]interface{}, error) {
		return ev.generateItems(e, expr)
	}
}

func (ev *evaluator) evalTop(expr *generate.ComprehensionSpec) ([]interface{}, error) {
	generatedValues := make([]generated, len(expr.For))
	var e *env
	for i := range expr.For {
		// TODO: detect duplicate var names
		values := makeGeneratorFunc(e, &expr.For[i].In)
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
	return ev.instantiateTemplate(t, map[string]interface{}{}, generatedValues, nil)
}

func (ev *evaluator) instantiateTemplate(t *template, ar map[string]interface{}, rest []generated, out []interface{}) ([]interface{}, error) {
	if len(rest) == 0 {
		val, err := t.evaluate(ar)
		if err != nil {
			return nil, err
		}
		return append(out, val), nil
	}

	g := rest[0]
	values, err := g.values(ev, ar)
	if err != nil {
		return nil, err
	}
	for i := range values {
		ar[g.name] = values[i]

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
		out, err = ev.instantiateTemplate(t, ar, rest[1:], out)
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
