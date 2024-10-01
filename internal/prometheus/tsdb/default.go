package tsdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

type DefaultTSDBBackfiller struct {
	logger             *slog.Logger
	prometheusExecName string
	promtoolExecName   string
}

var _ TSDBBackfiller = &DefaultTSDBBackfiller{}

func NewDefaultTSDBBackfiller(
	logger *slog.Logger,
	prometheusExecName string,
	promtoolExecName string,
) *DefaultTSDBBackfiller {
	return &DefaultTSDBBackfiller{
		logger:             logger,
		prometheusExecName: prometheusExecName,
		promtoolExecName:   promtoolExecName,
	}
}

func (b *DefaultTSDBBackfiller) BackfillOpenMetricsSeries(
	ctx context.Context,
	openMetricsSeriesFileName string,
	tsdbDirName string,
) error {
	logger := b.logger.WithGroup("backfillOpenMetricsSeries")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := b.runPromtool(&stdout, &stderr, "tsdb", "create-blocks-from", "openmetrics", openMetricsSeriesFileName, tsdbDirName)
	logger.Debug("got promtool output", "stdout", stdout.String(), "stderr", stderr.String())

	if err != nil {
		return fmt.Errorf("failed to run promtool tsdb create-blocks-from openmetrics: %w", err)
	}

	return nil
}

func (b *DefaultTSDBBackfiller) BackfillRule(
	ctx context.Context,
	prometheusRuleFileName string,
	start time.Time,
	end time.Time,
	tsdbDirName string,
) error {
	logger := b.logger.WithGroup("backfillRule")

	tempTSDBDirName, err := os.MkdirTemp("", "slom-tsdb-*")
	if err != nil {
		return fmt.Errorf("failed to create a temporary directory for TSDB: %w", err)
	}
	defer os.RemoveAll(tempTSDBDirName)
	logger.Debug("created a temporary directory for TSDB", "tempTSDBDirName", tempTSDBDirName)

	if err := b.backfillRule(ctx, prometheusRuleFileName, start, end, tsdbDirName, tempTSDBDirName); err != nil {
		return err
	}

	tempTSDBDir, err := os.Open(tempTSDBDirName)
	if err != nil {
		return fmt.Errorf("failed to open the temporary TSDB directory %s: %w", tempTSDBDirName, err)
	}
	tempTSDBBlockNames, err := tempTSDBDir.Readdirnames(-1)
	if err != nil {
		return fmt.Errorf("failed to list up TSDB blocks in the temporary TSDB directory: %w", err)
	}

	for _, tsdbBlockName := range tempTSDBBlockNames {
		srcTSDBBlockName := filepath.Join(tempTSDBDirName, tsdbBlockName)
		dstTSDBBlockName := filepath.Join(tsdbDirName, tsdbBlockName)
		if err := os.Rename(srcTSDBBlockName, dstTSDBBlockName); err != nil {
			return fmt.Errorf("failed to move a TSDB block from %s to %s: %w", srcTSDBBlockName, dstTSDBBlockName, err)
		}
	}
	return nil
}

func (b *DefaultTSDBBackfiller) backfillRule(
	ctx context.Context,
	prometheusRuleFileName string,
	start time.Time,
	end time.Time,
	srcTSDBDirName string,
	dstTSDBDirName string,
) error {
	logger := b.logger.WithGroup("backfillRule")

	emptyPrometheusConfigFile, err := os.CreateTemp("", "slom-tsdb-prometheus-config-*")
	if err != nil {
		return fmt.Errorf("failed to create a temporary file for Prometheus config: %w", err)
	}
	defer os.Remove(emptyPrometheusConfigFile.Name())

	port, err := findAvailablePort()
	if err != nil {
		return err
	}
	address := net.JoinHostPort("127.0.0.1", fmt.Sprint(port))

	process := newPrometheusProcess(logger, b.prometheusExecName, address, emptyPrometheusConfigFile.Name(), srcTSDBDirName)
	if err := process.Start(ctx); err != nil {
		return fmt.Errorf("failed to start prometheus process: %w", err)
	}
	defer func() {
		if err := process.Close(); err != nil {
			logger.Error("failed to close prometheus process: %w", err)
		}
	}()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err = b.runPromtool(
		&stdout,
		&stderr,
		"tsdb",
		"create-blocks-from",
		"rules",
		"--start", start.Format(time.RFC3339),
		"--end", end.Format(time.RFC3339),
		"--output-dir", dstTSDBDirName,
		"--url", fmt.Sprintf("http://%s", address),
		prometheusRuleFileName,
	)
	logger.Debug("got promtool output", "stdout", stdout.String(), "stderr", stderr.String())
	if err != nil {
		return fmt.Errorf("failed to run promtool tsdb create-blocks-from rules: %w", err)
	}

	return nil
}

func (b *DefaultTSDBBackfiller) runPromtool(
	stdout io.Writer,
	stderr io.Writer,
	arg ...string,
) error {
	cmd := exec.Command(b.promtoolExecName, arg...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe for promtool process: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe for promtool process: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to run promtool process: %w", err)
	}
	if _, err := io.Copy(stdout, stdoutPipe); err != nil {
		return fmt.Errorf("failed to read stdout of promtool process: %w", err)
	}
	if _, err := io.Copy(stderr, stderrPipe); err != nil {
		return fmt.Errorf("failed to read stderr of promtool process: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("promtool command did not complete gracefully: %w", err)
	}
	return nil
}

func findAvailablePort() (int, error) {
	startPort := 30000 + rand.Intn(10000)

	for port := startPort; port < 65535; port++ {
		addr := fmt.Sprintf(":%d", port)
		// Attempt to listen on the port
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			defer listener.Close()
			return port, nil
		}
	}

	return 0, errors.New("could not find an available port")
}

type prometheusProcess struct {
	logger  *slog.Logger
	cmd     *exec.Cmd
	address string
}

func newPrometheusProcess(
	logger *slog.Logger,
	execName string,
	address string,
	configFileName string,
	tsdbDirName string,
) *prometheusProcess {
	return &prometheusProcess{
		logger: logger.WithGroup("prometheusProcess").With("address", address),
		cmd: exec.Command(
			execName,
			"--web.listen-address", address,
			"--config.file", configFileName,
			"--storage.tsdb.path", tsdbDirName,
		),
		address: address,
	}
}

func (p *prometheusProcess) Start(ctx context.Context) error {
	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start prometheus process: %w", err)
	}
	p.logger = p.logger.With("pid", p.cmd.Process.Pid)

	if err := p.checkPrometheusReadyWithRetry(ctx); err != nil {
		return fmt.Errorf("prometheus could not be ready: %w", err)
	}

	p.logger.Info("started prometheus process")
	return nil
}

func (p *prometheusProcess) Close() error {
	if err := p.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM to the prometheus process: %w", err)
	}
	if err := p.cmd.Wait(); err != nil {
		return fmt.Errorf("failed to wait prometheus process: %w", err)
	}
	p.logger.Info("stopped prometheus process")

	return nil
}

func (p *prometheusProcess) checkPrometheusReadyWithRetry(
	ctx context.Context,
) error {
	numChecks := 5
	checkInterval := time.Duration(1) * time.Second
	url := &url.URL{Scheme: "http", Host: p.address, Path: "-/ready"}
	urlString := url.String()

	for i := 0; i < numChecks; i++ {
		select {
		case <-ctx.Done():
			return errors.New("context cancelled")
		default:
		}
		err := p.checkPrometheusReady(ctx, urlString)
		if err != nil {
			p.logger.Debug("prometheus is not ready so rerunning the check", "err", err, "attempt", i+1)
			time.Sleep(checkInterval)
		} else {
			return nil
		}
	}

	return errors.New("exceeded the max check attempt")
}

func (p *prometheusProcess) checkPrometheusReady(
	ctx context.Context,
	url string,
) error {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("prometheus is not ready because of an error: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("prometheus is not ready according to the status code %d", resp.StatusCode)
	}

	return nil
}
