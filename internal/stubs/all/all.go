// Package all imports all stub packages to ensure they register via init().
// Import this package in session setup to enable all stubs.
//
// Example:
//
//	import _ "github.com/zboralski/galago/internal/stubs/all"
package all

import (
	// Import all stub packages for side effects (init registration)
	_ "github.com/zboralski/galago/internal/stubs/android"
	_ "github.com/zboralski/galago/internal/stubs/cxxabi"
	_ "github.com/zboralski/galago/internal/stubs/internal"
	_ "github.com/zboralski/galago/internal/stubs/jni"
	_ "github.com/zboralski/galago/internal/stubs/libc"
	_ "github.com/zboralski/galago/internal/stubs/lua"
	_ "github.com/zboralski/galago/internal/stubs/network"
	_ "github.com/zboralski/galago/internal/stubs/pthread"
	_ "github.com/zboralski/galago/internal/stubs/setters"
	_ "github.com/zboralski/galago/internal/stubs/tolua"
)
