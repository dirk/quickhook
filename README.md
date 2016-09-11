# quickhook

[![Build Status](https://travis-ci.org/dirk/quickhook.svg?branch=master)](https://travis-ci.org/dirk/quickhook)

Quickhook is a fast, Unix'y, opinionated Git hook runner. It handles running all user-defined hooks, collecting their output, reporting failures, and exiting with a non-zero status code if appropriate.

## Installation

If you're on Mac there is a [Homebrew tap for Quickhook](https://github.com/dirk/homebrew-quickhook):

```sh
brew tap dirk/quickhook
# Tapped 1 formula (26 files, 20.4K)

brew install quickhook
# /usr/local/Cellar/quickhook/1.2.0: 2 files, 7M, built in 8 seconds
```

## Usage

First you'll need to set Quickhook to be called in your Git hooks. The `quickhook install` command will discover hooks defined in the `.quickhook` directory and create Git hook shims for those. For example, the below is what you can expect if you clone this repository and install the shims:

```
$ quickhook install
Create file .git/hooks/commit-msg? [yn] y
Installed shim .git/hooks/commit-msg
Create file .git/hooks/pre-commit? [yn] y
Installed shim .git/hooks/pre-commit
```

The `hook` sub-commands have a some hook-specific options. For example, these are some of the options you can use with the pre-commit hook command:

```sh
# Run the pre-commit hooks on all Git-tracked files in the repository
quickhook hook pre-commit --all

# Run them on just one or more files
quickhook hook pre-commit --files hooks/commit_msg.go hooks/pre_commit.go
```

## Writing hooks

Quickhook will look for hooks in a corresponding sub-directory of the `.quickhook` directory in your repository. For example, it will look for pre-commit hooks in `.quickhook/pre-commit/`. A hook is any executable file in that directory. See the [`go-vet`](.quickhook/pre-commit/go-vet) file for an example.

### pre-commit

Pre-commit hooks receive the list of staged files separated by newlines on stdin. They are expected to write their result to stdout/stderr (Quickhook doesn't care). If they exit with a non-zero exit code then the commit will be aborted and their output displayed to the user.

**Note**: Pre-commit hooks will be executed in parallel and should not mutate the local repository state.

File-and-line-specific errors should be written in the following format:

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

This informal Unix convention is already followed by many programming languages, linters, and so forth.

### commit-msg

Commit-message hooks are run sequentially. They receive a single argument: a path to a temporary file containing the message for the commit. If they exit with a non-zero exit code the commit will be aborted and any stdout/stderr output displayed to the user.

Given that they are run sequentially, `commit-msg` hooks are allowed to mutate the commit message temporary file.

## Performance

Quickhook is designed to be as fast and lightweight as possible. There are a few guiding principles for this:

- Ship as a small, self-contained executable.
- Eschew configuration in favor of rigid adherence to Unix'y approach of composing programs.
- Do as much as possible in parallel.

## License

Released under the Modified BSD license, see [LICENSE](LICENSE) for details.
