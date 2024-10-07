package cmd

import (
	"io"

	"github.com/ajalab/slom/cmd/common"
	"github.com/ajalab/slom/cmd/generate"
	"github.com/ajalab/slom/cmd/version"
	"github.com/spf13/cobra"
)

func Execute(args []string, stdout, stderr io.Writer) error {
	commonFlags := common.CommonFlags{}
	rootCmd := &cobra.Command{
		Use: "slom",
	}
	rootCmd.SilenceUsage = true
	rootCmd.PersistentFlags().BoolVarP(&commonFlags.Debug, "debug", "d", false, "enable debug logging")
	rootCmd.AddCommand(generate.NewCommand(&commonFlags))
	rootCmd.AddCommand(version.NewCommand())

	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}
