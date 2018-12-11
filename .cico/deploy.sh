#!/bin/bash
set -xe

# Make sure we are in the right directory
cd $(dirname $(readlink -f $0))/../
source .cico/setup.sh

setup

deploy

# Just deploy once!
if [ "$TARGET" != "rhel" ]; then
    deploy_devcluster build || true # Don't fail on deploying the cluster it's not critical yet
fi
