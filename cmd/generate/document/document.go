package document

import (
	"fmt"
	"io"
	"os"

	"github.com/ajalab/slom/cmd/common"
	configspec "github.com/ajalab/slom/internal/config/spec"
	"github.com/ajalab/slom/internal/document"
	"github.com/ajalab/slom/internal/print"
	"github.com/ajalab/slom/internal/spec"
	"github.com/spf13/cobra"
)

func run(
	specConfigFileName string,
	output string,
	stdout io.Writer,
) error {
	specConfigFile, err := os.Open(specConfigFileName)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", specConfigFileName, err)
	}
	defer specConfigFile.Close()

	specConfig, err := configspec.ParseSpecConfig(specConfigFile)
	if err != nil {
		return fmt.Errorf("failed to parse %s as spec config file: %w", specConfigFileName, err)
	}
	spec, err := spec.ToSpec(specConfig)
	if err != nil {
		return fmt.Errorf("failed to convert a spec config %s into spec: %w", specConfigFileName, err)
	}
	document := document.ToDocument(spec)

	printer, err := print.NewPrinter(stdout, output)
	if err != nil {
		return fmt.Errorf("failed to get a printer: %w", err)
	}
	defer printer.Close()

	return printer.Print(document)
}

func NewCommand(flags *common.CommonFlags) *cobra.Command {
	var output string

	command := &cobra.Command{
		Use:   "document [-o output] specFileName",
		Short: "Generate an SLO document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(args[0], output, cmd.OutOrStdout())
		},
	}
	command.Flags().StringVarP(&output, "output", "o", "json", "output format of the generated document. Either \"json\", \"yaml\", \"go-template-file=<filename>\"")

	return command
}
