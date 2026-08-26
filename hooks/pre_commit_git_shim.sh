#!/bin/sh
set -e
COMMAND=$1
shift
if
    [ "$COMMAND" = "cat-file" ] ||
    [ "$COMMAND" = "check-attr" ] ||
    [ "$COMMAND" = "check-ignore" ] ||
    [ "$COMMAND" = "diff" ] ||
    [ "$COMMAND" = "diff-files" ] ||
    [ "$COMMAND" = "diff-index" ] ||
    [ "$COMMAND" = "diff-tree" ] ||
    [ "$COMMAND" = "grep" ] ||
    [ "$COMMAND" = "log" ] ||
    [ "$COMMAND" = "ls-files" ] ||
    [ "$COMMAND" = "ls-tree" ] ||
    [ "$COMMAND" = "merge-base" ] ||
    [ "$COMMAND" = "rev-list" ] ||
    [ "$COMMAND" = "rev-parse" ] ||
    [ "$COMMAND" = "show" ] ||
    [ "$COMMAND" = "status" ];
then
    # The Git executable below will be replaced at runtime when shimming.
    ACTUAL_GIT "$COMMAND" "$@"
    exit $?
fi
COMBINED=$(echo "$COMMAND  $*" | xargs)
echo "git is not allowed in parallel hooks (git $COMBINED)"
exit 1
