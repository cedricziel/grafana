package tracing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/setting"
)

func TestSplitCustomAttribs(t *testing.T) {
	tests := []struct {
		input    string
		expected []attribute.KeyValue
	}{
		{
			input:    "key1:value:1",
			expected: []attribute.KeyValue{attribute.String("key1", "value:1")},
		},
		{
			input: "key1:value1,key2:value2",
			expected: []attribute.KeyValue{
				attribute.String("key1", "value1"),
				attribute.String("key2", "value2"),
			},
		},
		{
			input:    "",
			expected: []attribute.KeyValue{},
		},
	}

	for _, test := range tests {
		attribs, err := splitCustomAttribs(test.input)
		assert.NoError(t, err)
		assert.EqualValues(t, test.expected, attribs)
	}
}

func TestSplitCustomAttribs_Malformed(t *testing.T) {
	tests := []struct {
		input    string
		expected []attribute.KeyValue
	}{
		{input: "key1=value1"},
		{input: "key1"},
	}

	for _, test := range tests {
		_, err := splitCustomAttribs(test.input)
		assert.Error(t, err)
	}
}

func TestOptentelemetry_ParseSettingsOpentelemetry(t *testing.T) {
	cfg := setting.NewCfg()
	otel := NewOpenTelemetry(cfg, log.NewNopLogger())

	otelsect := cfg.Raw.Section("tracing.opentelemetry")
	jaegersect := cfg.Raw.Section("tracing.opentelemetry.jaeger")
	otlpsect := cfg.Raw.Section("tracing.opentelemetry.otlp")
	cfg.Raw.Section("tracing.opentelemetry.otlphttp")

	assert.NoError(t, otel.parseSettingsOpentelemetry())
	assert.Equal(t, noopExporter, otel.enabled)

	otelsect.Key("custom_attributes")
	assert.NoError(t, otel.parseSettingsOpentelemetry())
	assert.Empty(t, otel.customAttribs)

	otelsect.Key("custom_attributes").SetValue("key1:value1,key2:value2")
	assert.NoError(t, otel.parseSettingsOpentelemetry())
	expected := []attribute.KeyValue{
		attribute.String("key1", "value1"),
		attribute.String("key2", "value2"),
	}
	assert.Equal(t, expected, otel.customAttribs)

	jaegersect.Key("address").SetValue("somehost:6831")
	assert.NoError(t, otel.parseSettingsOpentelemetry())
	assert.Equal(t, "somehost:6831", otel.address)
	assert.Equal(t, jaegerExporter, otel.enabled)

	jaegersect.Key("address").SetValue("")
	otlpsect.Key("address").SetValue("somehost:4317")
	assert.NoError(t, otel.parseSettingsOpentelemetry())
	assert.Equal(t, "somehost:4317", otel.address)
	assert.Equal(t, otlpExporter, otel.enabled)
}

func TestOpenTelemetry_ParseSettingsOpentelemetryOTLPExporter(t *testing.T) {
	cfg := setting.NewCfg()
	otel := NewOpenTelemetry(cfg, log.NewNopLogger())

	otelsect := cfg.Raw.Section("tracing.opentelemetry")
	jaegersect := cfg.Raw.Section("tracing.opentelemetry.jaeger")
	otlpsect := cfg.Raw.Section("tracing.opentelemetry.otlp")
	otlphttpsect := cfg.Raw.Section("tracing.opentelemetry.otlphttp")

	assert.NoError(t, otel.parseSettingsOpentelemetry())
	assert.Equal(t, noopExporter, otel.enabled)

	otelsect.Key("custom_attributes")
	assert.NoError(t, otel.parseSettingsOpentelemetry())
	assert.Empty(t, otel.customAttribs)

	otelsect.Key("custom_attributes").SetValue("key1:value1,key2:value2")
	assert.NoError(t, otel.parseSettingsOpentelemetry())
	expected := []attribute.KeyValue{
		attribute.String("key1", "value1"),
		attribute.String("key2", "value2"),
	}
	assert.Equal(t, expected, otel.customAttribs)

	jaegersect.Key("address").SetValue("somehost:6831")
	assert.NoError(t, otel.parseSettingsOpentelemetry())
	assert.Equal(t, "somehost:6831", otel.address)
	assert.Equal(t, jaegerExporter, otel.enabled)

	jaegersect.Key("address").SetValue("")
	otlpsect.Key("address").SetValue("somehost:4317")
	assert.NoError(t, otel.parseSettingsOpentelemetry())
	assert.Equal(t, "somehost:4317", otel.address)
	assert.Equal(t, otlpExporter, otel.enabled)

	// set OTLPHTTP config
	otlphttpsect.Key("endpoint").SetValue("http://collector:4317")
	otlphttpsect.Key("headers").SetValue("x-foo: bar, x-baz: qux")
	assert.NoError(t, otel.parseSettingsOpentelemetry())
	assert.Equal(t, "http://collector:4317", otel.exporters[otlpHttpExporter].(OTLPHTTPExporterConfig).Endpoint)
	assert.Contains(t, otel.exporters[otlpHttpExporter].(OTLPHTTPExporterConfig).Headers, "x-foo: bar")
}
