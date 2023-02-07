#!/bin/bash

# Copyright 2023 Democratized Data Foundation
#
# Use of this software is governed by the Business Source License
# included in the file licenses/BSL.txt.
#
# As of the Change Date specified in that file, in accordance with
# the Business Source License, use of this software will be governed
# by the Apache License, Version 2.0, included in the file
# licenses/APL.txt.

# Pre-requirements:
#    - GoLang
#    - Git

# Usage: ./deploy_defradb.sh <PAT> <RELEASE_TAG_OR_COMMIT>
# Example: ./deploy_defradb.sh "v0.4.0" "github_pat_xyz"

COMMIT_TO_DEPLOY="$1";
READ_ONLY_DEFRADB_PAT="$2";
DEFRADB_GIT_REPO="github.com/sourcenetwork/defradb.git";

\git clone "https://git:${READ_ONLY_DEFRADB_PAT}@${DEFRADB_GIT_REPO}";

\cd ./defradb || { printf "\ncd into defradb failed.\n" && exit 2; };

\git checkout "${COMMIT_TO_DEPLOY}" || { printf "\nchecking out commit failed.\n" && exit 3; };

\make deps:modules || { printf "\nbuilding defradb dependencies failed.\n" && exit 4; };

\make install || { printf "\nfailure while installing defradb.\n" && exit 5; };

\defradb version || { printf "\ndefradb installed but not working properly.\n" && exit 6; };

printf "\ndefradb successfully installed.\n";
