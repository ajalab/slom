package tsdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type DockerTSDBBackfiller struct {
	logger          *slog.Logger
	prometheusImage string
	cli             *client.Client
}

func NewDockerTSDBBackfiller(
	logger *slog.Logger,
	prometheusImage string,
) (*DockerTSDBBackfiller, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create a Docker client: %w", err)
	}

	return &DockerTSDBBackfiller{
		logger:          logger,
		prometheusImage: prometheusImage,
		cli:             cli,
	}, nil
}

func (b *DockerTSDBBackfiller) Initialize(
	ctx context.Context,
) error {
	logger := b.logger.WithGroup("initialize")

	reader, err := b.cli.ImagePull(ctx, b.prometheusImage, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull an image \"%s\": %w", b.prometheusImage, err)
	}
	logger.Info("started pulling a prometheus image", "image", b.prometheusImage)
	defer reader.Close()

	_, err = io.Copy(io.Discard, reader)
	logger.Info("finished pulling a prometheus image", "image", b.prometheusImage)
	return err
}

func (b *DockerTSDBBackfiller) BackfillOpenMetricsSeries(
	ctx context.Context,
	openMetricsSeriesFileName string,
	tsdbDirName string,
) error {
	logger := b.logger.WithGroup("backfillOpenMetricsSeries")

	tsdbDirAbsPath, err := filepath.Abs(tsdbDirName)
	if err != nil {
		return fmt.Errorf("failed to get the absolute path of tsdbDirName \"%s\": %w", tsdbDirAbsPath, err)
	}

	resp, err := b.cli.ContainerCreate(ctx, &container.Config{
		Image:      b.prometheusImage,
		Entrypoint: strslice.StrSlice{"promtool"},
		Cmd: []string{
			"tsdb",
			"create-blocks-from",
			"openmetrics",
			"/etc/prometheus/metrics",
			"/prometheus",
		},
		Tty: false,
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: tsdbDirAbsPath,
				Target: "/prometheus",
			},
			{
				Type:   mount.TypeBind,
				Source: openMetricsSeriesFileName,
				Target: "/etc/prometheus/metrics",
			},
		},
	}, nil, nil, "")
	if err != nil {
		return fmt.Errorf("failed to create a new container: %w", err)
	}
	logger = logger.With("containerId", resp.ID)
	logger.Info("created a new prometheus container to run promtool")

	if err := b.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return err
	}
	logger.Info("started the prometheus container to run promtool")

	defer func() {
		if err := b.cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{}); err != nil {
			slog.Error("failed to remove the prometheus container", "err", err)
		}
		logger.Info("removed the prometheus container")
	}()

	statusCh, errCh := b.cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case status := <-statusCh:
		logger.Info("finished running the prometheus container to run promtool", "statusCode", status.StatusCode, "err", status.Error)
	}

	out, err := b.cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return err
	}

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	stdcopy.StdCopy(&stdout, &stderr, out)
	logger.Debug("got promtool output", "stdout", stdout.String(), "stderr", stderr.String())

	return nil
}

func (b *DockerTSDBBackfiller) BackfillRule(
	ctx context.Context,
	prometheusRuleFileName string,
	start time.Time,
	end time.Time,
	tsdbDirName string,
) error {
	return nil
}

func (b *DockerTSDBBackfiller) backfillRule(
	ctx context.Context,
	prometheusRuleFileName string,
	start time.Time,
	end time.Time,
	srcTSDBDirName string,
	dstTSDBDirName string,
) error {
	logger := b.logger.WithGroup("backfillRule")

	srcTSDBDirAbsPath, err := filepath.Abs(srcTSDBDirName)
	if err != nil {
		return fmt.Errorf("failed to get the absolute path of tsdbDirName \"%s\": %w", srcTSDBDirAbsPath, err)
	}

	emptyPrometheusConfigFile, err := os.CreateTemp("", "slogen-tsdb-prometheus-config-*")
	if err != nil {
		return fmt.Errorf("failed to create a temporary file for Prometheus config: %w", err)
	}
	defer os.Remove(emptyPrometheusConfigFile.Name())

	var (
		containerTsdbPath   = "/prometheus"
		containerConfigPath = "/etc/prometheus/prometheus.yml"
		containerRulePath   = "/etc/prometheus/rule.yml"
	)
	createResp, err := b.cli.ContainerCreate(ctx, &container.Config{
		Image: b.prometheusImage,
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: srcTSDBDirAbsPath,
				Target: containerTsdbPath,
			},
			{
				Type:   mount.TypeBind,
				Source: emptyPrometheusConfigFile.Name(),
				Target: containerConfigPath,
			},
			{
				Type:   mount.TypeBind,
				Source: prometheusRuleFileName,
				Target: containerRulePath,
			},
		},
		AutoRemove: true,
	}, nil, nil, "")
	if err != nil {
		return fmt.Errorf("failed to create a new prometheus container: %w", err)
	}
	logger = logger.With("containerId", createResp.ID)
	logger.Info("created a new prometheus container")

	if err := b.cli.ContainerStart(ctx, createResp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start the prometheus container %s: %w", createResp.ID, err)
	}
	logger.Info("started the prometheus container")

	defer func() {
		if err := b.cli.ContainerStop(ctx, createResp.ID, container.StopOptions{}); err != nil {
			slog.Error("failed failed to stop the prometheus container", "err", err)
			return
		}
		logger.Info("stopped the prometheus container")

		if err := b.cli.ContainerRemove(ctx, createResp.ID, container.RemoveOptions{}); err != nil {
			slog.Error("failed to remove the prometheus container", "err", err)
			return
		}
		logger.Info("removed the prometheus container")
	}()

	if err := b.runExecUntilSuccess(ctx, createResp.ID, []string{"wget", "-qO-", "localhost:9090/-/ready"}); err != nil {
		return fmt.Errorf("prometheus container %s could not be ready until the deadline: %w", createResp.ID, err)
	}
	logger.Info("ensured prometheus process is ready")

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	cmd := []string{
		"promtool",
		"tsdb",
		"create-blocks-from",
		"rules",
		"--start", start.Format(time.RFC3339),
		"--end", end.Format(time.RFC3339),
		"--output-dir", containerTsdbPath,
		containerRulePath,
	}
	exitCode, err := b.runExec(ctx, createResp.ID, cmd, &stdout, &stderr)
	logger.Debug("finished running promtool", "stdout", stdout.String(), "stderr", stderr.String(), "exitCode", exitCode)
	if err != nil {
		return fmt.Errorf("failed to run exec promtool: %w", err)
	}
	if exitCode != 0 {
		return fmt.Errorf("failed to run exec promtool: non-zero return code %d", exitCode)
	}

	return nil
}

func (b *DockerTSDBBackfiller) runExecUntilSuccess(
	ctx context.Context,
	containerId string,
	cmd []string,
) error {
	logger := b.logger.WithGroup("runExecUntilSuccess")

	execMaxAttempt := 5
	execInterval := time.Duration(1) * time.Second
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	for i := 0; i < execMaxAttempt; i++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled")
		default:
			exitCode, err := b.runExec(ctx, containerId, cmd, &stdout, &stderr)
			if err != nil {
				return fmt.Errorf("failed to run exec: %w", err)
			}
			if exitCode == 0 {
				return nil
			}
			logger.Debug("exec is not successful so retrying exec", "exitCode", exitCode, "stdout", stdout.String(), "stderr", stderr.String(), "attempt", i+1)
			time.Sleep(execInterval)
		}
	}
	return fmt.Errorf("exceeded the max exec attempt %d", execMaxAttempt)
}

func (b *DockerTSDBBackfiller) runExec(
	ctx context.Context,
	containerId string,
	cmd []string,
	stdout io.Writer,
	stderr io.Writer,
) (int, error) {
	logger := b.logger.WithGroup("runExec")

	execInspectMaxAttempt := 5
	execInspectInterval := time.Duration(1) * time.Second

	idResp, err := b.cli.ContainerExecCreate(ctx, containerId, types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create container exec: %w", err)
	}

	hijackedResp, err := b.cli.ContainerExecAttach(ctx, idResp.ID, types.ExecStartCheck{})
	if err != nil {
		return 0, fmt.Errorf("failed to attach container exec: %w", err)
	}
	defer hijackedResp.Close()

	stdcopy.StdCopy(stdout, stderr, hijackedResp.Reader)

	for i := 0; i < execInspectMaxAttempt; i++ {
		select {
		case <-ctx.Done():
			return 0, fmt.Errorf("context cancelled")
		default:
			execInspect, err := b.cli.ContainerExecInspect(ctx, idResp.ID)
			if err != nil {
				return 0, fmt.Errorf("failed to inspect an exec: %w", err)
			}
			if !execInspect.Running {
				return execInspect.ExitCode, nil
			}

			logger.Debug("exec process is running so retrying inspection after the interval", "execId", idResp.ID, "attempt", i+1)
			time.Sleep(execInspectInterval)
		}
	}
	return 0, errors.New("exceeded the max exec inspect attempt")
}

func (b *DockerTSDBBackfiller) Close() error {
	return b.cli.Close()
}
