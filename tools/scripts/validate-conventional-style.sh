#!/bin/bash

#========================================================================================
# Script that would validate that the given string of commit or pull request title,
#  adheres to our sub-set of conventional style commit labels. In addition to that also
#  makes sure that the first letter after the `:` separator is capitalized.
# Usage: ./validate-conventional-style.sh "feat: Add a new feature"
#========================================================================================

# Declare a non-mutable indexed array that contains all the subset of conventional style
#  labels that we deem valid for our use case. There should always be insync with the
#  labels we have defined for the change log in: `defradb/tools/configs/chglog/config.yml`.
readonly -a VALID_LABELS=("chore"
                          "ci"
                          "docs"
                          "feat"
                          "fix"
                          "perf"
                          "refactor"
                          "test"
                          "tools");

if [ "${#}" -ne 1 ]; then
    printf "Error: Invalid number of arguments (pass title as 1 string argument).\n";
    exit 2;
fi

TITLE=${1};
IS_DEPENDABOT=false;

# In case the title is known to be from dependabot, then we skip the title validation.
if [[ "${TITLE}" == *"Bump"* ]]; then
    printf "Info: Title is from dependabot, skipping title validation.\n";
    IS_DEPENDABOT=true;
    exit 0;
fi

# Validate that the entire length of the title is less than or equal to our character limit.
if [ "${#TITLE}" -gt 60 ] && [ "${IS_DEPENDABOT}" = false ]; then
    printf "Error: The length of the title is too long (should be 60 or less).\n";
    exit 3;
fi

# Split the title at ':' and store the result in ${SPLIT_TOKENS}.
# Doing eval to ensure the split works for elements that contain spaces.
eval "SPLIT_TOKENS=($(echo "\"$TITLE\"" | sed 's/:/" "/g'))";

# Validate the `:` token exists exactly once.
if [ "${#SPLIT_TOKENS[*]}" -ne 2 ]; then
    printf "Error: Splitting title at ':' didn't result in 2 elements.\n";
    exit 4;
fi

LABEL="${SPLIT_TOKENS[0]}";
DESCRIPTION="${SPLIT_TOKENS[1]}";

printf "Info: label = [%s]\n" "${LABEL}";
printf "Info: description = [%s]\n" "${DESCRIPTION}";

# Validate that description isn't too short.
if [ "${#DESCRIPTION}" -le 2 ]; then
    printf "Error: Description is too short.\n";
    exit 5;
fi

CHECK_SPACE="${DESCRIPTION::1}"; # First character
CHECK_FIRST_UPPER_CASE="${DESCRIPTION:1:1}"; # Second character
CHECK_LAST_LOWER_CASE_OR_NUM="${DESCRIPTION: -1}"; # Last character

# Validate that there is a space between the label and description.
if [ "${CHECK_SPACE}" != " " ]; then
    printf "Error: There is no space between label and description.\n";
    exit 6;
fi

# Validate that the first character after the label + ' ' is an uppercase alphabet character.
if [[ "${CHECK_FIRST_UPPER_CASE}" != [A-Z] ]]; then
    printf "Error: First character after the label is not an uppercase alphabet.\n";
    exit 7;
fi

# Validate that the last character is a lower case alphabet or a number character.
if [[ "${CHECK_LAST_LOWER_CASE_OR_NUM}" != [a-z0-9] ]]; then
    printf "Error: Last character is neither a lowercase alphabet nor a number.\n";
    exit 8;
fi

# Validate that ${LABEL} is one of the valid labels.
for validLabel in "${VALID_LABELS[@]}"; do
    if [ "${LABEL}" == "${validLabel}" ]; then
        printf "Success: Title's label and description style is valid.\n";
        exit 0;
    fi
done

# Should only reach here if the label was invalid.
printf "Error: The label used in the title isn't a valid label.\n";
exit 9;
