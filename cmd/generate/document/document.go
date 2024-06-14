package document

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/ajalab/slogen/cmd/common"
	configspec "github.com/ajalab/slogen/internal/config/spec"
	"github.com/ajalab/slogen/internal/document"
	"github.com/ajalab/slogen/internal/spec"
	"github.com/spf13/cobra"
)

const (
	outputJSON           = "json"
	outputYAML           = "yaml"
	outputGoTemplateFile = "go-template-file"
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

	switch {
	case output == outputJSON:
		return common.PrintJSON(document, stdout)
	case output == outputYAML:
		return common.PrintYAML(document, stdout)
	case strings.HasPrefix(output, outputGoTemplateFile+"="):
		goTemplateFileName := strings.TrimPrefix(output, outputGoTemplateFile+"=")
		tmpl, err := template.ParseFiles(goTemplateFileName)
		if err != nil {
			return err
		}
		return tmpl.Execute(stdout, document)
	}
	return fmt.Errorf("unsupported format: %s", output)
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
