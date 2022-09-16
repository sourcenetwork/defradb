# Make current inline-array types non-nullable

Our inline array types were arrays of nullable types, however internally we failed to handle them being nullable resulting in runtime errors if nul values were provided.  This change removes support for nullable inline arrays, and adds support for non-nullable inline arrays.

The decision to do it this way around was made in part due to the impending introduction of generics to Go (v1.18), which would likely make handling nullable scalars much cleaner and it would be more time efficient to implement that once instead of implementing without generics followed by an immediate refactor.    
