package generate

import (
	"github.com/ajalab/slogen/cmd/common"
	"github.com/ajalab/slogen/cmd/generate/prometheus/rule"
	"github.com/ajalab/slogen/cmd/generate/prometheus/series"
	"github.com/ajalab/slogen/cmd/generate/prometheus/tsdb"
	"github.com/spf13/cobra"
)

func NewCommand(flags *common.CommonFlags) *cobra.Command {
	command := &cobra.Command{
		Use:   "generate",
		Short: "Generate files from a config",
	}
	command.AddCommand(rule.NewCommand(flags))
	command.AddCommand(series.NewCommand(flags))
	command.AddCommand(tsdb.NewCommand(flags))

	return command
}
