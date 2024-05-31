package tsdb

import (
	"context"
	"time"
)

type TSDBBackfiller interface {
	BackfillOpenMetricsSeries(
		ctx context.Context,
		openMetricsSeriesFileName string,
		tsdbDirName string,
	) error

	BackfillRule(
		ctx context.Context,
		prometheusRuleFileName string,
		start time.Time,
		end time.Time,
		tsdbDirName string,
	) error
}
