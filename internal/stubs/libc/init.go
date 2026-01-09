// Package libc provides stub implementations for libc functions.
// Import this package to register all libc stubs with the default registry.
package libc

// This file exists to ensure all libc stubs are registered via init().
// Each file in this package registers its stubs in its own init() function.
