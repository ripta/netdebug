package result

import "context"

type contextKey int

var resultKey contextKey

func FromContext(ctx context.Context) Result {
	return ctx.Value(resultKey).(Result)
}

func WithResult(ctx context.Context, res Result) context.Context {
	return context.WithValue(ctx, resultKey, res)
}
