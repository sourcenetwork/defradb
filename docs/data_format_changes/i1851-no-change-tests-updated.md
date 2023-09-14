# Make existing mutation tests use mutation test system

This is not a breaking change, tests were changed from using gql requests to CreateDoc and UpdateDoc actions, meaning the point at which the change detector split setup/assert shifted.
