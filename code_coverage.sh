#!/usr/bin/env bash

if [ -z $GO_TEST_TIMEOUT ]; then
  GO_TEST_TIMEOUT="4s"
fi

if [ -z $TBN_CI_BUILD ]; then
  go test $GO_TEST_RUNNER $@  -cover -timeout $GO_TEST_TIMEOUT
  exit $?
fi

function die() {
  echo $*
  exit 1
}

ERROR=""
COVERAGE_FILE=coverage.txt
TMP_FILE=coverage_tmp.txt

rm -rf $COVERAGE_FILE
touch $COVERAGE_FILE
rm -rf $TMP_FILE

for PKG in $@; do
    touch $TMP_FILE
    go test $GO_TEST_RUNNER \
      -timeout $GO_TEST_TIMEOUT \
      -covermode=count \
      -coverprofile=$TMP_FILE \
      $PKG || ERROR="$ERROR $PKG"

    cat $TMP_FILE | grep -v "/mock_" >> $COVERAGE_FILE
    rm -rf $TMP_FILE
done

if [ ! -z "$ERROR" ]
then
    die "Encountered error for one or more packages: $ERROR"
fi
