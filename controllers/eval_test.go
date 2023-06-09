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
	"testing"

	. "github.com/onsi/gomega"

	generate "github.com/squaremo/comprehension-controller/api/v1alpha1"
)

func Test_Eval(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		t.Parallel()
		g := NewGomegaWithT(t)
		forExpr := generate.ForExpr{
			For: "foo",
			In: generate.Generator{
				List: []string{},
			},
			Do: generate.Expr{
				TemplateExpr: &generate.TemplateExpr{
					APIVersion: "v1",
					Kind:       "Foo",
					Rest:       "blat!",
				},
			},
		}
		g.Expect(eval(nil, &forExpr)).To(HaveLen(0))
	})

	t.Run("const template", func(t *testing.T) {
		t.Parallel()
		g := NewGomegaWithT(t)
		forExpr := generate.ForExpr{
			For: "foo",
			In: generate.Generator{
				List: []string{"1", "2", "3"},
			},
			Do: generate.Expr{
				TemplateExpr: &generate.TemplateExpr{
					APIVersion: "v1",
					Kind:       "Foo",
					Rest:       "blat",
				},
			},
		}
		g.Expect(eval(nil, &forExpr)).To(Equal([]string{"blat", "blat", "blat"}))
	})

	t.Run("nested expr", func(t *testing.T) {
		t.Parallel()
		g := NewGomegaWithT(t)
		forExpr := generate.ForExpr{
			For: "foo",
			In: generate.Generator{
				List: []string{"1", "2", "3"},
			},
			Do: generate.Expr{
				ForExpr: &generate.ForExpr{
					For: "bar",
					In: generate.Generator{
						List: []string{"a", "b"},
					},
					Do: generate.Expr{
						TemplateExpr: &generate.TemplateExpr{
							APIVersion: "v1",
							Kind:       "Foo",
							Rest:       "blat",
						},
					},
				},
			},
		}
		g.Expect(eval(nil, &forExpr)).To(Equal([]string{
			"blat", "blat", "blat", "blat", "blat", "blat",
		}))
	})
}
