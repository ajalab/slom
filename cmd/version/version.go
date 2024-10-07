package version

import (
	"errors"
	"fmt"
	"io"
	"runtime/debug"

	"github.com/spf13/cobra"
)

func run(stdout io.Writer) error {
	info, err := getBuildInfo()
	if err != nil {
		return err
	}

	out := fmt.Sprintf(`slom version %s (revision %s)
  go version: %s
  platform: %s`,
		info.moduleVersion, info.revision, info.goVersion, info.platform)
	_, err = fmt.Fprintln(stdout, out)
	return err
}

type buildInfo struct {
	moduleVersion string
	revision      string
	goVersion     string
	platform      string
}

func getBuildInfo() (*buildInfo, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return nil, errors.New("build info is not available possibly because the binary is built without module support")
	}

	var arch, os, revision string
	for _, s := range info.Settings {
		switch s.Key {
		case "GOARCH":
			arch = s.Value
		case "GOOS":
			os = s.Value
		case "vcs.revision":
			revision = s.Value
		}
	}

	return &buildInfo{
		moduleVersion: info.Main.Version,
		revision:      revision,
		goVersion:     info.GoVersion,
		platform:      fmt.Sprintf("%s/%s", os, arch),
	}, nil
}

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Prints the build information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.OutOrStdout())
		},
	}
}
