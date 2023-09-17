#!/bin/sh
set -e
COMMAND=$1
shift
if
    [ "$COMMAND" = "diff" ] ||
    [ "$COMMAND" = "ls-files" ] ||
    [ "$COMMAND" = "rev-list" ] ||
    [ "$COMMAND" = "rev-parse" ] ||
    [ "$COMMAND" = "show" ] ||
    [ "$COMMAND" = "status" ];
then
    # The Git executable below will be replaced at runtime when shimming.
    ACTUAL_GIT "$COMMAND" "$@"
    exit 0
fi
COMBINED=$(echo "$COMMAND  $*" | xargs)
echo "git is not allowed in parallel hooks (git $COMBINED)"
exit 1
