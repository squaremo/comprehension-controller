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

func (ev *evaluator) evalTop(expr *generate.ForExpr) ([]interface{}, error) {
	return ev.eval(nil, expr)
}

func (ev *evaluator) eval(e *env, expr *generate.ForExpr) ([]interface{}, error) {
	ins, err := ev.generateItems(e, &expr.In)
	if err != nil {
		return nil, err
	}
	var outs []interface{}
	for i := range ins {
		newE := &env{name: expr.For, value: ins[i], next: e}
		// TODO use explicit stack?
		if forExpr := expr.Do.ForExpr; forExpr != nil {
			nestedOuts, err := ev.eval(newE, forExpr)
			if err != nil {
				return outs, err
			}
			outs = append(outs, nestedOuts...)
		} else if templateExpr := expr.Do.TemplateExpr; templateExpr != nil {
			var template interface{}
			if err := json.Unmarshal(templateExpr.Template.Raw, &template); err != nil {
				return nil, err
			}
			out, err := interpolateTemplate(newE, template)
			if err != nil {
				return outs, err
			}
			outs = append(outs, out)
		}
	}
	return outs, nil
}

type env struct {
	name  string
	value interface{}
	next  *env
}

func (e *env) lookup(name string) (interface{}, bool) {
	for {
		if e == nil {
			return "", false
		}
		if e.name == name {
			return e.value, true
		}
		e = e.next
	}
}
