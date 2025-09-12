package llm

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gelimerv1alpha1 "zzzinho.busan/api/v1alpha1"
)

type RuntimeHandler interface {
	Do(ctx context.Context, llm *gelimerv1alpha1.LLM) (ctrl.Result, error)
}

type BaseHandler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (h *BaseHandler) ensureFinalizer(ctx context.Context, llm *gelimerv1alpha1.LLM) error {
	// TODO: Implement
	return nil
}
