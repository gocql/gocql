#!/bin/bash

set -e

function run_tests() {
	local clusterSize=3
	local version=$1
    local auth=$2

    local keypath="$(pwd)/testdata/pki"

    if [ "$auth" = true ]; then
		clusterSize=1
	fi

	local conf=(
	    "client_encryption_options.enabled: true"
	    "client_encryption_options.keystore: $keypath/.keystore"
	    "client_encryption_options.keystore_password: cassandra"
	    "client_encryption_options.require_client_auth: true"
	    "client_encryption_options.truststore: $keypath/.truststore"
	    "client_encryption_options.truststore_password: cassandra"
	    "concurrent_reads: 2"
	    "concurrent_writes: 2"
	    "rpc_server_type: sync"
	    "rpc_min_threads: 2"
	    "rpc_max_threads: 2"
	    "write_request_timeout_in_ms: 5000"
	    "read_request_timeout_in_ms: 5000"
	)

	ccm create test -v binary:$version -n $clusterSize -d --vnodes
    ccm updateconf "${conf[@]}"

	sed -i '/#MAX_HEAP_SIZE/c\MAX_HEAP_SIZE="256M"' ~/.ccm/repository/$version/conf/cassandra-env.sh
	sed -i '/#HEAP_NEWSIZE/c\HEAP_NEWSIZE="100M"' ~/.ccm/repository/$version/conf/cassandra-env.sh

	if [ "$auth" = true ]
    then
		ccm updateconf 'authenticator: PasswordAuthenticator' 'authorizer: CassandraAuthorizer'
	fi

	ccm start -v
	ccm status
	ccm node1 nodetool status

	local proto=2
	if [[ $version == 1.2.* ]]; then
		proto=1
	elif [[ $version == 2.1.* ]]; then
		proto=3
	fi

	if [ "$auth" = true ]
    then
    	echo "list users;" | CQLSH_HOST=127.0.0.1 cqlsh -u cassandra -p cassandra
    	go test -v . -timeout 15s -run=TestAuthentication -tags integration -runssl -runauth -proto=$proto -cluster=$(ccm liveset) -clusterSize=$clusterSize -autowait=1000ms
	else
		go test -timeout 5m -tags integration -cover -v -runssl -proto=$proto -rf=3 -cluster=$(ccm liveset) -clusterSize=$clusterSize -autowait=2000ms -compressor=snappy ./... | tee results.txt

		if [ ${PIPESTATUS[0]} -ne 0 ]; then
			echo "--- FAIL: ccm status follows:"
			ccm status
			ccm node1 nodetool status
			ccm node1 showlog > status.log
			cat status.log
			echo "--- FAIL: Received a non-zero exit code from the go test execution, please investigate this"
			exit 1
		fi

		cover=`cat results.txt | grep coverage: | grep -o "[0-9]\{1,3\}" | head -n 1`

		if [[ $cover -lt "55" ]]; then
			echo "--- FAIL: expected coverage of at least 60 %, but coverage was $cover %"
			exit 1
		fi
	fi

	ccm clear
}

run_tests $1 $2
