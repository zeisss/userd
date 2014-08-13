#!/bin/bash

function run_test_suite() {
	local args=$1
	local filter=$2

	shift

	echo "================================"
	echo "= Args: $args"
	echo "================================"
	echo

	./userd $args >> $LOG_FILE 2>&1 &
	local PID=$!

	trap 'kill $PID' TERM KILL QUIT

	go test ./client -run $filter $*
	kill $PID
}

function cleanup() {
	if [ -f $LOG_FILE ]; then
		rm $LOG_FILE
	fi	
}

LOG_FILE=/tmp/userd-test.log

cleanup
run_test_suite "-auth-email=true" ".+Integration.+__Suite(All|AuthEmailTrue)" $*
run_test_suite "-auth-email=false" ".+Integration.+__Suite(All|AuthEmailFalse)" $*

echo
echo "Logoutput can be found at $LOG_FILE"