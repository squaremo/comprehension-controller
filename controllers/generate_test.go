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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	generate "github.com/squaremo/comprehension-controller/api/v1alpha1"
)

func expectGeneratorItems(y string, ev *evaluator, expected []interface{}) {
	var gen generate.Generator
	ExpectWithOffset(1, yaml.Unmarshal([]byte(y), &gen)).To(Succeed())
	e := &env{}
	objs, err := ev.generateItems(e, &gen)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	ExpectWithOffset(1, objs).To(BeEquivalentTo(expected))
}

var _ = Describe("generators", func() {

	When("there's a list generator with objects", func() {
		const generatorYAML = `
list:
- foo: bar
- baz: [1,2,3]
`
		It("generates objects", func() {
			ev := &evaluator{}
			expectGeneratorItems(generatorYAML, ev, []interface{}{
				map[string]interface{}{
					"foo": "bar",
				},
				map[string]interface{}{
					"baz": []interface{}{
						float64(1), float64(2), float64(3),
					},
				},
			})
		})
	})

})
