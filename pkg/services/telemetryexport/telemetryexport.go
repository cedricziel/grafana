package telemetryexport

import (
	"context"
	"strings"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/setting"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/config/configtelemetry"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Service interface {
	ExportTraces(ctx context.Context, trace ptrace.Traces) error
	ExportMetrics(ctx context.Context, metrics pmetric.Metrics) error
	ExportLogs(ctx context.Context, logs plog.Logs) error

	Start(ctx context.Context, host component.Host) error
}

type TelemetryExporterService struct {
	traceExporter  exporter.Traces
	metricExporter exporter.Metrics
	logExporter    exporter.Logs

	logger log.Logger
}

func (s TelemetryExporterService) ExportTraces(ctx context.Context, trace ptrace.Traces) error {
	return s.traceExporter.ConsumeTraces(ctx, trace)
}

func (s TelemetryExporterService) ExportMetrics(ctx context.Context, metrics pmetric.Metrics) error {
	return nil
}

func (s TelemetryExporterService) ExportLogs(ctx context.Context, logs plog.Logs) error {
	return s.logExporter.ConsumeLogs(ctx, logs)
}

func (s TelemetryExporterService) Start(ctx context.Context, host component.Host) error {
	return nil
}

func ProvideTelemetryExporterService(cfg *setting.Cfg) (Service, error) {
	localLogger := log.New("telemetryexporter.ProvideTelemetryExporterService")
	section, err := cfg.Raw.GetSection("telemetryexport")
	if err != nil {
		localLogger.Error("failed to get telemetryexport section: %v", err)

		return &TelemetryExporterService{}, err
	}

	telemtryExporterConfig := OtlpHttpExporterConfig{}
	if section.MapTo(&telemtryExporterConfig) != nil {
		localLogger.Error("failed to map telemetryexport section: %v", err)

		return &TelemetryExporterService{}, err
	}

	if telemtryExporterConfig.Validate() != nil {
		localLogger.Error("failed to validate telemetryexport section: %v", err)

		return &TelemetryExporterService{}, err
	}

	headerMap := make(map[string]configopaque.String, len(telemtryExporterConfig.Headers))
	for _, header := range telemtryExporterConfig.Headers {
		splitted := strings.Split(header, ":")
		headerMap[splitted[0]] = configopaque.String(splitted[1])
	}

	exporterFactory := otlphttpexporter.NewFactory()
	defaultConfig := exporterFactory.CreateDefaultConfig().(*otlphttpexporter.Config)
	defaultConfig.Endpoint = telemtryExporterConfig.Endpoint
	defaultConfig.TLSSetting.Insecure = telemtryExporterConfig.Insecure
	defaultConfig.Headers = headerMap

	zapper, err := zap.NewDevelopment()
	if err != nil {
		localLogger.Error("failed to create zap logger: %v", err)

		return &TelemetryExporterService{}, err
	}

	settings := exporter.CreateSettings{
		TelemetrySettings: component.TelemetrySettings{
			Logger:         zapper,
			TracerProvider: trace.NewNoopTracerProvider(),
			MeterProvider:  metric.NewNoopMeterProvider(),
			MetricsLevel:   configtelemetry.LevelNone,
		},
		BuildInfo: component.NewDefaultBuildInfo(),
	}

	exporter, err := exporterFactory.CreateTracesExporter(context.Background(), settings, defaultConfig)
	if err != nil {
		localLogger.Error("failed to create exporter: %v", err)

		return &TelemetryExporterService{}, err
	}
	exporter.Start(context.Background(), nil)

	logExporter, err := exporterFactory.CreateLogsExporter(context.Background(), settings, defaultConfig)
	if err != nil {
		localLogger.Error("failed to create logs exporter: %v", err)

		return &TelemetryExporterService{}, err
	}
	logExporter.Start(context.Background(), nil)

	metricsExporter, err := exporterFactory.CreateMetricsExporter(context.Background(), settings, defaultConfig)
	if err != nil {
		localLogger.Error("failed to create metrics exporter: %v", err)

		return &TelemetryExporterService{}, err
	}
	metricsExporter.Start(context.Background(), nil)

	return &TelemetryExporterService{
		traceExporter:  exporter,
		metricExporter: metricsExporter,
		logExporter:    logExporter,

		logger: log.New("server"),
	}, nil
}

type NoopTelemetryExporterService struct {
}

func (s NoopTelemetryExporterService) ExportTraces(ctx context.Context, trace ptrace.Traces) error {

	return nil
}

func (s NoopTelemetryExporterService) ExportMetrics(ctx context.Context, metrics pmetric.Metrics) error {
	return nil
}

func (s NoopTelemetryExporterService) ExportLogs(ctx context.Context, logs plog.Logs) error {
	return nil
}

func (s NoopTelemetryExporterService) Start(ctx context.Context, host component.Host) error {
	return nil
}

func ProvideTelemetryExporterServiceForTest(cfg *setting.Cfg) (Service, error) {

	return NoopTelemetryExporterService{}, nil
}
