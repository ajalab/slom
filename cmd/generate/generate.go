package generate

import (
	"github.com/ajalab/slom/cmd/common"
	"github.com/ajalab/slom/cmd/generate/document"
	"github.com/ajalab/slom/cmd/generate/prometheus/rule"
	"github.com/ajalab/slom/cmd/generate/prometheus/series"
	"github.com/ajalab/slom/cmd/generate/prometheus/tsdb"
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
	command.AddCommand(document.NewCommand(flags))

	return command
}
