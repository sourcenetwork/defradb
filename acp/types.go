package acp

// RegistrationResult is an enum type which indicates the result of a RegisterObject call to SourceHub / ACP Core
type RegistrationResult int32

const (
	// NoOp indicates no action was take. The operation failed or the Object already existed and was active
	RegistrationResult_NoOp RegistrationResult = 0
	// Registered indicates the Object was sucessfuly registered to the Actor.
	RegistrationResult_Registered RegistrationResult = 1
	// Unarchived indicates that a previously deleted Object is active again.
	// Only the original owners can Unarchive an object.
	RegistrationResult_Unarchived RegistrationResult = 2
)

// Policy defines the minimum set of methods required in order to verify whether it meets DPI requirements
type Policy interface {
	GetResourceByName(name string) Resource
}

// Resource defines the minimum set of methods required in order to verify whether
// a Policy's Resource meets DPI requirements
type Resource interface {
	GetPermissionByName(name string) Permission
}

// Permission defines the minimum set of methods required in order to verify whether
// a Resource's Permission meets DPI requirements
type Permission interface {
	GetExpression() string
}
