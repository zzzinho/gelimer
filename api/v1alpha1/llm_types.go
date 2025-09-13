/*
Copyright 2025.

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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: 다른 타입도 추가
// +kubebuilder:validation:Enum=vLLM
type RuntimeType string

const (
	RuntimeTypeVLLM RuntimeType = "vLLM"
)

// LLMSpec defines the desired state of LLM.
type LLMSpec struct {
	// +kubebuilder:validation:Required
	Model string `json:"model"`

	// +kubebuilder:validation:Required
	Runtime RuntimeType `json:"runtime"`

	// +optional
	RuntimeConfig RuntimeConfig `json:"runtimeConfig,omitempty"`

	// +kubebuilder:validation:Required
	Replicas int32 `json:"replicas"`

	// +kubebuilder:validation:Required
	Image string `json:"image"`

	// +kubebuilder:validation:Required
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy"`

	// +kubebuilder:validation:Required
	Port int32 `json:"port"`

	// +kubebuilder:validation:Required
	Resources corev1.ResourceRequirements `json:"resources"`

	// +optional
	Env []corev1.EnvVar `json:"env"`

	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts"`

	// +optional
	Volumes []corev1.Volume `json:"volumes"`

	// +optional
	ServiceAccountName string `json:"serviceAccountName"`
}

type RuntimeConfig struct {
	// +optional
	VLLM *VLLMConfig `json:"vLLM,omitempty"`

	// +optional
	Custom *CustomConfig `json:"custom,omitempty"`
}

// vLLM Config
// https://docs.vllm.ai/en/stable/cli/index.html#serve
type VLLMConfig struct {
	// Model configuration (optional, can use default model from LLMSpec)
	// +optional
	Model *ModelConfig `json:"model,omitempty"`

	// Load configuration
	// +optional
	Load *LoadConfig `json:"load,omitempty"`

	// Decoding configuration
	// +optional
	Decoding *DecodingConfig `json:"decoding,omitempty"`

	// Parallel configuration
	// +optional
	Parallel *ParallelConfig `json:"parallel,omitempty"`

	// Cache configuration
	// +optional
	Cache *CacheConfig `json:"cache,omitempty"`

	// LoRA configuration
	// +optional
	LoRA *LoRAConfig `json:"lora,omitempty"`

	// Scheduler configuration
	// +optional
	Scheduler *SchedulerConfig `json:"scheduler,omitempty"`

	// Additional vLLM serve arguments as key-value pairs
	// +optional
	ExtraArgs map[string]string `json:"extraArgs,omitempty"`
}

// ModelConfig contains essential vLLM model configuration options
type ModelConfig struct {
	// Task type for the model
	// +kubebuilder:validation:Enum=auto;classify;draft;embed;embedding;generate;reward;score;transcription
	// +kubebuilder:default=auto
	Task string `json:"task,omitempty"`

	// Tokenizer name or path (if different from model)
	// +optional
	Tokenizer *string `json:"tokenizer,omitempty"`

	// Tokenizer mode
	// +kubebuilder:validation:Enum=auto;slow;fast
	// +kubebuilder:default=auto
	TokenizerMode string `json:"tokenizerMode,omitempty"`

	// Trust remote code
	// +optional
	TrustRemoteCode *bool `json:"trustRemoteCode,omitempty"`

	// Maximum model sequence length
	// +optional
	MaxModelLen *int32 `json:"maxModelLen,omitempty"`
}

// LoadConfig contains model loading configuration
type LoadConfig struct {
	// Download directory for cached weights and files
	// +optional
	DownloadDir *string `json:"downloadDir,omitempty"`

	// Load format of the model weights
	// +kubebuilder:validation:Enum=auto;pt;safetensors;npcache;dummy;tensorizer;sharded_state;gguf;bitsandbytes;mistral
	// +kubebuilder:default=auto
	LoadFormat string `json:"loadFormat,omitempty"`

	// Data type for model weights and activations
	// +kubebuilder:validation:Enum=auto;half;float16;bfloat16;float;float32
	// +kubebuilder:default=auto
	Dtype string `json:"dtype,omitempty"`

	// Quantization method
	// +kubebuilder:validation:Enum=aqlm;awq;deepspeedfp;tpu_int8;fp8;fbgemm_fp8;modelopt;marlin;gguf;gptq_marlin_24;gptq_marlin;awq_marlin;gptq;squeezellm;compressed-tensors;bitsandbytes;qqq;experts_int8;neuron_quant;ipex
	// +optional
	QuantizationMethod *string `json:"quantizationMethod,omitempty"`

	// Device type for vLLM execution
	// +kubebuilder:validation:Enum=auto;cuda;neuron;cpu;openvino;tpu;xpu
	// +kubebuilder:default=auto
	Device string `json:"device,omitempty"`
}

// DecodingConfig contains decoding configuration
type DecodingConfig struct {
	// Guided decoding backend
	// +kubebuilder:validation:Enum=outlines;lm-format-enforcer
	// +kubebuilder:default=outlines
	GuidedDecodingBackend string `json:"guidedDecodingBackend,omitempty"`

	// Maximum number of log probabilities to return per output token
	// +optional
	MaxLogprobs *int32 `json:"maxLogprobs,omitempty"`

	// Disable sliding window attention
	// +optional
	DisableSlidingWindow *bool `json:"disableSlidingWindow,omitempty"`
}

// ParallelConfig contains parallel processing configuration
type ParallelConfig struct {
	// Tensor parallel size
	// +kubebuilder:default=1
	TensorParallelSize int32 `json:"tensorParallelSize,omitempty"`

	// Pipeline parallel size
	// +kubebuilder:default=1
	PipelineParallelSize int32 `json:"pipelineParallelSize,omitempty"`

	// Distributed executor backend
	// +kubebuilder:validation:Enum=ray;mp
	// +optional
	DistributedExecutorBackend *string `json:"distributedExecutorBackend,omitempty"`

	// Worker use Ray for distributed serving
	// +optional
	WorkerUseRay *bool `json:"workerUseRay,omitempty"`
}

// CacheConfig contains KV cache configuration
type CacheConfig struct {
	// Data type for KV cache storage
	// +kubebuilder:validation:Enum=auto;fp8;fp8_e5m2;fp8_e4m3
	// +kubebuilder:default=auto
	KvCacheDtype string `json:"kvCacheDtype,omitempty"`

	// GPU memory utilization for KV cache
	// +kubebuilder:default="0.9"
	GpuMemoryUtilization string `json:"gpuMemoryUtilization,omitempty"`

	// Swap space for KV cache (in GiB)
	// +optional
	SwapSpace *int32 `json:"swapSpace,omitempty"`

	// Block size for paged attention
	// +kubebuilder:default=16
	BlockSize int32 `json:"blockSize,omitempty"`
}

// LoRAConfig contains LoRA configuration
type LoRAConfig struct {
	// LoRA modules
	// +optional
	LoraModules []LoraModule `json:"loraModules,omitempty"`

	// Maximum LoRAs
	// +kubebuilder:default=1
	MaxLoras int32 `json:"maxLoras,omitempty"`

	// Maximum LoRA rank
	// +kubebuilder:default=16
	MaxLoraRank int32 `json:"maxLoraRank,omitempty"`

	// LoRA dtype
	// +kubebuilder:validation:Enum=auto;float16;bfloat16;float32
	// +kubebuilder:default=auto
	LoraDtype string `json:"loraDtype,omitempty"`
}

// SchedulerConfig contains scheduler configuration
type SchedulerConfig struct {
	// Maximum number of sequences in a batch
	// +kubebuilder:default=256
	MaxNumSeqs int32 `json:"maxNumSeqs,omitempty"`

	// Maximum number of batched tokens
	// +optional
	MaxNumBatchedTokens *int32 `json:"maxNumBatchedTokens,omitempty"`

	// Maximum number of padding tokens
	// +kubebuilder:default=256
	MaxPaddings int32 `json:"maxPaddings,omitempty"`

	// Enable chunked prefill
	// +optional
	EnableChunkedPrefill *bool `json:"enableChunkedPrefill,omitempty"`

	// Preemption mode
	// +kubebuilder:validation:Enum=swap;recompute
	// +kubebuilder:default=recompute
	PreemptionMode string `json:"preemptionMode,omitempty"`
}

// LoraModule represents a LoRA module configuration
type LoraModule struct {
	Name string `json:"name"`
	Path string `json:"path"`
	// +optional
	BaseModelName *string `json:"baseModelName,omitempty"`
}

type CustomConfig struct {
	// +optional
	Command []string `json:"command,omitempty"`
	// +optional
	Args []string `json:"args,omitempty"`
	// +optional
	Config map[string]string `json:"config,omitempty"`
}

// LLMStatus defines the observed state of LLM.
type LLMStatus struct {
	// +optional
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// LLM is the Schema for the llms API.
type LLM struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LLMSpec   `json:"spec,omitempty"`
	Status LLMStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LLMList contains a list of LLM.
type LLMList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LLM `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LLM{}, &LLMList{})
}
