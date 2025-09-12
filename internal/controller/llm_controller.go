/*
Copyright 2025 zzzinho.

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

package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	gelimerv1alpha1 "zzzinho.busan/api/v1alpha1"
	reconciler "zzzinho.busan/internal/reconciler/llm"
)

const containerName = "model-container"

// LLMReconciler reconciles a LLM object
type LLMReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	handlers map[gelimerv1alpha1.RuntimeType]reconciler.RuntimeHandler
}

// +kubebuilder:rbac:groups=gelimer.zzzinho.busan,resources=llms,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gelimer.zzzinho.busan,resources=llms/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=gelimer.zzzinho.busan,resources=llms/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LLM object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *LLMReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	logger.Info("Reconciling LLM")

	llm := &gelimerv1alpha1.LLM{}

	if err := r.Get(ctx, req.NamespacedName, llm); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("LLM", "name", llm.Name)
	handler, ok := r.handlers[llm.Spec.Runtime]
	if !ok {
		return ctrl.Result{}, fmt.Errorf("unknown runtime: %s", llm.Spec.Runtime)
	}
	return handler.Do(ctx, llm)
}

// SetupWithManager sets up the controller with the Manager.
func (r *LLMReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gelimerv1alpha1.LLM{}).
		Named("llm").
		Complete(r)
}
