package llm

import (
	"context"
	"strconv"

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
	log.Info("vLLM 리콘사일링 시작", "namespace", llm.Namespace, "name", llm.Name)

	// 디플로이먼트 생성
	deployment := r.createDeploymentManifest(llm)
	if err := r.Create(ctx, deployment); err != nil {
		log.Error(err, "디플로이먼트 생성 실패")
		return ctrl.Result{}, err
	}

	// 서비스 생성 (추후 구현)

	return ctrl.Result{}, nil
}

func (r *VLLMReconciler) createDeploymentManifest(llm *gelimerv1alpha1.LLM) *appsv1.Deployment {
	// 서비스 계정 이름 설정
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
							Name:    containerName,
							Image:   llm.Spec.Image,
							Command: r.createVLLMCommand(llm),
							Args:    r.createVLLMArgs(llm),
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

// createVLLMCommand는 vLLM 컨테이너의 명령어를 생성합니다
func (r *VLLMReconciler) createVLLMCommand(llm *gelimerv1alpha1.LLM) []string {
	// vLLM 설정이 있고 Command가 지정된 경우 사용
	if llm.Spec.RuntimeConfig.VLLM != nil && len(llm.Spec.RuntimeConfig.VLLM.Command) > 0 {
		return llm.Spec.RuntimeConfig.VLLM.Command
	}

	// 기본 vLLM 명령어
	return []string{"vllm", "serve"}
}

// createVLLMArgs는 vLLM 컨테이너의 인수를 생성합니다
func (r *VLLMReconciler) createVLLMArgs(llm *gelimerv1alpha1.LLM) []string {
	var args []string

	// 모델 이름 (필수)
	args = append(args, llm.Spec.Model)

	// 포트 설정 (필수)
	args = append(args, "--port", strconv.Itoa(int(llm.Spec.Port)))

	// vLLM 설정에서 추가 인수 처리
	if llm.Spec.RuntimeConfig.VLLM != nil && len(llm.Spec.RuntimeConfig.VLLM.Args) > 0 {
		args = append(args, llm.Spec.RuntimeConfig.VLLM.Args...)
	}

	return args
}
