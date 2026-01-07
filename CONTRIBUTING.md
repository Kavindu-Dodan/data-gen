# Contributing

All contributions are welcome and encouraged.

Keep things simple and straightforward and follow the _simple & flexible_ philosophy whenever possible.

## Pre-checks

Open an issue and simply work on it. Once your changes are ready,

- Install necessary tools by running `make install-tools`.
- Check for linting errors by running `make lint`.
- Check for import ordering by running `make check-imports`.
- Check unit tests by running `make test`.
- To validate end to end functionality, run `make check-run`.

## Opening a Pull Request

When pre-requisites are met, open a pull request against the `main` branch.
Ensure your PR title follows conventional commit guidelines for automated release management.

You can read more about conventional commits guide from the [official documentation](https://www.conventionalcommits.org/en/v1.0.0/).

For example, if the change is _adding export support to destination ABC_.
Then your PR title can be `feat: add export support to destination ABC`.