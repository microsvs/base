package utils

import (
	opentracing "github.com/opentracing/opentracing-go"
)

func SetGlobalTracer(tracer opentracing.Tracer) {
	opentracing.SetGlobalTracer(tracer)
}

func GetGlobalTracer() opentracing.Tracer {
	return opentracing.GlobalTracer()
}
