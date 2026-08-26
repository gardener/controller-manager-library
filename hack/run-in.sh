#!/usr/bin/env bash

# SPDX-FileCopyrightText: Copyright Contributors to the Gardener project
#
# SPDX-License-Identifier: Apache-2.0

# run something in a directory
# goes hand-in .run-controller-gen.sh

cd "$1"
echo "DIR: $1"
shift
echo "CMD: $@"
"$@"
