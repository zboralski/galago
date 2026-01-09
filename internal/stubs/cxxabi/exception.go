// Package cxxabi provides stub implementations for C++ ABI functions.
// This file implements stubs for C++ exception handling.
package cxxabi

import (
	"sync"

	"github.com/zboralski/galago/internal/emulator"
	"github.com/zboralski/galago/internal/stubs"
)

var (
	// guardState tracks static initialization guard states
	guardState   = make(map[uint64]bool)
	guardStateMu sync.Mutex
)

func init() {
	// Exception handling
	stubs.RegisterFunc("cxxabi", "__cxa_throw", stubCxaThrow)
	stubs.RegisterFunc("cxxabi", "__cxa_rethrow", stubCxaRethrow)
	stubs.RegisterFunc("cxxabi", "__cxa_begin_catch", stubCxaBeginCatch)
	stubs.RegisterFunc("cxxabi", "__cxa_end_catch", stubCxaEndCatch)
	stubs.RegisterFunc("cxxabi", "__cxa_allocate_exception", stubCxaAllocateException)
	stubs.RegisterFunc("cxxabi", "__cxa_free_exception", stubCxaFreeException)
	stubs.RegisterFunc("cxxabi", "__cxa_get_exception_ptr", stubCxaGetExceptionPtr)
	stubs.RegisterFunc("cxxabi", "__cxa_current_exception_type", stubCxaCurrentExceptionType)
	stubs.RegisterFunc("cxxabi", "__cxa_call_unexpected", stubCxaCallUnexpected)
	stubs.RegisterFunc("cxxabi", "__cxa_bad_cast", stubCxaBadCast)
	stubs.RegisterFunc("cxxabi", "__cxa_bad_typeid", stubCxaBadTypeid)

	// Static initialization guards
	stubs.RegisterFunc("cxxabi", "__cxa_guard_acquire", stubCxaGuardAcquire)
	stubs.RegisterFunc("cxxabi", "__cxa_guard_release", stubCxaGuardRelease)
	stubs.RegisterFunc("cxxabi", "__cxa_guard_abort", stubCxaGuardAbort)

	// Exit handlers
	stubs.RegisterFunc("cxxabi", "__cxa_atexit", stubCxaAtexit)
	stubs.RegisterFunc("cxxabi", "__cxa_finalize", stubCxaFinalize)
	stubs.RegisterFunc("cxxabi", "__cxa_thread_atexit", stubCxaThreadAtexit)
	stubs.RegisterFunc("cxxabi", "__cxa_thread_atexit_impl", stubCxaThreadAtexit)

	// Pure virtual
	stubs.RegisterFunc("cxxabi", "__cxa_pure_virtual", stubCxaPureVirtual)
	stubs.RegisterFunc("cxxabi", "__cxa_deleted_virtual", stubCxaDeletedVirtual)

	// Personality routines (for unwinding)
	stubs.RegisterFunc("cxxabi", "__gxx_personality_v0", stubGxxPersonality)
	stubs.RegisterFunc("cxxabi", "_Unwind_Resume", stubUnwindResume)
	stubs.RegisterFunc("cxxabi", "_Unwind_RaiseException", stubUnwindRaiseException)
	stubs.RegisterFunc("cxxabi", "_Unwind_DeleteException", stubUnwindDeleteException)
	stubs.RegisterFunc("cxxabi", "_Unwind_GetLanguageSpecificData", stubUnwindGetLSDA)
	stubs.RegisterFunc("cxxabi", "_Unwind_GetRegionStart", stubUnwindGetRegionStart)
	stubs.RegisterFunc("cxxabi", "_Unwind_SetGR", stubUnwindSetGR)
	stubs.RegisterFunc("cxxabi", "_Unwind_SetIP", stubUnwindSetIP)
	stubs.RegisterFunc("cxxabi", "_Unwind_GetIP", stubUnwindGetIP)

	// RTTI
	stubs.RegisterFunc("cxxabi", "__dynamic_cast", stubDynamicCast)
}

// Exception handling stubs

func stubCxaThrow(emu *emulator.Emulator) bool {
	excPtr := emu.X(0)
	stubs.DefaultRegistry.Log("cxxabi", "__cxa_throw", stubs.FormatHex(excPtr))

	// Stop emulation on throw - we don't support real exception handling
	// In a full implementation, this would trigger stack unwinding
	emu.Stop()
	return true
}

func stubCxaRethrow(emu *emulator.Emulator) bool {
	stubs.DefaultRegistry.Log("cxxabi", "__cxa_rethrow", "")
	emu.Stop()
	return true
}

func stubCxaBeginCatch(emu *emulator.Emulator) bool {
	excPtr := emu.X(0)
	stubs.DefaultRegistry.Log("cxxabi", "__cxa_begin_catch", stubs.FormatHex(excPtr))
	emu.SetX(0, excPtr)
	stubs.ReturnFromStub(emu)
	return false
}

func stubCxaEndCatch(emu *emulator.Emulator) bool {
	stubs.ReturnFromStub(emu)
	return false
}

func stubCxaAllocateException(emu *emulator.Emulator) bool {
	size := emu.X(0)
	if size == 0 {
		size = 64
	}

	// Allocate exception object with header space
	ptr := emu.Malloc(size + 128)
	stubs.DefaultRegistry.Log("cxxabi", "__cxa_allocate_exception", stubs.FormatPtrPair("size", size, "ptr", ptr))
	emu.SetX(0, ptr)
	stubs.ReturnFromStub(emu)
	return false
}

func stubCxaFreeException(emu *emulator.Emulator) bool {
	stubs.ReturnFromStub(emu)
	return false
}

func stubCxaGetExceptionPtr(emu *emulator.Emulator) bool {
	excPtr := emu.X(0)
	emu.SetX(0, excPtr)
	stubs.ReturnFromStub(emu)
	return false
}

func stubCxaCurrentExceptionType(emu *emulator.Emulator) bool {
	emu.SetX(0, 0) // No current exception
	stubs.ReturnFromStub(emu)
	return false
}

func stubCxaCallUnexpected(emu *emulator.Emulator) bool {
	stubs.DefaultRegistry.Log("cxxabi", "__cxa_call_unexpected", "")
	emu.Stop()
	return true
}

func stubCxaBadCast(emu *emulator.Emulator) bool {
	stubs.DefaultRegistry.Log("cxxabi", "__cxa_bad_cast", "")
	emu.Stop()
	return true
}

func stubCxaBadTypeid(emu *emulator.Emulator) bool {
	stubs.DefaultRegistry.Log("cxxabi", "__cxa_bad_typeid", "")
	emu.Stop()
	return true
}

// Static initialization guard stubs

func stubCxaGuardAcquire(emu *emulator.Emulator) bool {
	guardPtr := emu.X(0)

	guardStateMu.Lock()
	initialized := guardState[guardPtr]
	guardStateMu.Unlock()

	if initialized {
		emu.SetX(0, 0) // Already initialized
	} else {
		emu.SetX(0, 1) // Need to initialize
	}
	stubs.ReturnFromStub(emu)
	return false
}

func stubCxaGuardRelease(emu *emulator.Emulator) bool {
	guardPtr := emu.X(0)

	guardStateMu.Lock()
	guardState[guardPtr] = true
	guardStateMu.Unlock()

	stubs.ReturnFromStub(emu)
	return false
}

func stubCxaGuardAbort(emu *emulator.Emulator) bool {
	// Abort initialization - just return
	stubs.ReturnFromStub(emu)
	return false
}

// Exit handler stubs

func stubCxaAtexit(emu *emulator.Emulator) bool {
	// Register destructor for static objects - we just ignore it
	emu.SetX(0, 0) // Success
	stubs.ReturnFromStub(emu)
	return false
}

func stubCxaFinalize(emu *emulator.Emulator) bool {
	// Run registered destructors - we don't track them
	stubs.ReturnFromStub(emu)
	return false
}

func stubCxaThreadAtexit(emu *emulator.Emulator) bool {
	emu.SetX(0, 0) // Success
	stubs.ReturnFromStub(emu)
	return false
}

// Pure virtual stubs

func stubCxaPureVirtual(emu *emulator.Emulator) bool {
	stubs.DefaultRegistry.Log("cxxabi", "__cxa_pure_virtual", "FATAL: pure virtual call")
	emu.Stop()
	return true
}

func stubCxaDeletedVirtual(emu *emulator.Emulator) bool {
	stubs.DefaultRegistry.Log("cxxabi", "__cxa_deleted_virtual", "FATAL: deleted virtual call")
	emu.Stop()
	return true
}

// Personality routine stubs

func stubGxxPersonality(emu *emulator.Emulator) bool {
	// Return _URC_CONTINUE_UNWIND (8)
	emu.SetX(0, 8)
	stubs.ReturnFromStub(emu)
	return false
}

func stubUnwindResume(emu *emulator.Emulator) bool {
	stubs.DefaultRegistry.Log("cxxabi", "_Unwind_Resume", "")
	emu.Stop()
	return true
}

func stubUnwindRaiseException(emu *emulator.Emulator) bool {
	// Return _URC_END_OF_STACK
	emu.SetX(0, 5)
	stubs.ReturnFromStub(emu)
	return false
}

func stubUnwindDeleteException(emu *emulator.Emulator) bool {
	stubs.ReturnFromStub(emu)
	return false
}

func stubUnwindGetLSDA(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubUnwindGetRegionStart(emu *emulator.Emulator) bool {
	emu.SetX(0, 0)
	stubs.ReturnFromStub(emu)
	return false
}

func stubUnwindSetGR(emu *emulator.Emulator) bool {
	stubs.ReturnFromStub(emu)
	return false
}

func stubUnwindSetIP(emu *emulator.Emulator) bool {
	stubs.ReturnFromStub(emu)
	return false
}

func stubUnwindGetIP(emu *emulator.Emulator) bool {
	emu.SetX(0, emu.PC())
	stubs.ReturnFromStub(emu)
	return false
}

// RTTI stubs

func stubDynamicCast(emu *emulator.Emulator) bool {
	// void* __dynamic_cast(const void* src, const __class_type_info* src_type,
	//                      const __class_type_info* dst_type, ptrdiff_t src2dst_offset)
	src := emu.X(0)
	// srcType := emu.X(1)
	// dstType := emu.X(2)
	// offset := emu.X(3)

	// Just return the source pointer (no real type checking)
	emu.SetX(0, src)
	stubs.ReturnFromStub(emu)
	return false
}

// ClearGuardState resets all guard states.
func ClearGuardState() {
	guardStateMu.Lock()
	guardState = make(map[uint64]bool)
	guardStateMu.Unlock()
}
