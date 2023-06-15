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
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	generate "github.com/squaremo/comprehension-controller/api/v1alpha1"
)

func expectGeneratorItems(y string, ev *evaluator, match types.GomegaMatcher) {
	var gen generate.Generator
	ExpectWithOffset(1, yaml.Unmarshal([]byte(y), &gen)).To(Succeed())
	e := &env{}
	objs, err := ev.generateItems(e, &gen)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	ExpectWithOffset(1, objs).To(match)
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
			expectGeneratorItems(generatorYAML, ev, Equal([]interface{}{
				map[string]interface{}{
					"foo": "bar",
				},
				map[string]interface{}{
					"baz": []interface{}{
						float64(1), float64(2), float64(3),
					},
				},
			}))
		})
	})

	When("there's a query generator using a name", func() {

		var namespace string
		var count int
		var k8s client.Client
		var ev *evaluator

		BeforeEach(func() {
			namespace = fmt.Sprintf("testns-%d", count)
			count++
			var ns corev1.Namespace
			ns.Name = namespace
			Expect(k8sClient.Create(context.TODO(), &ns)).To(Succeed())
			k8s = client.NewNamespacedClient(k8sClient, namespace)
			ev = &evaluator{Client: k8s}
		})

		When("there are no such objects", func() {

			const nosuchQuery = `
query:
   apiVersion: v1
   kind: Service
   matchLabels: { bluuaurgh: ten }
`
			It("generates an empty list", func() {
				expectGeneratorItems(nosuchQuery, ev, BeNil())
			})
		})

		When("there is a named object", func() {
			const namedObject = `
query:
  apiVersion: v1
  kind: ConfigMap
  name: test
`
			var obj map[string]interface{}

			BeforeEach(func() {
				var cm unstructured.Unstructured
				cm.SetAPIVersion("v1")
				cm.SetKind("ConfigMap")
				cm.SetName("test")
				cm.SetNamespace(namespace)
				obj = cm.Object
				Expect(k8s.Create(context.TODO(), &cm)).To(Succeed())
			})

			It("returns just that object", func() {
				expectGeneratorItems(namedObject, ev, ConsistOf(matchKeys(obj)))
			})
		})
	})
})

func matchKeys(obj map[string]interface{}) types.GomegaMatcher {
	keys := Keys{}
	for k := range obj {
		if m, ok := obj[k].(map[string]interface{}); ok {
			keys[k] = matchKeys(m)
		} else {
			keys[k] = Equal(obj[k])
		}
	}
	return MatchKeys(IgnoreExtras, keys)
}
