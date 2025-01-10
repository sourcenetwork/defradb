# Add policy
defradb client acp policy add -f organization_policy.yml --identity f0308aa7ffb6a22ceaf5e3d0c1f3a054e6996caede12f85c2fad68f8af8b963f

# add schema
defradb client schema add -f ./organization.graphql

# get current defined indexes
defradb client collection describe --name AnymalOrganization | jq -r '.[0].description.Indexes'

# add AnymalOrganization document
defradb client collection create --name AnymalOrganization -f organization_document.json

# Run a explain query to verify the planner is using
# the correct index
defradb client query -f explain_query.graphql 2> /dev/null | jq '.data.explain.operationNode[0].selectTopNode.selectNode.scanNode.indexFetches' | awk '$0 == "1"' | grep -q "" && echo 'Used correct Index (first)!!!' || echo "MISSING INDEX"

# apply the patch
# (Note: This script behaves the same if --set-active is provided at the time
# of the patch, or later using the set-active command)
defradb client schema patch -p add_step_to_anymal_organization_patch.json --set-active

# Run a explain query again to verify the planner is 
# still using the correct index
# (this wont print the correct string)
defradb client query -f explain_query.graphql 2> /dev/null | jq '.data.explain.operationNode[0].selectTopNode.selectNode.scanNode.indexFetches' | awk '$0 == "1"' | grep -q "" && echo 'Used correct Index (second)!!!' || echo "MISSING INDEX"

# Get current indexes
# (will return an empty set)
defradb client collection describe --name AnymalOrganization | jq -r '.[0].description.Indexes'

# Re-add the index
# THIS STEP SHOULDN'T BE NECESSARY
defradb client index create --collection AnymalOrganization --fields name

# Get current indexes
# (will return a valid entry now)
defradb client collection describe --name AnymalOrganization | jq -r '.[0].description.Indexes'

# Final check to see if the index is being used in the planner
defradb client query -f explain_query.graphql 2> /dev/null | jq '.data.explain.operationNode[0].selectTopNode.selectNode.scanNode.indexFetches' | awk '$0 == "1"' | grep -q "" && echo 'Used correct Index (third)!!!' || echo "MISSING INDEX"