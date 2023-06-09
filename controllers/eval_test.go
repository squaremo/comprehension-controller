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

	"sigs.k8s.io/yaml"

	generate "github.com/squaremo/comprehension-controller/api/v1alpha1"
)

func printEval(eyaml string) {
	var expr generate.ForExpr
	if err := yaml.Unmarshal([]byte(eyaml), &expr); err != nil {
		panic(err)
	}
	for _, out := range evalTop(&expr) {
		fmt.Println(out)
	}
}

func Example_empty() {
	printEval(`
for: foo
in:
  list: []
do:
  rest: "blah"
`)
	// Output:
}

func Example_const() {
	printEval(`
for: foo
in:
  list:
  - a
  - b
  - c
do:
  rest: "blat"
`)
	// Output:
	// blat
	// blat
	// blat
}

func Example_nest() {
	printEval(`
for: foo
in:
  list: [1,2,3]
do:
  for: bar
  in:
    list: [a, b]
  do:
    rest: "blah"
`)
	// Output:
	// blah
	// blah
	// blah
	// blah
	// blah
	// blah
}
