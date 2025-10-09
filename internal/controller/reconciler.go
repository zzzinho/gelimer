package controller

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)



type Reconciler interface {
	Do(ctx context.Context, obj client.Object) (string, error)
}
