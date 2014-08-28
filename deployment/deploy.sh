#!/bin/bash

shopt -s expand_aliases

# Print simple usage information.
usage()
{
    echo "Usage: $0 [--stack] [--userd-version]

where:
    --stack          stack name, used to create cloudformation stack and domain names
    --userd-version  version of userd service docker image
    " 1>&2
    exit 1
}

STACK=""
USERD_VERSION=""

set -- `getopt -u -a --longoptions="stack: userd-version:" "h" "$@"` || usage

while [ $# -gt 0 ]; do
    if [ "$2" = "--" ]; then
        echo "ERROR: $1 cannot be empty"
        echo
        usage
    fi

    case "$1" in
        --stack ) STACK="$2"; shift;;
        --userd-version ) USERD_VERSION="$2"; shift;;
        --)        shift;break;;
        -*)        usage;;
        *)         break;;
    esac
    shift
done

DEFAULT_STACK=cluster-01
STACK=${STACK:-$DEFAULT_STACK}
TUNNEL=${STACK}.fleet.giantswarm.io

if [ ! -n "$STACK" ] || [ ! -n "$USERD_VERSION" ]
then
     usage
fi

# take the first machine to talk to the cluster
alias fleetctl="fleetctl --strict-host-key-checking=false --request-timeout=5 --tunnel $TUNNEL"

# replace variables in service template (set correct userd version)
cat deployment/units/userd\@.template | \
  sed -e "s,\%\%USERD_VERSION\%\%,$USERD_VERSION,g" \
  > deployment/units/userd\@.service

for $i in $(seq 2); do
  fleetctl destroy deployment/units/userd-presence@${i}
  fleetctl destroy deployment/units/userd@${1}
  fleetctl start deployment/units/userd@${1}
  fleetctl start deployment/units/userd-presence@${1}
done

# cleanup
rm deployment/units/userd\@.service
