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
	// vLLM 런타임 인수들
	// +optional
	VLLM *VLLMConfig `json:"vLLM,omitempty"`

	// 커스텀 런타임 설정
	// +optional
	Custom *CustomConfig `json:"custom,omitempty"`
}

// VLLMConfig는 vLLM 런타임의 명령줄 인수를 설정합니다
// https://docs.vllm.ai/en/stable/cli/index.html#serve
type VLLMConfig struct {
	// vLLM 서브명령어 (기본값: ["vllm", "serve"])
	// +optional
	Command []string `json:"command,omitempty"`

	// vLLM serve에 전달될 모든 인수들
	// 예: ["--tensor-parallel-size", "2", "--max-model-len", "4096"]
	// +optional
	Args []string `json:"args,omitempty"`
}

// CustomConfig는 사용자 정의 런타임 설정을 제공합니다
type CustomConfig struct {
	// 실행할 명령어
	// +optional
	Command []string `json:"command,omitempty"`
	// 명령어에 전달할 인수들
	// +optional
	Args []string `json:"args,omitempty"`
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
