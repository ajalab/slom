package rule

import "github.com/ajalab/slogen/internal/spec"

func GenerateErrorRateQuery(
	indicator *spec.PrometheusIndicator,
	window spec.Window,
) string {
	var query string
	switch w := window.(type) {
	case *spec.RollingWindow:
		query = reWindow.ReplaceAllString(indicator.ErrorRatio(), w.Duration().String())
	case *spec.CalendarWindow:
		query = ""
	}

	return query
}
