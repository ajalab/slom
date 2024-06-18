package print

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

const (
	outputJSON           = "json"
	outputYAML           = "yaml"
	outputGoTemplateFile = "go-template-file"
)

type Printer interface {
	Print(v any) error
	io.Closer
}

func NewPrinter(w io.Writer, output string) (Printer, error) {
	switch {
	case output == outputJSON:
		return NewJSONPrinter(w), nil
	case output == outputYAML:
		return NewYAMLPrinter(w), nil
	case strings.HasPrefix(output, outputGoTemplateFile+"="):
		goTemplateFileName := strings.TrimPrefix(output, outputGoTemplateFile+"=")
		return NewGoTemplatePrinter(w, goTemplateFileName)
	}
	return nil, fmt.Errorf("unsupported format: %s", output)
}

type JSONPrinter struct {
	e *json.Encoder
}

func NewJSONPrinter(w io.Writer) *JSONPrinter {
	e := json.NewEncoder(w)
	e.SetEscapeHTML(false)
	e.SetIndent("", "    ")

	return &JSONPrinter{e}
}

var _ Printer = &JSONPrinter{}

func (p *JSONPrinter) Print(v any) error {
	return p.e.Encode(v)
}

func (p *JSONPrinter) Close() error {
	return nil
}

type YAMLPrinter struct {
	e *yaml.Encoder
}

func NewYAMLPrinter(w io.Writer) *YAMLPrinter {
	e := yaml.NewEncoder(w)
	e.SetIndent(2)

	return &YAMLPrinter{e}
}

var _ Printer = &YAMLPrinter{}

func (p *YAMLPrinter) Print(v any) error {
	return p.e.Encode(v)
}

func (p *YAMLPrinter) Close() error {
	return p.e.Close()
}

type GoTemplatePrinter struct {
	tmpl *template.Template
	w    io.Writer
}

func NewGoTemplatePrinter(w io.Writer, goTemplateFileName string) (*GoTemplatePrinter, error) {
	tmpl, err := template.ParseFiles(goTemplateFileName)
	if err != nil {
		return nil, err
	}
	return &GoTemplatePrinter{tmpl, w}, nil
}

var _ Printer = &GoTemplatePrinter{}

func (p *GoTemplatePrinter) Print(v any) error {
	return p.tmpl.Execute(p.w, v)
}

func (p *GoTemplatePrinter) Close() error {
	return nil
}
