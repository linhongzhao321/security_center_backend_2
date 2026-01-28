package contextextension

import (
	"context"

	"github.com/google/uuid"
)

const modKey = `run_mod`

func WithMod(ctx context.Context, mod RunMod) context.Context {
	return context.WithValue(ctx, modKey, mod)
}

// FindCtxKey
func currentMod(ctx context.Context) RunMod {
	val := ctx.Value(modKey)
	if val == nil {
		return RunModLocal
	}

	valString, ok := val.(RunMod)
	if !ok {
		return RunModLocal
	}

	return valString
}

func IsDebug(ctx context.Context) bool {
	return currentMod(ctx) != RunModRelease
}

const TraceIDKey = `_vino_trace_id`

func GenTraceID(ctx context.Context) context.Context {
	traceID := uuid.New().String()
	return WithTraceID(ctx, traceID)
}

func WithTraceID(ctx context.Context, logID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, logID)
}
