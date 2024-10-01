package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestGenerateDocumentOutput(t *testing.T) {
	dir := "testdata/generate-document-output"

	specFilesPattern := filepath.Join(dir, "spec/*.yaml")
	specFiles, err := filepath.Glob(specFilesPattern)
	if err != nil {
		t.Fatalf("failed to look up spec files %s: %s", specFilesPattern, err)
	}

	for _, specFile := range specFiles {
		specId := filepath.Base(specFile[:len(specFile)-len(filepath.Ext(specFile))])

		t.Run(specId, func(t *testing.T) {
			outFileDocumentJson := filepath.Join(dir, "out/document-json", specId+".json")
			runTestWithOutFile(t, outFileDocumentJson, "document-json", func(t *testing.T) {
				args := []string{"generate", "document", "-o", "json", specFile}
				checkSlomOutput(t, args, outFileDocumentJson)
			})
			outFileDocumentYaml := filepath.Join(dir, "out/document-yaml", specId+".yaml")
			runTestWithOutFile(t, outFileDocumentYaml, "document-yaml", func(t *testing.T) {
				args := []string{"generate", "document", "-o", "yaml", specFile}
				checkSlomOutput(t, args, outFileDocumentYaml)
			})
			outFileDocumentGoTemplateFile := filepath.Join(dir, "out/document-go-template-file", specId+".md")
			runTestWithOutFile(t, outFileDocumentGoTemplateFile, "document-go-template-file", func(t *testing.T) {
				goTemplateFile := filepath.Join(dir, "go-template-file", specId+".tmpl")
				args := []string{"generate", "document", "-o", "go-template-file=" + goTemplateFile, specFile}
				checkSlomOutput(t, args, outFileDocumentGoTemplateFile)
			})
		})
	}
}

func TestGeneratePrometheusRuleOutput(t *testing.T) {
	dir := "testdata/generate-prometheus-rule-output"

	specFilesPattern := filepath.Join(dir, "spec/*.yaml")
	specFiles, err := filepath.Glob(specFilesPattern)
	if err != nil {
		t.Fatalf("failed to look up spec files %s: %s", specFilesPattern, err)
	}

	for _, specFile := range specFiles {
		specId := filepath.Base(specFile[:len(specFile)-len(filepath.Ext(specFile))])

		t.Run(specId, func(t *testing.T) {
			outFilePrometheusRulePrometheus := filepath.Join(dir, "out/prometheus-rule-prometheus", specId+".yaml")
			runTestWithOutFile(t, outFilePrometheusRulePrometheus, "prometheus-rule-prometheus", func(t *testing.T) {
				args := []string{"generate", "prometheus-rule", "-o", "prometheus", specFile}
				checkSlomOutput(t, args, outFilePrometheusRulePrometheus)
			})

			outFilePrometheusRuleJson := filepath.Join(dir, "out/prometheus-rule-json", specId+".json")
			runTestWithOutFile(t, outFilePrometheusRuleJson, "prometheus-rule-json", func(t *testing.T) {
				args := []string{"generate", "prometheus-rule", "-o", "json", specFile}
				checkSlomOutput(t, args, outFilePrometheusRuleJson)
			})
		})
	}
}

func TestGeneratePrometheusSeriesOutput(t *testing.T) {
	dir := "testdata/generate-prometheus-series-output"

	seriesFilesPattern := filepath.Join(dir, "series/*.yaml")
	seriesFiles, err := filepath.Glob(seriesFilesPattern)
	if err != nil {
		t.Fatalf("failed to look up series files %s: %s", seriesFilesPattern, err)
	}

	for _, seriesFile := range seriesFiles {
		seriesId := filepath.Base(seriesFile[:len(seriesFile)-len(filepath.Ext(seriesFile))])

		t.Run(seriesId, func(t *testing.T) {
			outFilePrometheusSeriesOpenMetrics := filepath.Join(dir, "out/prometheus-series-openmetrics", seriesId+".openmetrics")
			runTestWithOutFile(t, outFilePrometheusSeriesOpenMetrics, "prometheus-series-openmetrics", func(t *testing.T) {
				args := []string{"generate", "prometheus-series", "-o", "openmetrics", seriesFile}
				checkSlomOutput(t, args, outFilePrometheusSeriesOpenMetrics)
			})

			outFilePrometheusSeriesUnitTest := filepath.Join(dir, "out/prometheus-series-unittest", seriesId+".yaml")
			runTestWithOutFile(t, outFilePrometheusSeriesUnitTest, "prometheus-series-unittest", func(t *testing.T) {
				args := []string{"generate", "prometheus-series", "-o", "unittest", seriesFile}
				checkSlomOutput(t, args, outFilePrometheusSeriesUnitTest)
			})
		})
	}
}

func runTestWithOutFile(t *testing.T, outFile string, name string, f func(t *testing.T)) bool {
	_, err := os.Stat(outFile)
	if errors.Is(err, os.ErrNotExist) {
		t.Fatalf("failed to look up output file %s: does not exist", outFile)
		return false
	} else if err != nil {
		t.Fatalf("failed to look up output file %s: %s", outFile, err)
		return false
	}

	return t.Run(name, f)
}

func checkSlomOutput(t *testing.T, args []string, expectedOutputFileName string) {
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
	dir := "testdata/generate-prometheus-rule-promtool"

	specFilesPattern := filepath.Join(dir, "spec/*.yaml")
	specFiles, err := filepath.Glob(specFilesPattern)
	if err != nil {
		t.Fatalf("failed to look up spec directory %s: %s", specFilesPattern, err)
	}

	for _, specFile := range specFiles {
		specId := filepath.Base(specFile[:len(specFile)-len(filepath.Ext(specFile))])

		unitTestFilesPattern := filepath.Join(dir, fmt.Sprintf("unittest/%s-*.yaml", specId))
		unitTestFiles, err := filepath.Glob(unitTestFilesPattern)
		if err != nil {
			t.Fatalf("failed to look up unit test directory %s: %s", unitTestFilesPattern, err)
		}

		for _, unitTestFile := range unitTestFiles {
			unitTestId := filepath.Base(unitTestFile[:len(unitTestFile)-len(filepath.Ext(unitTestFile))])
			t.Run(unitTestId, func(t *testing.T) {
				args := []string{"generate", "prometheus-rule", "-o", "prometheus", specFile}
				checkPromtool(t, args, unitTestFile, specFile)
			})
		}
	}
}

func checkPromtool(t *testing.T, args []string, unitTestFileName string, specConfigFileName string) {
	ruleFile := bytes.Buffer{}
	stderr := bytes.Buffer{}
	if err := run(args, &ruleFile, &stderr); err != nil {
		t.Fatalf("failed to run slom: %v", err)
	}

	unitTestFile, err := os.Open(unitTestFileName)
	if err != nil {
		t.Fatalf("failed to open a unit test file: %s: %v", unitTestFileName, err)
	}
	defer unitTestFile.Close()

	ctx := context.Background()
	containerUnitTestFilePath := "/slom/" + path.Base(unitTestFileName)
	cmd := []string{"promtool", "test", "rules", containerUnitTestFilePath}

	var ec int
	var out bytes.Buffer
	req := testcontainers.ContainerRequest{
		Image: "prom/prometheus",
		Files: []testcontainers.ContainerFile{
			{
				Reader:            bytes.NewBuffer(ruleFile.Bytes()),
				ContainerFilePath: "/slom/" + path.Base(specConfigFileName),
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
