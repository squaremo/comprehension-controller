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
	generate "github.com/squaremo/comprehension-controller/api/v1alpha1"
)

func evalTop(expr *generate.ForExpr) []string {
	return eval(nil, expr)
}

func eval(e *env, expr *generate.ForExpr) []string {
	ins := generateItems(e, &expr.In)
	var outs []string
	for i := range ins {
		newE := &env{name: expr.For, value: ins[i], next: e}
		// TODO use explicit stack?
		if forExpr := expr.Do.ForExpr; forExpr != nil {
			nestedOuts := eval(newE, forExpr)
			outs = append(outs, nestedOuts...)
		} else if templateExpr := expr.Do.TemplateExpr; templateExpr != nil {
			outs = append(outs, interpolateString(newE, templateExpr.Rest))
		}
	}
	return outs
}

func generateItems(_ *env, gen *generate.Generator) []string {
	switch {
	case gen.List != nil:
		return gen.List
	default:
		panic("unknown generator")
	}
}

type env struct {
	name  string
	value string // stand-in until I have a representation of objects
	next  *env
}

func (e *env) lookup(name string) (string, bool) {
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

func interpolateString(e *env, templateString string) string {
	return templateString // TODO!
}
