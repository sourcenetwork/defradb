#!/bin/bash

#========================================================================================
# Script to take the file containing the HTML output of benchstat and seperate
#  out better, worse and unchanged benchmark results such that the summary is
#  much nicer to read. The result is dumped to stdout.
#========================================================================================

#========================================================================================
# Helper Variables
#========================================================================================

BETTER_COMPARISONS="$(cat ${1} | grep -v "class='group'" | grep "class='better'")";
WORSE_COMPARISONS="$(cat ${1} | grep -v "class='group'" | grep "class='worse'")";
UNCHANGED_COMPARISONS="$(cat ${1} | grep -v "class='group'" | grep "class='unchanged'")";

BETTER_COUNT="$(cat ${1} | grep -v "class='group'" | grep -c "class='better'")";
WORSE_COUNT="$(cat ${1} | grep -v "class='group'" | grep -c "class='worse'")";
UNCHANGED_COUNT="$(cat ${1} | grep -v "class='group'" | grep -c "class='unchanged'")";
TOTAL_COUNT="$((BETTER_COUNT+WORSE_COUNT+UNCHANGED_COUNT))";

#========================================================================================
# Script Execution.
#========================================================================================

PRETTY_RESULT="

## Benchmark Results

### Summary
* "${TOTAL_COUNT}" Benchmarks successfully compared.
* "${BETTER_COUNT}" Benchmarks were ✅ Better.
* "${WORSE_COUNT}" Benchmarks were ❌ Worse .
* "${UNCHANGED_COUNT}" Benchmarks were ✨ Unchanged.


<details>
<summary> ✅ See Better Results...</summary>
<table class='benchstat better'>
<tbody>
<tr><th><th colspan='2' class='metric'>time/op<th>delta
"${BETTER_COMPARISONS}"
<tr><td>&nbsp;
</tbody>
</table>
</details>

<details>
<summary> ❌ See Worse Results...</summary>
<table class='benchstat worse'>
<tbody>
<tr><th><th colspan='2' class='metric'>time/op<th>delta
"${WORSE_COMPARISONS}"
<tr><td>&nbsp;
</tbody>
</table>
</details>

<details>
<summary> ✨ See Unchanged Results...</summary>
<table class='benchstat unchanged'>
<tbody>
<tr><th><th colspan='2' class='metric'>time/op<th>delta
"${UNCHANGED_COMPARISONS}"
<tr><td>&nbsp;
</tbody>
</table>
</details>

<details>
<summary> 🐋 See Full Results...</summary>
"$(cat ${1})"
</details>

"

echo "${PRETTY_RESULT}";
