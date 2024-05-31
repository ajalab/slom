package prometheus

import (
	"fmt"
	"strings"
)

func GenerateLabels(labels map[string]string, commaSpace bool) string {
	var sep string
	if commaSpace {
		sep = ", "
	} else {
		sep = ","
	}

	var ls []string
	for name, value := range labels {
		ls = append(ls, fmt.Sprintf("%s=\"%s\"", name, value))
	}
	return "{" + strings.Join(ls, sep) + "}"
}
