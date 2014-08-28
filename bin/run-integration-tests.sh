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

	echo "===================================" >> $LOG_FILE
	echo "= ARGS: ${args}" >> $LOG_FILE
	echo >> $LOG_FILE
	
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

	run_test_suite "--auth-email=true --eventstream=log" ".+Integration.+__Suite(All|AuthEmailTrue)" $*
	run_test_suite "--auth-email=false --eventstream=log" ".+Integration.+__Suite(All|AuthEmailFalse)" $*

	run_test_suite "--auth-email=true" ".+Integration.+__Suite(All|AuthEmailTrue)" $*
	run_test_suite "--auth-email=false" ".+Integration.+__Suite(All|AuthEmailFalse)" $*

	if [ ! -z $REDIS ]; then
		run_test_suite "--auth-email=true --storage=redis --redis-address=$REDIS" ".+Integration.+__Suite(All|AuthEmailTrue)" $*
		run_test_suite "--auth-email=false --storage=redis --redis-address=$REDIS" ".+Integration.+__Suite(All|AuthEmailFalse)" $*
	fi

	if [ ! -z $ETCD ]; then
		run_test_suite "--auth-email=true --storage=etcd --storage-etcd-peer=$ETCD" ".+Integration.+__Suite(All|AuthEmailTrue)" $*
		run_test_suite "--auth-email=false --storage=etcd --storage-etcd-peer=$ETCD" ".+Integration.+__Suite(All|AuthEmailFalse)" $*
	fi

}

LOG_FILE=/tmp/userd-test.log
DEFAULT_ARGS=${DEFAULT_ARGS:-}

cleanup
run_suites

echo
echo "Logoutput can be found at $LOG_FILE"
