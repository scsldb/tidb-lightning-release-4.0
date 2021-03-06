# How to contribute

This document outlines some of the conventions on development workflow, commit
message formatting, contact points and other resources to make it easier to get
your contribution accepted.

## Getting started

- Fork the repository on GitHub.
- Read the README.md for build instructions.
- Play with the project, submit bugs, submit patches!

## Building TiDB-Lightning

Developing TiDB-Lightning requires:

* [Go 1.13.5+](http://golang.org/doc/code.html)
* An internet connection to download the dependencies

Simply run `make` to build the program.

```sh
make
```

### Running tests

This project contains unit tests and integration tests with coverage collection.
See [tests/README.md](./tests/README.md) for how to execute and add tests.

### Updating generated code

TiDB-Lightning contains some generated source code for parsing. To modify them,
you also need to install:

* [Ragel 6.10+](https://www.colm.net/open-source/ragel/)
* [protoc 3.6+](https://github.com/protocolbuffers/protobuf/releases)
* [protoc-gen-gogofaster 1.2+](https://github.com/gogo/protobuf#more-speed-and-more-generated-code)

Run `make data_parsers` to regenerate these source code.

### Updating dependencies

TiDB-Lightning manages dependencies using [Go 1.11 module](https://github.com/golang/go/wiki/Modules).
To add or update a dependency, either

* Use the `go mod edit` command to change the dependency, or
* Edit `go.mod` and then run `make update` to update the checksum.

## Contribution flow

This is a rough outline of what a contributor's workflow looks like:

- Create a topic branch from where you want to base your work. This is usually `master`.
- Make commits of logical units and add test case if the change fixes a bug or adds new functionality.
- Run tests and make sure all the tests are passed.
- Make sure your commit messages are in the proper format (see below).
- Push your changes to a topic branch in your fork of the repository.
- Submit a pull request.
- Your PR must receive LGTMs from two maintainers.

Thanks for your contributions!

### Code style

The coding style suggested by the Golang community is used in TiDB-Lightning.
See the [style doc](https://github.com/golang/go/wiki/CodeReviewComments) for details.

Please follow this style to make TiDB-Lightning easy to review, maintain and develop.

### Format of the Commit Message

We follow a rough convention for commit messages that is designed to answer two
questions: what changed and why. The subject line should feature the what and
the body of the commit should describe the why.

```
restore: add comment for variable declaration

Improve documentation.
```

The format can be described more formally as follows:

```
<subsystem>: <what changed>
<BLANK LINE>
<why this change was made>
<BLANK LINE>
<footer>(optional)
```

The first line is the subject and should be no longer than 70 characters, the
second line is always blank, and other lines should be wrapped at 80 characters.
This allows the message to be easier to read on GitHub as well as in various
git tools.

If the change affects more than one subsystem, you can use comma to separate them like `restore,mydump:`.

If the change affects many subsystems, you can use ```*``` instead, like ```*:```.

For the why part, if no specific reason for the change,
you can use one of some generic reasons like "Improve documentation.",
"Improve performance.", "Improve robustness.", "Improve test coverage."

## Related projects

This repository is one of the many components forming the whole TiDB-Lightning
Toolset.

The source code of `tikv-importer` can be found in the TiKV project:
<https://github.com/tikv/tikv/tree/master/src/import/>

The SQL???KV encoder is part of the TiDB project:
<https://github.com/pingcap/tidb>

The gRPC interface between `tidb-lightning` and `tikv-importer` is found in the
`kvproto` repository: <https://github.com/pingcap/kvproto>
