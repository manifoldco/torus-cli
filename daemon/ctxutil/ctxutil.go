// Package ctxutil contains utilities for using Contexts.
package ctxutil

import "context"

// ErrIfDone returns the Context's error if done. It is a convenience method
// for  long-running routines that don't have cancellable goroutines.
func ErrIfDone(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
