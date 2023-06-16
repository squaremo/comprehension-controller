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
	"fmt"
	//"reflect" // useful for println debugging
	"strings"

	"github.com/google/cel-go/cel"
)

// evaluationFunc runs an expression given the variable values.
type evaluationFunc func(map[string]interface{}) error

// replaceFunc is a func for replacing the value at some site
type replaceFunc func(v interface{})

// template is the result of compiling a template, which you can use
// to instantiate the template with evaluate(). It is not at all
// threadsafe! In fact, it mutates the value given to it.
type template struct {
	blank        interface{}
	replacements []evaluationFunc
}

// evaluate the template with a map representing the activation
// record; that is, the values for each of the variables in the
// expression.
func (t *template) evaluate(ar map[string]interface{}) (interface{}, error) {
	for i := range t.replacements {
		if err := t.replacements[i](ar); err != nil {
			return nil, err
		}
	}
	return t.blank, nil
}

func (e *env) vars() map[string]interface{} {
	out := make(map[string]interface{})
	for e != nil {
		if _, ok := out[e.name]; !ok {
			out[e.name] = e.value
		}
		e = e.next
	}
	return out
}

func (e *env) celEnv() (*cel.Env, error) {
	ce, err := cel.NewEnv()
	if err != nil {
		return nil, err
	}
	for k := range e.vars() {
		ce, err = ce.Extend(cel.Variable(k, cel.AnyType))
		if err != nil {
			return nil, err
		}
	}
	return ce, nil
}

func replacePointer(site *interface{}) replaceFunc {
	return func(v interface{}) {
		*site = v
	}
}

func compileTemplate(e *env, t interface{}) (*template, error) {
	ce, err := e.celEnv()
	if err != nil {
		return nil, err
	}

	var templ template
	templ.blank = t
	replacements, err := compileAny(ce, t, replacePointer(&templ.blank))
	if err != nil {
		return nil, err
	}
	templ.replacements = replacements
	return &templ, nil
}

func compileAny(ce *cel.Env, t interface{}, r replaceFunc) ([]evaluationFunc, error) {
	switch obj := t.(type) {
	case string:
		fn, err := compileString(ce, obj, r)
		if err != nil {
			return nil, err
		}
		if fn != nil {
			return []evaluationFunc{fn}, nil
		}
		return nil, nil
	case map[string]interface{}:
		return compileMap(ce, obj)
	case []interface{}:
		return compileSlice(ce, obj)
	default:
		//fmt.Printf("Type = %s\n", reflect.TypeOf(t))
		return nil, nil
	}
}

// compileString takes an environment in which to compile CEL, the
// potential program-containing string, and the replacement site; and
// returns any funcs needed to do template replacements.
func compileString(ce *cel.Env, template string, rfn replaceFunc) (evaluationFunc, error) {
	parts, err := parseInterpolation(template)
	if err != nil {
		return nil, err
	}

	// shortcut: any values without refs will just be [token{text: <s>}]
	if len(parts) == 1 && parts[0].expr == "" {
		return nil, nil
	}

	// semantics: if there's one part, and it's a ref, substitute the
	// whole value in, as it is.
	if len(parts) == 1 && parts[0].expr != "" {
		expr := parts[0].expr
		prog, err := compileExpr(ce, expr)
		if err != nil {
			return nil, err
		}
		fn := func(ar map[string]interface{}) error {
			ref, _, err := prog.Eval(ar)
			if err != nil {
				return err
			}
			rfn(ref.Value())
			return nil
		}
		return fn, nil
	}

	out := make([]string, len(parts))
	var outReplacements []evaluationFunc
	replace := func(i int, prog cel.Program) evaluationFunc {
		return func(ar map[string]interface{}) error {
			ref, _, err := prog.Eval(ar)
			if err != nil {
				return err
			}
			v := ref.Value()
			if s, ok := v.(string); ok {
				out[i] = s
			} else if s, ok := v.(interface{ String() string }); ok {
				out[i] = s.String()
			} else {
				out[i] = fmt.Sprintf("%#v", v) // TODO ???
			}
			return nil
		}
	}

	for i := range parts {
		if parts[i].expr == "" {
			out[i] = parts[i].text
			continue
		}
		prog, err := compileExpr(ce, parts[i].expr)
		if err != nil {
			return nil, err
		}
		outReplacements = append(outReplacements, replace(i, prog))
	}

	fn := func(ar map[string]interface{}) error {
		for i := range outReplacements {
			if err := outReplacements[i](ar); err != nil {
				return err
			}
		}
		rfn(strings.Join(out, ""))
		return nil
	}
	return fn, nil
}

func replaceMapItem(m map[string]interface{}, k string) replaceFunc {
	return func(v interface{}) {
		m[k] = v
	}
}

// compileMap descends through a map value, and returns any funcs
// needed to do replacements within.
func compileMap(ce *cel.Env, t map[string]interface{}) ([]evaluationFunc, error) {
	var replacements []evaluationFunc
	for k, v := range t {
		fieldReplacements, err := compileAny(ce, v, replaceMapItem(t, k))
		if err != nil {
			return nil, err
		}
		replacements = append(replacements, fieldReplacements...)
	}
	return replacements, nil
}

// compileSlice descends through a slice value, returning any funcs
// needed to do replacements within.
func compileSlice(ce *cel.Env, t []interface{}) ([]evaluationFunc, error) {
	var replacements []evaluationFunc
	for i := range t {
		itemReplacements, err := compileAny(ce, t[i], replacePointer(&t[i]))
		if err != nil {
			return nil, err
		}
		replacements = append(replacements, itemReplacements...)
	}
	return replacements, nil
}

func compileExpr(ce *cel.Env, expr string) (cel.Program, error) {
	ast, issues := ce.Compile(expr)
	if err := issues.Err(); err != nil {
		return nil, err
	}
	prog, err := ce.Program(ast)
	if err != nil {
		return nil, err
	}
	return prog, nil
}

// --- old interpolation stuff

func interpolateTemplate(e *env, t interface{}) (interface{}, error) {
	switch obj := t.(type) {
	case string:
		return interpolateString(e, obj)
	case map[string]interface{}:
		return interpolateMap(e, obj)
	case []interface{}:
		return interpolateSlice(e, obj)
	default:
		//fmt.Printf("Type = %s\n", reflect.TypeOf(t))
		return t, nil
	}
}

func interpolateMap(e *env, t map[string]interface{}) (interface{}, error) {
	result := map[string]interface{}{}
	for k, v := range t {
		out, err := interpolateTemplate(e, v)
		if err != nil {
			return nil, err
		}
		result[k] = out
	}
	return result, nil
}

func interpolateSlice(e *env, t []interface{}) (interface{}, error) {
	result := make([]interface{}, len(t))
	for i := range t {
		out, err := interpolateTemplate(e, t[i])
		if err != nil {
			return nil, err
		}
		result[i] = out
	}
	return result, nil
}

// Token represents a part of a string that contains interpolated
// bits. Either `.text` is set, meaning "just text", or `.expr` is
// set, meaning "something to be evaluated in the runtime
// environment".
type token struct {
	text string
	expr string
}

// For tests and troubleshooting
func (t token) String() string {
	if t.expr != "" {
		return fmt.Sprintf("${%s}", t.expr)
	}
	return fmt.Sprintf("%q", t.text)
}

// parseInterpolation takes a string value and tries to parse it as an
// interpolated string. If it cannot parse it, an error is
// returned. If it can be parsed, the list of tokens is returned
func parseInterpolation(s string) ([]token, error) {
	var parts []token
	i := 0
	var sb strings.Builder

	const (
		stateText = iota
		stateDollar
		stateRef
	)
	state := stateText

	for i < len(s) {
		switch state {
		case stateText:
			if s[i] == '$' {
				state = stateDollar
			} else {
				sb.WriteByte(s[i])
			}
		case stateDollar:
			if s[i] == '{' { // var ref ${
				if sb.Len() > 0 {
					parts = append(parts, token{text: sb.String()})
					sb.Reset()
				}
				state = stateRef
			} else if s[i] == '$' { // escaped '$' e.g., $$
				state = stateText
				sb.WriteByte('$')
			} else { // '$' in regular text, e.g., $a
				state = stateText
				sb.WriteByte('$')
				sb.WriteByte(s[i])
			}
		case stateRef:
			if s[i] == '}' {
				parts = append(parts, token{expr: sb.String()})
				sb.Reset()
				state = stateText
			} else {
				sb.WriteByte(s[i])
			}
		}
		i++
	}

	// the only problem is if you don't close a ref
	if state == stateRef {
		return nil, fmt.Errorf("malformed string value %q", s)
	}
	if sb.Len() > 0 {
		parts = append(parts, token{text: sb.String()})
	}
	return parts, nil
}

func interpolateString(e *env, template string) (interface{}, error) {
	parts, err := parseInterpolation(template)
	if err != nil {
		return nil, err
	}

	// shortcut: any values without refs will just be [token{text: <s>}]
	if len(parts) == 1 && parts[0].expr == "" {
		return parts[0].text, nil
	}

	// semantics: if there's one part, and it's a ref, substitute the
	// whole value in, as it is.
	if len(parts) == 1 && parts[0].expr != "" {
		expr := parts[0].expr
		return interpretExpr(e, expr)
	}

	var sb strings.Builder
	for i := range parts {
		if expr := parts[i].expr; expr != "" {
			v, err := interpretExpr(e, expr)
			if err != nil {
				return nil, err
			}
			if s, ok := v.(string); ok {
				sb.WriteString(s)
			} else if s, ok := v.(interface{ String() string }); ok {
				sb.WriteString(s.String())
			} else {
				sb.WriteString(fmt.Sprintf("%#v", v)) // TODO ???
			}
		} else {
			sb.WriteString(parts[i].text)
		}
	}
	return sb.String(), nil
}

// interpreting expressions

func interpretExpr(e *env, expr string) (interface{}, error) {
	values := e.vars()
	celEnv, err := cel.NewEnv()
	for k := range values {
		celEnv, err = celEnv.Extend(cel.Variable(k, cel.AnyType))
		if err != nil {
			return nil, err
		}
	}
	ast, issues := celEnv.Compile(expr)
	if err := issues.Err(); err != nil {
		return nil, err
	}
	prog, err := celEnv.Program(ast)
	if err != nil {
		return nil, err
	}
	ref, _, err := prog.Eval(values)
	if err != nil {
		return nil, err
	}
	return ref.Value(), nil
}
