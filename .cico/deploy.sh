#!/bin/bash
set -xe

# Make sure we are in the right directory
cd $(dirname $(readlink -f $0))/../
source .cico/setup.sh

setup

# Do the test on deploy if it become too slow separate it to another job i.e: https://git.io/fpCrg
dotest

deploy
