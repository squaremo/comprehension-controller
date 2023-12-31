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
	"fmt"

	"sigs.k8s.io/yaml"

	generate "github.com/squaremo/comprehension-controller/api/v1alpha1"
)

func printEval(eyaml string) {
	var expr generate.ComprehensionSpec
	if err := yaml.Unmarshal([]byte(eyaml), &expr); err != nil {
		panic(err)
	}
	ev := &Evaluator{}
	outs, err := ev.Eval(&expr)
	if err != nil {
		panic(err)
	}
	for _, out := range outs {
		fmt.Println(out)
	}
}

func Example_eval_empty() {
	printEval(`
for:
- var: foo
  in:
    list: []
yield:
  template: "blah"
`)
	// Output:
}

func Example_eval_const() {
	printEval(`
for:
- var: foo
  in:
    list:
    - a
    - b
    - c
yield:
  template: "blat"
`)
	// Output:
	// blat
	// blat
	// blat
}

func Example_eval_nest() {
	printEval(`
for:
- var: foo
  in:
    list: [1,2,3]
- var: bar
  in:
    list: [a, b]
yield:
  template: "blah"
`)
	// Output:
	// blah
	// blah
	// blah
	// blah
	// blah
	// blah
}

func Example_eval_varref() {
	printEval(`
for:
- var: foo
  in:
    list: [bar, boo]
yield:
  template: value=${foo}
`)
	// Output:
	// value=bar
	// value=boo
}

func Example_eval_nested_varref() {
	printEval(`
for:
- var: outer
  in:
    list: [a, b]
- var: inner
  in:
    list: ["1", "2"]
yield:
  template: "[${outer}, ${inner}]"
`)
	// Unordered output:
	// [a, 1]
	// [b, 1]
	// [a, 2]
	// [b, 2]
}

func Example_eval_when() {
	printEval(`
yield:
  template: ${x * x}
for:
- var: x
  in:
    list: [1,2,3]
  when: int(x) % 2 == 1
`)
	// Output:
	// 1
	// 9
}

func Example_eval_when_nested() {
	printEval(`
yield:
  template: "${a}^2 + ${b}^2 = ${c}^2"
for:
- var: "a"
  in:
    list: [1,2,3,4,5,6,7,8,9,10]
- var: "b"
  in:
    list: [1,2,3,4,5,6,7,8,9,10]
- var: "c"
  in:
    list: [1,2,3,4,5,6,7,8,9,10]
  when: c*c == a*a + b*b
`)
	// Output:
	// 3^2 + 4^2 = 5^2
	// 4^2 + 3^2 = 5^2
	// 6^2 + 8^2 = 10^2
	// 8^2 + 6^2 = 10^2
}

func Example_eval_dependent() {
	printEval(`
yield:
    template: ${a} -> ${b}
for:
- var: a
  in:
    list: [1,2,3]
- var: b
  in:
    list:
    - ${int(a) * 2}
    - ${int(a) * 3}
    - ${int(a) * 4}
`)
	// Output:
	// 1 -> 2
	// 1 -> 3
	// 1 -> 4
	// 2 -> 4
	// 2 -> 6
	// 2 -> 8
	// 3 -> 6
	// 3 -> 9
	// 3 -> 12
}

// demonstrates that you can interpolate into a `list:` generator to
// flatten a list-of-lists value into the inner items.
func Example_eval_flatten() {
	printEval(`
yield:
  template: ${x}
for:
- var: xs
  in:
    list: [[1,2,3], [4,5,6]]
- var: x
  in:
    list: ${xs} # i.e., lift the value into a list generator
`)
	// Output:
	// 1
	// 2
	// 3
	// 4
	// 5
	// 6
}

// demonstrates that you can issue an HTTP request with interpolated
// parameters (in this case, the repo comes from a literal list). This
// relies on the httptest server run in main_test.go.
func Example_eval_request() {
	printEval(`
for:
- var: repo
  in:
    list: ["flux-whatif", "flux2"]
- var: pulls
  in:
    request:
      url: ` + baseurl + `/${repo}-pulls.json
- var: pr
  in:
    list: ${pulls}
yield:
  template: ${pr.url}
`)
	// Output:
	// https://api.github.com/repos/squaremo/flux-whatif-example/pulls/1
	// https://api.github.com/repos/fluxcd/flux2/pulls/4017
	// https://api.github.com/repos/fluxcd/flux2/pulls/4006
	// https://api.github.com/repos/fluxcd/flux2/pulls/3368
	// https://api.github.com/repos/fluxcd/flux2/pulls/3366
	// https://api.github.com/repos/fluxcd/flux2/pulls/3254
	// https://api.github.com/repos/fluxcd/flux2/pulls/3158
	// https://api.github.com/repos/fluxcd/flux2/pulls/2639
	// https://api.github.com/repos/fluxcd/flux2/pulls/2370
	// https://api.github.com/repos/fluxcd/flux2/pulls/2222
	// https://api.github.com/repos/fluxcd/flux2/pulls/2188
	// https://api.github.com/repos/fluxcd/flux2/pulls/2187
	// https://api.github.com/repos/fluxcd/flux2/pulls/2086
	// https://api.github.com/repos/fluxcd/flux2/pulls/1620
	// https://api.github.com/repos/fluxcd/flux2/pulls/1350
}
