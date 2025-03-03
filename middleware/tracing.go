package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TracingConfig 追踪配置
type TracingConfig struct {
	ServiceName    string // 服务名称
	ServiceVersion string // 服务版本
}

// DefaultTracingConfig 返回默认追踪配置
func DefaultTracingConfig() *TracingConfig {
	return &TracingConfig{
		ServiceName:    "unknown-service",
		ServiceVersion: "v0.0.0",
	}
}

// Tracing 返回OpenTelemetry追踪中间件
func Tracing(cfg *TracingConfig) gin.HandlerFunc {
	if cfg == nil {
		cfg = DefaultTracingConfig()
	}

	tracer := otel.GetTracerProvider().Tracer(
		cfg.ServiceName,
		trace.WithInstrumentationVersion(cfg.ServiceVersion),
	)

	propagator := otel.GetTextMapPropagator()

	return func(c *gin.Context) {
		// 从请求头中提取上下文
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// 创建新的span
		ctx, span := tracer.Start(ctx, c.Request.URL.Path)
		defer span.End()

		// 设置span属性
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.client_ip", c.ClientIP()),
		)

		// 将span上下文注入请求头
		propagator.Inject(ctx, propagation.HeaderCarrier(c.Request.Header))

		// 将追踪上下文传递给后续处理器
		c.Request = c.Request.WithContext(ctx)

		// 将 trace_id 设置到上下文中
		traceID := span.SpanContext().TraceID().String()
		c.Set("trace_id", traceID)

		c.Next()

		// 记录响应状态
		span.SetAttributes(
			attribute.Int64("http.status_code", int64(c.Writer.Status())),
			attribute.String("trace_id", traceID),
		)
	}
}