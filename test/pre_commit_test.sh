setup() {
  removeAllTrace

  ( cd test/tmp ;
    git init --quiet . ;
    echo "Changed!" > example ;
    git add example )
}

teardown() {
  removeAllTrace
}

removeAllTrace() {
  rm -rf test/tmp/.git
  rm -rf test/tmp/.quickhook
  rm -f test/tmp/example
}

runHook() {
  cd test/tmp
  run ../../quickhook hook pre-commit --no-color
  cd ../..
}

HOOK_DIR=test/tmp/.quickhook/pre-commit

@test "pre-commit fails if any of the hooks failed" {
  mkdir -p $HOOK_DIR

  echo $'#!/bin/bash \n echo "passed" \n exit 0' > $HOOK_DIR/passes
  echo $'#!/bin/bash \n echo "failed" \n exit 1' > $HOOK_DIR/fails
  chmod +x $HOOK_DIR/*

  runHook

  [ "$status" -ne 0 ]
  [ "${lines[0]}" = "fails: fail" ]
  [ "${lines[1]}" = "failed" ]
  [ "${lines[2]}" = "passes: ok" ]
}

@test "pre-commit passes if all hooks pass" {
  mkdir -p $HOOK_DIR

  echo $'#!/bin/bash \n echo "passed" \n exit 0' > $HOOK_DIR/passes1
  echo $'#!/bin/bash \n echo "passed" \n exit 0' > $HOOK_DIR/passes2
  chmod +x $HOOK_DIR/*

  runHook

  [ "$status" -eq 0 ]
  [ "${lines[0]}" = "passes1: ok" ]
  [ "${lines[1]}" = "passes2: ok" ]
}
