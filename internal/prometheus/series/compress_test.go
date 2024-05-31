package series

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

func TestCompressingSeriesWriter(t *testing.T) {
	type tc struct {
		series   []int
		expected string
	}
	tcs := []tc{
		{[]int{1}, "1 "},
		{[]int{1, 1}, "1 1 "},
		{[]int{1, 2}, "1 2 "},
		{[]int{1, 1, 1}, "1x2 "},
		{[]int{1, 2, 2}, "1 2 2 "},
		{[]int{1, 2, 3}, "1+1x2 "},
		{[]int{1, 2, 4}, "1 2 4 "},
		{[]int{1, 2, 4, 6}, "1 2+2x2 "},
		{[]int{1, 2, 4, 8}, "1 2 4 8 "},
		{[]int{1, 2, 3, 6, 7}, "1+1x2 6 7 "},
		{[]int{1, 2, 4, 6, 7}, "1 2+2x2 7 "},
		{[]int{1, 2, 4, 6, 8}, "1 2+2x3 "},
		{[]int{1, 2, 3, 6, 7, 8}, "1+1x2 6+1x2 "},
		{[]int{1, 2, 4, 6, 7, 8}, "1 2+2x2 7 8 "},
		{[]int{1, 2, 4, 6, 7, 8, 9}, "1 2+2x2 7+1x2 "},
		{[]int{1, 2, 4, 6, 7, 6, 5}, "1 2+2x2 7-1x2 "},
	}
	for _, tc := range tcs {
		t.Run(string(fmt.Sprint(tc.series)), func(t *testing.T) {
			var actual bytes.Buffer
			now := time.Now()
			csw := newCompressingSeriesWriter(&actual)
			f := csw.writerFunc()
			for _, v := range tc.series {
				f(v, now)
			}
			csw.Close()

			a := actual.String()
			if a != tc.expected {
				t.Errorf("the actual output doesn't match the expected output. expected=\"%s\", actual=\"%s\", csw=%#v", tc.expected, a, csw)
			}
		})
	}
}
