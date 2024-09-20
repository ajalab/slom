package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestGeneratePrometheusRuleOutput(t *testing.T) {
	type testCase struct {
		specConfigFileName              string
		jsonRuleFileName                string
		jsonRecordingRuleFileName       string
		prometheusRuleFileName          string
		prometheusRecordingRuleFileName string
	}

	testCases := []testCase{
		{
			"testdata/spec/availability99.yaml",
			"testdata/out/prometheus-rule-json/availability99.json",
			"testdata/out/prometheus-rule-json/availability99-recording.json",
			"testdata/out/prometheus-rule-prometheus/availability99.yaml",
			"testdata/out/prometheus-rule-prometheus/availability99-recording.yaml",
		},
		{
			"testdata/spec/availability99_availability99.yaml",
			"testdata/out/prometheus-rule-json/availability99_availability99.json",
			"testdata/out/prometheus-rule-json/availability99_availability99-recording.json",
			"testdata/out/prometheus-rule-prometheus/availability99_availability99.yaml",
			"testdata/out/prometheus-rule-prometheus/availability99_availability99-recording.yaml",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.specConfigFileName, func(t *testing.T) {
			t.Run("json-all", func(t *testing.T) {
				args := []string{"generate", "prometheus-rule", "-o", "json", "-t", "all", tc.specConfigFileName}
				checkSlogenResult(t, args, tc.jsonRuleFileName)
			})
			t.Run("json-recording", func(t *testing.T) {
				args := []string{"generate", "prometheus-rule", "-o", "json", "-t", "record", tc.specConfigFileName}
				checkSlogenResult(t, args, tc.jsonRecordingRuleFileName)
			})
			t.Run("prometheus-all", func(t *testing.T) {
				args := []string{"generate", "prometheus-rule", "-o", "prometheus", "-t", "all", tc.specConfigFileName}
				checkSlogenResult(t, args, tc.prometheusRuleFileName)
			})
			t.Run("prometheus-recording", func(t *testing.T) {
				args := []string{"generate", "prometheus-rule", "-o", "prometheus", "-t", "record", tc.specConfigFileName}
				checkSlogenResult(t, args, tc.prometheusRecordingRuleFileName)
			})
		})
	}
}

func TestGeneratePrometheusSeriesOutput(t *testing.T) {
	type testCase struct {
		seriesConfigFileName      string
		seriesOpenMetricsFileName string
		seriesUnitTestFileName    string
	}

	testCases := []testCase{
		{
			"testdata/series/constant_availability999_small.yaml",
			"testdata/out/prometheus-series-openmetrics/constant_availability999_small.openmetrics",
			"testdata/out/prometheus-series-unittest/constant_availability999_small.yaml",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.seriesConfigFileName, func(t *testing.T) {
			t.Run("openmetrics", func(t *testing.T) {
				args := []string{"generate", "prometheus-series", "-o", "openmetrics", tc.seriesConfigFileName}
				checkSlogenResult(t, args, tc.seriesOpenMetricsFileName)
			})
			t.Run("unittest", func(t *testing.T) {
				args := []string{"generate", "prometheus-series", "--output", "unittest", tc.seriesConfigFileName}
				checkSlogenResult(t, args, tc.seriesUnitTestFileName)
			})
		})
	}
}

func TestGenerateDocumentOutput(t *testing.T) {
	type testCase struct {
		specConfigFileName             string
		jsonDocumentFileName           string
		yamlDocumentFileName           string
		goTemplateFileName             string
		goTemplateFileDocumentFileName string
	}

	testCases := []testCase{
		{
			"testdata/spec/availability99.yaml",
			"testdata/out/document-json/availability99.json",
			"testdata/out/document-yaml/availability99.yaml",
			"testdata/document-go-template-file/md.tmpl",
			"testdata/out/document-go-template-file/availability99.md",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.specConfigFileName, func(t *testing.T) {
			t.Run("json", func(t *testing.T) {
				args := []string{"generate", "document", "-o", "json", tc.specConfigFileName}
				checkSlogenResult(t, args, tc.jsonDocumentFileName)
			})
			t.Run("yaml", func(t *testing.T) {
				args := []string{"generate", "document", "-o", "yaml", tc.specConfigFileName}
				checkSlogenResult(t, args, tc.yamlDocumentFileName)
			})
			t.Run("go-template-file", func(t *testing.T) {
				args := []string{"generate", "document", "-o", "go-template-file=" + tc.goTemplateFileName, tc.specConfigFileName}
				checkSlogenResult(t, args, tc.goTemplateFileDocumentFileName)
			})
		})

	}
}

func checkSlogenResult(t *testing.T, args []string, expectedOutputFileName string) {
	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	if err := run(args, &stdout, &stderr); err != nil {
		t.Fatalf("failed to run: %v", err)
	}

	expectedOutputFile, err := os.ReadFile(expectedOutputFileName)
	if err != nil {
		t.Fatalf("failed to load a file %s: %v", expectedOutputFileName, err)
	}
	if !bytes.Equal(expectedOutputFile, stdout.Bytes()) {
		t.Errorf("output does not match the expected content. %s", cmp.Diff(expectedOutputFile, stdout.Bytes()))
	}
}

func TestGeneratePrometheusRulePromtool(t *testing.T) {
	type testCase struct {
		unitTestFileName   string
		specConfigFileName string
	}

	testCases := []testCase{
		{
			"testdata/prometheus-unittest/availability99-constant_availability999.yaml",
			"testdata/spec/availability99.yaml",
		},
		{
			"testdata/prometheus-unittest/availability99-constant_availability999_86.yaml",
			"testdata/spec/availability99.yaml",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(path.Base(tc.unitTestFileName), func(t *testing.T) {
			args := []string{"generate", "prometheus-rule", "-o", "prometheus", tc.specConfigFileName}
			checkPromtool(t, args, tc.unitTestFileName, tc.specConfigFileName)
		})
	}
}

func checkPromtool(t *testing.T, args []string, unitTestFileName string, specConfigFileName string) {
	ruleFile := bytes.Buffer{}
	stderr := bytes.Buffer{}
	if err := run(args, &ruleFile, &stderr); err != nil {
		t.Fatalf("failed to run slogen: %v", err)
	}

	unitTestFile, err := os.Open(unitTestFileName)
	if err != nil {
		t.Fatalf("failed to open a unit test file: %s: %v", unitTestFileName, err)
	}
	defer unitTestFile.Close()

	ctx := context.Background()
	containerUnitTestFilePath := "/slogen/" + path.Base(unitTestFileName)
	cmd := []string{"promtool", "test", "rules", containerUnitTestFilePath}

	var ec int
	var out bytes.Buffer
	req := testcontainers.ContainerRequest{
		Image: "prom/prometheus",
		Files: []testcontainers.ContainerFile{
			{
				Reader:            bytes.NewBuffer(ruleFile.Bytes()),
				ContainerFilePath: "/slogen/" + path.Base(specConfigFileName),
				FileMode:          0o666,
			},
			{
				Reader:            unitTestFile,
				ContainerFilePath: containerUnitTestFilePath,
				FileMode:          0o666,
			},
		},
		WaitingFor: wait.ForExec(cmd).WithExitCodeMatcher(func(exitCode int) bool {
			ec = exitCode
			return true
		}).WithResponseMatcher(func(body io.Reader) bool {
			out.ReadFrom(body)
			return true
		}),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start prometheus container: %v", err)
	}

	if ec != 0 {
		t.Fatalf("promtool unit test failed: ec=%d\n%s", ec, out.String())
	}

	t.Cleanup(func() { container.Terminate(ctx) })
}
