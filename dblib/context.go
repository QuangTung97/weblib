package dblib

import "context"

var ctxKey = new(int)

type contextValueType struct {
	isReadonly bool
	tx         Transaction
}

func getFromContext(ctx context.Context) (*contextValueType, bool) {
	val, ok := ctx.Value(ctxKey).(*contextValueType)
	return val, ok
}

func setToContext(ctx context.Context, val *contextValueType) context.Context {
	return context.WithValue(ctx, ctxKey, val)
}
