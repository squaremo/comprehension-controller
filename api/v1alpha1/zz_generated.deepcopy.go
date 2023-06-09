//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Comprehension) DeepCopyInto(out *Comprehension) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Comprehension.
func (in *Comprehension) DeepCopy() *Comprehension {
	if in == nil {
		return nil
	}
	out := new(Comprehension)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Comprehension) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ComprehensionList) DeepCopyInto(out *ComprehensionList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Comprehension, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ComprehensionList.
func (in *ComprehensionList) DeepCopy() *ComprehensionList {
	if in == nil {
		return nil
	}
	out := new(ComprehensionList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ComprehensionList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ComprehensionSpec) DeepCopyInto(out *ComprehensionSpec) {
	*out = *in
	in.ForExpr.DeepCopyInto(&out.ForExpr)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ComprehensionSpec.
func (in *ComprehensionSpec) DeepCopy() *ComprehensionSpec {
	if in == nil {
		return nil
	}
	out := new(ComprehensionSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ComprehensionStatus) DeepCopyInto(out *ComprehensionStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ComprehensionStatus.
func (in *ComprehensionStatus) DeepCopy() *ComprehensionStatus {
	if in == nil {
		return nil
	}
	out := new(ComprehensionStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Expr) DeepCopyInto(out *Expr) {
	*out = *in
	if in.ForExpr != nil {
		in, out := &in.ForExpr, &out.ForExpr
		*out = new(ForExpr)
		(*in).DeepCopyInto(*out)
	}
	if in.TemplateExpr != nil {
		in, out := &in.TemplateExpr, &out.TemplateExpr
		*out = new(TemplateExpr)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Expr.
func (in *Expr) DeepCopy() *Expr {
	if in == nil {
		return nil
	}
	out := new(Expr)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ForExpr) DeepCopyInto(out *ForExpr) {
	*out = *in
	in.In.DeepCopyInto(&out.In)
	in.Do.DeepCopyInto(&out.Do)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ForExpr.
func (in *ForExpr) DeepCopy() *ForExpr {
	if in == nil {
		return nil
	}
	out := new(ForExpr)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Generator) DeepCopyInto(out *Generator) {
	*out = *in
	if in.List != nil {
		in, out := &in.List, &out.List
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Generator.
func (in *Generator) DeepCopy() *Generator {
	if in == nil {
		return nil
	}
	out := new(Generator)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TemplateExpr) DeepCopyInto(out *TemplateExpr) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TemplateExpr.
func (in *TemplateExpr) DeepCopy() *TemplateExpr {
	if in == nil {
		return nil
	}
	out := new(TemplateExpr)
	in.DeepCopyInto(out)
	return out
}
