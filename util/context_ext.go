package util

import (
	"context"
	"time"
)

const (
	ctxData = "ctx_data"
)

func WithRequestInfo(ctx context.Context, requestId string, requestTime time.Time) context.Context {
	m := map[string]interface{}{
		"request_id":   requestId,
		"request_time": requestTime,
	}
	return context.WithValue(ctx, ctxData, m)
}

func FromRequestContext(ctx context.Context) (string, time.Time, bool) {
	if v := ctx.Value(ctxData); v != nil {
		m, ok := v.(map[string]interface{})
		if !ok {
			return "", time.Time{}, false
		}
		reqId, _ := m["request_id"].(string)
		reqTime, _ := m["request_time"].(time.Time)
		return reqId, reqTime, true
	}
	return "", time.Time{}, false
}
