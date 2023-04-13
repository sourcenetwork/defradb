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

#######################################################################################################################
#                                                DESCRIPTION
#######################################################################################################################
#
# This script helps change your expected values to actual values using your go test output. This script should be
# therefore ran from the top-level / root directory. It is recommended to use this script in safe mode (DEFAULT),
# be carefull if running in unsafe mode as it may overwrite your files that weren't commited.
#
# Pre-requirements:
#    - awk
#    - git
#    - go (GoLang)
#    - jq
#    - rg (Rip Grep)
#    - sed
#    - sort
#    - sponge (or `moreutils` package)
#    - tee
#
# USAGE: ./fix_failing_cids.sh
# USAGE: ./fix_failing_cids.sh unsafe
#######################################################################################################################

SAFE_OR_UNSAFE="${1:-safe}"
if [ "${SAFE_OR_UNSAFE}" != "safe" ] && [ "${SAFE_OR_UNSAFE}" != "unsafe" ]; then
    printf "\nError: Invalid argument [%s], can be either: 'safe' or 'unsafe'\n" "${SAFE_OR_UNSAFE}";
    exit;
fi

# Helper function to find if a package / program exists or not.
Exists() { which "${1}" &> /dev/null; echo ${?}; }

if [[ $(Exists "awk") -ne 0 ]]; then echo "Missing dependency: awk"; exit; fi
if [[ $(Exists "git") -ne 0 ]]; then echo "Missing dependency: git"; exit; fi
if [[ $(Exists "go") -ne 0 ]]; then echo "Missing dependency: go"; exit; fi
if [[ $(Exists "jq") -ne 0 ]]; then echo "Missing dependency: jq"; exit; fi
if [[ $(Exists "rg") -ne 0 ]]; then echo "Missing dependency: rg"; exit; fi
if [[ $(Exists "sed") -ne 0 ]]; then echo "Missing dependency: sed"; exit; fi
if [[ $(Exists "sort") -ne 0 ]]; then echo "Missing dependency: sort"; exit; fi
if [[ $(Exists "sponge") -ne 0 ]]; then echo "Missing dependency: sponge"; exit; fi
if [[ $(Exists "tee") -ne 0 ]]; then echo "Missing dependency: tee"; exit; fi

# If the caller is currently in a git repository, and not running in unsafe mode, where
# they also have a dirty git workspace (uncommited/staged changes) then don't let them
# execute this script. This is to protect their current work from getting overwritten.
if [ "${SAFE_OR_UNSAFE}" == "safe" ] && \git rev-parse --is-inside-work-tree > /dev/null 2>&1; then
    # Now that we know this is a git repository, ensure they are calling the script
    # from top-level, and doesn't have uncommited work.
    if [[ "${PWD}" == "$(git rev-parse --show-toplevel)" ]]; then
        \git diff-index --quiet HEAD -- || { printf "\nFatal: There are uncommited changes.\n" && exit; };
    else
        printf "\nFatal: Must be in toplevel directory, but was not.\n";
        exit;
    fi
fi

# If no go root files, then probably not a go directory, so exit.
if [ ! -f "./go.mod" ]; then
    printf "\nFatal: Must be a go workspace.\n";
    exit;
fi

#######################################################################################################################
#                                               VARIABLES
#######################################################################################################################

# Temporary directory to store our output files to.
readonly SCRIPT_NAME="${0##*/}";
readonly SCRIPT_NAME_WITHOUT_EXT="${SCRIPT_NAME%%.*}";
DIR_TEMPORARY=$(mktemp "defradb_${SCRIPT_NAME_WITHOUT_EXT}_XXXXX" -dtq;);

readonly FILE_WITH_GO_TEST_OUTPUT="${DIR_TEMPORARY}/go_test_output.json";
readonly FILE_WITH_ONLY_FAILS="${DIR_TEMPORARY}/only_failed_tests.json";
readonly DIR_DEFRADB_TESTS="./tests/";

# Stores the pair of the expected/actul value, and name of the test: <EXPECTED>|<ACTUAL>|<TEST_NAME>
# Example: "bafybeiebgsbhyoi4spgbuw5jkjfsa"|"bafybeifh7ztz46q2awyyk2yql2fqy"|TestEventsSimpleWithUpdate
readonly FILE_WITH_REPLACING_ACTIONS="${DIR_TEMPORARY}/replacing_actions.txt";

# Normal regex patterns.
readonly REGEX_RG_NORMAL_EXPECTED="^.*expected.*:\s";
readonly REGEX_RG_NORMAL_ACTUAL="^.*actual.*:\s";
readonly REGEX_RG_SYMBOL_EXPECTED="\s*\-\s*.*\s*\"cid\":\s*.*\s*.*\s\"";
readonly REGEX_RG_SYMBOL_ACTUAL="\s*\+\s*.*\s*\"cid\":\s*.*\s*.*\s\"";

# Regex patterns that are bash comparable.
readonly REGEX_TO_FIND_NORMAL_EXPECTED="^.*expected.*:[[:blank:]]*\"";
readonly REGEX_TO_FIND_SYMBOL_EXPECTED="[[:blank:]]*\-[[:blank:]]*.*[[:blank:]]*\"cid\":[[:blank:]]*.*[[:blank:]]*.*[[:blank:]]\"";
readonly REGEX_TO_FIND_NORMAL_ACTUAL="^.*actual.*:[[:blank:]]*\"";
readonly REGEX_TO_FIND_SYMBOL_ACTUAL="[[:blank:]]*\+[[:blank:]]*.*[[:blank:]]*\"cid\":[[:blank:]]*.*[[:blank:]]*.*[[:blank:]]\"";
readonly REGEX_TO_CHECK_COMPLEX=".*map\[string\]interface[[:blank:]]\{\}.*";

# Just an additional protection to avoid changing smaller values.
# ONLY replace expected with actual if it's length is atleast this number.
# Want to avoid situation of replacing all `1` in a file to `2` for example.
# Note: this includes the character count for `"` as well. i.e. "a" is length 3.
readonly MINIMUM_EXPECTED_LENGTH=8;

# Map that will store the test name as key, and the file path of that test as value.
declare -A MAP_OF_TEST_FILES

# Caches a map with all exepected keys and their actual values that we found so far.
# These are used to help fix/replace cids in complex files.
declare -A MAP_OF_EXPECTED_ACTUAL_PAIR;

# Map to store all test file paths that had a complex expected value to replace.
# These tests then we can try to fix using our cached cid pairs.
declare -A MAP_OF_COMPLEX_TESTS;


#######################################################################################################################
#                                               STAGE 1
#######################################################################################################################
# We start by runing the go tests and dumping the json output to a file, then we trim that file filtering out
# the passed tests, so that only the failed tests remain. Then we find where each test is located, and store
# the location path mapped to the test name.
#######################################################################################################################

# Run the tests in a serial manner and store the output.
go test -p 1 -json ./... | tee "${FILE_WITH_GO_TEST_OUTPUT}";

# Gather names of all the tests that passed successfully.
PASSED_TESTS=$(\rg '^.*\-\-\- PASS:' -r "" "${FILE_WITH_GO_TEST_OUTPUT}" | awk '{print $1}';);

# Gather names of all the tests that failed.
FAILED_TESTS=$(\rg '^.*\-\-\- FAIL:' -r "" "${FILE_WITH_GO_TEST_OUTPUT}" | awk '{print $1}';);

# Copy the file, so the trimmings don't effect the original file.
\cp "${FILE_WITH_GO_TEST_OUTPUT}" "${FILE_WITH_ONLY_FAILS}"

printf "\nInfo: Remove / trim all irrelevent information...\n";

# Remove all lines that don't have a "Test" field.
\rg "\"Test\":\"(.*?)\"" "${FILE_WITH_ONLY_FAILS}" | sponge "${FILE_WITH_ONLY_FAILS}";

# Remove all tests that were passed.
for TEST in ${PASSED_TESTS}; do
    printf "Trim Passed Test = [ %s ]\n" "${TEST}";
    \rg -v "\"Test\":\"${TEST}\"" "${FILE_WITH_ONLY_FAILS}" | sponge "${FILE_WITH_ONLY_FAILS}";
done

# Find file paths of all the failed tests, and populate the appropriate map.
printf "\nInfo: Find file paths of every failed test...\n";
for TEST in ${FAILED_TESTS}; do
    PATH_TO_TEST=$(\rg -Fl "${TEST}" "${DIR_DEFRADB_TESTS}" --type go | head -n 1;);
    if [[ -z "${PATH_TO_TEST}" ]]; then
        printf "\nInfo: No file found for [ %s ] under [ %s ], trying to find in root dir...\n" "${TEST}" "${DIR_DEFRADB_TESTS}";
        PATH_TO_TEST=$(\rg -Fl "${TEST}" --type go | head -n 1;);
    fi

    if [[ -n "${PATH_TO_TEST}" ]]; then
        printf "\nInfo: Found file for [ %s ] under [ %s ]\n" "${TEST}" "${PATH_TO_TEST}";
        # Hash map of test name mapped to it's file path.
        MAP_OF_TEST_FILES[${TEST}]="${PATH_TO_TEST}";
    else
        printf "\nWarning: No test file found for test=[ %s ]\n" "${TEST}";
    fi

done


#######################################################################################################################
#                                               STAGE 2
#######################################################################################################################
# In this stage we now go through the trimmed test output that only has failed tests, and try to find matching
# expected / actual pairs. If a pair is found we store that, with the test's file name where we need to use the
# pair on. Additionally we also cache the pair in a key-value map where key=expected, and value=actual. This
# cached map can help us later convert some complex cases which weren't easy to replace normally. In the end
# of this stage we should have all the actions we need to take to replace the simple cases.
#######################################################################################################################

# Action can be one of: "SEARCHING", "RUNNING", "PAIRING".
# Where each action is defined as:
#   "SEARCHING" - Looking for a test run.
#   "RUNNING" - Found the start of a a test run, now find the 'expected'/'actual' pair.
#   "PAIRING" - Find the expected (to replace) and actual (to replace with) values and pair them.
CURRENT_ACTION="SEARCHING";

# Name of last test, if suddenly a new test name is encountered while running, then reset the action type.
LAST_TEST_NAME="";

# Store the found expected value, until we encounter the actual value to make a pair.
EXPECTED_SO_FAR="";

# Pairs can be one of: "" (default is empty), "NORMAL", "SYMBOL", 
# Where each pair type is defined as:
#   "" - Looking for a test run.
#   "NORMAL" -  Matching pair pattern contains `expected :`, and `actual :` lines..
#   "SYMBOL" - Matching pair pattern contains `-` for expected, and `+` for actual.
PAIR_TYPE="";

while IFS= read -r LINE; do

    # Every line may have these fields:
    #   - Time    time.Time // encodes as an RFC3339-format string
    #   - Elapsed float64 // seconds
    #   - Action  string
    #   - Package string
    #   - Test    string
    #   - Output  string
    #
    # We are only interested in `Test`, `Output` and `Action`.
    CURRENT_LINE_TEST_NAME=$(echo "${LINE}" | jq --join-output '.Test // empty';)
    CURRENT_LINE_OUTPUT=$(echo "${LINE}" | jq --join-output '.Output // empty';)
    CURRENT_LINE_ACTION=$(echo "${LINE}" | jq --join-output '.Action // empty';)

    if [[ -z "${CURRENT_LINE_TEST_NAME}" ]]; then
        # Skip all lines with empty test name, and don't forget to reset the action.
        CURRENT_ACTION="SEARCHING";
        LAST_TEST_NAME="";
        EXPECTED_SO_FAR="";
        PAIR_TYPE="";

    elif [[ "${CURRENT_ACTION}" == "SEARCHING" ]]; then
        printf "\nInfo: Seaching a test run...\n";

        if [[ "${CURRENT_LINE_ACTION}" == "run" ]]; then
            CURRENT_ACTION="RUNNING";
            LAST_TEST_NAME=${CURRENT_LINE_TEST_NAME};
            EXPECTED_SO_FAR="";
            PAIR_TYPE="";
        fi

    elif [[ "${CURRENT_ACTION}" == "RUNNING" ]]; then
        printf "\nInfo: Inside a test run...\n";

        if [[ "${LAST_TEST_NAME}" != "${CURRENT_LINE_TEST_NAME}" ]]; then
            CURRENT_ACTION="SEARCHING";
            LAST_TEST_NAME="";
            EXPECTED_SO_FAR="";
            PAIR_TYPE="";

        elif [[ "${CURRENT_LINE_OUTPUT}" =~ ${REGEX_TO_FIND_NORMAL_EXPECTED} ]]; then
            EXPECTED_1=$(echo "${CURRENT_LINE_OUTPUT}" |  rg "${REGEX_RG_NORMAL_EXPECTED}" -r "" | awk '{ printf "%s", $0 }';);
            # Trim everything after comma (,) if there is one.
            EXPECTED_1="${EXPECTED_1%%,*}";

            # If it is a more complex value and not just a string.
            if [[ ${EXPECTED_1} =~ ${REGEX_TO_CHECK_COMPLEX} ]]; then
                # Track this test's file name to later apply our cached expected/actual pairs to.
                MAP_OF_COMPLEX_TESTS[${MAP_OF_TEST_FILES[${CURRENT_LINE_TEST_NAME}]}]="${CURRENT_LINE_TEST_NAME}";

                # Keep going, but discard this complex case, but we tracked this test for later.
                CURRENT_ACTION="RUNNING";
                LAST_TEST_NAME=${CURRENT_LINE_TEST_NAME};
                EXPECTED_SO_FAR="";
                PAIR_TYPE="";

            # Can do more validation of this expected result before caching it in variable.
            elif [[ ${#EXPECTED_1} -ge ${MINIMUM_EXPECTED_LENGTH} ]]; then
                CURRENT_ACTION="PAIRING";
                LAST_TEST_NAME=${CURRENT_LINE_TEST_NAME};
                EXPECTED_SO_FAR=${EXPECTED_1};
                PAIR_TYPE="NORMAL";

            else
                # Keep going, without a reset to find other possible failed tests in this run.
                CURRENT_ACTION="RUNNING";
                LAST_TEST_NAME=${CURRENT_LINE_TEST_NAME};
                EXPECTED_SO_FAR="";
                PAIR_TYPE="";
            fi

        elif [[ "${CURRENT_LINE_OUTPUT}" =~ ${REGEX_TO_FIND_SYMBOL_EXPECTED} ]]; then
            EXPECTED_2=$(echo "${CURRENT_LINE_OUTPUT}" |  rg "${REGEX_RG_SYMBOL_EXPECTED}" -r "\"" | awk '{ printf "%s", $0 }';);
            # Trim everything after comma (,) if there is one.
            EXPECTED_2="${EXPECTED_2%%,*}";

            # Can do more validation of this expected result before caching it in variable.
            if [[ ${#EXPECTED_2} -ge ${MINIMUM_EXPECTED_LENGTH} ]]; then
                CURRENT_ACTION="PAIRING";
                LAST_TEST_NAME=${CURRENT_LINE_TEST_NAME};
                EXPECTED_SO_FAR=${EXPECTED_2};
                PAIR_TYPE="SYMBOL";

            else
                # Keep going, without a reset to find other possible failed tests in this run.
                CURRENT_ACTION="RUNNING";
                LAST_TEST_NAME=${CURRENT_LINE_TEST_NAME};
                EXPECTED_SO_FAR="";
                PAIR_TYPE="";
            fi

        elif [ "${CURRENT_LINE_ACTION}" == "fail" ] || [ "${CURRENT_LINE_ACTION}" == "pass" ]; then
                # Test run has finished, go back to seaching for another run.
                CURRENT_ACTION="SEARCHING";
                LAST_TEST_NAME="";
                EXPECTED_SO_FAR="";
                PAIR_TYPE="";
        fi

    elif [ "${CURRENT_ACTION}" == "PAIRING" ] && [ "${PAIR_TYPE}" == "NORMAL" ]; then
        printf "\nInfo: Pairing up normal expected/actual...\n";

        if [[ "${LAST_TEST_NAME}" != "${CURRENT_LINE_TEST_NAME}" ]]; then
            CURRENT_ACTION="SEARCHING";
            LAST_TEST_NAME="";
            EXPECTED_SO_FAR="";
            PAIR_TYPE="";

        elif [[ "${CURRENT_LINE_OUTPUT}" =~ ${REGEX_TO_FIND_NORMAL_ACTUAL} ]]; then
            ACTUAL_1=$(echo "${CURRENT_LINE_OUTPUT}" |  rg "${REGEX_RG_NORMAL_ACTUAL}" -r "" | awk '{ printf "%s", $0 }';);
            # Trim everything after comma (,) if there is one.
            ACTUAL_1="${ACTUAL_1%%,*}";

            # Can do more validation of this actual result before adding the pair.
            if [[ -n "${ACTUAL_1}" ]]; then
                printf "\nInfo: Found a normal pair, store it...\n";

                # Only store pair if a test file path was found previously for this test.
                if [[ -v MAP_OF_TEST_FILES[${CURRENT_LINE_TEST_NAME}] ]]; then
                    PAIR_WITH_TEST_FILE="${EXPECTED_SO_FAR}|${ACTUAL_1}|${MAP_OF_TEST_FILES[${CURRENT_LINE_TEST_NAME}]}";
                    printf "\n--> pair=[ %s ]\n" "${PAIR_WITH_TEST_FILE}";
                    echo "${PAIR_WITH_TEST_FILE}" >> "${FILE_WITH_REPLACING_ACTIONS}";

                else
                    printf "\nWarning: Found bad normal pair, with invalid file path.\n"
                fi

                # Always cache this pair in a map, we can use this info to fix other tests that were too complex to fix.
                MAP_OF_EXPECTED_ACTUAL_PAIR[${EXPECTED_SO_FAR}]="${ACTUAL_1}";
            fi

            # Go back to "RUNNING" action and keep going, but discard the cached expected result.
            CURRENT_ACTION="RUNNING";
            LAST_TEST_NAME=${CURRENT_LINE_TEST_NAME};
            EXPECTED_SO_FAR="";
            PAIR_TYPE="";

        elif [ "${CURRENT_LINE_ACTION}" == "fail" ] || [ "${CURRENT_LINE_ACTION}" == "pass" ]; then
                # Test run has finished, go back to seaching for another run.
                CURRENT_ACTION="SEARCHING";
                LAST_TEST_NAME="";
                EXPECTED_SO_FAR="";
                PAIR_TYPE="";
        fi

    elif [ "${CURRENT_ACTION}" == "PAIRING" ] && [ "${PAIR_TYPE}" == "SYMBOL" ]; then
        printf "\nInfo: Pairing up symbol expected/actual...\n";

        if [[ "${LAST_TEST_NAME}" != "${CURRENT_LINE_TEST_NAME}" ]]; then
            CURRENT_ACTION="SEARCHING";
            LAST_TEST_NAME="";
            EXPECTED_SO_FAR="";
            PAIR_TYPE="";

        elif [[ "${CURRENT_LINE_OUTPUT}" =~ ${REGEX_TO_FIND_SYMBOL_ACTUAL} ]]; then
            # We replace the match with '"' in the end because we match up to the beggining of the string.
            ACTUAL_2=$(echo "${CURRENT_LINE_OUTPUT}" | rg "${REGEX_RG_SYMBOL_ACTUAL}" -r "\"" | awk '{ printf "%s", $0 }';);
            # Trim everything after comma (,) if there is one.
            ACTUAL_2="${ACTUAL_2%%,*}";

            # Can do more validation of this actual result before adding the pair.
            if [[ -n "${ACTUAL_2}" ]]; then
                printf "\nInfo: Found a symbol pair, store it...\n";

                # Only store pair if a test file path was found previously for this test.
                if [[ -v MAP_OF_TEST_FILES[${CURRENT_LINE_TEST_NAME}] ]]; then
                    PAIR_WITH_TEST_FILE="${EXPECTED_SO_FAR}|${ACTUAL_2}|${MAP_OF_TEST_FILES[${CURRENT_LINE_TEST_NAME}]}";
                    printf "\n--> pair=[ %s ]\n" "${PAIR_WITH_TEST_FILE}";
                    echo "${PAIR_WITH_TEST_FILE}" >> "${FILE_WITH_REPLACING_ACTIONS}";

                else
                    printf "\nWarning: Found bad symbol pair, with invalid file path.\n"
                fi

                # Always cache this pair in a map, we can use this info to fix other tests that were too complex to fix.
                MAP_OF_EXPECTED_ACTUAL_PAIR[${EXPECTED_SO_FAR}]="${ACTUAL_2}";
            fi

            # Go back to "RUNNING" action and keep going, but discard the cached expected result.
            CURRENT_ACTION="RUNNING";
            LAST_TEST_NAME=${CURRENT_LINE_TEST_NAME};
            EXPECTED_SO_FAR="";
            PAIR_TYPE="";

        elif [ "${CURRENT_LINE_ACTION}" == "fail" ] || [ "${CURRENT_LINE_ACTION}" == "pass" ]; then
                CURRENT_ACTION="SEARCHING";
                LAST_TEST_NAME="";
                EXPECTED_SO_FAR="";
                PAIR_TYPE="";
        fi

    else
        printf "\nError: encountered an unknown action.\n";
        CURRENT_ACTION="SEARCHING";
        LAST_TEST_NAME="";
        EXPECTED_SO_FAR="";
        PAIR_TYPE="";
    fi

done < "${FILE_WITH_ONLY_FAILS}";

#######################################################################################################################
#                                               STAGE 3
#######################################################################################################################
# We start by trimming the accumulated actions to be unique so we don't apply the same action again. Then we
# start applying these unique replacing actions. After all the simple actions are done/applied then we try
# to fix the complex cases. We visit each complex test file, and see if there is any match using our cached
# expected/actual pairs, if there is then change all occurances of expected with actual in that file. At the
# end of this stage we should have swapped most (if not all), failing expected values with their actual ones.
#######################################################################################################################

# If replacing actions exist then do them.
if [ -f "${FILE_WITH_REPLACING_ACTIONS}" ] && [ -n "$(cat "${FILE_WITH_REPLACING_ACTIONS}")" ]; then
    printf "\nInfo: Trying to apply replacing actions...\n"

    # Trim all the actions to a list of only unique actions.
    sort "${FILE_WITH_REPLACING_ACTIONS}" --unique | sponge "${FILE_WITH_REPLACING_ACTIONS}";

    # Replace all expected values in the test file, with the actual value.
    while IFS= read -r REPLACE_ACTION_LINE; do

        # Must set IFS on the same line as the read with no semicolon or other separator, to scope it to this command.
        IFS='|' read -ra ACTION_ARGS <<< "${REPLACE_ACTION_LINE}"
        if [[ "${#ACTION_ARGS[@]}" == "3" ]]; then
            printf "\nReplacing [ %s ], with [ %s ], in file [ %s ]}\n" "${ACTION_ARGS[0]}" "${ACTION_ARGS[1]}" "${ACTION_ARGS[2]}";
            sed "s/${ACTION_ARGS[0]}/${ACTION_ARGS[1]}/g" "${ACTION_ARGS[2]}" | sponge "${ACTION_ARGS[2]}";
        else
            printf "\nWarning: Skipping action because of missing information.\n"
        fi

    done < "${FILE_WITH_REPLACING_ACTIONS}";

else
    printf "\nInfo: There are NO replacing actions to apply.\n"
fi

if [ ${#MAP_OF_COMPLEX_TESTS[@]} -ne 0 ]; then
    printf "\nInfo: Trying to fix the complex test files...\n";

    # Use our cached pairs to replace any occurance in files that had complex expected and actual pairs.
    for TEST_TO_FIX in "${!MAP_OF_COMPLEX_TESTS[@]}"; do
        printf "\nFixing test = [ %s ]\n" "${MAP_OF_COMPLEX_TESTS[${TEST_TO_FIX}]}";
        printf "In file = [ %s ]\n" "${TEST_TO_FIX}";

        for EXPECTED_KEY in "${!MAP_OF_EXPECTED_ACTUAL_PAIR[@]}"; do
            sed "s/${EXPECTED_KEY}/${MAP_OF_EXPECTED_ACTUAL_PAIR[${EXPECTED_KEY}]}/g" "${TEST_TO_FIX}" | sponge "${TEST_TO_FIX}";
        done
    done
else
    printf "\nInfo: There are NO complex expect/actual pairs to fix.\n"
fi

 
printf "\n======================================================================";
printf "\nTemporary Working Directory = [ %s ]" "${DIR_TEMPORARY}";
printf "\n======================================================================\n";
