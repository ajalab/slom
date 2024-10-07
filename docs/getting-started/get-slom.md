# Get slom

This document outlines how to install slom in your environment.

## Using pre-compiled binaries

Pre-compiled binaries for slom are available on [GitHub Releases](https://github.com/ajalab/slom/releases).

## Using `go install`

You can install slom with the `go install` command:

```shell
go install github.com/ajalab/slom
```

To install a specific version, append one of the following suffixes:

- `@vX.Y.Z`: installs version `vX.Y.Z`
- `@latest`: install the latest published version
- `@main`: install from the latest commit on the `main` branch

## Using container images

Container images for slom are available on the [@GitHub container registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry).

For instance, you can pull an image using Docker:

```shell
docker pull ghcr.io/ajalab/slom
```

For a list of available versions, see the [slom](https://github.com/ajalab/slom/pkgs/container/slom) GitHub package.

## Building from the source

You can build slom from the source with the `go build` command.
