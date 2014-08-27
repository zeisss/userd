#!/bin/bash

function run_test_suite() {
	local args=$1
	local filter=$2

	shift 2

	echo "================================"
	echo "= Args: ${args}"
	echo "= Default: ${DEFAULT_ARGS}"
	echo "================================"
	echo

	./userd ${DEFAULT_ARGS} ${args} >> ${LOG_FILE} 2>&1 &
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

function run_suites() {

	run_test_suite "-auth-email=true -eventstream=log" ".+Integration.+__Suite(All|AuthEmailTrue)" $*
	run_test_suite "-auth-email=false -eventstream=log" ".+Integration.+__Suite(All|AuthEmailFalse)" $*

	run_test_suite "-auth-email=true" ".+Integration.+__Suite(All|AuthEmailTrue)" $*
	run_test_suite "-auth-email=false" ".+Integration.+__Suite(All|AuthEmailFalse)" $*

}

LOG_FILE=/tmp/userd-test.log
DEFAULT_ARGS=${DEFAULT_ARGS:-}

cleanup
run_suites

echo
echo "Logoutput can be found at $LOG_FILE"
