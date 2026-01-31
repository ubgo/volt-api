package gotel

import (
	"context"
	"os"
	"saas/pkg/utils"
	"time"

	"github.com/cohesivestack/valgo"
	"github.com/spf13/cast"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	globallog "go.opentelemetry.io/otel/log/global"
	nooplog "go.opentelemetry.io/otel/log/noop"
	noopmeter "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
)

// Config holds the configuration for the telemetry.
type Config struct {
	ServiceName     string
	ServiceVersion  string
	LoggerExporters []log.Processor
	TracerExporters []trace.SpanExporter
	MetricExporters []metric.Reader
	EnableMetrics   bool
	EnableTraces    bool
	EnableLogs      bool
}

type Telemetry struct {
	propagator propagation.TextMapPropagator
	res        *resource.Resource
	tp         *trace.TracerProvider
	lp         *log.LoggerProvider
	mp         *metric.MeterProvider
	config     Config
}

var _telemetry *Telemetry

func MustGetTelemetry() *Telemetry {
	if _telemetry == nil {
		panic("telemetry not initialized")
	}
	return _telemetry
}

func NewTelemetry(ctx context.Context, cfg Config) (*Telemetry, error) {
	val := valgo.Is(
		valgo.String(cast.ToString(cfg.ServiceName), "ServiceName").Not().Blank(),
		valgo.String(cast.ToString(cfg.ServiceVersion), "ServiceVersion").Not().Blank(),
	)

	if !val.Valid() {
		return nil, utils.ValgoFirstError(val)
	}

	rp := newResource(cfg.ServiceName, cfg.ServiceVersion)

	tp, err := newTracerProvider(rp, cfg.TracerExporters)
	if err != nil {
		return nil, err
	}

	lp, err := newLoggerProvider(rp, cfg.LoggerExporters)
	if err != nil {
		return nil, err
	}

	mp, err := newMeterProvider(rp, cfg.MetricExporters)
	if err != nil {
		return nil, err
	}

	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	_telemetry = &Telemetry{
		propagator: propagator,
		res:        rp,
		tp:         tp,
		lp:         lp,
		mp:         mp,
		config:     cfg,
	}

	return _telemetry, nil
}

// set global
func (t *Telemetry) SetGlobal() {
	if t.config.EnableTraces {
		otel.SetTracerProvider(t.tp)
	} else {
		otel.SetTracerProvider(nooptrace.NewTracerProvider())
	}

	if t.config.EnableLogs {
		globallog.SetLoggerProvider(t.lp)
	} else {
		globallog.SetLoggerProvider(nooplog.NewLoggerProvider())
	}

	if t.config.EnableMetrics {
		otel.SetMeterProvider(t.mp)
	} else {
		otel.SetMeterProvider(noopmeter.NewMeterProvider())
	}

	otel.SetTextMapPropagator(t.propagator)
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	var err error

	err = t.lp.Shutdown(ctx)
	if err != nil {
		return err
	}

	err = t.mp.Shutdown(ctx)
	if err != nil {
		return err
	}

	err = t.tp.Shutdown(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (t *Telemetry) GetLoggerProvider() *log.LoggerProvider {
	return t.lp
}

// TraceStart starts a new span with the given name. The span must be ended by calling End.
/*
	_, span := a.telem.TraceStart(c.Request.Context(), "get_something")
	defer span.End()
*/
func (t *Telemetry) TraceStart(ctx context.Context, name string) (context.Context, oteltrace.Span) {
	return t.tp.Tracer(name).Start(ctx, name)
}

// newResource creates a new OTEL resource with the service name and version.
func newResource(serviceName string, serviceVersion string) *resource.Resource {
	hostName, _ := os.Hostname()

	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
		semconv.ServiceVersion(serviceVersion),
		semconv.HostName(hostName),
	)
}

// newLoggerProvider creates a new logger provider with the OTLP gRPC exporter.
func newLoggerProvider(res *resource.Resource, exporters []log.Processor) (*log.LoggerProvider, error) {
	args := []log.LoggerProviderOption{
		log.WithResource(res),
	}

	// Add each processor
	for _, processor := range exporters {
		args = append(args, log.WithProcessor(processor))
	}

	lp := log.NewLoggerProvider(args...)

	return lp, nil
}

func newTracerProvider(res *resource.Resource, exporters []trace.SpanExporter) (*trace.TracerProvider, error) {
	args := []trace.TracerProviderOption{
		trace.WithResource(res),
	}

	// Add each exporter as a separate batcher
	for _, exporter := range exporters {
		args = append(args, trace.WithBatcher(exporter, trace.WithBatchTimeout(time.Second)))
	}

	tracerProvider := trace.NewTracerProvider(args...)
	return tracerProvider, nil
}

// newMeterProvider creates a new meter provider with the OTLP gRPC exporter.
func newMeterProvider(res *resource.Resource, exporters []metric.Reader) (*metric.MeterProvider, error) {
	args := []metric.Option{
		metric.WithResource(res),
	}

	// Add each reader
	for _, reader := range exporters {
		args = append(args, metric.WithReader(reader))
	}

	mp := metric.NewMeterProvider(args...)
	otel.SetMeterProvider(mp)

	return mp, nil
}

type PreExporterConfig struct {
	Config       Config
	EnableStdout bool
	EnableHttp   bool

	// there is weird situation we got into grafana as it needs this url `otlp-gateway-prod-eu-west-2.grafana.net/otlp` so when using
	// as endpoint we got error invalid URL escape "%2F" i created a github issue here: https://github.com/open-telemetry/opentelemetry-go/issues/7180
	// so for now the solution is to pass full endpoint URLS like `https://otlp-gateway-prod-eu-west-2.grafana.net/otlp/v1/traces`
	// but we will use `isEndpoingUrl=true` and endpoing=https://otlp-gateway-prod-eu-west-2.grafana.net/otlp and the we will auto append `/v1/traces`
	// for hyperdx we will use `isEndpoingUrl=false` and endpoing=in-otel.hyperdx.io
	IsEndpoingUrl bool
	Endpoint      string

	ApiKey string

	// if true then add "BASIC token" to authorization header as hyperdx does not works with basic
	// hyperdx(authorization=<YOUR_HYPERDX_API_KEY_HERE>) but grafana requires (Basic <GRAFANA_API_KEY>)
	PrefixBasic bool

	Insecure bool
}

func NewTelemetryWithPreExporters(ctx context.Context, cfg PreExporterConfig) (*Telemetry, error) {
	if cfg.EnableHttp {
		val := valgo.Is(
			valgo.String(cast.ToString(cfg.Endpoint), "Endpoint").Not().Blank(),
			valgo.String(cast.ToString(cfg.ApiKey), "ApiKey").Not().Blank(),
		)

		if !val.Valid() {
			return nil, utils.ValgoFirstError(val)
		}
	}

	logExports := []log.Processor{}
	if cfg.EnableStdout && cfg.Config.EnableLogs {
		err := newStdoutLogExporter(&logExports)
		if err != nil {
			return nil, err
		}
	}

	if cfg.EnableHttp && cfg.Config.EnableLogs {
		err := newHttpLoggerExporter(cfg.Endpoint, cfg.IsEndpoingUrl, cfg.PrefixBasic, cfg.ApiKey, cfg.Insecure, &logExports)
		if err != nil {
			return nil, err
		}
	}

	tracerExports := []trace.SpanExporter{}
	if cfg.EnableStdout && cfg.Config.EnableTraces {
		err := newStdoutTracerExporter(&tracerExports)
		if err != nil {
			return nil, err
		}
	}

	if cfg.EnableHttp && cfg.Config.EnableTraces {
		err := newHttpTracerExporter(cfg.Endpoint, cfg.IsEndpoingUrl, cfg.PrefixBasic, cfg.ApiKey, cfg.Insecure, &tracerExports)
		if err != nil {
			return nil, err
		}
	}
	metricExports := []metric.Reader{}
	// stdout metric exporter should never be added to the slice if disabled if you disable on global level
	// it won't work it will still stdout in the console not sure why it seems exporter works independently of global level
	if cfg.EnableStdout && cfg.Config.EnableMetrics {
		err := newStdoutMetricExporter(&metricExports)
		if err != nil {
			return nil, err
		}
	}

	if cfg.EnableHttp && cfg.Config.EnableMetrics {
		err := newHttpMetricExporter(cfg.Endpoint, cfg.IsEndpoingUrl, cfg.PrefixBasic, cfg.ApiKey, cfg.Insecure, &metricExports)
		if err != nil {
			return nil, err
		}
	}

	return NewTelemetry(ctx, Config{
		ServiceName:     cfg.Config.ServiceName,
		ServiceVersion:  cfg.Config.ServiceVersion,
		EnableLogs:      cfg.Config.EnableLogs,
		EnableTraces:    cfg.Config.EnableTraces,
		EnableMetrics:   cfg.Config.EnableMetrics,
		LoggerExporters: logExports,
		TracerExporters: tracerExports,
		MetricExporters: metricExports,
	})
}

func newStdoutLogExporter(logExports *[]log.Processor) error {
	stdoutLogExporter, err := stdoutlog.New()
	if err != nil {
		return err
	}

	*logExports = append(*logExports, log.NewBatchProcessor(stdoutLogExporter))

	return nil
}

func newStdoutTracerExporter(tracerExports *[]trace.SpanExporter) error {
	stdoutTracerExporter, err := stdouttrace.New()
	if err != nil {
		return err
	}

	*tracerExports = append(*tracerExports, stdoutTracerExporter)

	return nil
}

func newStdoutMetricExporter(metricExports *[]metric.Reader) error {
	stdoutMetricExporter, err := stdoutmetric.New()
	if err != nil {
		return err
	}

	*metricExports = append(*metricExports, metric.NewPeriodicReader(stdoutMetricExporter, metric.WithInterval(time.Second)))

	return nil
}

func newHttpTracerExporter(endpoint string, isEndpoingUrl bool, prefixBasic bool, apiKey string, insecure bool, tracerExports *[]trace.SpanExporter) error {
	authKey := apiKey
	if prefixBasic {
		authKey = "Basic " + apiKey
	}

	options := []otlptracehttp.Option{
		otlptracehttp.WithHeaders(map[string]string{
			"authorization": authKey, // Replace with your actual API key
		}),
	}

	if isEndpoingUrl {
		options = append(options, otlptracehttp.WithEndpointURL(endpoint+"/v1/traces"))
	} else {
		options = append(options, otlptracehttp.WithEndpoint(endpoint))
	}

	if insecure {
		options = append(options, otlptracehttp.WithInsecure())
	}

	client := otlptracehttp.NewClient(options...)

	traceExporter, err := otlptrace.New(context.Background(), client)
	if err != nil {
		return err
	}

	*tracerExports = append(*tracerExports, traceExporter)

	return nil
}

func newHttpMetricExporter(endpoint string, isEndpoingUrl bool, prefixBasic bool, apiKey string, insecure bool, metricExports *[]metric.Reader) error {
	authKey := apiKey
	if prefixBasic {
		authKey = "Basic " + apiKey
	}

	options := []otlpmetrichttp.Option{
		otlpmetrichttp.WithHeaders(map[string]string{
			"authorization": authKey, // Replace with your actual API key
		}),
	}

	if isEndpoingUrl {
		options = append(options, otlpmetrichttp.WithEndpointURL(endpoint+"/v1/metrics"))
	} else {
		options = append(options, otlpmetrichttp.WithEndpoint(endpoint))
	}

	if insecure {
		options = append(options, otlpmetrichttp.WithInsecure())
	}

	metricExporter, err := otlpmetrichttp.New(context.Background(), options...)
	if err != nil {
		return err
	}

	*metricExports = append(*metricExports, metric.NewPeriodicReader(metricExporter, metric.WithInterval(time.Second)))

	return nil
}

func newHttpLoggerExporter(endpoint string, isEndpoingUrl bool, prefixBasic bool, apiKey string, insecure bool, loggerExports *[]log.Processor) error {
	authKey := apiKey
	if prefixBasic {
		authKey = "Basic " + apiKey
	}

	options := []otlploghttp.Option{
		otlploghttp.WithHeaders(map[string]string{
			"authorization": authKey, // Replace with your actual API key
		}),
	}

	if isEndpoingUrl {
		options = append(options, otlploghttp.WithEndpointURL(endpoint+"/v1/logs"))
	} else {
		options = append(options, otlploghttp.WithEndpoint(endpoint))
	}

	if insecure {
		options = append(options, otlploghttp.WithInsecure())
	}

	loggerExporter, err := otlploghttp.New(context.Background(), options...)
	if err != nil {
		return err
	}

	*loggerExports = append(*loggerExports, log.NewBatchProcessor(loggerExporter))

	return nil
}
