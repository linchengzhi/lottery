package middleware

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"time"
)

// TracingMiddleware returns a middleware that enables distributed tracing
func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tracer := opentracing.GlobalTracer()
		// 从请求头中提取 span context
		spanCtx, err := tracer.Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(c.Request.Header),
		)

		// 创建新的 span
		operationName := fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())
		var span opentracing.Span
		if err != nil {
			// 如果无法从头部提取，创建新的 root span
			span = tracer.StartSpan(operationName)
		} else {
			// 否则创建子 span
			span = tracer.StartSpan(operationName, opentracing.ChildOf(spanCtx))
		}
		defer span.Finish()

		// 设置 span 标签
		//ext.HTTPMethod.Set(span, c.Request.Method)
		//ext.HTTPUrl.Set(span, c.Request.URL.String())
		//ext.Component.Set(span, "gin")

		//// 设置请求头
		//if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		//	span.SetTag("request.id", requestID)
		//}

		//// 设置用户相关信息（如果有的话）
		//if userID := c.GetHeader("X-User-ID"); userID != "" {
		//	span.SetTag("user.id", userID)
		//}

		// 将 span 注入到请求上下文中
		c.Set("span", span)
		c.Set("tracingContext", opentracing.ContextWithSpan(c.Request.Context(), span))

		// 处理请求
		c.Next()

		// 记录响应状态
		status := c.Writer.Status()
		ext.HTTPStatusCode.Set(span, uint16(status))
		if status >= 400 {
			ext.Error.Set(span, true)
		}

		// 记录任何错误
		if len(c.Errors) > 0 {
			span.SetTag("gin.errors", c.Errors.String())
		}
	}
}

// StartSpanFromContext 用于在处理函数中创建子 span
func StartSpanFromContext(c *gin.Context, operationName string) (opentracing.Span, error) {
	spanInterface, exists := c.Get("span")
	if !exists {
		return nil, fmt.Errorf("no parent span found in context")
	}

	parentSpan := spanInterface.(opentracing.Span)
	span := opentracing.StartSpan(
		operationName,
		opentracing.ChildOf(parentSpan.Context()),
	)

	return span, nil
}

// WithTimeoutAndSpan 创建一个同时包含超时和 span 信息的 context
func WithTimeoutAndSpan(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	// 1. 保存原始的 span
	//span := opentracing.SpanFromContext(ctx)

	// 2. 创建超时 context
	timeoutCtx, cancel := context.WithTimeout(context.Background(), timeout)

	// 3. 如果有 span，则添加到新的 context 中
	//if span != nil {
	//	return opentracing.ContextWithSpan(timeoutCtx, span), cancel
	//}

	return timeoutCtx, cancel
}
