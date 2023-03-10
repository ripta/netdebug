package result

import "context"

type contextKey int

var resultKey contextKey

func FromContext(ctx context.Context) Result {
	val := ctx.Value(resultKey)
	if val == nil {
		return Result{}
	}

	return val.(Result)
}

func WithResult(ctx context.Context, res Result) context.Context {
	return context.WithValue(ctx, resultKey, res)
}
