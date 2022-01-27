#!/bin/bash

TESTROOT=test
TESTBED=tests
TESTGROUP=$1

if [ -d "$TESTBED" ]
then
    echo Removing old directory $TESTBED
    rm -r "$TESTBED"
fi

echo Creating directory $TESTBED
mkdir "$TESTBED"

echo Running all tests...
ECODE=0

# stats
test/bin/run_stdin.sh test tests EmptyHDOSImages 1s40t-stats test/EmptyHDOSImages/data/1s40t.h8d test/bin/stdin_hdos_stats.txt
test/bin/run_stdin.sh test tests EmptyHDOSImages 1s80t-stats test/EmptyHDOSImages/data/1s80t.h8d test/bin/stdin_hdos_stats.txt

# cat
test/bin/run_stdin.sh test tests EmptyHDOSImages 1s40t-cat test/EmptyHDOSImages/data/1s40t.h8d test/bin/stdin_hdos_cat.txt
test/bin/run_stdin.sh test tests EmptyHDOSImages 1s80t-cat test/EmptyHDOSImages/data/1s80t.h8d test/bin/stdin_hdos_cat.txt

# dir
test/bin/run_stdin.sh test tests EmptyHDOSImages 1s40t-dir test/EmptyHDOSImages/data/1s40t.h8d test/bin/stdin_hdos_dir.txt
test/bin/run_stdin.sh test tests EmptyHDOSImages 1s80t-dir test/EmptyHDOSImages/data/1s80t.h8d test/bin/stdin_hdos_dir.txt
