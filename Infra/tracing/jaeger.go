// pkg/jaeger/jaeger.go

package tracing

import (
	"fmt"
	"io"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-client-go/rpcmetrics"
	"github.com/uber/jaeger-lib/metrics"
)

type Config struct {
	ServiceName  string
	JaegerHost   string  // Jaeger collector host, e.g., "localhost"
	JaegerPort   string  // Jaeger collector port, e.g., "14250"
	SamplingRate float64 // 0.0 to 1.0
	LogSpans     bool
}

type Tracer struct {
	opentracing.Tracer
	closer io.Closer
}

func InitJaeger(cfg Config) error {
	if cfg.SamplingRate == 0 {
		cfg.SamplingRate = 1.0
	}

	jaegerCfg := jaegercfg.Configuration{
		ServiceName: cfg.ServiceName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: cfg.SamplingRate,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:          cfg.LogSpans,
			CollectorEndpoint: fmt.Sprintf("http://%s:%s/api/traces", cfg.JaegerHost, cfg.JaegerPort),
		},
	}

	tracer, _, err := jaegerCfg.NewTracer(
		jaegercfg.Logger(jaegerlog.StdLogger),
		jaegercfg.Metrics(metrics.NullFactory),
		jaegercfg.Observer(rpcmetrics.NewObserver(metrics.NullFactory, rpcmetrics.DefaultNameNormalizer)),
	)
	if err != nil {
		return fmt.Errorf("failed to init jaeger: %w", err)
	}

	opentracing.SetGlobalTracer(tracer)

	return nil
}

// Close closes the tracer
func (t *Tracer) Close() error {
	if t.closer != nil {
		return t.closer.Close()
	}
	return nil
}
