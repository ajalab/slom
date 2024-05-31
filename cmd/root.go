package cmd

import (
	"io"

	"github.com/ajalab/slogen/cmd/common"
	"github.com/ajalab/slogen/cmd/generate"
	"github.com/spf13/cobra"
)

func Execute(args []string, stdout, stderr io.Writer) error {
	commonFlags := common.CommonFlags{}
	rootCmd := &cobra.Command{
		Use: "slogen",
	}
	rootCmd.SilenceUsage = true
	rootCmd.PersistentFlags().BoolVarP(&commonFlags.Debug, "debug", "d", false, "enable debug logging")
	rootCmd.AddCommand(generate.NewCommand(&commonFlags))

	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}
