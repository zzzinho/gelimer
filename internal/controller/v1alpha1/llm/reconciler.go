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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	gelimerv1alpha1 "zzzinho.busan/api/v1alpha1"
)

const (
	LLMFinalizerName = "gelimer.zzzinho.busan/llm-finalizer"
	containerName    = "model-container"
)

type LLMReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *LLMReconciler) Do(ctx context.Context, llm *gelimerv1alpha1.LLM) (ctrl.Result, error) {
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

	// add finalizer if not exists
	if llm.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(llm, LLMFinalizerName) {
			log.Info("No finalizer found, adding finalizer")
			if err := r.addFinalizer(ctx, llm); err != nil {
				log.Error(err, "failed to add finalizer")
				return ctrl.Result{}, err
			}
			log.Info("Finalizer added", "namespace", llm.Namespace, "name", llm.Name)
		}
	} else { // delete LLM
		log.Info("Deleting LLM", "namespace", llm.Namespace, "name", llm.Name)
		if !controllerutil.ContainsFinalizer(llm, LLMFinalizerName) {
			log.Info("No finalizer found, skippling cleanup")
			return ctrl.Result{}, nil
		}

		if err := r.updateStatus(ctx, llm, "Deleting", "Deleting LLM"); err != nil {
			log.Error(err, "failed to update LLM status")
			return ctrl.Result{}, err
		}

		if err := r.cleanupResources(ctx, llm); err != nil {
			log.Error(err, "failed to cleanup resources")
			return ctrl.Result{}, err
		}
		if err := r.removeFinalizer(ctx, llm); err != nil {
			log.Error(err, "failed to remove finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if llm.Status.Status == "" {
		if err := r.updateStatus(ctx, llm, "Pending", "LLM status is empty"); err != nil {
			log.Error(err, "failed to update LLM status")
			return ctrl.Result{}, err
		}

		// finalizer
		if err := r.addFinalizer(ctx, llm); err != nil {
			log.Error(err, "failed to add finalizer")
			return ctrl.Result{}, err
		}

		// cleanup resources
		if err := r.cleanupResources(ctx, llm); err != nil {
			log.Error(err, "failed to cleanup resources")
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

func (r *LLMReconciler) addFinalizer(ctx context.Context, llm *gelimerv1alpha1.LLM) error {
	log := logf.FromContext(ctx)
	log.Info("Adding finalizer", "namespace", llm.Namespace, "name", llm.Name)
	patch := client.MergeFrom(llm.DeepCopy())
	controllerutil.AddFinalizer(llm, LLMFinalizerName)
	if err := r.Patch(ctx, llm, patch); err != nil {
		log.Error(err, "failed to patch LLM")
		return err
	}
	return nil
}

func (r *LLMReconciler) removeFinalizer(ctx context.Context, llm *gelimerv1alpha1.LLM) error {
	log := logf.FromContext(ctx)
	log.Info("Removing finalizer", "namespace", llm.Namespace, "name", llm.Name)
	patch := client.MergeFrom(llm.DeepCopy())
	controllerutil.RemoveFinalizer(llm, LLMFinalizerName)
	if err := r.Patch(ctx, llm, patch); err != nil {
		log.Error(err, "failed to patch LLM for finalizer removal")
		return err
	}

	log.Info("Finalizer removed", "namespace", llm.Namespace, "name", llm.Name)
	return nil
}

func (r *LLMReconciler) createResources(ctx context.Context, llm *gelimerv1alpha1.LLM) error {
	log := logf.FromContext(ctx)
	log.Info("Creating resources", "namespace", llm.Namespace, "name", llm.Name)

	return nil
}

func (r *LLMReconciler) createDeployment(ctx context.Context, llm *gelimerv1alpha1.LLM) *appsv1.Deployment {
	saName := llm.Spec.ServiceAccountName
	if saName == "" {
		saName = "default"
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
					ServiceAccountName: saName,
					Containers: []corev1.Container{
						{
							Name:  containerName,
							Image: llm.Spec.Image,
							// Args:  createVllmArgs(llm), TODO: add args
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

func (r *LLMReconciler) createService(ctx context.Context, llm *gelimerv1alpha1.LLM) error {
	log := logf.FromContext(ctx)
	log.Info("Creating service", "namespace", llm.Namespace, "name", llm.Name)
	return nil
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
