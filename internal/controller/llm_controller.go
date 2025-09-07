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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	gelimerv1alpha1 "zzzinho.busan/api/v1alpha1"
)

const containerName = "model-container"

// LLMReconciler reconciles a LLM object
type LLMReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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

	switch llm.Spec.Runtime {
	case gelimerv1alpha1.RuntimeTypeVLLM:
		if err := r.createVLLMDeployment(ctx, llm); err != nil {
			logger.Error(err, "Failed to create VLLM deployment")
			return ctrl.Result{}, err
		}
		logger.Info("VLLM deployment created")
		llm.Status.Status = "Running"
		if err := r.Status().Update(ctx, llm); err != nil {
			logger.Error(err, "Failed to update LLM status")
			return ctrl.Result{}, err
		}
		logger.Info("LLM status updated")
		return ctrl.Result{}, nil
	default:
		logger.Info("Unknown runtime", "runtime", llm.Spec.Runtime)
	}

	return ctrl.Result{}, nil
}

func (r *LLMReconciler) createVLLMDeployment(ctx context.Context, llm *gelimerv1alpha1.LLM) error {
	logger := logf.FromContext(ctx)

	logger.Info("Creating VLLM")
	dpmnt := r.createVllmDeploymentManifest(llm)
	if err := r.Create(ctx, dpmnt); err != nil {
		logger.Error(err, "Failed to create VLLM deployment")
		return err
	}
	return nil
}

func (r *LLMReconciler) createVllmDeploymentManifest(llm *gelimerv1alpha1.LLM) *appsv1.Deployment {
	serviceAccountName := llm.Spec.ServiceAccountName
	if serviceAccountName == "" {
		serviceAccountName = "default"
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      llm.Name,
			Namespace: llm.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &llm.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": llm.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": llm.Name,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: serviceAccountName,
					Containers: []corev1.Container{
						{
							Name:  containerName,
							Image: llm.Spec.Image,
							Args:  createVllmArgs(llm),
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: llm.Spec.Port,
									Name:          "http",
								},
							},
							Resources:    llm.Spec.Resources,
							Env:          llm.Spec.Env,
							VolumeMounts: llm.Spec.VolumeMounts,
						},
					},
					Volumes: llm.Spec.Volumes,
				},
			},
		},
	}
}

func createVllmArgs(llm *gelimerv1alpha1.LLM) []string {
	args := []string{"--model", llm.Spec.Model}

	// TODO: vLLM Config to args

	return args
}

// SetupWithManager sets up the controller with the Manager.
func (r *LLMReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gelimerv1alpha1.LLM{}).
		Named("llm").
		Complete(r)
}
