package llm

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	gelimerv1alpha1 "zzzinho.busan/api/v1alpha1"
)

const (
	containerName = "model-container"
)

type LLMReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *LLMReconciler) Reconcile(ctx context.Context, llm *gelimerv1alpha1.LLM) (ctrl.Result, error) {
	logr := logf.FromContext(ctx)
	logr.Info("Reconciling LLM", "name", llm.Name)

	switch llm.Spec.Runtime {
	case gelimerv1alpha1.RuntimeTypeVLLM:
		return r.handleVLLM(ctx, llm)
	default:
		return ctrl.Result{}, fmt.Errorf("unknown runtime: %s", llm.Spec.Runtime)
	}
}

func (r *LLMReconciler) handleVLLM(ctx context.Context, llm *gelimerv1alpha1.LLM) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	if !llm.DeletionTimestamp.IsZero() {
		log.Info("Deleting LLM", "namespace", llm.Namespace, "name", llm.Name)

		if err := r.updateStatus(ctx, llm, "Deleting", "Deleting LLM"); err != nil {
			log.Error(err, "failed to update LLM status")
			return ctrl.Result{}, err
		}

		if err := r.cleanupResources(ctx, llm); err != nil {
			log.Error(err, "failed to cleanup resources")
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if llm.Status.Status == "" {
		if err := r.updateStatus(ctx, llm, "Pending", "LLM status is empty"); err != nil {
			log.Error(err, "failed to update LLM status")
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

func (r *LLMReconciler) updateStatus(ctx context.Context, llm *gelimerv1alpha1.LLM, status, message string) error {
	log := logf.FromContext(ctx)

	// Status 부분 업데이트를 위한 Patch 생성
	patch := client.MergeFrom(llm.DeepCopy())

	llm.Status.Status = status
	llm.Status.Message = message

	// Status subresource에 대한 부분 업데이트 (Patch) 수행
	if err := r.Status().Patch(ctx, llm, patch); err != nil {
		log.Error(err, "failed to patch LLM status")
		return err
	}
	return nil
}

func (r *LLMReconciler) cleanupResources(ctx context.Context, llm *gelimerv1alpha1.LLM) error {
	log := logf.FromContext(ctx)
	log.Info("Cleaning up resources", "namespace", llm.Namespace, "name", llm.Name)

	// Delete the Service
	if err := r.Delete(ctx, &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      llm.Name,
			Namespace: llm.Namespace,
		},
	}); err != nil {
		log.Error(err, "failed to delete Service")
		return err
	}

	// Delete Deployment
	if err := r.Delete(ctx, &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      llm.Name,
			Namespace: llm.Namespace,
		},
	}); err != nil {
		log.Error(err, "failed to delete Deployment")
		return err
	}
	return nil
}
