#!/bin/bash

echo
TESTROOT=$1
TESTBED=$2
TESTGROUP=$3
TESTNAME=$4
H8DFILE=$5
SCRIPT=$6

echo Start test $TESTNAME

# create testbed
echo Create testbed...
mkdir "$TESTBED/$TESTNAME"

# run h8d-examiner with stdin script, capture output
echo Running h8d-examiner...
go run h8d-examiner.go $H8DFILE <$SCRIPT >$TESTBED/$TESTNAME/stdout.txt

# compare output
echo Compare output...
diff "$TESTROOT/$TESTGROUP/ref/$TESTNAME.txt" "$TESTBED/$TESTNAME/stdout.txt"
((ECODE=$?))

# if different copy stdout to ref directory
if [ $ECODE -ne 0 ]
then
    ((NUM_FAIL+=1))
    cp "$TESTBED/$TESTNAME/stdout.txt" "$TESTROOT/$TESTGROUP/ref/$TESTNAME.txt"
fi

echo End test $TESTNAME
exit $NUM_FAIL
