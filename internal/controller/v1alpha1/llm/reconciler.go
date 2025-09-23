package llm

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	gelimerv1alpha1 "zzzinho.busan/api/v1alpha1"
)

const (
	containerName = "model-container"
	vllmPort      = 8000
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

	// Initialize status if empty
	if llm.Status.Status == "" {
		if err := r.updateStatus(ctx, llm, "Pending", "Initializing LLM resources"); err != nil {
			log.Error(err, "failed to update LLM status")
			return ctrl.Result{}, err
		}
	}

	// Always create or update resources to handle spec changes
	if err := r.createOrUpdateResources(ctx, llm); err != nil {
		log.Error(err, "failed to create or update resources")
		// Update status to Error on failure
		if statusErr := r.updateStatus(ctx, llm, "Error", fmt.Sprintf("Failed to create/update resources: %v", err)); statusErr != nil {
			log.Error(statusErr, "failed to update error status")
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *LLMReconciler) createOrUpdateResources(ctx context.Context, llm *gelimerv1alpha1.LLM) error {
	log := logf.FromContext(ctx)
	log.Info("Creating resources", "namespace", llm.Namespace, "name", llm.Name)

	switch llm.Spec.Runtime {
	case gelimerv1alpha1.RuntimeTypeVLLM:
		return r.createOrUpdateVLLMResources(ctx, llm)
	default:
		return fmt.Errorf("unknown runtime: %s", llm.Spec.Runtime)
	}
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

// createOrUpdateVLLMResources creates or updates vLLM deployment and service
func (r *LLMReconciler) createOrUpdateVLLMResources(ctx context.Context, llm *gelimerv1alpha1.LLM) error {
	log := logf.FromContext(ctx)
	log.Info("Creating or updating vLLM resources", "namespace", llm.Namespace, "name", llm.Name)

	// Create or update Deployment
	deployment := r.buildVLLMDeployment(llm)
	if err := controllerutil.SetControllerReference(llm, deployment, r.Scheme); err != nil {
		log.Error(err, "failed to set controller reference for deployment")
		return err
	}

	foundDeployment := &appsv1.Deployment{}
	err := r.Get(ctx, client.ObjectKey{Name: deployment.Name, Namespace: deployment.Namespace}, foundDeployment)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating new Deployment", "namespace", deployment.Namespace, "name", deployment.Name)
		if err := r.Create(ctx, deployment); err != nil {
			log.Error(err, "failed to create Deployment")
			return err
		}
	} else if err != nil {
		log.Error(err, "failed to get Deployment")
		return err
	} else {
		// Update existing deployment
		foundDeployment.Spec = deployment.Spec
		if err := r.Update(ctx, foundDeployment); err != nil {
			log.Error(err, "failed to update Deployment")
			return err
		}
		log.Info("Updated Deployment", "namespace", deployment.Namespace, "name", deployment.Name)
	}

	// Create or update Service
	service := r.buildVLLMService(llm)
	if err := controllerutil.SetControllerReference(llm, service, r.Scheme); err != nil {
		log.Error(err, "failed to set controller reference for service")
		return err
	}

	foundService := &corev1.Service{}
	err = r.Get(ctx, client.ObjectKey{Name: service.Name, Namespace: service.Namespace}, foundService)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating new Service", "namespace", service.Namespace, "name", service.Name)
		if err := r.Create(ctx, service); err != nil {
			log.Error(err, "failed to create Service")
			return err
		}
	} else if err != nil {
		log.Error(err, "failed to get Service")
		return err
	} else {
		// Update existing service (preserve ClusterIP)
		service.Spec.ClusterIP = foundService.Spec.ClusterIP
		foundService.Spec = service.Spec
		if err := r.Update(ctx, foundService); err != nil {
			log.Error(err, "failed to update Service")
			return err
		}
		log.Info("Updated Service", "namespace", service.Namespace, "name", service.Name)
	}

	// Update status to Running
	if err := r.updateStatus(ctx, llm, "Running", "vLLM deployment is running"); err != nil {
		log.Error(err, "failed to update LLM status")
		return err
	}

	return nil
}

// buildVLLMDeployment builds the deployment spec for vLLM
func (r *LLMReconciler) buildVLLMDeployment(llm *gelimerv1alpha1.LLM) *appsv1.Deployment {
	labels := map[string]string{
		"app":     llm.Name,
		"llm":     llm.Name,
		"runtime": string(llm.Spec.Runtime),
	}

	// Generate spec hash for change detection
	specHash := r.generateSpecHash(llm)

	// Create annotations for template to force pod restart on spec changes
	templateAnnotations := map[string]string{
		"gelimer.zzzinho.busan/spec-hash":  specHash,
		"gelimer.zzzinho.busan/generation": fmt.Sprintf("%d", llm.Generation),
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      llm.Name,
			Namespace: llm.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &llm.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: templateAnnotations,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: llm.Spec.ServiceAccountName,
					Containers: []corev1.Container{
						{
							Name:            containerName,
							Image:           llm.Spec.Image,
							ImagePullPolicy: llm.Spec.ImagePullPolicy,
							Args:            r.buildVLLMArgs(llm),
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: llm.Spec.Port,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env:          r.buildVLLMEnvVars(llm),
							Resources:    llm.Spec.Resources,
							VolumeMounts: llm.Spec.VolumeMounts,
						},
					},
					Volumes: llm.Spec.Volumes,
				},
			},
		},
	}

	return deployment
}

// buildVLLMService builds the service spec for vLLM
func (r *LLMReconciler) buildVLLMService(llm *gelimerv1alpha1.LLM) *corev1.Service {
	labels := map[string]string{
		"app":     llm.Name,
		"llm":     llm.Name,
		"runtime": string(llm.Spec.Runtime),
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      llm.Name,
			Namespace: llm.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       llm.Spec.Port,
					TargetPort: intstr.FromInt(int(llm.Spec.Port)),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	return service
}

// buildVLLMArgs builds the arguments for vLLM container
func (r *LLMReconciler) buildVLLMArgs(llm *gelimerv1alpha1.LLM) []string {
	args := []string{
		"--model", llm.Spec.Model,
		"--host", "0.0.0.0",
		"--port", strconv.Itoa(int(llm.Spec.Port)),
	}

	// Add vLLM specific configuration if provided
	if llm.Spec.RuntimeConfig.VLLM != nil {
		vllmConfig := llm.Spec.RuntimeConfig.VLLM

		// Model configuration
		if vllmConfig.Model != nil {
			if vllmConfig.Model.Task != "" && vllmConfig.Model.Task != "auto" {
				args = append(args, "--task", vllmConfig.Model.Task)
			}
			if vllmConfig.Model.Tokenizer != nil {
				args = append(args, "--tokenizer", *vllmConfig.Model.Tokenizer)
			}
			if vllmConfig.Model.TokenizerMode != "" && vllmConfig.Model.TokenizerMode != "auto" {
				args = append(args, "--tokenizer-mode", vllmConfig.Model.TokenizerMode)
			}
			if vllmConfig.Model.TrustRemoteCode != nil && *vllmConfig.Model.TrustRemoteCode {
				args = append(args, "--trust-remote-code")
			}
			if vllmConfig.Model.MaxModelLen != nil {
				args = append(args, "--max-model-len", strconv.Itoa(int(*vllmConfig.Model.MaxModelLen)))
			}
		}

		// Load configuration
		if vllmConfig.Load != nil {
			if vllmConfig.Load.DownloadDir != nil {
				args = append(args, "--download-dir", *vllmConfig.Load.DownloadDir)
			}
			if vllmConfig.Load.LoadFormat != "" && vllmConfig.Load.LoadFormat != "auto" {
				args = append(args, "--load-format", vllmConfig.Load.LoadFormat)
			}
			if vllmConfig.Load.Dtype != "" && vllmConfig.Load.Dtype != "auto" {
				args = append(args, "--dtype", vllmConfig.Load.Dtype)
			}
			if vllmConfig.Load.QuantizationMethod != nil {
				args = append(args, "--quantization", *vllmConfig.Load.QuantizationMethod)
			}
			if vllmConfig.Load.Device != "" && vllmConfig.Load.Device != "auto" {
				args = append(args, "--device", vllmConfig.Load.Device)
			}
		}

		// Decoding configuration
		if vllmConfig.Decoding != nil {
			if vllmConfig.Decoding.GuidedDecodingBackend != "" {
				args = append(args, "--guided-decoding-backend", vllmConfig.Decoding.GuidedDecodingBackend)
			}
			if vllmConfig.Decoding.MaxLogprobs != nil {
				args = append(args, "--max-logprobs", strconv.Itoa(int(*vllmConfig.Decoding.MaxLogprobs)))
			}
			if vllmConfig.Decoding.DisableSlidingWindow != nil && *vllmConfig.Decoding.DisableSlidingWindow {
				args = append(args, "--disable-sliding-window")
			}
		}

		// Parallel configuration
		if vllmConfig.Parallel != nil {
			if vllmConfig.Parallel.TensorParallelSize > 1 {
				args = append(args, "--tensor-parallel-size", strconv.Itoa(int(vllmConfig.Parallel.TensorParallelSize)))
			}
			if vllmConfig.Parallel.PipelineParallelSize > 1 {
				args = append(args, "--pipeline-parallel-size", strconv.Itoa(int(vllmConfig.Parallel.PipelineParallelSize)))
			}
			if vllmConfig.Parallel.DistributedExecutorBackend != nil {
				args = append(args, "--distributed-executor-backend", *vllmConfig.Parallel.DistributedExecutorBackend)
			}
			if vllmConfig.Parallel.WorkerUseRay != nil && *vllmConfig.Parallel.WorkerUseRay {
				args = append(args, "--worker-use-ray")
			}
		}

		// Cache configuration
		if vllmConfig.Cache != nil {
			if vllmConfig.Cache.KvCacheDtype != "" && vllmConfig.Cache.KvCacheDtype != "auto" {
				args = append(args, "--kv-cache-dtype", vllmConfig.Cache.KvCacheDtype)
			}
			if vllmConfig.Cache.GpuMemoryUtilization != "" {
				args = append(args, "--gpu-memory-utilization", vllmConfig.Cache.GpuMemoryUtilization)
			}
			if vllmConfig.Cache.SwapSpace != nil {
				args = append(args, "--swap-space", strconv.Itoa(int(*vllmConfig.Cache.SwapSpace)))
			}
			if vllmConfig.Cache.BlockSize > 0 {
				args = append(args, "--block-size", strconv.Itoa(int(vllmConfig.Cache.BlockSize)))
			}
		}

		// LoRA configuration
		if vllmConfig.LoRA != nil {
			if len(vllmConfig.LoRA.LoraModules) > 0 {
				args = append(args, "--enable-lora")
				for _, module := range vllmConfig.LoRA.LoraModules {
					loraSpec := module.Name + "=" + module.Path
					if module.BaseModelName != nil {
						loraSpec += ":" + *module.BaseModelName
					}
					args = append(args, "--lora-modules", loraSpec)
				}
			}
			if vllmConfig.LoRA.MaxLoras > 0 {
				args = append(args, "--max-loras", strconv.Itoa(int(vllmConfig.LoRA.MaxLoras)))
			}
			if vllmConfig.LoRA.MaxLoraRank > 0 {
				args = append(args, "--max-lora-rank", strconv.Itoa(int(vllmConfig.LoRA.MaxLoraRank)))
			}
			if vllmConfig.LoRA.LoraDtype != "" && vllmConfig.LoRA.LoraDtype != "auto" {
				args = append(args, "--lora-dtype", vllmConfig.LoRA.LoraDtype)
			}
		}

		// Scheduler configuration
		if vllmConfig.Scheduler != nil {
			if vllmConfig.Scheduler.MaxNumSeqs > 0 {
				args = append(args, "--max-num-seqs", strconv.Itoa(int(vllmConfig.Scheduler.MaxNumSeqs)))
			}
			if vllmConfig.Scheduler.MaxNumBatchedTokens != nil {
				args = append(args, "--max-num-batched-tokens", strconv.Itoa(int(*vllmConfig.Scheduler.MaxNumBatchedTokens)))
			}
			if vllmConfig.Scheduler.EnableChunkedPrefill != nil {
				if *vllmConfig.Scheduler.EnableChunkedPrefill {
					args = append(args, "--enable-chunked-prefill")
				} else {
					args = append(args, "--disable-chunked-prefill")
				}
			}
			if vllmConfig.Scheduler.PreemptionMode != "" {
				args = append(args, "--preemption-mode", vllmConfig.Scheduler.PreemptionMode)
			}
		}

		// Extra arguments
		if len(vllmConfig.ExtraArgs) > 0 {
			for key, value := range vllmConfig.ExtraArgs {
				args = append(args, "--"+key, value)
			}
		}
	}

	return args
}

// buildVLLMEnvVars builds environment variables for vLLM container
func (r *LLMReconciler) buildVLLMEnvVars(llm *gelimerv1alpha1.LLM) []corev1.EnvVar {
	// Start with user-provided env vars
	envVars := llm.Spec.Env
	if envVars == nil {
		envVars = []corev1.EnvVar{}
	}

	// Add default vLLM environment variables if not already set
	defaultEnvs := map[string]string{
		"VLLM_API_KEY": "", // Empty means no auth required
	}

	// Check if env var already exists before adding default
	for key, value := range defaultEnvs {
		found := false
		for _, env := range envVars {
			if env.Name == key {
				found = true
				break
			}
		}
		if !found {
			envVars = append(envVars, corev1.EnvVar{
				Name:  key,
				Value: value,
			})
		}
	}

	return envVars
}

// generateSpecHash generates a hash of the LLM spec for change detection
func (r *LLMReconciler) generateSpecHash(llm *gelimerv1alpha1.LLM) string {
	// Create a struct with only the relevant spec fields for hashing
	specForHash := struct {
		Image              string                        `json:"image"`
		Model              string                        `json:"model"`
		Port               int32                         `json:"port"`
		Replicas           int32                         `json:"replicas"`
		Runtime            gelimerv1alpha1.RuntimeType   `json:"runtime"`
		RuntimeConfig      gelimerv1alpha1.RuntimeConfig `json:"runtimeConfig,omitempty"`
		Resources          corev1.ResourceRequirements   `json:"resources,omitempty"`
		Env                []corev1.EnvVar               `json:"env,omitempty"`
		VolumeMounts       []corev1.VolumeMount          `json:"volumeMounts,omitempty"`
		Volumes            []corev1.Volume               `json:"volumes,omitempty"`
		ImagePullPolicy    corev1.PullPolicy             `json:"imagePullPolicy,omitempty"`
		ServiceAccountName string                        `json:"serviceAccountName,omitempty"`
	}{
		Image:              llm.Spec.Image,
		Model:              llm.Spec.Model,
		Port:               llm.Spec.Port,
		Replicas:           llm.Spec.Replicas,
		Runtime:            llm.Spec.Runtime,
		RuntimeConfig:      llm.Spec.RuntimeConfig,
		Resources:          llm.Spec.Resources,
		Env:                llm.Spec.Env,
		VolumeMounts:       llm.Spec.VolumeMounts,
		Volumes:            llm.Spec.Volumes,
		ImagePullPolicy:    llm.Spec.ImagePullPolicy,
		ServiceAccountName: llm.Spec.ServiceAccountName,
	}

	// Convert to JSON and hash
	specBytes, err := json.Marshal(specForHash)
	if err != nil {
		// Fallback to generation if JSON marshaling fails
		return fmt.Sprintf("gen-%d", llm.Generation)
	}

	hash := sha256.Sum256(specBytes)
	return fmt.Sprintf("%x", hash)[:16] // Use first 16 characters of hash
}
