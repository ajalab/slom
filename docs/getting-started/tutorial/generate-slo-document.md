# Generate SLO document

It's recommended to document your defined SLO to ensure that other teams and stakeholders can review and understand it.
slom supports generating an SLO document from the spec file using a template.

## Add annotations and labels to the spec

First, update the [previous SLO spec](./alert-on-error-budget-consumption.md) to add labels and annotations.

```yaml title="example.yaml"
--8<-- "examples/tutorial/spec/document.yaml"
```

1. Added `labels` to indicate the groups to which this SLO specification belongs.
2. Added `annotations` to include arbitrary metadata for this SLO spec.
3. Added `annotations` to include arbitrary metadata for this availability SLO.

## Prepare an SLO document template

To render an SLO document, you can use a template file that follows Go's [template syntax](https://pkg.go.dev/text/template).
Below is an exmaple of an SLO document template, similar to the [SLO Documentation Template](https://docs.google.com/document/d/1SNgnAjRT1jrMa7vGHK0J_0jJEDvKJ5JmTEXFvNRDaHE/edit#heading=h.x9snb54sjlu9) provided by Google Cloud.

````md title="example.tmpl"
--8<-- "examples/tutorial/go-template-file/document.tmpl"
````

!!! info
    See the [Document](../../references/document/index.md) reference for the available fields.


## Generate an SLO document

After preparing an SLO document template file, run [`slom generate document`](../../references/cli/generate/document.md) command to generate a Prometheus rule file based on the SLO spec.

```sh
slom generate document -o go-template-file=document.tmpl example.yaml
```

See the [appendix](./appendix-example-of-generated-slo-document.md) for the result.

!!! info
    slom can also generate an SLO document in formats such as JSON, enabling other tools to render the document as needed.
    For details, refer to the [`slom generate document`](../../references/cli/generate/document.md) command reference.
