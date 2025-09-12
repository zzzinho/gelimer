package llm

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gelimerv1alpha1 "zzzinho.busan/api/v1alpha1"
)

const LLMFinalizerName = "gelimer.zzzinho.busan/llm-finalizer"

type LLMReconciler struct {
	BaseHandler
}

func (r *LLMReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := logf.FromContext(ctx)

	var llm gelimerv1alpha1.LLM
	if err := r.Get(ctx, req.NamespacedName, &llm); err != nil {
		log.Error(err, "failed to get LLM")
		return ctrl.Result{}, err
	}

	switch llm.Spec.Runtime {
	case gelimerv1alpha1.RuntimeTypeVLLM:
		return r.handleVLLM(ctx, &llm)
	default:
		log.Error(errors.NewBadRequest("unknown runtime"), "Unknown runtime", "runtime", llm.Spec.Runtime)
		return reconcile.Result{}, errors.NewBadRequest("Not supported runtime")
	}
}

func (r *LLMReconciler) handleVLLM(ctx context.Context, llm *gelimerv1alpha1.LLM) (reconcile.Result, error) {
	// Handle deletion
	if !llm.DeletionTimestamp.IsZero() {
		return r.handleDelete(ctx, llm)
	}

	// Handle creation and update
	if llm.Status.Status == "" {
		return r.handleCreate(ctx, llm)
	} else {
		return r.handleUpdate(ctx, llm)
	}
}

func (r *LLMReconciler) handleCreate(ctx context.Context, llm *gelimerv1alpha1.LLM) (reconcile.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Creating LLM", "name", llm.Name)

	// Add finalizer to handle deletion
	if !controllerutil.ContainsFinalizer(llm, LLMFinalizerName) {
		log.Info("No finalizer found, adding finalizer")
		controllerutil.AddFinalizer(llm, LLMFinalizerName)
		if err := r.Update(ctx, llm); err != nil {
			log.Error(err, "failed to update LLM")
			return reconcile.Result{}, err
		}
		log.Info("Finalizer added", "name", llm.Name)
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, nil
}

func (r *LLMReconciler) handleUpdate(ctx context.Context, llm *gelimerv1alpha1.LLM) (reconcile.Result, error) {
	// TODO: Implement update logic
	return reconcile.Result{}, nil
}

func (r *LLMReconciler) handleDelete(ctx context.Context, llm *gelimerv1alpha1.LLM) (reconcile.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Deleting LLM", "name", llm.Name)

	if !controllerutil.ContainsFinalizer(llm, LLMFinalizerName) {
		log.Info("No finalizer found, skippling cleanup")
		return reconcile.Result{}, nil
	}

	if err := r.updateStatus(ctx, llm, "Deleting", "Deleting LLM"); err != nil {
		log.Error(err, "failed to update LLM status")
		return reconcile.Result{}, err
	}

	if err := r.cleanupResources(ctx, llm); err != nil {
		log.Error(err, "failed to cleanup resources")
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *LLMReconciler) cleanupResources(ctx context.Context, llm *gelimerv1alpha1.LLM) error {
	log := logf.FromContext(ctx)
	log.Info("Cleaning up resources", "name", llm.Name)

	// Delete deployment
	deployments := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      llm.Name,
			Namespace: llm.Namespace,
		},
	}

	if err := r.Delete(ctx, deployments); err != nil && !errors.IsNotFound(err) {
		log.Error(err, "failed to delete deployment")
		return err
	}

	// Delete service
	services := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      llm.Name,
			Namespace: llm.Namespace,
		},
	}

	if err := r.Delete(ctx, services); err != nil && !errors.IsNotFound(err) {
		log.Error(err, "failed to delete service")
		return err
	}

	log.Info("Resources cleaned up", "name", llm.Name)
	return nil
}

func (r *LLMReconciler) updateStatus(ctx context.Context, llm *gelimerv1alpha1.LLM, status, message string) error {
	log := logf.FromContext(ctx)

	llm.Status.Status = status
	if err := r.Status().Update(ctx, llm); err != nil {
		log.Error(err, "failed to update LLM status")
		return err
	}
	log.Info("LLM status updated", "status", status, "message", message)
	return nil
}

func (r *LLMReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gelimerv1alpha1.LLM{}).
		Named("llm").
		Complete(r)
}
