# [![Go Report Card](https://goreportcard.com/badge/github.com/antham/ghokin)](https://goreportcard.com/report/github.com/antham/ghokin) [![codecov](https://codecov.io/gh/antham/ghokin/branch/master/graph/badge.svg)](https://codecov.io/gh/antham/ghokin) [![GitHub tag](https://img.shields.io/github/tag/antham/ghokin.svg)]()

Ghokin format and apply transformation on gherkin files.

---

- [Install](#install)
- [Usage](#usage)
- [Documentation](#documentation)
- [Contribute](#contribute)

---

## Install

Download the latest binary for your achitecture [here](https://github.com/antham/ghokin/releases/latest).

If you can't find a binary for your architecture, install the go toolchain, clone the repository and run : `go install .`.

## Usage

```
Clean and/or apply transformation on gherkin files

Usage:
  ghokin [command]

Available Commands:
  check       Check files/folders are well formatted
  completion  Generate the autocompletion script for the specified shell
  fmt         Format stdin or feature files/folders
  help        Help about any command
  version     App version

Flags:
      --config string   config file
  -h, --help            help for ghokin

Use "ghokin [command] --help" for more information about a command.
```

⚠️ Ghokin works only on `UTF-8` encoded files, it will detect and convert automatically files that are not encoded in this charset.

### fmt stdout

Dump stdin or a feature file formatted on stdout

```
ghokin fmt stdout features/test.feature
```

or

```
cat features/test.feature|ghokin fmt stdout
```

### fmt replace

Format and replace files or all files in one or several directory

```
ghokin fmt replace features/test-1.feature features/test-2.feature
```

or

```
ghokin fmt replace features-1/ features-2/
```

### check

Ensure files or all files in one or several directory are well formatted, exit with an error code otherwise

```
ghokin check features/test-1.feature features/test-2.feature
```

or

```
ghokin check features-1/ features-2/
```

## Documentation

### Shell commands

You can run shell commands from within your feature file to transform some datas with annotations, to do so you need first to define in the config an alias and afterwards you can simply "comment" the line before the line you want to transform with that alias.
For instance let say `@json` calls behind the curtain `jq`, we could validate and format some json in our feature like so :

```
Feature: A Feature
  Description

  Scenario: A scenario to test
    Given a thing
    # @json
    """
    {
      "test": "test"
    }
    """
```

### Config

Defaut config is to use 2 spaces for indentation.

It's possible to override configuration by defining a `.ghokin.yml` file in the home directory or in the current directory where we are running the binary from :

```
indent: 2
aliases:
  json: "jq ."
```

Aliases key defined [shell commands](#shell-commands) callable in comments as we discussed earlier.

It's possible to use environments variables instead of a static config file :

```
export GHOKIN_INDENT=2
export GHOKIN_ALIASES='{"json":"jq ."}'
```

## Contribute

If you want to add a new feature to ghokin project, the best way is to open a ticket first to know exactly how to implement your changes in code.

### Setup

After cloning the repository you need to install vendors with `go mod vendor`
To test your changes locally you can run go tests with : `make test-all`
