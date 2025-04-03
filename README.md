# Locust Core

Core libraries for Locust, a framework for creating automated software for
Cosmos SDK chains.

## Developing

This project uses [Conventional Commits][1] and linting opinions defined in
[.golangci.yml][2].

Git `pre-commit` and `commit-msg` scripts are provided to help align with
project standards.

```sh
ln -sf ../../scripts/commit-msg .git/hooks/commit-msg                                                 (main)
ln -sf ../../scripts/pre-commit .git/hooks/pre-commit
```

[1]: https://www.conventionalcommits.org/en/v1.0.0/
[2]: .golangci.yml
