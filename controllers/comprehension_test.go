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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	generate "github.com/squaremo/comprehension-controller/api/v1alpha1"
	//+kubebuilder:scaffold:imports
)

// It's often easier to write out examples in YAML, then as Go values.
func loadFromYAML(s string, obj client.Object) {
	ExpectWithOffset(1, yaml.Unmarshal([]byte(s), obj)).To(Succeed())
}

var _ = Describe("simple comprehension", func() {

	const c = `
apiVersion: generate.squaremo.dev/v1alpha1
kind: Comprehension
metadata:
  name: testcase
spec:
  for: var
  in:
    list:
    - foo
    - bar
    - baz
  do:
    template:
      apiVersion: v1
      kind: ConfigMap
      metadata:
        name: cm-${var}
      data:
        value: ${var}
`

	When("there's a comprehension using a list", func() {

		var namespace string

		BeforeEach(func() {
			namespace = "foo"
			var ns corev1.Namespace
			ns.Name = namespace
			Expect(k8sClient.Create(context.TODO(), &ns)).To(Succeed())

			var obj generate.Comprehension
			loadFromYAML(c, &obj)
			obj.Namespace = namespace
			Expect(k8sClient.Create(context.TODO(), &obj)).To(Succeed())
		})

		It("instantiates the template", func() {
			var configmaps corev1.ConfigMapList
			Eventually(func() int {
				Expect(k8sClient.List(context.TODO(), &configmaps, &client.ListOptions{
					Namespace: namespace,
				})).To(Succeed())
				return len(configmaps.Items)
			}, "5s", "1s").Should(Equal(3))

			configmapMatch := func(name string) gomegatypes.GomegaMatcher {
				return SatisfyAll(
					HaveField("Data", HaveKeyWithValue("value", name)),
					HaveField("Name", "cm-"+name),
				)
			}

			// these expectations are tied to the template, of course
			Expect(configmaps.Items).To(ConsistOf(
				configmapMatch("foo"),
				configmapMatch("bar"),
				configmapMatch("baz"),
			))
		})

	})
})
