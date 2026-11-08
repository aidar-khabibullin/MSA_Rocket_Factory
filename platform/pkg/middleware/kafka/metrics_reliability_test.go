package kafka

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// sumByName возвращает сумму всех значений Int64-счётчика с указанным именем.
func sumByName(t *testing.T, reader *sdkmetric.ManualReader, name string) int64 {
	t.Helper()
	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &rm))

	var total int64
	for _, scope := range rm.ScopeMetrics {
		for _, m := range scope.Metrics {
			if m.Name != name {
				continue
			}
			sum, ok := m.Data.(metricdata.Sum[int64])
			require.True(t, ok, "%s должен быть Int64 Sum", name)
			for _, point := range sum.DataPoints {
				total += point.Value
			}
		}
	}
	return total
}

func TestKafkaMetrics_ReliabilityCounters(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	otel.SetMeterProvider(sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)))

	metrics, err := NewKafkaMetrics()
	require.NoError(t, err)
	ctx := context.Background()

	metrics.RecordConsumerRetry(ctx, "order.paid", "assembly-service")
	metrics.RecordConsumerRetry(ctx, "order.paid", "assembly-service")
	metrics.RecordConsumerDeadLetter(ctx, "order.paid", "assembly-service")
	metrics.RecordDedupDuplicate(ctx, "order.paid", "assembly-service")
	metrics.RecordDedupStoreError(ctx, "assembly")
	metrics.RecordDedupStoreError(ctx, "assembly")
	metrics.RecordDedupStoreError(ctx, "assembly")

	require.EqualValues(t, 2, sumByName(t, reader, "kafka_consumer_retries_total"))
	require.EqualValues(t, 1, sumByName(t, reader, "kafka_consumer_dead_letters_total"))
	require.EqualValues(t, 1, sumByName(t, reader, "kafka_dedup_duplicates_total"))
	require.EqualValues(t, 3, sumByName(t, reader, "kafka_dedup_store_errors_total"))
}
