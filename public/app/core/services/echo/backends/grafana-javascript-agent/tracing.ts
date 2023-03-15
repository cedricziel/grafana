import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
import { BatchSpanProcessor } from '@opentelemetry/sdk-trace-web';

import { TracingInstrumentation } from '@grafana/faro-web-tracing';

interface Options {
  otlpEndpoint?: string;
}

export function initTracingInstrumentation(options: Options) {
  return new TracingInstrumentation();
}
