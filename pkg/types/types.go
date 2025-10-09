package types

import "errors"

const (
	InProgress     = "reconciliation in progress"
	FinalizerAdded = "finalizer added"
)


var (
	ErrPermanent = errors.New("permanent failure")
)
