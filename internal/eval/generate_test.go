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
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	generate "github.com/squaremo/comprehension-controller/api/v1alpha1"
)

func expectGeneratorItems(g Gomega, y string, ev *Evaluator, match types.GomegaMatcher) {
	var gen generate.Generator
	g.ExpectWithOffset(1, yaml.Unmarshal([]byte(y), &gen)).To(Succeed())
	e := &env{}
	generate, err := compileGenerator(e, &gen)
	g.ExpectWithOffset(1, err).NotTo(HaveOccurred())

	objs, err := generate(ev, map[string]interface{}{})
	g.ExpectWithOffset(1, err).NotTo(HaveOccurred())
	g.ExpectWithOffset(1, objs).To(match)
}

func Test_list(t *testing.T) {

	t.Run("list generator with objects", func(t *testing.T) {
		const generatorYAML = `
list:
- foo: bar
- baz: [1,2,3]
`
		t.Run("generates objects", func(t *testing.T) {
			g := NewWithT(t)

			ev := &Evaluator{}
			expectGeneratorItems(g, generatorYAML, ev, Equal([]interface{}{
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
}

var count int

func newNamespaceAndEval(g Gomega) (string, *Evaluator) {
	namespace := fmt.Sprintf("testns-%d", count)
	count++
	var ns corev1.Namespace
	ns.Name = namespace
	g.Expect(k8sClient.Create(context.TODO(), &ns)).To(Succeed())

	k8s := client.NewNamespacedClient(k8sClient, namespace)
	ev := &Evaluator{Client: k8s}

	return namespace, ev
}

func Test_query(t *testing.T) {

	t.Run("there's a query generator using a name", func(t *testing.T) {

		t.Run("there are no such objects", func(t *testing.T) {
			g := NewWithT(t)

			const nosuchQuery = `
query:
   apiVersion: v1
   kind: Service
   matchLabels: { bluuaurgh: ten }
`
			_, ev := newNamespaceAndEval(g)
			expectGeneratorItems(g, nosuchQuery, ev, BeNil())
		})

		t.Run("there is a named object", func(t *testing.T) {
			g := NewWithT(t)

			const namedObject = `
query:
  apiVersion: v1
  kind: ConfigMap
  name: test
`
			namespace, ev := newNamespaceAndEval(g)

			var obj map[string]interface{}

			var cm unstructured.Unstructured
			cm.SetAPIVersion("v1")
			cm.SetKind("ConfigMap")
			cm.SetName("test")
			cm.SetNamespace(namespace)
			obj = cm.Object
			g.Expect(k8sClient.Create(context.TODO(), &cm)).To(Succeed())

			expectGeneratorItems(g, namedObject, ev, ConsistOf(matchKeys(obj)))
		})
	})
}

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
