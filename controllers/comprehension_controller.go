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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	generate "github.com/squaremo/comprehension-controller/api/v1alpha1"
	"github.com/squaremo/comprehension-controller/internal/eval"
)

// ComprehensionReconciler reconciles a Comprehension object
type ComprehensionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=generate.squaremo.dev,resources=comprehensions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=generate.squaremo.dev,resources=comprehensions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=generate.squaremo.dev,resources=comprehensions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Comprehension object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ComprehensionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var obj generate.Comprehension
	if err := r.Get(ctx, req.NamespacedName, &obj); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	ev := &eval.Evaluator{Client: client.NewNamespacedClient(r.Client, req.Namespace)}

	outs, err := ev.Eval(&obj.Spec)
	if err != nil {
		log.Error(err, "failed to evaluate comprehension")
	}

	for i := range outs {
		switch out := outs[i].(type) {
		case map[string]interface{}:
			if err := r.createObject(ctx, req.Namespace, out); err != nil {
				return ctrl.Result{}, err // TODO do better
			}
		case []interface{}:
			for i := range out {
				obj, ok := out[i].(map[string]interface{})
				if !ok {
					log.Info("item in instanatiated template is not an object") // TODO better
					continue
				}
				if err := r.createObject(ctx, req.Namespace, obj); err != nil {
					return ctrl.Result{}, err // TODO can do better here
				}
			}
		default:
			log.Info("instantiated template does not result in an object or list of objects")
			continue // TODO better than this
		}
	}

	return ctrl.Result{}, nil
}

func (r *ComprehensionReconciler) createObject(ctx context.Context, namespace string, fields map[string]interface{}) error {
	log := log.FromContext(ctx)
	var instance unstructured.Unstructured
	instance.Object = fields
	instance.SetNamespace(namespace)
	err := r.Create(ctx, &instance) // FIXME Upsert instead
	if err != nil {
		return err
	}
	log.Info("created object", "apiVersion", instance.GetAPIVersion(), "kind", instance.GetKind(), "name", instance.GetName())
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ComprehensionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&generate.Comprehension{}).
		Complete(r)
}
