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
	gomegatypes "github.com/onsi/gomega/types"

	corev1 "k8s.io/api/core/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	generate "github.com/squaremo/comprehension-controller/api/v1alpha1"
	//+kubebuilder:scaffold:imports
)

// It's often easier to write out examples in YAML, then as Go values.
func loadFromYAML(s string, obj client.Object) {
	ExpectWithOffset(2, yaml.Unmarshal([]byte(s), obj)).To(Succeed())
}

var namespaceBase = "test-comprehension-"
var namespaceCount int

func newNamespace() string {
	namespace := fmt.Sprintf("%s%d", namespaceBase, namespaceCount)
	namespaceCount++
	var ns corev1.Namespace
	ns.Name = namespace
	ExpectWithOffset(1, k8sClient.Create(context.TODO(), &ns)).To(Succeed())
	return namespace
}

func createObjectsInNamespace(ns string, objs ...client.Object) {
	for i := range objs {
		objs[i].SetNamespace(ns)
		ExpectWithOffset(1, k8sClient.Create(context.TODO(), objs[i])).To(Succeed())
	}
}

func createComprehension(ns string, y string) *generate.Comprehension {
	var obj generate.Comprehension
	loadFromYAML(y, &obj)
	obj.Namespace = ns
	obj.Name = "testcase"
	ExpectWithOffset(1, k8sClient.Create(context.TODO(), &obj)).To(Succeed())
	return &obj
}

var _ = Describe("simple comprehension", func() {

	When("there's a comprehension using a list", func() {

		const listCompro = `
apiVersion: generate.squaremo.dev/v1alpha1
kind: Comprehension
spec:
  for:
  - var: v
    in:
      list:
      - foo
      - bar
      - baz
  yield:
    template:
      apiVersion: v1
      kind: ConfigMap
      metadata:
        name: cm-${v}
      data:
        value: ${v}
`

		var namespace string
		var configmaps corev1.ConfigMapList
		var compro *generate.Comprehension

		configmapMatch := func(name string) gomegatypes.GomegaMatcher {
			return SatisfyAll(
				HaveField("Data", HaveKeyWithValue("value", name)),
				HaveField("Name", "cm-"+name),
			)
		}

		BeforeEach(func() {
			namespace = newNamespace()
			compro = createComprehension(namespace, listCompro)
			Eventually(func() int {
				Expect(k8sClient.List(context.TODO(), &configmaps, &client.ListOptions{
					Namespace: namespace,
				})).To(Succeed())
				return len(configmaps.Items)
			}, "5s", "1s").Should(Equal(3))
		})

		It("instantiates the template", func() {
			// these expectations are tied to the template, of course
			Expect(configmaps.Items).To(ConsistOf(
				configmapMatch("foo"),
				configmapMatch("bar"),
				configmapMatch("baz"),
			))
		})

		It("sets the owner of each object", func() {
			hasController := Satisfy(func(cm corev1.ConfigMap) bool {
				return metav1.IsControlledBy(&cm, compro)
			})
			Expect(configmaps.Items).To(HaveEach(hasController))
		})

		When("an item is added to the generator", func() {
			BeforeEach(func() {
				Expect(k8sClient.Get(context.TODO(), client.ObjectKeyFromObject(compro), compro)).To(Succeed())
				compro.Spec.For[0].In.List = &apiextensions.JSON{
					Raw: []byte(`["foo", "bar", "baz", "boo"]`),
				}
				Expect(k8sClient.Update(context.TODO(), compro)).To(Succeed())

				Eventually(func() int {
					Expect(k8sClient.List(context.TODO(), &configmaps, &client.ListOptions{
						Namespace: namespace,
					})).To(Succeed())
					return len(configmaps.Items)
				}, "5s", "1s").Should(Equal(4))
			})

			It("should have successfully rerun the comprehension", func() {
				Expect(configmaps.Items).To(ConsistOf(
					configmapMatch("foo"),
					configmapMatch("bar"),
					configmapMatch("baz"),
					configmapMatch("boo"),
				))
			})

			It("should record the objects created in the inventory", func() {
				Expect(k8sClient.Get(context.TODO(), client.ObjectKeyFromObject(compro), compro)).To(Succeed())
				Expect(compro.Status.Inventory).NotTo(BeNil())
				// the list.Items don't have full TypeMeta, otherwise
				// I could get the GroupKind from that. Never mind.
				gv := schema.GroupVersion{Version: "v1"}.String()
				k := "ConfigMap"
				nsn := func(i int) string {
					return client.ObjectKeyFromObject(&configmaps.Items[i]).String()
				}
				Expect(compro.Status.Inventory.Entries).To(ConsistOf(
					generate.ObjectRef{NamespacedName: nsn(0), GroupVersion: gv, Kind: k},
					generate.ObjectRef{NamespacedName: nsn(1), GroupVersion: gv, Kind: k},
					generate.ObjectRef{NamespacedName: nsn(2), GroupVersion: gv, Kind: k},
					generate.ObjectRef{NamespacedName: nsn(3), GroupVersion: gv, Kind: k},
				))
			})
		})

		When("the items are changed", func() {
			BeforeEach(func() {
				Expect(k8sClient.Get(context.TODO(), client.ObjectKeyFromObject(compro), compro)).To(Succeed())
				compro.Spec.For[0].In.List = &apiextensions.JSON{
					Raw: []byte(`["foo", "jam"]`),
				}
				Expect(k8sClient.Update(context.TODO(), compro)).To(Succeed())

				Eventually(func() int {
					Expect(k8sClient.List(context.TODO(), &configmaps, &client.ListOptions{
						Namespace: namespace,
					})).To(Succeed())
					return len(configmaps.Items)
				}, "5s", "1s").Should(Equal(2))
			})

			It("removes the unneeded items", func() {
				Expect(configmaps.Items).To(ConsistOf(
					configmapMatch("foo"),
					configmapMatch("jam"),
				))
			})
		})
	})

	When("there's a comprehension using a named object", func() {

		const objCompro = `
apiVersion: generate.squaremo.dev/v1alpha1
kind: Comprehension
metadata:
  name: object-query
spec:
  for:
  - var: cm
    in:
      query:
        apiVersion: v1
        kind: ConfigMap
        name: source
  yield:
    template:
      apiVersion: v1
      kind: Secret
      metadata:
        name: target
      stringData: ${cm.data}
`
		var namespace string
		var compro *generate.Comprehension
		var secret corev1.Secret

		BeforeEach(func() {
			namespace = newNamespace()

			var cm corev1.ConfigMap
			cm.Name = "source"
			cm.Data = map[string]string{
				"foo": "bar",
			}
			createObjectsInNamespace(namespace, &cm)
			compro = createComprehension(namespace, objCompro)

			Eventually(func() error {
				return k8sClient.Get(context.TODO(), types.NamespacedName{
					Namespace: namespace,
					Name:      "target",
				}, &secret)
			}, "2s", "0.5s").Should(BeNil())
		})

		It("instantiates the template", func() {
			// these expectations are tied to the template, of course
			Expect(secret).To(SatisfyAll(
				HaveField("Data", HaveKeyWithValue("foo", []byte("bar"))),
				HaveField("Name", "target"),
			))
		})

		It("sets the owner to the comprehension object", func() {
			hasController := Satisfy(func(s *corev1.Secret) bool {
				return metav1.IsControlledBy(s, compro)
			})
			Expect(&secret).To(hasController)
		})
	})

	When("there's a comprehension using an object query", func() {
		const objCompro = `
apiVersion: generate.squaremo.dev/v1alpha1
kind: Comprehension
spec:
  for:
  - var: cm
    in:
      query:
        apiVersion: v1
        kind: ConfigMap
        matchLabels:
          app: foo
  yield:
    template:
      apiVersion: v1
      kind: Secret
      metadata:
        name: target-${cm.metadata.name}
      stringData: ${cm.data}
`
		var namespace string

		BeforeEach(func() {
			namespace = newNamespace()

			names := []string{"foo", "bar", "baz"}
			for i := range names {
				var cm corev1.ConfigMap
				cm.Namespace = namespace
				cm.Name = names[i]
				cm.Labels = map[string]string{
					"app": "foo",
				}
				cm.Data = map[string]string{
					"name": names[i],
				}
				Expect(k8sClient.Create(context.TODO(), &cm)).To(Succeed())
			}
			createComprehension(namespace, objCompro)
		})

		It("instantiates the template", func() {
			var secrets corev1.SecretList
			Eventually(func() int {
				Expect(k8sClient.List(context.TODO(), &secrets, &client.ListOptions{
					Namespace: namespace,
				})).To(Succeed())
				return len(secrets.Items)
			}, "2s", "0.5s").Should(Equal(3))
		})
	})

	When("there's a comprehension with a template of a list", func() {
		const compro = `
apiVersion: generate.squaremo.dev/v1alpha1
kind: Comprehension
spec:
  yield:
    template:
    - apiVersion: v1
      kind: Secret
      metadata:
        name: secret-${i}
      stringData:
        i: ${i}
    - apiVersion: v1
      kind: ConfigMap
      metadata:
        name: cm-${i}
      data:
        i: ${i}

  for:
  - var: i
    in: { list: [a,b,c] }
`

		var namespace string

		BeforeEach(func() {
			namespace = newNamespace()
			createComprehension(namespace, compro)
		})

		It("flattens the list and creates each item of each template result", func() {
			var cms corev1.ConfigMapList
			Eventually(func() int {
				Expect(k8sClient.List(context.TODO(), &cms, &client.ListOptions{
					Namespace: namespace,
				})).To(Succeed())
				return len(cms.Items)
			}, "3s", "0.5s").Should(Equal(3))
			// TODO other assertions?
			var secrets corev1.SecretList
			Eventually(func() int {
				Expect(k8sClient.List(context.TODO(), &secrets, &client.ListOptions{
					Namespace: namespace,
				})).To(Succeed())
				return len(secrets.Items)
			}, "3s", "0.5s").Should(Equal(3))
			// TODO other assertions?

		})
	})

	When("there's a query generator using an expression", func() {
		const compro = `
apiVersion: generate.squaremo.dev/v1alpha1
kind: Comprehension
spec:
  yield:
    template:
    - apiVersion: v1
      kind: Secret
      metadata:
        name: secret-${name}
      stringData: ${existing.data}
  for:
  - var: name
    in: { list: ["testcase-expr-0", "testcase-expr-1"] }
  - var: existing
    in:
      query:
        apiVersion: v1
        kind: ConfigMap
        name: ${name}
`
		var namespace string
		BeforeEach(func() {
			namespace = newNamespace()
			cm0 := &corev1.ConfigMap{
				Data: map[string]string{
					"foo": "bar",
				},
			}
			cm0.Name = "testcase-expr-0"
			cm1 := &corev1.ConfigMap{
				Data: map[string]string{
					"bar": "foo",
				},
			}
			cm1.Name = "testcase-expr-1"
			createObjectsInNamespace(namespace, cm0, cm1)
			createComprehension(namespace, compro)
		})

		It("executes the query with values as given by expressions", func() {
			var secrets corev1.SecretList
			Eventually(func() int {
				Expect(k8sClient.List(context.TODO(), &secrets, &client.ListOptions{
					Namespace: namespace,
				})).To(Succeed())
				return len(secrets.Items)
			}, "3s", "0.5s").Should(Equal(2)) // TODO more assertions?
		})
	})

})
