package audit

import "context"

type requestMetaKey struct{}

type RequestMeta struct {
	IPAddress string
	UserAgent string
}

func WithRequestMeta(ctx context.Context, meta RequestMeta) context.Context {
	return context.WithValue(ctx, requestMetaKey{}, meta)
}

func RequestMetaFromContext(ctx context.Context) RequestMeta {
	meta, ok := ctx.Value(requestMetaKey{}).(RequestMeta)
	if !ok {
		return RequestMeta{}
	}
	return meta
}
