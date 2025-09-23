package llm

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	gelimerv1alpha1 "zzzinho.busan/api/v1alpha1"
)

type VLLMReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *VLLMReconciler) Reconcile(ctx context.Context, llm *gelimerv1alpha1.LLM) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Reconciling VLLM", "namespace", llm.Namespace, "name", llm.Name)

	// create deployment
	deployment := r.createDeploymentManifest(llm)
	if err := r.Create(ctx, deployment); err != nil {
		log.Error(err, "failed to create deployment")
		return ctrl.Result{}, err
	}

	// create service

	return ctrl.Result{}, nil
}

func (r *VLLMReconciler) createDeploymentManifest(llm *gelimerv1alpha1.LLM) *appsv1.Deployment {
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
							Args:  r.createVLLMArgs(llm),
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

func (r *VLLMReconciler) createVLLMArgs(llm *gelimerv1alpha1.LLM) []string {
	args := []string{"vllm", "serve"}

	// Base model argument (required)
	args = append(args, llm.Spec.Model)

	// Port argument (required)
	args = append(args, "--port", strconv.Itoa(int(llm.Spec.Port)))

	// Process VLLM specific configuration if present
	if llm.Spec.RuntimeConfig.VLLM != nil {
		vllmConfig := llm.Spec.RuntimeConfig.VLLM
		args = append(args, r.buildVLLMConfigArgs(vllmConfig)...)
	}

	return args
}

func (r *VLLMReconciler) buildVLLMConfigArgs(config *gelimerv1alpha1.VLLMConfig) []string {
	var args []string

	// Model configuration
	if config.Model != nil {
		args = append(args, r.buildModelConfigArgs(config.Model)...)
	}

	// Load configuration
	if config.Load != nil {
		args = append(args, r.buildLoadConfigArgs(config.Load)...)
	}

	// Decoding configuration
	if config.Decoding != nil {
		args = append(args, r.buildDecodingConfigArgs(config.Decoding)...)
	}

	// Parallel configuration
	if config.Parallel != nil {
		args = append(args, r.buildParallelConfigArgs(config.Parallel)...)
	}

	// Cache configuration
	if config.Cache != nil {
		args = append(args, r.buildCacheConfigArgs(config.Cache)...)
	}

	// LoRA configuration
	if config.LoRA != nil {
		args = append(args, r.buildLoRAConfigArgs(config.LoRA)...)
	}

	// Scheduler configuration
	if config.Scheduler != nil {
		args = append(args, r.buildSchedulerConfigArgs(config.Scheduler)...)
	}

	// Extra arguments
	if config.ExtraArgs != nil {
		for key, value := range config.ExtraArgs {
			args = append(args, fmt.Sprintf("--%s", key), value)
		}
	}

	return args
}

func (r *VLLMReconciler) buildModelConfigArgs(config *gelimerv1alpha1.ModelConfig) []string {
	var args []string

	if config.Task != "" && config.Task != "auto" {
		args = append(args, "--task", config.Task)
	}

	if config.Tokenizer != nil {
		args = append(args, "--tokenizer", *config.Tokenizer)
	}

	if config.TokenizerMode != "" && config.TokenizerMode != "auto" {
		args = append(args, "--tokenizer-mode", config.TokenizerMode)
	}

	if config.TrustRemoteCode != nil {
		if *config.TrustRemoteCode {
			args = append(args, "--trust-remote-code")
		}
	}

	if config.MaxModelLen != nil {
		args = append(args, "--max-model-len", strconv.Itoa(int(*config.MaxModelLen)))
	}

	return args
}

func (r *VLLMReconciler) buildLoadConfigArgs(config *gelimerv1alpha1.LoadConfig) []string {
	var args []string

	if config.DownloadDir != nil {
		args = append(args, "--download-dir", *config.DownloadDir)
	}

	if config.LoadFormat != "" && config.LoadFormat != "auto" {
		args = append(args, "--load-format", config.LoadFormat)
	}

	if config.Dtype != "" && config.Dtype != "auto" {
		args = append(args, "--dtype", config.Dtype)
	}

	if config.QuantizationMethod != nil {
		args = append(args, "--quantization", *config.QuantizationMethod)
	}

	if config.Device != "" && config.Device != "auto" {
		args = append(args, "--device", config.Device)
	}

	return args
}

func (r *VLLMReconciler) buildDecodingConfigArgs(config *gelimerv1alpha1.DecodingConfig) []string {
	var args []string

	if config.GuidedDecodingBackend != "" && config.GuidedDecodingBackend != "outlines" {
		args = append(args, "--guided-decoding-backend", config.GuidedDecodingBackend)
	}

	if config.MaxLogprobs != nil {
		args = append(args, "--max-logprobs", strconv.Itoa(int(*config.MaxLogprobs)))
	}

	if config.DisableSlidingWindow != nil && *config.DisableSlidingWindow {
		args = append(args, "--disable-sliding-window")
	}

	return args
}

func (r *VLLMReconciler) buildParallelConfigArgs(config *gelimerv1alpha1.ParallelConfig) []string {
	var args []string

	if config.TensorParallelSize > 1 {
		args = append(args, "--tensor-parallel-size", strconv.Itoa(int(config.TensorParallelSize)))
	}

	if config.PipelineParallelSize > 1 {
		args = append(args, "--pipeline-parallel-size", strconv.Itoa(int(config.PipelineParallelSize)))
	}

	if config.DistributedExecutorBackend != nil {
		args = append(args, "--distributed-executor-backend", *config.DistributedExecutorBackend)
	}

	if config.WorkerUseRay != nil && *config.WorkerUseRay {
		args = append(args, "--worker-use-ray")
	}

	return args
}

func (r *VLLMReconciler) buildCacheConfigArgs(config *gelimerv1alpha1.CacheConfig) []string {
	var args []string

	if config.KvCacheDtype != "" && config.KvCacheDtype != "auto" {
		args = append(args, "--kv-cache-dtype", config.KvCacheDtype)
	}

	if config.GpuMemoryUtilization != "" && config.GpuMemoryUtilization != "0.9" {
		args = append(args, "--gpu-memory-utilization", config.GpuMemoryUtilization)
	}

	if config.SwapSpace != nil {
		args = append(args, "--swap-space", strconv.Itoa(int(*config.SwapSpace)))
	}

	if config.BlockSize != 16 {
		args = append(args, "--block-size", strconv.Itoa(int(config.BlockSize)))
	}

	return args
}

func (r *VLLMReconciler) buildLoRAConfigArgs(config *gelimerv1alpha1.LoRAConfig) []string {
	var args []string

	if len(config.LoraModules) > 0 {
		var moduleSpecs []string
		for _, module := range config.LoraModules {
			spec := fmt.Sprintf("%s=%s", module.Name, module.Path)
			if module.BaseModelName != nil {
				spec = fmt.Sprintf("%s=%s=%s", module.Name, module.Path, *module.BaseModelName)
			}
			moduleSpecs = append(moduleSpecs, spec)
		}
		args = append(args, "--lora-modules", strings.Join(moduleSpecs, ","))
	}

	if config.MaxLoras != 1 {
		args = append(args, "--max-loras", strconv.Itoa(int(config.MaxLoras)))
	}

	if config.MaxLoraRank != 16 {
		args = append(args, "--max-lora-rank", strconv.Itoa(int(config.MaxLoraRank)))
	}

	if config.LoraDtype != "" && config.LoraDtype != "auto" {
		args = append(args, "--lora-dtype", config.LoraDtype)
	}

	return args
}

func (r *VLLMReconciler) buildSchedulerConfigArgs(config *gelimerv1alpha1.SchedulerConfig) []string {
	var args []string

	if config.MaxNumSeqs != 256 {
		args = append(args, "--max-num-seqs", strconv.Itoa(int(config.MaxNumSeqs)))
	}

	if config.MaxNumBatchedTokens != nil {
		args = append(args, "--max-num-batched-tokens", strconv.Itoa(int(*config.MaxNumBatchedTokens)))
	}

	if config.EnableChunkedPrefill != nil && *config.EnableChunkedPrefill {
		args = append(args, "--enable-chunked-prefill")
	}

	if config.PreemptionMode != "" && config.PreemptionMode != "recompute" {
		args = append(args, "--preemption-mode", config.PreemptionMode)
	}

	return args
}
