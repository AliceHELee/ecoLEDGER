#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#
# Exit on first error
set -e

# don't rewrite paths for Windows Git Bash users
export MSYS_NO_PATHCONV=1

# clean the keystore
rm -rf ./hfc-key-store
rm -rf ./wallet

# launch network; create channel and join peer to channel
pushd ../basic-network
./start.sh
popd

pushd ../chaincode
./cc_waste.sh
./cc_recycle.sh
popd

