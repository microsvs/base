package tracing

import (
	"time"

	"github.com/microsvs/base/pkg/log"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/rpcmetrics"
	"github.com/uber/jaeger-lib/metrics"
)

// Init creates a new instance of Jaeger tracer.
func Init(serviceName string, metricsFactory metrics.Factory, backendHostPort string) opentracing.Tracer {
	var err error
	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  jaeger.SamplerTypeRateLimiting,
			Param: 10 * 10000,
		},
	}
	var sender jaeger.Transport
	if sender, err = jaeger.NewUDPTransport(backendHostPort, 0); err != nil {
		log.ErrorRaw("cannot initialize UDP sender", err)
	}
	tracer, _, err := cfg.New(
		serviceName,
		config.Reporter(jaeger.NewRemoteReporter(
			sender,
			jaeger.ReporterOptions.BufferFlushInterval(1*time.Second),
		)),
		config.Metrics(metricsFactory),
		config.Observer(rpcmetrics.NewObserver(metricsFactory, rpcmetrics.DefaultNameNormalizer)),
	)
	if err != nil {
		log.ErrorRaw("cannot initialize Jaeger Tracer", err)
	}
	return tracer
}
