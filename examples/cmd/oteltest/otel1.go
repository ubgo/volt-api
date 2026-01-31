package main

import (
	"context"
	"errors"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
)

// https://opentelemetry.io/docs/languages/go/getting-started/
// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func setupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := newTracerProvider()
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterProvider, err := newMeterProvider()
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Set up logger provider.
	loggerProvider, err := newLoggerProvider()
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	return
}

func newResource() (*resource.Resource, error) {
	hostName, _ := os.Hostname()
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("wildai"), // Replace with your actual service name
		semconv.ServiceVersion("1.0.0"),
		semconv.HostName(hostName),
	), nil
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTracerProvider() (*trace.TracerProvider, error) {
	// Configure OTLP exporter for HyperDX
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint("in-otel.hyperdx.io"),
		// otlptracehttp.WithInsecure(), // Remove this if using HTTPS
		otlptracehttp.WithHeaders(map[string]string{
			"authorization": "xxxx", // Replace with your actual API key
		}),
	)

	traceExporter, err := otlptrace.New(context.Background(), client)
	if err != nil {
		return nil, err
	}

	// Create resource with service attributes
	res, err := newResource()
	if err != nil {
		return nil, err
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			trace.WithBatchTimeout(time.Second)),
		trace.WithResource(res),
	)
	return tracerProvider, nil
}

func newMeterProvider() (*metric.MeterProvider, error) {
	// Create OTLP metric exporter for HyperDX
	otlpMetricExporter, err := otlpmetrichttp.New(context.Background(),
		otlpmetrichttp.WithEndpoint("in-otel.hyperdx.io"),
		// otlpmetrichttp.WithInsecure(), // Remove this if using HTTPS
		otlpmetrichttp.WithHeaders(map[string]string{
			"authorization": "xxxx", // Replace with your actual API key
		}),
	)
	if err != nil {
		return nil, err
	}

	// Create stdout metric exporter for local debugging
	stdoutMetricExporter, err := stdoutmetric.New()
	if err != nil {
		return nil, err
	}

	// Create resource with service attributes
	res, err := newResource()
	if err != nil {
		return nil, err
	}

	// Create meter provider with multiple readers
	meterProvider := metric.NewMeterProvider(
		// OTLP reader for HyperDX
		metric.WithReader(metric.NewPeriodicReader(otlpMetricExporter,
			metric.WithInterval(3*time.Second))),
		// Stdout reader for local debugging
		metric.WithReader(metric.NewPeriodicReader(stdoutMetricExporter,
			metric.WithInterval(3*time.Second))),
		metric.WithResource(res),
	)
	return meterProvider, nil
}

func newLoggerProvider() (*log.LoggerProvider, error) {

	logExporter, err := otlploghttp.New(context.Background(),
		otlploghttp.WithEndpoint("in-otel.hyperdx.io"),
		// otlploghttp.WithInsecure(), // Remove this if using HTTPS
		otlploghttp.WithHeaders(map[string]string{
			"authorization": "xxxx", // Replace with your actual API key
		}),
	)

	// // logExporter, err := stdoutlog.New()
	// if err != nil {
	// 	return nil, err
	// }

	// Create resource with service attributes
	res, err := newResource()
	if err != nil {
		return nil, err
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
		log.WithResource(res),
	)
	return loggerProvider, nil
}

// func newTracerProvider() (*trace.TracerProvider, error) {
// 	traceExporter, err := stdouttrace.New(
// 		stdouttrace.WithPrettyPrint())
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Create resource with service attributes
// 	res, err := newResource()
// 	if err != nil {
// 		return nil, err
// 	}

// 	tracerProvider := trace.NewTracerProvider(
// 		trace.WithBatcher(traceExporter,
// 			// Default is 5s. Set to 1s for demonstrative purposes.
// 			trace.WithBatchTimeout(time.Second)),
// 		trace.WithResource(res),
// 	)
// 	return tracerProvider, nil
// }

// func newMeterProvider() (*metric.MeterProvider, error) {
// 	metricExporter, err := stdoutmetric.New()
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Create resource with service attributes
// 	res, err := newResource()
// 	if err != nil {
// 		return nil, err
// 	}

// 	meterProvider := metric.NewMeterProvider(
// 		metric.WithReader(metric.NewPeriodicReader(metricExporter,
// 			// Default is 1m. Set to 3s for demonstrative purposes.
// 			metric.WithInterval(3*time.Second))),
// 		metric.WithResource(res),
// 	)
// 	return meterProvider, nil
// }

// func newLoggerProvider() (*log.LoggerProvider, error) {
// 	logExporter, err := stdoutlog.New()
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Create resource with service attributes
// 	res, err := newResource()
// 	if err != nil {
// 		return nil, err
// 	}

// 	loggerProvider := log.NewLoggerProvider(
// 		log.WithProcessor(log.NewBatchProcessor(logExporter)),
// 		log.WithResource(res),
// 	)
// 	return loggerProvider, nil
// }
