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

# HDOS
# stats
test/bin/run_stdin.sh test tests EmptyHDOSImages 1s40t-stats test/EmptyHDOSImages/data/1s40t.h8d test/bin/stdin_hdos_stats.txt
test/bin/run_stdin.sh test tests EmptyHDOSImages 1s80t-stats test/EmptyHDOSImages/data/1s80t.h8d test/bin/stdin_hdos_stats.txt

# cat
test/bin/run_stdin.sh test tests EmptyHDOSImages 1s40t-cat test/EmptyHDOSImages/data/1s40t.h8d test/bin/stdin_hdos_cat.txt
test/bin/run_stdin.sh test tests EmptyHDOSImages 1s80t-cat test/EmptyHDOSImages/data/1s80t.h8d test/bin/stdin_hdos_cat.txt

# dir
test/bin/run_stdin.sh test tests EmptyHDOSImages 1s40t-dir test/EmptyHDOSImages/data/1s40t.h8d test/bin/stdin_hdos_dir.txt
test/bin/run_stdin.sh test tests EmptyHDOSImages 1s80t-dir test/EmptyHDOSImages/data/1s80t.h8d test/bin/stdin_hdos_dir.txt

# CP/M
# stats
test/bin/run_stdin.sh test tests CPM_Apps c80_1-stats test/CPM_Apps/data/C80CPM1.h8d test/bin/stdin_cpm_stats.txt
test/bin/run_stdin.sh test tests CPM_Apps c80_2-stats test/CPM_Apps/data/C80CPM2.h8d test/bin/stdin_cpm_stats.txt
test/bin/run_stdin.sh test tests CPM_Apps c80_3-stats test/CPM_Apps/data/C80CPM3.h8d test/bin/stdin_cpm_stats.txt

# cat
test/bin/run_stdin.sh test tests CPM_Apps c80_1-cat test/CPM_Apps/data/C80CPM1.h8d test/bin/stdin_cpm_cat.txt
test/bin/run_stdin.sh test tests CPM_Apps c80_2-cat test/CPM_Apps/data/C80CPM2.h8d test/bin/stdin_cpm_cat.txt
test/bin/run_stdin.sh test tests CPM_Apps c80_3-cat test/CPM_Apps/data/C80CPM3.h8d test/bin/stdin_cpm_cat.txt

# dir
test/bin/run_stdin.sh test tests CPM_Apps c80_1-dir test/CPM_Apps/data/C80CPM1.h8d test/bin/stdin_cpm_dir.txt
test/bin/run_stdin.sh test tests CPM_Apps c80_2-dir test/CPM_Apps/data/C80CPM2.h8d test/bin/stdin_cpm_dir.txt
test/bin/run_stdin.sh test tests CPM_Apps c80_3-dir test/CPM_Apps/data/C80CPM3.h8d test/bin/stdin_cpm_dir.txt
