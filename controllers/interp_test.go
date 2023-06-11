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
	"fmt"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

func printTokens(s string) {
	parts, err := parseInterpolation(s)
	if err != nil {
		panic(err)
	}
	// fmt.Printf("%#v\n", parts) // for sanity checking
	for i := range parts {
		fmt.Println(parts[i].String())
	}
}

func Example_parseInterpolation_norefs() {
	printTokens("this is just a string")
	// Output:
	// "this is just a string"
}

func Example_parseInterpolation_justref() {
	printTokens("${var}")
	// Output:
	// ${var}
}

func Example_parseInterpolation_escapes() {
	printTokens("text $and $${escaped dollar")
	// Output:
	// "text $and ${escaped dollar"
}

func Example_parseInterpolation_embedref() {
	printTokens("text${var}more")
	// Output:
	// "text"
	// ${var}
	// "more"
}

func printTemplate(t string, name string, value interface{}) {
	// I do a bit of a dance here because I want to replicate how a
	// template is procssed by eval. It gets an apiextension.JSON, so
	// I start with that.
	var templateField apiextensions.JSON
	if err := yaml.Unmarshal([]byte(t), &templateField); err != nil {
		panic(err)
	}

	var template interface{}
	if err := json.Unmarshal(templateField.Raw, &template); err != nil {
		panic(err)
	}

	out, err := interpolateTemplate(&env{name: name, value: value}, template)
	if err != nil {
		panic(err)
	}
	asJSON, err := json.Marshal(out)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", asJSON)
}

func Example_interpolateTemplate_string() {
	printTemplate(`"${foo}"`, "foo", "bar")
	// Output:
	// "bar"
}

func Example_interpolateTemplate_map() {
	t := `
foo: ${v}
`
	printTemplate(t, "v", "bar")
	// Output:
	// {"foo":"bar"}
}

func Example_interpolateTemplate_nested() {
	t := `
foo:
  bar: ${v}
`
	printTemplate(t, "v", "boink")
	// Output:
	// {"foo":{"bar":"boink"}}
}

func Example_interpolateTemplate_array() {
	t := `
- foo
- bar
- ${v}
`
	printTemplate(t, "v", "boo")
	// Output:
	// ["foo","bar","boo"]
}

func Example_interpolateTemplate_mapvalue() {
	t := `
foo: ${v}
`
	printTemplate(t, "v", map[string]interface{}{
		"bar": "baz",
	})
	// Output:
	// {"foo":{"bar":"baz"}}
}

func Example_interpolateTemplate_slicevalue() {
	t := `
foo:
- bar
- ${v}
- baz
`
	printTemplate(t, "v", []interface{}{"boo", "boom"})
	// Output:
	// {"foo":["bar",["boo","boom"],"baz"]}
}

func Example_interpolateTemplate_indexexpr() {
	t := `
foo: ${v[1]}
`
	printTemplate(t, "v", []interface{}{"baz", "bar"})
	// Output:
	// {"foo":"bar"}
}
