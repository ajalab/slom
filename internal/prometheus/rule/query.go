package rule

import (
	"regexp"

	"github.com/ajalab/slogen/internal/spec"
)

var reWindow = regexp.MustCompile(`\$window\b`)

func generateErrorRateQuery(
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
