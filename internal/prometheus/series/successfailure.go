package series

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

type successFailureMetricPointsGenerator interface {
	generate(
		start time.Time,
		end time.Time,
		interval time.Duration,
		writerSuccess func(v int, t int64) error,
		writerFailure func(v int, t int64) error,
	) error
}

type constantSuccessFailureMetricPointsGenerator struct {
	throughputSuccess int
	throughputFailure int
	overrides         []constantSuccessFailureMetricPointsOverride
}

type constantSuccessFailureMetricPointsOverride struct {
	start             time.Time
	end               time.Time
	throughputSuccess int
	throughputFailure int
}

var _ successFailureMetricPointsGenerator = &constantSuccessFailureMetricPointsGenerator{}

func (g *constantSuccessFailureMetricPointsGenerator) generate(
	start time.Time,
	end time.Time,
	interval time.Duration,
	writerSuccess func(v int, t int64) error,
	writerFailure func(v int, t int64) error,
) error {
	var (
		totalSuccess int
		totalFailure int
	)
	t := start

	for t.Before(end) {
		success := g.throughputSuccess
		failure := g.throughputFailure
		for i := len(g.overrides) - 1; i >= 0; i-- {
			o := g.overrides[i]
			if (t.Equal(o.start) || t.After(o.start)) && t.Before(o.end) {
				success = o.throughputSuccess
				failure = o.throughputFailure
				break
			}
		}
		totalSuccess += success
		totalFailure += failure

		tUnix := t.Unix()
		if err := writerSuccess(totalSuccess, tUnix); err != nil {
			return err
		}
		if err := writerFailure(totalFailure, tUnix); err != nil {
			return err
		}

		t = t.Add(interval)
	}

	return nil
}

type binomialSuccessFailureMetricPointsGenerator struct {
	throughput    int
	baseErrorRate float64
	baseDist      distuv.Binomial
}

func newBinomialSuccessFailureMetricPointsGenerator(
	throughput int,
	baseErrorRate float64,
	source rand.Source,
) *binomialSuccessFailureMetricPointsGenerator {
	baseDist := distuv.Binomial{
		N:   float64(throughput),
		P:   baseErrorRate,
		Src: rand.New(source),
	}
	return &binomialSuccessFailureMetricPointsGenerator{
		throughput:    throughput,
		baseErrorRate: baseErrorRate,
		baseDist:      baseDist,
	}
}

var _ successFailureMetricPointsGenerator = &binomialSuccessFailureMetricPointsGenerator{}

func (g *binomialSuccessFailureMetricPointsGenerator) generate(
	start time.Time,
	end time.Time,
	interval time.Duration,
	writerSuccess func(v int, t int64) error,
	writerFailure func(v int, t int64) error,
) error {
	var (
		totalSuccess int
		totalFailure int
	)
	t := start

	for t.Before(end) {
		failure := int(g.baseDist.Rand())
		success := g.throughput - failure
		totalSuccess += success
		totalFailure += failure

		tUnix := t.Unix()
		if err := writerSuccess(totalSuccess, tUnix); err != nil {
			return err
		}
		if err := writerFailure(totalFailure, tUnix); err != nil {
			return err
		}

		t = t.Add(interval)
	}

	return nil
}

type successFailureSeriesGenerator struct {
	metricPointsGenerator successFailureMetricPointsGenerator
	name                  string
	labelNameStatus       string
	labelValueSuccess     string
	labelValueFailure     string
	labels                map[string]string
}

func (sg *successFailureSeriesGenerator) generateLabels(commaSpace bool) (string, string) {
	var ls []string
	for name, value := range sg.labels {
		ls = append(ls, fmt.Sprintf("%s=\"%s\"", name, value))
	}

	lsSuccess := append(slices.Clone(ls), fmt.Sprintf("%s=\"%s\"", sg.labelNameStatus, sg.labelValueSuccess))
	lsFailure := append(ls, fmt.Sprintf("%s=\"%s\"", sg.labelNameStatus, sg.labelValueFailure))

	var sep = ","
	if commaSpace {
		sep = ", "
	}

	return "{" + strings.Join(lsSuccess, sep) + "}", "{" + strings.Join(lsFailure, sep) + "}"
}

var _ seriesGenerator = &successFailureSeriesGenerator{}

func (g *successFailureSeriesGenerator) generateOpenMetricsSeries(
	start time.Time,
	end time.Time,
	interval time.Duration,
	w io.Writer,
) error {
	labelsSuccess, labelsFailure := g.generateLabels(false)

	var buf bytes.Buffer
	g.metricPointsGenerator.generate(
		start,
		end,
		interval,
		func(v int, t int64) error {
			_, err := fmt.Fprintf(w, "%s%s %d %d\n", g.name, labelsSuccess, v, t)
			return err
		},
		func(v int, t int64) error {
			_, err := fmt.Fprintf(&buf, "%s%s %d %d\n", g.name, labelsFailure, v, t)
			return err
		},
	)
	if _, err := io.Copy(w, &buf); err != nil {
		return err
	}
	return nil
}

func (g *successFailureSeriesGenerator) generateUnitTestSeries(
	start time.Time,
	end time.Time,
	interval time.Duration,
) []series {
	var valuesBufSuccess bytes.Buffer
	var valuesBufFailure bytes.Buffer

	cswSuccess := newCompressingSeriesWriter(&valuesBufSuccess)
	cswFailure := newCompressingSeriesWriter(&valuesBufFailure)

	g.metricPointsGenerator.generate(start, end, interval, cswSuccess.writerFunc(), cswFailure.writerFunc())
	cswSuccess.Close()
	cswFailure.Close()

	labelsSuccess, labelsFailure := g.generateLabels(true)

	seriesSuccess := series{
		Series: g.name + labelsSuccess,
		Values: valuesBufSuccess.String(),
	}
	seriesFailure := series{
		Series: g.name + labelsFailure,
		Values: valuesBufFailure.String(),
	}
	return []series{seriesSuccess, seriesFailure}
}
