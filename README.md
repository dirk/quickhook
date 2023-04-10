# quickhook

![Build Status](https://github.com/dirk/quickhook/actions/workflows/push.yml/badge.svg)
![codecov](https://codecov.io/github/dirk/quickhook/branch/main/graph/badge.svg?token=FRMS9TRJ93)

Quickhook is a Git hook runner designed for speed. It is opinionated where it matters: hooks are executables organized by directory and must exit with a non-zero code on error. Everything else is up to you!

## Installation

### `go install`

If you have your $PATH set up for Go then it's as simple as:

```sh
$ go install github.com/dirk/quickhook
$ quickhook --version
1.5.0
```

To uninstall use `clean -i`:

```sh
$ go clean -i github.com/dirk/quickhook
```

### Homebrew

If you're on Mac there is a [Homebrew tap for Quickhook](https://github.com/dirk/homebrew-quickhook):

```sh
$ brew tap dirk/quickhook
==> Tapping dirk/quickhook
...
Tapped 1 formula (14 files, 12.6KB).

$ brew install quickhook
==> Fetching dirk/quickhook/quickhook
==> Downloading https://github.com/dirk/quickhook/archive/v1.5.0.tar.gz
...
/opt/homebrew/Cellar/quickhook/1.5.0: 5 files, 3.1MB, built in 2 seconds
```

### Linux

Installable debs and RPMs are available for the [latest release](https://github.com/dirk/quickhook/releases/latest).

```sh
# Installing a .deb
wget https://github.com/dirk/quickhook/releases/download/v1.5.0/quickhook-1.5.0-amd64.deb
sudo apt install ./quickhook-1.5.0-amd64.deb

# Installing a .rpm
wget https://github.com/dirk/quickhook/releases/download/v1.5.0/quickhook-1.5.0-amd64.rpm
sudo rpm --install quickhook-1.5.0-amd64.rpm
```

## Usage

First you'll need to install Quickhook in your repository: `quickhook install` command will discover hooks defined in the `.quickhook` directory and create Git hook shims for those. For example, the below is what you can expect from running installation in this repository:

```sh
$ quickhook install
Create file .git/hooks/commit-msg? [yn] y
Installed shim .git/hooks/commit-msg
Create file .git/hooks/pre-commit? [yn] y
Installed shim .git/hooks/pre-commit
```

Quickhook provides some options to run various hooks directly for development and testing. This way you don't have to follow the whole Git commit workflow just to exercise the new hook you're working on.

```sh
# Run the pre-commit hooks on all Git-tracked files in the repository
$ quickhook hook pre-commit --all

# Run them on just one or more files
$ quickhook hook pre-commit --files=hooks/commit_msg.go,hooks/pre_commit.go
```

You can see all of the options by passing `--help` to the sub-command:

```sh
$ quickhook hook pre-commit --help
...
OPTIONS:
   --all, -a    Run on all Git-tracked files
   --files, -F  Run on the given comma-separated list of files
```

## Writing hooks

Quickhook will look for hooks in a corresponding sub-directory of the `.quickhook` directory in your repository. For example, it will look for pre-commit hooks in `.quickhook/pre-commit/`. A hook is any executable file in that directory.

### pre-commit

Pre-commit hooks receive the list of staged files separated by newlines on stdin. They are expected to write their result to stdout/stderr (Quickhook doesn't care). If they exit with a non-zero exit code then the commit will be aborted and their output displayed to the user. See the [`go-vet`](.quickhook/pre-commit/go-vet) file for an example.

**Note**: Pre-commit hooks will be executed in parallel and should not mutate the local repository state. For this reason `git` is shimmed on the hooks' $PATH to be unavailable.

#### Mutating hooks

You can also add executables to `.quickhook/pre-commit-mutating/`. These will be run _sequentially_, without Git shimmed, and may mutate the local repository state.

#### Suggested formatting

If you're unsure how to format your lines, there's an informal Unix convention which is already followed by many programming languages, linters, and so forth.

```
some/directory/and/file.go:123: Something doesn't look right
```

A more formal definition of an error line is:

- Sequence of characters representing a valid path
- A colon (`:`) character
- Integer of the line where the error occurred
- A color character followed by a space character
- Any printable character describing the error
- A newline (`\n`) terminating the error line

### commit-msg

Commit-message hooks are run sequentially. They receive a single argument: a path to a temporary file containing the message for the commit. If they exit with a non-zero exit code the commit will be aborted and any stdout/stderr output displayed to the user.

Given that they are run sequentially, `commit-msg` hooks are allowed to mutate the commit message temporary file.

## Performance

Quickhook is designed to be as fast and lightweight as possible. There are a few guiding principles for this:

- Ship as a small, self-contained executable.
- No configuration.
- Do as much as possible in parallel.

## Contributing

Contributions are welcome. If you want to use the locally-built version of Quickhook in the Git hooks, there's a simple 3-line script that will set that up:

```sh
$ ./scripts/install.sh
Installed shim .git/hooks/commit-msg
Installed shim .git/hooks/pre-commit
```

Building and testing should be straightforward:

```sh
# Build a quickhook executable:
$ go build

# Run all tests:
$ go test ./...
```

**Warning**: Many of the tests are integration-style tests which depend on a locally-built Quickhook executable. If you see unexpected test failures, please first try running `go build` before you rerun tests.

There's also a script that will generate and open an HTML page with coverage:

```sh
$ ./scripts/coverage.sh
```

## License

Released under the Modified BSD license, see [LICENSE](LICENSE) for details.
