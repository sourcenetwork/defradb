#!/bin/bash

# TestReturnCode prints Pass if the expected return code matches the
#  evaluated command, otherwise prints false and exits on first failure.
# Usage: <Command> <ExpectedReturnCode>
# Example: TestReturnCode "cat a.txt" 0
TestReturnCode() {
    eval "${1}" &> /dev/null;

    local ACTUAL="${?}";
    local EXPECTED="${2}";

    if [ "${ACTUAL}" -eq "${EXPECTED}" ]; then
        printf "PASS\n";
    else
        printf "FAIL ...\n";
        printf "> Command  : [%s] \n" "${1}";
        printf "> Expected : [%s] \n" "${EXPECTED}";
        printf "> ACTUAL   : [%s] \n" "${ACTUAL}";
        exit 1;
    fi
}

# Test the script that is responsible for the validation of pr title.
readonly T1="./validate-conventional-style.sh"

TestReturnCode "${T1}" 2;

TestReturnCode "${T1} 'chore: This title  has  everything     valid except that    its too long'" 3;
TestReturnCode "${T1} 'bot Bump github.com/alternativesourcenetwork/defradb from 1.1.0.1.0.0 to 1.1.0.1.0.1'" 3;

TestReturnCode "${T1} 'chore: This title has more than one : colon'" 4;
TestReturnCode "${T1} 'chore This title has no colon'" 4;
TestReturnCode "${T1} 'bot Bump github.com/short/short from 1.2.3 to 1.2.4'" 4; 

TestReturnCode "${T1} 'feat: a'" 5;
TestReturnCode "${T1} 'feat: '" 5;
TestReturnCode "${T1} 'feat:'" 5;

TestReturnCode "${T1} 'feat:There is no space between label & desc.'" 6;
TestReturnCode "${T1} 'feat:there is no space between label & desc.'" 6;

TestReturnCode "${T1} 'ci: lowercase first character after label'" 7;

TestReturnCode "${T1} 'ci: Last character should not be period.'" 8;
TestReturnCode "${T1} 'ci(i): Last character should not be period.'" 8;
TestReturnCode "${T1} 'ci: Last character is a space '" 8;
TestReturnCode "${T1} 'ci: Last character is a \\\`tick\\\`'" 8;

TestReturnCode "${T1} 'bug: This is an invalid label'" 9;
TestReturnCode "${T1} 'bug(i): This is an invalid label'" 9;

TestReturnCode "${T1} 'ci: Last character is a number v1.5.0'" 0;
TestReturnCode "${T1} 'ci: Last character is not lowercase alphabeT'" 0;
TestReturnCode "${T1} 'chore: This is a valid title'" 0;
TestReturnCode "${T1} 'ci: This is a valid title'" 0;
TestReturnCode "${T1} 'docs: This is a valid title'" 0;
TestReturnCode "${T1} 'feat: This is a valid title'" 0;
TestReturnCode "${T1} 'fix: This is a valid title'" 0;
TestReturnCode "${T1} 'perf: This is a valid title'" 0;
TestReturnCode "${T1} 'refactor: This is a valid title'" 0;
TestReturnCode "${T1} 'test: This is a valid title'" 0;
TestReturnCode "${T1} 'tools: This is a valid title'" 0;
TestReturnCode "${T1} 'bot: Bump github.com/alternativesourcenetwork/defradb from 1.1.0.1.0.0 to 1.1.0.1.0.1'" 0;
TestReturnCode "${T1} 'ci(i): Valid ignore title'" 0;
TestReturnCode "${T1} 'chore(i): Valid ignore title'" 0;
TestReturnCode "${T1} 'docs(i): Valid ignore title'" 0;
TestReturnCode "${T1} 'feat(i): Valid ignore title'" 0;
TestReturnCode "${T1} 'fix(i): Valid ignore title'" 0;
TestReturnCode "${T1} 'perf(i): Valid ignore title'" 0;
TestReturnCode "${T1} 'refactor(i): Valid ignore title'" 0;
TestReturnCode "${T1} 'test(i): Valid ignore title'" 0;
TestReturnCode "${T1} 'tools(i): Valid ignore title'" 0;
TestReturnCode "${T1} 'bot(i): Bump github.com/alternativesourcenetwork/defradb from 1.1.0.1.0.0 to 1.1.0.1.0.1'" 0;
TestReturnCode "${T1} 'bot(i): Bump githurk/defradb from 1.1.0.1.0.0 to 1.1.0.1.0.1'" 0;
