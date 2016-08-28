# quickhook

Quickhook is a fast, Unix'y, mildly opinionated Git hook runner. It is designed to be called from within Git hook files; it handles run all user-defined hooks, collecting their output, and reporting failures to output and a non-zero exit status.

## Installation and usage

First you'll need to set quickhook to be called in your Git hooks. To call quickhook before committing you should have a `.git/hooks/pre-commit` file like:

```
#!/bin/sh
quickhook hook pre-commit
```

Quickhook will look for hooks in the `.quickhook/pre-commit/` directory in your repository. See the [`go-vet`](.quickhook/pre-commit/go-vet) file for an example.

## Writing hooks

Right now quickhook only supports pre-commit hooks.

### pre-commit

Pre-commit hooks will receive the list of staged files on stdin. They are expected to write their result to stdout/stderr (quickhook doesn't care), and exit with a non-zero exit code if the commit should be aborted.

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

This informal convention is already followed by many programming languages, linters, and so forth.

## Performance

Quickhook is designed to be as fast and lightweight as possible. There are a few guiding principles for this:

- Ships as a small, self-contained executable.
- Eschew configuration in favor of rigid adherence to Unix'y approach of composing programs.
- Do as much as possible in parallel.
