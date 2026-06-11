package vault

import "fmt"

// errInvalidArgs returns a formatted error when the number of ABI-decoded
// arguments does not match the expected count for a method.
func errInvalidArgs(method string, want, got int) error {
	return fmt.Errorf("precompile %s: expected %d arg(s), got %d", method, want, got)
}

// errInvalidArgType returns a formatted error when an ABI-decoded argument
// cannot be type-asserted to its expected Go type. This typically indicates a
// mismatch between the ABI definition and the on-wire encoding from the caller.
func errInvalidArgType(method, argName string) error {
	return fmt.Errorf("precompile %s: invalid type for argument %q", method, argName)
}
